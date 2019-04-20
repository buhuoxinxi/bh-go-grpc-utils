package balancer

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/rs/xid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

// etcd try confirm cancel config
const (
	defaultTryConfirmCancelETCDAliveTTL  int64 = 10  // etcd alive ttl(10s)
	defaultTryConfirmCancelStatusInitial       = "0" // initial status
	defaultTryConfirmCancelStatusSuccess       = "1" // success status
	defaultTryConfirmCancelStatusFail          = "2" // fail status
)

// err
var (
	ErrTCCInvalidLease = status.Error(codes.Internal, "ETCDClient.Grant lease fail")
	ErrTCCInvalidPut   = status.Error(codes.Internal, "ETCDClient.Put fail")
	ErrTCCInvalidGet   = status.Error(codes.Internal, "ETCDClient.Get fail")
	ErrTCCInvalidTCC   = status.Error(codes.Internal, "try confirm cancel fail")
)

// TryConfirmCancel 尝试-确认-取消
type TryConfirmCancel struct {
	ETCDClient   *clientv3.Client // etcd client
	TCCKeyPrefix string           // etcd key prefix
	TCCKeySlice  []string         // etcd key slice
	TCCKeyTTL    int64            // tcc key etcd ttl
}

// NewETCDClient etcd client
func (e *TryConfirmCancel) NewETCDClient() {
	e.ETCDClient = etcdClient
}

// PutTryKeyValue put try
func (e *TryConfirmCancel) PutTryKeyValue(tryNumber int) error {
	// invalid lock number
	if tryNumber <= 1 {
		return nil
	}

	// client
	if e.ETCDClient == nil {
		e.NewETCDClient()
	}

	// etcd ttl
	if e.TCCKeyTTL <= 0 {
		e.TCCKeyTTL = defaultTryConfirmCancelETCDAliveTTL
	}

	// key
	e.TCCKeySlice = make([]string, tryNumber)

	// key prefix
	e.TCCKeyPrefix = xid.New().String()

	// etcd alive ttl
	leaseResp, err := e.ETCDClient.Grant(context.Background(), e.TCCKeyTTL)
	if err != nil {
		return ErrTCCInvalidLease
	}

	// put
	for i := 0; i < tryNumber; i++ {
		// lock ley
		e.TCCKeySlice[i] = fmt.Sprintf("%s_%d", e.TCCKeyPrefix, i)
		// save to etcd
		_, err = e.ETCDClient.Put(context.Background(), e.TCCKeySlice[i], defaultTryConfirmCancelStatusInitial, clientv3.WithLease(leaseResp.ID))
		if err != nil {
			return ErrTCCInvalidPut
		}
	}
	return nil
}

// WatchTryKeyValue watch try
func (e *TryConfirmCancel) WatchTryKeyValue() (isReady bool, err error) {
	// invalid try number
	if len(e.TCCKeySlice) <= 1 {
		isReady = true
		return isReady, nil
	}

	// client
	if e.ETCDClient == nil {
		e.NewETCDClient()
	}

	// try key map
	var tccKeyMap sync.Map
	for i := range e.TCCKeySlice {
		tccKeyMap.Store(e.TCCKeySlice[i], false)
	}

	// polling
	return e.pollingTryKeyValue(tccKeyMap)

	// watch : if the processing time too short, goroutine will blocking
	//isReadyChannel := make(chan bool)
	//go e.watchTryKeyValue(tccKeyMap, isReadyChannel)
	// block
	//select {
	//case isReady := <-isReadyChannel:
	//	if !isReady {
	//		return isReady, ErrTCCInvalidTCC
	//	}
	//	return isReady, err
	//}
}

// pollingTryKeyValue polling try value
func (e *TryConfirmCancel) pollingTryKeyValue(tccKeyMap sync.Map) (isReady bool, err error) {
	// etcd key value
	getResp, err := e.ETCDClient.Get(context.Background(), e.TCCKeyPrefix, clientv3.WithPrefix())
	if err != nil {
		return isReady, ErrTCCInvalidGet
	} else if getResp.Count != int64(len(e.TCCKeySlice)) {
		return isReady, ErrTCCInvalidTCC
	}

	// check has success
	for i := range getResp.Kvs {
		// logrus.Printf("%q : %q\n", getResp.Kvs[i].Value, getResp.Kvs[i].Value)
		switch string(getResp.Kvs[i].Value) {

		case defaultTryConfirmCancelStatusSuccess:
			// is success
			if err = e.PutStatusSuccess(string(getResp.Kvs[i].Key)); err != nil {
				return isReady, err
			}
			tccKeyMap.Store(string(getResp.Kvs[i].Key), true)
			// ready
			if e.isReady(tccKeyMap) {
				isReady = true
				return isReady, err
			}

		case defaultTryConfirmCancelStatusFail:
			// is fail
			if err = e.PutStatusFail(string(getResp.Kvs[i].Key)); err != nil {
				isReady = false
				return isReady, err
			}
			isReady = false
			return isReady, err

		case defaultTryConfirmCancelStatusInitial:
			// is initial

		default:
			// default
			isReady = false
			return isReady, ErrTCCInvalidTCC
		}
	}

	// do again
	time.Sleep(time.Millisecond)

	return e.pollingTryKeyValue(tccKeyMap)
}

// watchTryKeyValue all try is ready
//
// if the processing time too short, the goroutine will blocking
func (e *TryConfirmCancel) watchTryKeyValue(tccKeyMap sync.Map, isReadyChannel chan bool) {
	// watch
	rch := e.ETCDClient.Watch(context.Background(), e.TCCKeyPrefix, clientv3.WithPrefix())
	for n := range rch {
		// etcd events
		for _, ev := range n.Events {
			// logrus.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			switch ev.Type {
			case mvccpb.PUT:
				switch string(ev.Kv.Value) {

				case defaultTryConfirmCancelStatusSuccess:
					// is success
					tccKeyMap.Store(string(ev.Kv.Key), true)
					// is ready
					if e.isReady(tccKeyMap) {
						isReadyChannel <- true
						return
					}

				case defaultTryConfirmCancelStatusFail:
					// is fail
					isReadyChannel <- false
					return

				case defaultTryConfirmCancelStatusInitial:
					// is initial

				default:
					// default
					isReadyChannel <- false
					return

				}

			case mvccpb.DELETE:
				// is fail
				isReadyChannel <- false
				return
			}
		}
	}
}

// isReady all try is ready
func (e *TryConfirmCancel) isReady(tccKeyMap sync.Map) bool {
	var isReady = true
	// all try is ready
	tccKeyMap.Range(func(key, value interface{}) bool {
		if ready, ok := value.(bool); ok {
			if !ready {
				isReady = ready
				return false
			}
		} else {
			isReady = false
		}
		return true
	})
	return isReady
}

// PutStatusSuccess set success
func (e *TryConfirmCancel) PutStatusSuccess(key string) error {
	// client
	if e.ETCDClient == nil {
		e.NewETCDClient()
	}

	if _, err := e.ETCDClient.Put(context.Background(), key, defaultTryConfirmCancelStatusSuccess); err != nil {
		return ErrTCCInvalidPut
	}
	return nil
}

// PutStatusFail set fail
func (e *TryConfirmCancel) PutStatusFail(key string) error {
	// client
	if e.ETCDClient == nil {
		e.NewETCDClient()
	}

	if _, err := e.ETCDClient.Put(context.Background(), key, defaultTryConfirmCancelStatusFail); err != nil {
		return ErrTCCInvalidPut
	}
	return nil
}
