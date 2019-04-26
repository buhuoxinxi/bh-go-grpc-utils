package bhgrpcutils

import (
	"fmt"
	"github.com/buhuoxinxi/bh-go-grpc-utils/etcd_balancer"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
)

func init() {
	// init config
	DefaultConfigFn()
}

// config
var (
	config *Config // config
)

// NewServer grpc server
func NewServer() *grpc.Server {
	// options
	var opts []grpc.ServerOption

	// ssl
	if config.SSLEnable {
		cred, err := credentials.NewServerTLSFromFile(config.SSLCertFile, config.SSLKeyFile)
		if err != nil {
			logrus.Panicf("credentials.NewServerTLSFromFile error : %v", err)
		}
		opts = append(opts, grpc.Creds(cred))
	}

	// unary interceptor
	opts = append(opts, DefaultUnaryInterceptorFn())

	// stream interceptor
	opts = append(opts, DefaultStreamInterceptorFn())

	return grpc.NewServer(opts...)
}

// RunServer start
func RunServer(server *grpc.Server) {
	// tcp
	lis, err := net.Listen("tcp", ":"+config.ServerPort)
	if err != nil {
		logrus.Panicf("RunServer net.Listen error : %v", err)
	}

	// server address
	serverAddr := net.JoinHostPort(config.ServerHost, config.ServerPort)
	logrus.Printf("server addr : %s", serverAddr)

	// register server to etcd
	if err := balancer.RegisterServer(serverAddr); err != nil {
		logrus.Panicf("balancer.RegisterServer error : %v", err)
	}

	// remove server from etcd
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch

		balancer.UnRegisterServer(serverAddr)

		if i, ok := s.(syscall.Signal); ok {
			os.Exit(int(i))
		} else {
			os.Exit(0)
		}
	}()

	// start
	if err := server.Serve(lis); err != nil {
		logrus.Panicf("RunServer server.Serve error : %v", err)
	}
}

// DefaultGRPCAuthorizationFn grpc auth
var DefaultGRPCAuthorizationFn = func(ctx context.Context) error {
	// metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// oauth2 authorization
		if data, ok := md["authorization"]; ok {
			_ = data
		}
	}
	// status.Error(codes.Unauthenticated, "auth fail")
	return nil
}

// DefaultUnaryInterceptorFn unary interceptor
var DefaultUnaryInterceptorFn = func() grpc.ServerOption {
	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		// request log
		go DefaultUnaryRequestLog(ctx, info.FullMethod, req)

		// auth
		if err := DefaultGRPCAuthorizationFn(ctx); err != nil {
			return nil, err
		}

		// recover
		defer func() {
			if e := recover(); e != nil {
				debug.PrintStack()
				err = status.Errorf(codes.Internal, "Panic err: %v", e)
			}
		}()

		// next
		return handler(ctx, req)
	}
	return grpc.UnaryInterceptor(interceptor)
}

// DefaultUnaryRequestLog unary request
var DefaultUnaryRequestLog = func(ctx context.Context, method string, reqParam interface{}) {
	// metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// authorization
		if data, ok := md["authorization"]; ok {
			_ = data
		}
	}
}

// DefaultStreamInterceptorFn stream interceptor
var DefaultStreamInterceptorFn = func() grpc.ServerOption {
	var interceptor = func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {

		// request log
		go DefaultStreamRequestLog(srv, ss, info)

		// auth
		if err := DefaultGRPCAuthorizationFn(ss.Context()); err != nil {
			return err
		}

		// recover
		defer func() {
			if e := recover(); e != nil {
				debug.PrintStack()
				err = status.Errorf(codes.Internal, "Panic err: %v", e)
			}
		}()

		// next
		return handler(srv, ss)
	}
	return grpc.StreamInterceptor(interceptor)
}

// DefaultStreamRequestLog unary request
var DefaultStreamRequestLog = func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo) {
	// unary request log
	DefaultUnaryRequestLog(
		ss.Context(),
		info.FullMethod,
		fmt.Sprintf("stream interceptor; isClient(%v), isServer(%v)", info.IsClientStream, info.IsServerStream),
	)
}
