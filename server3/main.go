package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm/module/apmgrpc/v2"
	"go.elastic.co/apm/v2"
	"google.golang.org/grpc"
	pb "grpc-tracer/proto/server23"
	"log"
	"net"
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
	err := callApi()
	if err != nil {
		log.Println(err)
		return nil, errors.New("")
	}
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

func callApi() error {
	req, resp := acquire()
	defer release(req, resp)
	req.SetRequestURI("http://localhost:8001/next")
	req.Header.SetMethod("GET")
	err := fasthttp.Do(req, resp)
	return err
}

func main() {
	flag.Parse()
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
