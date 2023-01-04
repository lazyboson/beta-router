package server

import (
	"context"
	"log"
	"net/http"

	pb "github.com/lazyboson/beta-router/pkg/pb/apipb"
	"github.com/lazyboson/beta-router/pkg/router"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type APIServer struct {
	grpcServerPort int
	httpServerPort int
	grpcServer     *grpc.Server
	httpMux        *http.ServeMux
	pb.UnimplementedAPIServiceServer

	r *router.Router
}

func NewAPIServer(port int, conf *router.Config) *APIServer {
	server := &APIServer{
		grpcServerPort: port,
		httpServerPort: port + 1,
	}
	gs := grpc.NewServer(grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(nil, grpcPrometheus.UnaryServerInterceptor)))
	server.grpcServer = gs
	reflection.Register(server.grpcServer)
	pb.RegisterAPIServiceServer(gs, server)

	gMux := runtime.NewServeMux()
	err := pb.RegisterAPIServiceHandlerServer(context.Background(), gMux, server)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gMux)
	server.httpMux = mux

	server.r = router.NewRouter(conf)

	return server
}

func (s *APIServer) StopServer() {
	s.grpcServer.GracefulStop()
}

func (s *APIServer) TaskEvents(ctx context.Context, req *pb.TaskCreationEventRequest) (*pb.TaskEventResponse, error) {
	res := s.r.ListenEvents(req)

	if res.Message == "" {
		return res, status.Error(codes.Internal, res.Message)
	}

	return res, nil
}
