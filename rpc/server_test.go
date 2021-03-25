package rpc

import (
	"context"
	"testing"

	"github.com/appleboy/gorush/gorush"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

const gRPCAddr = "localhost:9003"

func TestGracefulShutDownGRPCServer(t *testing.T) {
	// server configs
	gorush.InitLog()
	gorush.PushConf.GRPC.Enabled = true
	gorush.PushConf.GRPC.Port = "9003"
	gorush.PushConf.Log.Format = "json"

	// Run gRPC server
	ctx, gRPCContextCancel := context.WithCancel(context.Background())
	go func() {
		if err := RunGRPCServer(ctx); err != nil {
			panic(err)
		}
	}()

	// gRPC client conn
	conn, err := grpc.Dial(
		gRPCAddr,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	) // wait for server ready
	if err != nil {
		t.Error(err)
	}

	// Stop gRPC server
	go gRPCContextCancel()

	// wait for client connection would be closed
	for conn.GetState() != connectivity.TransientFailure {
	}
	conn.Close()
}
