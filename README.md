# Go Load Balancer (Learning Project)

This is a simple round-robin HTTP load balancer written in Go. The main objective is to learn about Go's concurrency, HTTP handling, and basic load balancing concepts.

## Features
- Forwards HTTP requests to a list of backend servers in round-robin order
- Handles concurrent requests safely using atomic operations
- Forwards request headers and response data
- Returns appropriate error responses if a backend is unavailable
- Periodically checks backend health with a `/ping` endpoint and only marks healthy servers as available

## Usage
1. Clone the repository
2. Run the tests:
   ```sh
   go test -v
   ```
3. Run the benchmarks:
   ```sh
   go test -bench=.
   ```
4. Explore the code in `balancer.go`, `balancer_test.go`, and `main.go`

## Note
This project is for educational purposes and is not production-ready.

## Next Steps
- Run the load balancer as an HTTP server and test with real backend services
- Add more advanced health checks and reporting
- Support for HTTPS backends
- Add logging and metrics
- Implement weighted round-robin or other balancing strategies
- Write more unit and integration tests
- Experiment with graceful shutdown and error handling
