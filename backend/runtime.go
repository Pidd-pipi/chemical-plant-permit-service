package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

var requestSequence uint64

func serveAddress(address string, shutdownTimeout time.Duration, handler http.Handler) error {
	server := newEnterpriseServer(address, handler)
	return serveHTTP(server, shutdownTimeout)
}

func serveHTTP(server *http.Server, shutdownTimeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-signals:
		return shutdownServer(server, shutdownTimeout)
	}
}

// shutdownTimeoutDefault 是未显式传入关停超时时的兜底值，
// 防止关停被慢请求无限期挂住、拖累重启。
const shutdownTimeoutDefault = 10 * time.Second

// requestTimeoutDefault 为每个请求附加的超时上限。
// 取值大于 WriteTimeout，避免与底层写超时打架；同时以 r.Context() 为父，
// 使客户端断连或服务关停产生的取消信号能立即下传到 handler。
const requestTimeoutDefault = 30 * time.Second

// shutdownServer 优雅关闭服务：用带 deadline 的上下文调用 Shutdown，
// 到期后强制返回，避免无限期等待慢请求而卡死关停、拖累重启。
func shutdownServer(server *http.Server, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = shutdownTimeoutDefault
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return server.Shutdown(ctx)
}

// requestTimeoutMiddleware 为每个请求附加超时上下文。
// 以 r.Context() 为父，使客户端断连或服务关停产生的取消信号能下传到下游 handler。
func requestTimeoutMiddleware(timeout time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if timeout <= 0 {
			next.ServeHTTP(w, r)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newEnterpriseServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:                address,
		Handler:             requestTimeoutMiddleware(requestTimeoutDefault, opsEnterpriseMiddleware(requestIDMiddleware(recoveryMiddleware(handler)))),
		ReadHeaderTimeout:   5 * time.Second,
		ReadTimeout:         15 * time.Second,
		WriteTimeout:        15 * time.Second,
		IdleTimeout:         60 * time.Second,
		MaxHeaderBytes:      1 << 20,
	}
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req-%d", atomic.AddUint64(&requestSequence, 1))
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic recovered request_id=%s method=%s path=%s panic=%v", w.Header().Get("X-Request-ID"), r.Method, r.URL.Path, recovered)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
