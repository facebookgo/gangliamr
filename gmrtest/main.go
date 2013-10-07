// Package gmrtest provides a sample application to experiment with gangliamr.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/4eek/gofaker/lorem"
	"github.com/daaku/go.ganglia/gmetric"
	"github.com/daaku/go.gangliamr"
	"github.com/daaku/go.metrics"
)

type Server struct {
	MaxSentences       int
	MaxSleep           time.Duration
	ConcurrentRequests metrics.Counter
	NumRequests        metrics.Meter
	ResponseTime       metrics.Timer
	PageSize           metrics.Histogram
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer s.ResponseTime.Start().Stop()
	defer s.ConcurrentRequests.Dec(1)
	s.ConcurrentRequests.Inc(1)
	s.NumRequests.Mark(1)
	time.Sleep(time.Duration(rand.Int63n(int64(s.MaxSleep))))
	bd := lorem.Sentences(rand.Intn(s.MaxSentences))
	s.PageSize.Update(int64(len(bd)))
	fmt.Fprint(w, bd, "\n")
}

func main() {
	server := &Server{
		ConcurrentRequests: &gangliamr.Counter{
			Name:   "concurrent_requests",
			Title:  "Number of concurrent requests",
			Units:  "requests",
			Groups: []string{"gmrtest"},
		},
		NumRequests: &gangliamr.Meter{
			Name:   "num_requests",
			Title:  "Number of requests",
			Units:  "requests",
			Groups: []string{"gmrtest"},
		},
		ResponseTime: &gangliamr.Timer{
			Name:       "num_requests",
			Resolution: time.Millisecond,
			Title:      "Response time",
			Groups:     []string{"gmrtest"},
		},
		PageSize: &gangliamr.Histogram{
			Name:   "page_size",
			Title:  "Page size",
			Units:  "bytes",
			Groups: []string{"gmrtest"},
		},
	}

	client := &gmetric.Client{
		Addr: []net.Addr{
			&net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8649},
		},
	}

	registry := &gangliamr.Registry{
		Prefix:            "gmrtest",
		WriteTickDuration: 20 * time.Second,
		Client:            client,
	}

	addr := flag.String("addr", "0.0.0.0:8077", "server address")
	gomaxprocs := flag.Int("gomaxprocs", runtime.NumCPU(), "gomaxprocs")
	flag.StringVar(&client.Host, "host", "localhost", "ganglia host")
	flag.DurationVar(&server.MaxSleep, "max-sleep", time.Second*5, "max sleep")
	flag.IntVar(&server.MaxSentences, "max-sentences", 500, "max sentences")

	flag.Parse()
	runtime.GOMAXPROCS(*gomaxprocs)

	if err := client.Open(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	registry.Register(server.ConcurrentRequests)
	registry.Register(server.NumRequests)
	registry.Register(server.ResponseTime)
	registry.Register(server.PageSize)

	fmt.Printf("Serving on http://%s/\n", *addr)

	if err := http.ListenAndServe(*addr, server); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := client.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
