package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/byytelope/stache/api/stache/v1/stachev1connect"
)

func main() {
	addr := flag.String("addr", "http://localhost:8080", "Daemon base URL")
	doList := flag.Bool("list", false, "List all items")
	setKey := flag.String("set", "", "Set value for key (requires -v)")
	getKey := flag.String("get", "", "Get value for key")
	val := flag.String("v", "", "Value to set (used with -set)")
	ct := flag.String("t", "text/plain", "MIME content type (used with -set)")
	ttlSec := flag.Int("l", 0, "TTL in seconds (0 = no expiry) (used with -set)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  stache -set <key> -v <value> [-t <content-type>] [-l <ttl-seconds>] [-addr <url>]\n")
		fmt.Fprintf(os.Stderr, "  stache -get <key> [-addr <url>]\n")
		fmt.Fprintf(os.Stderr, "  stache -list [-addr <url>]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	nActions := 0
	if *doList {
		nActions++
	}
	if *setKey != "" {
		nActions++
	}
	if *getKey != "" {
		nActions++
	}

	if nActions != 1 {
		flag.Usage()
		os.Exit(2)
	}

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	h := Handler{
		client: stachev1connect.NewCacheServiceClient(httpClient, *addr),
		out:    os.Stdout,
		err:    os.Stderr,
	}

	switch {
	case *doList:
		if err := h.List(); err != nil {
			os.Exit(1)
		}

	case *setKey != "":
		if *val == "" {
			fmt.Fprintln(os.Stderr, "error: -set requires -v <value>")
			flag.Usage()
			os.Exit(2)
		}
		if err := h.Set(*setKey, *val, *ct, int64(*ttlSec)); err != nil {
			os.Exit(1)
		}

	case *getKey != "":
		if err := h.Get(*getKey); err != nil {
			os.Exit(1)
		}
	}
}
