package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {

	healthCheckEnv := os.Getenv("GOLB_HEALTHCHECKS_ENABLED")
	if healthCheckEnv == "" {
		healthCheckEnv = "false"
	}
	hcEnabled, err := strconv.ParseBool(healthCheckEnv)
	if err != nil {
		log.Fatalf("error parsing bool for healcheck env var %v", err)
	}

	// GOLB_TARGET_SERVERS=http://localhost:8081,http://localhost:8082
	serversEnv := os.Getenv("GOLB_TARGET_SERVERS")
	servers := strings.Split(serversEnv, ",")

	// start loadbalancer
	lb := NewLoadBalancer(servers, hcEnabled)
	// listen
	http.HandleFunc("/", lb.ForwardToNextServer)

	port := os.Getenv("GOLB_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on Port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
