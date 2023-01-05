package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/lazyboson/beta-router/pkg/api/server"
	"github.com/lazyboson/beta-router/pkg/router"
	"github.com/subosito/gotenv"
)

type RouterService struct {
	printVersion *bool

	//grpc server port
	port      int
	conf      *router.Config
	apiServer *server.APIServer
}

const (
	envQueueBaseUrl = "QUEUE_BASE_URL"
	envCtiBaseUrl   = "CTI_BASE_URL"
	envGrpcPort     = "ROUTER_GRPC_PORT"
)

func new() (rs *RouterService) {
	rs = &RouterService{}
	return
}

func (rs *RouterService) start() {
	rs.apiServer = server.NewAPIServer(rs.port, rs.conf)
	rs.apiServer.StartServer()
}

func (rs *RouterService) Configure() {
	gotenv.Load("routerservice.env")
	grpcPort := os.Getenv(envGrpcPort)
	if grpcPort != "" {
		rs.port, _ = strconv.Atoi(grpcPort)
	}
	rs.conf = loadConf()
	flag.Parse()
}

func loadConf() *router.Config {
	routerConfig := &router.Config{}

	queueBaseUrl := os.Getenv(envQueueBaseUrl)
	if queueBaseUrl != "" {
		routerConfig.QueueBaseUrl = queueBaseUrl
	}

	ctiBaseUrl := os.Getenv(envCtiBaseUrl)
	if ctiBaseUrl != "" {
		routerConfig.CtiBaseUrl = ctiBaseUrl
	}

	return routerConfig
}

func (rs *RouterService) stop() {
	rs.apiServer.StopServer()
}

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	router := new()
	router.Configure()

	router.start()

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println(sig)
		done <- true
	}()

	<-done

	router.stop()
}
