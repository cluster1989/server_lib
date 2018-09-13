package libgrpc

import (
	"context"
	"net"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc"
)

type ServerOptions struct {
	Address string
	Cert    *ServerSecureOptions
	Auth    func(ctx context.Context) error
}

type ServerSecureOptions struct {
	CertFile string
	KeyFile  string
}

type Server struct {
	*grpc.Server

	serverOptions *ServerOptions
}

func NewConf() *ServerOptions {

	options := &ServerOptions{}
	return options
}

func NewServer(options *ServerOptions) *Server {
	s := &Server{}
	s.serverOptions = options
	var opts []grpc.ServerOption

	//tls
	if options.Cert != nil {
		creds, err := credentials.NewServerTLSFromFile(options.Cert.CertFile, options.Cert.KeyFile)
		if err != nil {
			panic(err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	//options auth
	if options.Auth != nil {
		var interceptor grpc.UnaryServerInterceptor
		interceptor = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			err = options.Auth(ctx)
			if err != nil {
				return
			}
			// 继续处理请求
			return handler(ctx, req)
		}

		opts = append(opts, grpc.UnaryInterceptor(interceptor))
	}
	s.Server = grpc.NewServer(opts...)
	return s
}

func (s *Server) RPCServe() error {

	listen, err := net.Listen("tcp", s.serverOptions.Address)

	if err != nil {
		panic(err)
	}
	return s.Serve(listen)
}
