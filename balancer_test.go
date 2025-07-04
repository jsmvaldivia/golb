package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestNewLoadBalancer(t *testing.T) {
	servers := []string{"http://localhost:8080", "http://localhost:8081"}
	lb := NewLoadBalancer(servers, false)

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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend2"))
	}))
	defer backend2.Close()

	servers := []string{backend1.URL, backend2.URL}
	lb := NewLoadBalancer(servers, false)

	req := httptest.NewRequest("GET", "/", nil)

	recorder := httptest.NewRecorder()
	lb.ForwardToNextServer(recorder, req)
	resp := recorder.Result()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if string(body) != "backend1" {
		t.Errorf("expected backend1, got %s", string(body))
	}

	recorder2 := httptest.NewRecorder()
	lb.ForwardToNextServer(recorder2, req)
	resp2 := recorder2.Result()
	body2, _ := io.ReadAll(resp2.Body)
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp2.StatusCode)
	}
	if string(body2) != "backend2" {
		t.Errorf("expected backend2, got %s", string(body2))
	}
}

func BenchmarkForwardToNextServer(b *testing.B) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok2"))
	}))
	defer backend2.Close()

	backend3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok3"))
	}))
	defer backend3.Close()

	servers := []string{backend1.URL, backend2.URL, backend3.URL}
	lb := NewLoadBalancer(servers, false)

	req := httptest.NewRequest("GET", "/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		lb.ForwardToNextServer(recorder, req)
	}
}

func TestHealthCheck(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("pong"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer healthy.Close()

	unhealthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer unhealthy.Close()

	servers := []string{healthy.URL, unhealthy.URL}
	lb := NewLoadBalancer(servers, true)

	lb.HealthCheck(true, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond) // Give time for health check goroutine to run

	for _, u := range lb.servers {
		if u.String() == healthy.URL {
			if !lb.healthMap[u] {
				t.Errorf("expected healthy server to be marked healthy")
			}
		} else if u.String() == unhealthy.URL {
			if lb.healthMap[u] {
				t.Errorf("expected unhealthy server to be marked unhealthy")
			}
		}
	}
}
