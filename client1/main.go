package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/module/apmgin/v2"
	"go.elastic.co/apm/module/apmgrpc/v2"
	"go.elastic.co/apm/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "grpc-tracer/proto/server12"
	"log"
	"net/http"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:8002", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(apmgrpc.NewUnaryClientInterceptor()))
	if err != nil {
		log.Println("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)
	tracer, err := apm.NewTracerOptions(apm.TracerOptions{
		ServiceName:    "client1",
		ServiceVersion: "0.1",
	})
	r := gin.New()
	r.Use(apmgin.Middleware(r, apmgin.WithTracer(tracer)))
	r.GET("/", func(cc *gin.Context) {
		//tx := apm.DefaultTracer().StartTransaction("name-transaction-1", "type-transaction")
		//ctx := apm.ContextWithTransaction(context.Background(), tx)
		//span, ctx := apm.StartSpan(cc.Request.Context(), "server2", "endpoint-server2")
		//defer span.End()
		r, err := c.CallTwo(cc.Request.Context(), &pb.HelloRequest{Name: *name})
		if err != nil {
			apm.CaptureError(cc.Request.Context(), err).Send()
			log.Printf("could not greet: %v", err)
		}
		log.Printf("Greeting: %s", r.GetMessage())
		//defer tx.End()
		cc.String(http.StatusOK, "Hello, %s!")
	})
	r.GET("/next", func(cc *gin.Context) {
		fmt.Println("next", cc.Request.Header)
		cc.String(http.StatusOK, "Hello, %s!")
	})
	r.Run(":8001")
}
