package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm/module/apmgrpc/v2"
	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"
	"golang.org/x/net/context/ctxhttp"
	"google.golang.org/grpc"
	pb "grpc-tracer/proto/server23"
	"log"
	"net"
	"net/http"
	"os"
)

var (
	port = flag.Int("port", 8003, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) CallThree(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	//span, ctx := apm.StartSpan(ctx, "server3", "endpoint-server3")
	//span.End()
	//err := callApi(ctx)
	//if err != nil {
	//	log.Println(err)
	//	return nil, errors.New("")
	//}
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func acquire() (*fasthttp.Request, *fasthttp.Response) {
	return fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
}

func release(req *fasthttp.Request, res *fasthttp.Response) {
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}

func callApi(ctx context.Context) error {
	span, ctx := apm.StartSpan(ctx, "call http api", "endpoint-server3")
	defer span.End()
	client := apmhttp.WrapClient(http.DefaultClient)
	_, err := ctxhttp.Get(ctx, client, "http://localhost:8001/next")

	//req, _ := http.NewRequest("GET", "http://localhost:8001/next", nil)
	//client := apmhttp.WrapClient(http.DefaultClient)
	//_, err := client.Do(req.WithContext(ctx))

	return err
	//req, resp := acquire()
	//defer release(req, resp)
	//req.SetRequestURI("http://localhost:8001/next")
	//req.Header.SetMethod("GET")
	//client := &fasthttp.Client{}
	//err := client.Do(req, resp)
	//return err
}

func main() {
	flag.Parse()

	os.Setenv("ELASTIC_APM_SERVER_URL", "http://192.168.55.54:9965")
	os.Setenv("ELASTIC_APM_SECRET_TOKEN", "UDVySFJZQUIxTlk3MzBzVVhwLTg6MmFkbkJsTExUd21CNldWa2NwaFhSdw==")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Println("failed to listen: %v", err)
	}
	tracer, err := apm.NewTracerOptions(apm.TracerOptions{
		ServiceName:    "server3",
		ServiceVersion: "0.1",
	})

	//s := grpc.NewServer(grpc.UnaryInterceptor(apmgrpc.NewUnaryServerInterceptor(apmgrpc.WithTracer(tracer))))
	s := grpc.NewServer(grpc.UnaryInterceptor(
		apmgrpc.NewUnaryServerInterceptor(
			apmgrpc.WithTracer(tracer),
			apmgrpc.WithRecovery(),
		)))
	defer s.GracefulStop()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Printf("failed to serve: %v", err)
	}
}
