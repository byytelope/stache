package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"connectrpc.com/connect"
	stachev1 "github.com/byytelope/stache/api/stache/v1"
	"github.com/byytelope/stache/api/stache/v1/stachev1connect"
)

type Handler struct {
	client stachev1connect.CacheServiceClient
	out    io.Writer
	err    io.Writer
}

func (h *Handler) Set(key string, value string, contentType string, ttlSeconds int64) error {
	req := &stachev1.SetRequest{
		Key:         &key,
		Value:       []byte(value),
		Ttl:         &ttlSeconds,
		ContentType: &contentType,
	}
	_, err := h.client.Set(context.Background(), connect.NewRequest(req))
	if err != nil {
		fmt.Fprintln(h.err, "Set error:", err)
		return err
	}

	fmt.Fprintf(h.out, "OK set key=%q ct=%q ttl=%ds\n", key, contentType, ttlSeconds)
	return nil
}

func (h *Handler) Get(key string) error {
	req := &stachev1.GetRequest{Key: &key}
	res, err := h.client.Get(context.Background(), connect.NewRequest(req))
	if err != nil {
		fmt.Fprintln(h.err, "Get error:", err)
		return err
	}

	switch res.Msg.GetContentType() {
	case "text/plain", "":
		fmt.Fprintf(h.out, "%s\n", string(res.Msg.GetValue()))
	case "application/json":
		var tmp any
		if err := json.Unmarshal(res.Msg.GetValue(), &tmp); err == nil {
			pretty, _ := json.MarshalIndent(tmp, "", "  ")
			fmt.Fprintf(h.out, "%s\n", pretty)
		} else {
			fmt.Fprintf(h.out, "%s\n", string(res.Msg.GetValue()))
		}
	default:
		fmt.Fprintf(h.out, "(%d bytes, %s)\n", len(res.Msg.GetValue()), res.Msg.GetContentType())
	}

	return nil
}

func (h *Handler) List() error {
	res, err := h.client.ListEntries(context.Background(), connect.NewRequest(&stachev1.ListEntriesRequest{}))
	if err != nil {
		fmt.Fprintln(h.err, "List error:", err)
		return err
	}

	ents := res.Msg.GetEntries()
	tw := tabwriter.NewWriter(h.out, 2, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "KEY\tSIZE\tCONTENT-TYPE\tEXPIRES")

	for _, e := range ents {
		exp := "-"
		if e.GetExpiresAtMs() > 0 {
			exp = time.UnixMilli(e.GetExpiresAtMs()).Format(time.RFC3339)
		}

		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\n", e.GetKey(), e.GetSize(), e.GetContentType(), exp)
	}

	tw.Flush()
	return nil
}
