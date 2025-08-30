package main

import (
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/byytelope/stache/api/stache/v1/stachev1connect"
	"github.com/byytelope/stache/pkg/stache"
)

type cacheServer struct {
	cache *stache.Cache
	stachev1connect.UnimplementedCacheServiceHandler
}

func main() {
	c := stache.NewCache()
	service := &cacheServer{cache: c}
	path, handler := stachev1connect.NewCacheServiceHandler(service, connect.WithInterceptors())

	checker := grpchealth.NewStaticChecker("stache.v1.CacheService")
	reflector := grpcreflect.NewStaticReflector("stache.v1.CacheService")

	mux := http.NewServeMux()
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	mux.Handle(grpchealth.NewHandler(checker))
	mux.Handle(path, handler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h2s := &http2.Server{}
	server := &http.Server{
		Addr:         ":8080",
		Handler:      h2c.NewHandler(mux, h2s),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Println("stached (Connect) listening on", server.Addr)
		if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println("serve error:", err)
		}
	}()

	waitForShutdown(server, time.Second*5)
}
