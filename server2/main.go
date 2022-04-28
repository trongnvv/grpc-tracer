package main

import (
	"context"
	"flag"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
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
	//tx := apm.TransactionFromContext(ctx)
	//tx.Context.SetLabel("request", in)
	//tx.Context.SetLabel("response", &pb.HelloReply{Message: "trongnv"})

	log.Printf("Received:")
	//cc := tx.TraceContext()
	//fmt.Println(cc.Trace.String())
	//fmt.Println(cc.Span.String())
	//span, ctx := apm.StartSpan(ctx, "server2", "endpoint-server2")
	//defer span.End()
	//md, ok := metadata.FromIncomingContext(ctx)
	//if ok {
	//	fmt.Println(md)
	//}

	r, err := client.CallThree(ctx, &pb23.HelloRequest{Name: "name 2"})

	if err != nil {
		log.Println("could not greet: %v", err)
		apm.CaptureError(ctx, err).Send()
		return nil, err
	}

	log.Printf("Greeting: %s", r.GetMessage())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func customUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, resp interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		apm.SpanFromContext(ctx).Context.SetLabel("request", req)
		return invoker(ctx, method, req, resp, cc, opts...)
	}
}

func customUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		//tx := apm.TransactionFromContext(ctx)
		fmt.Println("***********server*******", req)
		//tx.Context.SetLabel("request-s", req)
		//tx.Context.SetLabel("response-s", resp)
		return handler(ctx, req)
	}
}

func main() {
	flag.Parse()
	//conn, err := grpc.Dial("localhost:8003",
	//	grpc.WithTransportCredentials(insecure.NewCredentials()),
	//	grpc.WithUnaryInterceptor(
	//		apmgrpc.NewUnaryClientInterceptor(),
	//	),
	//)

	tracer, err := apm.NewTracer("server2", "0.1")

	conn, err := grpc.Dial("localhost:8003",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				apmgrpc.NewUnaryClientInterceptor(),
				customUnaryClientInterceptor(),
			),
		),
	)
	if err != nil {
		log.Println("err", err)
	}
	defer conn.Close()
	client = pb23.NewGreeterClient(conn)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Println("failed to listen: %v", err)
	}

	//s := grpc.NewServer(grpc.UnaryInterceptor(
	//	apmgrpc.NewUnaryServerInterceptor(
	//		apmgrpc.WithTracer(tracer),
	//		apmgrpc.WithRecovery(),
	//	),
	//))

	s := grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		apmgrpc.NewUnaryServerInterceptor(
			apmgrpc.WithTracer(tracer),
			apmgrpc.WithRecovery(),
		),
		customUnaryServerInterceptor()),
	))

	defer s.GracefulStop()
	defer tracer.Close()

	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Println("failed to serve: %v", err)
	}
}
