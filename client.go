package bhgrpcutils

import (
	"github.com/buhuoxinxi/bh-go-grpc-utils/etcd_balancer"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/resolver"
)

// NewClient grpc.Dail()
func NewClient() *grpc.ClientConn {
	// server config
	serverConfig := balancer.GetServerConfig()

	return newClient(serverConfig.ServerName)
}

// NewClientWithServerName grpc.Dail()
func NewClientWithServerName(serverName string) *grpc.ClientConn {
	return newClient(serverName)
}

// newClient grpc.Dail()
func newClient(serverName string) *grpc.ClientConn {
	// resolver
	r := balancer.NewResolver()
	resolver.Register(r)

	// options
	var opts []grpc.DialOption

	// ssl
	if config.SSLEnable {
		cred, err := credentials.NewClientTLSFromFile(config.SSLCertFile, config.SSLServerName)
		if err != nil {
			logrus.Panicf("NewClient credentials.NewClientTLSFromFile error : %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(cred))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// balancer name
	opts = append(opts, grpc.WithBalancerName(roundrobin.Name))

	// server address
	serverAddr := r.Scheme() + "://ikaiguang/" + serverName
	logrus.Printf("dial : %s", serverAddr)

	// client
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		logrus.Panicf("NewClient grpc.Dial error : %v", err)
	}
	//defer conn.Close()

	return conn
}
