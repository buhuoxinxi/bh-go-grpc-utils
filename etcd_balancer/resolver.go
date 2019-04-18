package balancer

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/resolver"
)

// NewResolver initialize an etcd client
func NewResolver() resolver.Builder {
	return new(etcdResolver)
}

// etcdResolver etcd resolver
type etcdResolver struct {
	cc resolver.ClientConn
}

// Build creates a new resolver for the given target.
//
// gRPC dial calls Build synchronously, and fails if the returned error is
// not nil.
func (r *etcdResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {

	r.cc = cc

	go r.watch("/" + target.Scheme + "/" + target.Endpoint + "/")

	return r, nil
}

// Scheme returns the scheme supported by this resolver.
// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
func (r etcdResolver) Scheme() string {
	return serverConfig.SchemaName
}

// ResolveNow will be called by gRPC to try to resolve the target name
// again. It's just a hint, resolver can ignore this if it's not necessary.
//
// It could be called multiple times concurrently.
func (r etcdResolver) ResolveNow(rn resolver.ResolveNowOption) {
	// resolve
}

// Close closes the resolver.
func (r etcdResolver) Close() {
	// close
}

func (r *etcdResolver) watch(keyPrefix string) {
	// server addr
	var addrList []resolver.Address

	// etcd key value
	getResp, err := etcdClient.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		logrus.Errorf("[E] etcdClient.Get error : " + err.Error())
	} else {
		for i := range getResp.Kvs {
			addrList = append(addrList, resolver.Address{Addr: string(getResp.Kvs[i].Value)})
		}
	}

	//r.cc.NewAddress(addrList)
	r.cc.UpdateState(resolver.State{Addresses: addrList})

	// watch
	rch := etcdClient.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	for n := range rch {
		// etcd events
		for _, ev := range n.Events {
			//logrus.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			switch ev.Type {
			case mvccpb.PUT:
				// put
				if !existFromResolverAddr(addrList, string(ev.Kv.Value)) {
					addrList = append(addrList, resolver.Address{Addr: string(ev.Kv.Value)})
					r.cc.UpdateState(resolver.State{Addresses: addrList})
				}

			case mvccpb.DELETE:
				// delete
				if s, ok := removeFromResolverAddr(addrList, string(ev.Kv.Value)); ok {
					addrList = s
					r.cc.UpdateState(resolver.State{Addresses: addrList})
				}
			}
		}
	}
}

// existFromResolverAddr check server addr
func existFromResolverAddr(l []resolver.Address, addr string) bool {
	for i := range l {
		if l[i].Addr == addr {
			return true
		}
	}
	return false
}

// removeFromResolverAddr remove server addr
func removeFromResolverAddr(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}
