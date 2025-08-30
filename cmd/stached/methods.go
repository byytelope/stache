package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"

	stachev1 "github.com/byytelope/stache/api/stache/v1"
	"github.com/byytelope/stache/pkg/stache"
)

func (s *cacheServer) Set(
	ctx context.Context,
	req *connect.Request[stachev1.SetRequest],
) (*connect.Response[stachev1.SetResponse], error) {
	r := req.Msg

	if r.GetKey() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("key required"))
	}

	ttl := time.Duration(r.GetTtl()) * time.Second
	ct := stache.ContentType(*r.ContentType)
	if ct == "" {
		ct = stache.Text
	}

	if err := s.cache.Set(r.GetKey(), r.GetValue(), stache.Meta{TTL: ttl, ContentType: ct}); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&stachev1.SetResponse{}), nil
}

func (s *cacheServer) Get(
	ctx context.Context,
	req *connect.Request[stachev1.GetRequest],
) (*connect.Response[stachev1.GetResponse], error) {
	key := req.Msg.GetKey()
	if key == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("key required"))
	}
	b, err := s.cache.GetBytes(key)
	if err != nil {
		if errors.Is(err, stache.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	entry, err := s.cache.GetEntry(key)
	if err != nil && !errors.Is(err, stache.ErrNotFound) {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	var expMs int64
	if !entry.ExpiresAt.IsZero() {
		expMs = entry.ExpiresAt.UnixMilli()
	}
	ct := string(entry.ContentType)
	res := &stachev1.GetResponse{
		Value:       b,
		ContentType: &ct,
		ExpiresAtMs: &expMs,
	}

	return connect.NewResponse(res), nil
}

func (s *cacheServer) Delete(ctx context.Context, req *connect.Request[stachev1.DeleteRequest]) (*connect.Response[stachev1.DeleteResponse], error) {
	key := req.Msg.GetKey()
	if key == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("key is required"))
	}

	_, ok := s.cache.Delete(key)

	return connect.NewResponse(&stachev1.DeleteResponse{Deleted: &ok}), nil
}

func (s *cacheServer) ListEntries(ctx context.Context, _ *connect.Request[stachev1.ListEntriesRequest]) (*connect.Response[stachev1.ListEntriesResponse], error) {
	ents := s.cache.Entries()
	out := make([]*stachev1.EntryInfo, 0, len(ents))
	for _, e := range ents {
		var expMs int64
		if !e.ExpiresAt.IsZero() {
			expMs = e.ExpiresAt.UnixMilli()
		}
		size := uint32(e.Size)
		ct := string(e.ContentType)
		out = append(out, &stachev1.EntryInfo{
			Key:         &e.Key,
			Size:        &size,
			ContentType: &ct,
			ExpiresAtMs: &expMs,
		})
	}

	return connect.NewResponse(&stachev1.ListEntriesResponse{Entries: out}), nil
}

func waitForShutdown(s *http.Server, timeout time.Duration) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_ = s.Shutdown(ctx)
}
