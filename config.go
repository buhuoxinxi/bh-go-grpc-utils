package bhgrpcutils

import (
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// server config
const (
	defaultServerPort = "50051" // default port
)

// server env
const (
	envKeyServerHost          = "BhServerHost"          // server host
	envKeyServerPort          = "BhServerPort"          // server port
	envKeyServerSSLEnable     = "BhServerSSLEnable"     // ssl enable
	envKeyServerSSLCaFile     = "BhServerSSLCaFile"     // ssl ca
	envKeyServerSSLCertFile   = "BhServerSSLCertFile"   // ssl cert
	envKeyServerSSLKeyFile    = "BhServerSSLKeyFile"    // ssl key
	envKeyServerSSLServerName = "BhServerSSLServerName" // ssl server name
)

// Config server config
type Config struct {
	ServerHost    string // server host
	ServerPort    string // server port
	SSLEnable     bool   // ssl enable
	SSLCaFile     string // ssl ca file path
	SSLCertFile   string // ssl cert file path
	SSLKeyFile    string // ssl key file path
	SSLServerName string // ssl name
}

// SetConfig set config
func SetConfig(cfg *Config) {
	config = cfg
}

// DefaultConfigFn init config
var DefaultConfigFn = func() {
	// init config
	var cfg Config

	// host
	if host := strings.TrimSpace(os.Getenv(envKeyServerHost)); len(host) > 0 {
		cfg.ServerHost = host
	} else {
		cfg.ServerHost = getLocalIPV4()
	}

	// port
	if port := strings.TrimSpace(os.Getenv(envKeyServerPort)); len(port) > 0 {
		cfg.ServerPort = strings.TrimPrefix(port, ":")
	} else {
		cfg.ServerPort = defaultServerPort
	}

	// ssl enable
	cfg.SSLEnable, _ = strconv.ParseBool(strings.TrimSpace(os.Getenv(envKeyServerSSLEnable)))
	if cfg.SSLEnable {
		parseServerSSLEnv(&cfg)
	}

	// init
	SetConfig(&cfg)
}

// parseServerSSLEnv parse server ssl env
func parseServerSSLEnv(cfg *Config) {
	// pwd
	pwdPath, err := os.Getwd()
	if err != nil {
		logrus.Errorf("os.Getwd error : %v", err)
	}

	// ssl ca
	if p := strings.TrimSpace(os.Getenv(envKeyServerSSLCaFile)); len(p) > 0 {
		if !filepath.IsAbs(p) {
			p = filepath.Join(pwdPath, p)
		}
		cfg.SSLCaFile = p
	}

	// ssl cert
	if p := strings.TrimSpace(os.Getenv(envKeyServerSSLCertFile)); len(p) > 0 {
		if !filepath.IsAbs(p) {
			p = filepath.Join(pwdPath, p)
		}
		cfg.SSLCertFile = p
	}

	// ssl key
	if p := strings.TrimSpace(os.Getenv(envKeyServerSSLKeyFile)); len(p) > 0 {
		if !filepath.IsAbs(p) {
			p = filepath.Join(pwdPath, p)
		}
		cfg.SSLKeyFile = p
	}

	// ssl server name
	if name := strings.TrimSpace(os.Getenv(envKeyServerSSLServerName)); len(name) > 0 {
		cfg.SSLServerName = name
	}
}

// getLocalIPV4 get local ipv4
func getLocalIPV4() string {
	addrList, err := net.InterfaceAddrs()
	if err != nil {
		logrus.Panicf("net.InterfaceAddrs error : %v", err)
	}

	for _, addr := range addrList {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ip.IP.To4() == nil {
				continue
			}
			return ip.IP.String()
		}
	}
	return "127.0.0.1"
}
