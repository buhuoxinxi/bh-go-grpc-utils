package bhtestdata

import (
	"google.golang.org/grpc/testdata"
	"os"
)

func init() {

	// host
	os.Setenv("BhServerHost", "")
	os.Setenv("BhServerPort", "50051")

	// resolver
	os.Setenv("BhServerResolverSchema", "bh_ikaigunag")
	os.Setenv("BhServerName", "bh_ikaigunag_server")
	os.Setenv("BhServerETCDAliveTTL", "5")

	// etcd
	os.Setenv("BhETCDEndpoints", "127.0.0.1:2379")
	os.Setenv("BhETCDDialTimeout", "3s")

	// ssl
	os.Setenv("BhServerSSLEnable", "true")
	os.Setenv("BhServerSSLCaFile", testdata.Path("ca.pem"))
	os.Setenv("BhServerSSLCertFile", testdata.Path("server1.pem"))
	os.Setenv("BhServerSSLKeyFile", testdata.Path("server1.key"))
	os.Setenv("BhServerSSLServerName", "x.test.youtube.com")
}
