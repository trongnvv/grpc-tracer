package main

import (
	"context"
	"flag"
	"fmt"
	"go.elastic.co/apm/module/apmgrpc/v2"
	"go.elastic.co/apm/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "grpc-tracer/proto/server12"
	pb23 "grpc-tracer/proto/server23"
	"log"
	"net"
)

var (
	port = flag.Int("port", 8002, "The server port")
)

var client pb23.GreeterClient

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) CallTwo(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	//span, ctx := apm.StartSpan(ctx, "server2", "endpoint-server2")
	//defer span.End()

	r, err := client.CallThree(ctx, &pb23.HelloRequest{Name: "name 2"})
	if err != nil {
		log.Println("could not greet: %v", err)
		apm.CaptureError(ctx, err).Send()
		return nil, err
	}
	log.Printf("Greeting: %s", r.GetMessage())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	flag.Parse()
	conn, err := grpc.Dial("localhost:8003", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(apmgrpc.NewUnaryClientInterceptor()))
	if err != nil {
		log.Println("err", err)
	}
	defer conn.Close()
	client = pb23.NewGreeterClient(conn)

	/////////////////////////
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Println("failed to listen: %v", err)
	}
	tracer, err := apm.NewTracerOptions(apm.TracerOptions{
		ServiceName:    "server2",
		ServiceVersion: "0.1",
	})
	s := grpc.NewServer(grpc.UnaryInterceptor(
		apmgrpc.NewUnaryServerInterceptor(
			apmgrpc.WithTracer(tracer),
			apmgrpc.WithRecovery(),
		)))
	defer s.GracefulStop()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Println("failed to serve: %v", err)
	}
}
