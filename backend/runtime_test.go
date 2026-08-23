package main

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func slowServer(t *testing.T, started chan struct{}) *http.Server {
	t.Helper()
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if started != nil {
			close(started)
		}
		time.Sleep(2 * time.Second)
	})}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv.Addr = ln.Addr().String()
	go func() {
		_ = srv.Serve(ln)
	}()
	return srv
}

func startSlowRequest(t *testing.T, srv *http.Server, started chan struct{}) {
	t.Helper()
	go func() {
		res, err := http.Get("http://" + srv.Addr + "/")
		if err == nil {
			res.Body.Close()
		}
	}()
	<-started
}

func TestShutdownServerRespectsTimeout(t *testing.T) {
	started := make(chan struct{})
	srv := slowServer(t, started)
	defer srv.Close()
	startSlowRequest(t, srv, started)
	done := make(chan error, 1)
	go func() { done <- shutdownServer(srv, 200*time.Millisecond) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("存在慢请求时关停应返回超时错误")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("关停忽略超时，无限挂起")
	}
}

func TestRequestTimeoutPropagatesCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var observed error
	done := make(chan struct{})
	handler := requestTimeoutMiddleware(time.Minute, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observed = r.Context().Err()
		close(done)
	}))
	cancel()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	handler.ServeHTTP(rec, req)
	<-done
	if observed == nil {
		t.Fatal("请求取消未向下游传播")
	}
}
