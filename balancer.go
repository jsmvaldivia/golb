package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"sync/atomic"
)

type LoadBalancer struct {
	servers []*url.URL
	current int64
}

func NewLoadBalancer(servers []string) *LoadBalancer {
	urls := make([]*url.URL, len(servers))
	for i, addr := range servers {
		url, err := url.Parse(addr)
		if err != nil {
			log.Fatalf("Invalid backend URL: %s", url)
		}

		urls[i] = url
	}

	return &LoadBalancer{servers: urls}
}

func (lb *LoadBalancer) NextServer() *url.URL {
	idx := atomic.AddInt64(&lb.current, 1)
	return lb.servers[int(idx)%len(lb.servers)]
}

func (lb *LoadBalancer) ForwardToNextServer(w http.ResponseWriter, r *http.Request) {
	srv := lb.NextServer()

	nr, err := http.NewRequest(r.Method, srv.String()+r.URL.Path, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("error creating new request: %v", err)
		return
	}

	nr.Header = r.Header.Clone()

	resp, err := http.DefaultClient.Do(nr)
	if err != nil {
		http.Error(w, "backend unavailable", http.StatusServiceUnavailable)
		return
	}

	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("error copying response body: %v", err)
	}
}
