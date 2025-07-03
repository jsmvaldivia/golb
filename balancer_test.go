package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestNewLoadBalancer(t *testing.T) {
	servers := []string{"http://localhost:8080", "http://localhost:8081"}
	lb := NewLoadBalancer(servers)

	if len(lb.servers) != len(servers) {
		t.Fatalf("expected %d servers, got %d", len(servers), len(lb.servers))
	}

	for i, s := range servers {
		u, _ := url.Parse(s)
		if !reflect.DeepEqual(lb.servers[i], u) {
			t.Errorf("expected server %d to be %v, got %v", i, u, lb.servers[i])
		}
	}
}

func TestForwardToNextServer(t *testing.T) {
	// Start two test backend servers
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend2"))
	}))
	defer backend2.Close()

	servers := []string{backend1.URL, backend2.URL}
	lb := NewLoadBalancer(servers)

	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	// First request should go to backend1
	lb.ForwardToNextServer(recorder, req)
	resp := recorder.Result()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "backend1" {
		t.Errorf("expected backend1, got %s", string(body))
	}

	// Second request should go to backend2
	recorder2 := httptest.NewRecorder()
	lb.ForwardToNextServer(recorder2, req)
	resp2 := recorder2.Result()
	body2, _ := io.ReadAll(resp2.Body)
	if string(body2) != "backend2" {
		t.Errorf("expected backend2, got %s", string(body2))
	}
}
