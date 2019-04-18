package balancer

import (
	"context"
	"errors"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
)

func init() {
	DefaultServerConfigFn()
}

// resolver config
var (
	serverConfig *ServerConfig // server config
)

// SetServerConfig server config
func SetServerConfig(cfg *ServerConfig) {
	serverConfig = cfg
}

// getServerETCDPrefix register server etcd key prefix
func getServerETCDPrefix(cfg *ServerConfig) string {
	return "/" + cfg.SchemaName + "/" + cfg.ServerName + "/"
}

// getServerETCDKey register server key
func getServerETCDKey(cfg *ServerConfig, serverAddr string) string {
	return getServerETCDPrefix(cfg) + serverAddr
}

// RegisterServer register service with name as prefix to etcd
func RegisterServer(serverAddr string) error {

	ticker := time.NewTicker(time.Duration(serverConfig.ETCDAliveTTL) * time.Second)

	go func() {
		for {
			// key exist
			getResp, err := etcdClient.Get(context.Background(), getServerETCDKey(serverConfig, serverAddr))
			if err != nil {
				logrus.Printf("[E] etcdClient.Get error : " + err.Error())
			} else if getResp.Count == 0 {
				// register server and keep alive
				if err = registerServerAndKeepAlive(serverAddr); err != nil {
					logrus.Error("registerServerAndKeepAlive error : " + err.Error())
				}
			} else {
				// do nothing
			}

			<-ticker.C
		}
	}()
	return nil
}

// registerServerAndKeepAlive register server and keep alive
func registerServerAndKeepAlive(serverAddr string) error {
	// lease TTL is ttl-second
	leaseResp, err := etcdClient.Grant(context.Background(), serverConfig.ETCDAliveTTL)
	if err != nil {
		return errors.New("[E] etcdClient.Grant error : " + err.Error())
	}

	// etcd key
	etcdKey := getServerETCDKey(serverConfig, serverAddr)
	logrus.Printf("[info] etcd key : %v\n", etcdKey)

	// save to etcd
	_, err = etcdClient.Put(context.Background(), etcdKey, serverAddr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return errors.New("[E] etcdClient.Put error : " + err.Error())
	}

	// keep alive
	if _, err = etcdClient.KeepAlive(context.Background(), leaseResp.ID); err != nil {
		return errors.New("[E] etcdClient.KeepAlive error : " + err.Error())
	}
	return nil
}

// UnRegisterServer remove server from etcd
func UnRegisterServer(serverAddr string) {
	etcdClient.Delete(context.Background(), getServerETCDKey(serverConfig, serverAddr))
}

// GetServerConfig get config
func GetServerConfig() *ServerConfig {
	return serverConfig
}
