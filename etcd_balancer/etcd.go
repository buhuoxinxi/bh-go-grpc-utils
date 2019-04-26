package balancer

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
)

func init() {
	// init config
	DefaultETCDConfigFn()

	// new client
	c, err := clientv3.New(*etcdConfig)
	if err != nil {
		logrus.Panicf("[E] etcd clientv3.New fail : %v", err)
	}
	etcdClient = c
}

// etcd
var (
	etcdConfig *clientv3.Config // config
	etcdClient *clientv3.Client // etcd client
)

// SetETCDConfig etcd config
func SetETCDConfig(cfg *clientv3.Config) {
	etcdConfig = cfg
}

// NewETCDClient new etcd client
func NewETCDClient(cfg *clientv3.Config) (*clientv3.Client, error) {
	return clientv3.New(*cfg)
}

// GetETCDClient get client
func GetETCDClient() *clientv3.Client {
	return etcdClient
}

// GetETCDConfig get config
//func GetETCDConfig() *clientv3.Config {
//	return etcdConfig
//}
