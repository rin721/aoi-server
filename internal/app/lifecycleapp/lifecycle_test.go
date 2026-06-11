package lifecycleapp

import (
	"context"
	"errors"
	"testing"

	"github.com/rei0721/go-scaffold/internal/app/initapp"
	"github.com/rei0721/go-scaffold/pkg/httpserver"
	"github.com/rei0721/go-scaffold/pkg/rpcserver"
)

func TestStartShutsDownHTTPWhenRPCStartFails(t *testing.T) {
	t.Parallel()

	httpSrv := &fakeLifecycleHTTPServer{}
	rpcSrv := &fakeLifecycleRPCServer{startErr: errors.New("bind failed")}

	err := Start(context.Background(), initapp.Transport{HTTPServer: httpSrv, RPCServer: rpcSrv})
	if err == nil {
		t.Fatal("Start() error = nil, want RPC start error")
	}
	if httpSrv.starts != 1 {
		t.Fatalf("HTTP starts = %d, want 1", httpSrv.starts)
	}
	if httpSrv.shutdowns != 1 {
		t.Fatalf("HTTP shutdowns = %d, want rollback shutdown", httpSrv.shutdowns)
	}
	if rpcSrv.starts != 1 {
		t.Fatalf("RPC starts = %d, want 1", rpcSrv.starts)
	}
}

func TestShutdownStopsHTTPAndRPC(t *testing.T) {
	t.Parallel()

	httpSrv := &fakeLifecycleHTTPServer{}
	rpcSrv := &fakeLifecycleRPCServer{}

	if err := Shutdown(context.Background(), initapp.Core{}, initapp.Infrastructure{}, initapp.Transport{
		HTTPServer: httpSrv,
		RPCServer:  rpcSrv,
	}); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if httpSrv.shutdowns != 1 {
		t.Fatalf("HTTP shutdowns = %d, want 1", httpSrv.shutdowns)
	}
	if rpcSrv.shutdowns != 1 {
		t.Fatalf("RPC shutdowns = %d, want 1", rpcSrv.shutdowns)
	}
}

type fakeLifecycleHTTPServer struct {
	httpserver.HTTPServer
	starts    int
	shutdowns int
}

func (s *fakeLifecycleHTTPServer) Start(context.Context) error {
	s.starts++
	return nil
}

func (s *fakeLifecycleHTTPServer) Shutdown(context.Context) error {
	s.shutdowns++
	return nil
}

type fakeLifecycleRPCServer struct {
	rpcserver.Server
	starts    int
	shutdowns int
	startErr  error
}

func (s *fakeLifecycleRPCServer) Start(context.Context) error {
	s.starts++
	return s.startErr
}

func (s *fakeLifecycleRPCServer) Shutdown(context.Context) error {
	s.shutdowns++
	return nil
}
