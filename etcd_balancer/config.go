package balancer

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
)

// etcd
//////////////////////////////////////////////////////////////////////////////////////////

// etcd config
const (
	defaultETCDEndpoints   = "127.0.0.1:2379" // etcd endpoints
	defaultETCDDialTimeout = 3 * time.Second  // dial timeout
)

// etcd env
const (
	envKeyETCDEndPoints   = "BhETCDEndpoints"   // endpoints
	envSepETCDEndPoints   = ","                 // endpoints separators
	envKeyEtCDDialTimeout = "BhETCDDialTimeout" // timeout
)

// server
//////////////////////////////////////////////////////////////////////////////////////////

// server config
const (
	defaultServerResolverSchema       = "bh_ikaigunag"        // schema name
	defaultServerName                 = "bh_ikaigunag_server" // server name
	defaultServerETCDAliveTTL   int64 = 5                     // etcd ttl(second)
)

// server env
const (
	envKeyServerResolverSchema = "BhServerResolverSchema" // schema
	envKeyServerName           = "BhServerName"           // server
	envKeyServerETCDAliveTTL   = "BhServerETCDAliveTTL"   // ttl
)

// etcd
//////////////////////////////////////////////////////////////////////////////////////////

// DefaultETCDConfigFn init config
var DefaultETCDConfigFn = func() {
	// config
	var cfg = clientv3.Config{
		DialTimeout: defaultETCDDialTimeout,
	}

	// etcd address
	if addr := strings.TrimSpace(os.Getenv(envKeyETCDEndPoints)); len(addr) > 0 {
		cfg.Endpoints = strings.Split(addr, envSepETCDEndPoints)
	} else {
		cfg.Endpoints = strings.Split(defaultETCDEndpoints, envSepETCDEndPoints)
	}

	// dial timeout
	if timeString := strings.TrimSpace(os.Getenv(envKeyEtCDDialTimeout)); len(timeString) > 0 {
		if duration, _ := time.ParseDuration(timeString); duration > 0 {
			cfg.DialTimeout = duration
		}
	}

	// init config
	SetETCDConfig(&cfg)
}

// server
//////////////////////////////////////////////////////////////////////////////////////////

// ServerConfig server config
type ServerConfig struct {
	SchemaName   string // resolver schema name
	ServerName   string // server name
	ETCDAliveTTL int64  // etcd ttl(second)
}

// DefaultServerConfigFn server config
var DefaultServerConfigFn = func() {
	// config
	var cfg = ServerConfig{
		SchemaName:   defaultServerResolverSchema,
		ServerName:   defaultServerName,
		ETCDAliveTTL: defaultServerETCDAliveTTL,
	}

	// resolver schema
	if schema := strings.TrimSpace(os.Getenv(envKeyServerResolverSchema)); len(schema) > 0 {
		cfg.SchemaName = schema
	}

	// server name
	if name := strings.TrimSpace(os.Getenv(envKeyServerName)); len(name) > 0 {
		cfg.ServerName = name
	}

	// etcd alive
	if int64String := strings.TrimSpace(os.Getenv(envKeyServerETCDAliveTTL)); len(int64String) > 0 {
		if n, _ := strconv.ParseInt(int64String, 10, 64); n > 0 {
			cfg.ETCDAliveTTL = n
		}
	}

	// init config
	SetServerConfig(&cfg)
}
