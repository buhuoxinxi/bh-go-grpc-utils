package balancer

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewDistributedLock new lock
func NewDistributedLock(lockKey string) (*DistributedLock, error) {
	return new(DistributedLock).GetLock(lockKey)
}

// etcd distributed lock config
const (
	defaultLockAliveTTL int64 = 3 // etcd alive ttl(3s)
)

// err
var (
	ErrLockInvalidLease    = status.Error(codes.Internal, "ETCDClient.Grant lease fail")
	ErrLockCannotKeepAlive = status.Error(codes.Internal, "lockLease.KeepAlive fail")
	ErrLockTxnCannotCommit = status.Error(codes.Internal, "ETCDClient.txn.Commit fail")
	ErrLockIsLocking       = status.Error(codes.Canceled, "get lock fail")
)

// DistributedLock distributed lock
type DistributedLock struct {
	ETCDClient   *clientv3.Client   // etcd client
	LockKeyTTL   int64              // lock key etcd ttl
	LockLease    clientv3.Lease     // lock lease
	LockLeaseId  clientv3.LeaseID   // lock key ttl lease id
	LockCancelFn context.CancelFunc // unlock
}

// NewETCDClient etcd client
func (d *DistributedLock) NewETCDClient() {
	d.ETCDClient = etcdClient
}

// UnLock unlock
func (d *DistributedLock) UnLock() {
	// Revoke lease
	d.LockLease.Revoke(context.TODO(), d.LockLeaseId)
	// cancel lock
	if d.LockCancelFn != nil {
		d.LockCancelFn()
	}
}

// GetLock get lock
func (d *DistributedLock) GetLock(lockKey string) (*DistributedLock, error) {
	// client
	if d.ETCDClient == nil {
		d.NewETCDClient()
	}

	// etcd ttl
	if d.LockKeyTTL <= 0 {
		d.LockKeyTTL = defaultLockAliveTTL
	}

	// read to lock
	d.LockLease = clientv3.NewLease(d.ETCDClient)

	// etcd ttl
	if leaseResp, err := d.LockLease.Grant(context.TODO(), d.LockKeyTTL); err != nil {
		return nil, ErrLockInvalidLease
	} else {
		d.LockLeaseId = leaseResp.ID
	}

	// set a cancel context
	cancelCtx, cancelFunc := context.WithCancel(context.TODO())
	d.LockCancelFn = cancelFunc

	// lease keep alive
	if _, err := d.LockLease.KeepAlive(cancelCtx, d.LockLeaseId); err != nil {
		d.UnLock()
		return nil, ErrLockCannotKeepAlive
	}

	// lock key value
	kv := clientv3.NewKV(d.ETCDClient)

	// start transaction
	txn := kv.Txn(context.TODO())

	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "locking", clientv3.WithLease(d.LockLeaseId))).
		Else(clientv3.OpGet(lockKey))

	// commit transaction
	if txtResp, err := txn.Commit(); err != nil {
		d.UnLock()
		return nil, ErrLockTxnCannotCommit
	} else {
		// cannot get lock
		if !txtResp.Succeeded {
			d.UnLock()
			//logrus.Println("is locking ：", string(txtResp.Responses[0].GetResponseRange().Kvs[0].Value))
			return nil, ErrLockIsLocking
		}
	}
	return d, nil
}

// listenLeaseChan listen lease
func (d *DistributedLock) listenLeaseChan(leaseRespChan <-chan *clientv3.LeaseKeepAliveResponse) {
	for {
		select {
		case leaseKeepResp := <-leaseRespChan:
			if leaseKeepResp == nil {
				//logrus.Infof("lease invalid")
				goto END
			}
		}
	}
END:
}
