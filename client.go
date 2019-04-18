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
	// resolver
	r := balancer.NewResolver()
	resolver.Register(r)

	// options
	var opts []grpc.DialOption

	// ssl
	if config.SSLEnable {
		cred, err := credentials.NewClientTLSFromFile(config.SSLCertFile, config.SSLServerName)
		if err != nil {
			logrus.Fatalf("NewClient credentials.NewClientTLSFromFile error : %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(cred))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// balancer name
	opts = append(opts, grpc.WithBalancerName(roundrobin.Name))

	// server config
	serverConfig := balancer.GetServerConfig()

	// server address
	serverAddr := r.Scheme() + "://ikaiguang/" + serverConfig.ServerName
	logrus.Printf("dial : %s", serverAddr)

	// client
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		logrus.Fatalf("NewClient grpc.Dial error : %v", err)
	}
	//defer conn.Close()

	return conn
}
