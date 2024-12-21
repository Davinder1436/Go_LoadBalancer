package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr) // Fixed: Correct capitalization
	handleErr(err)
	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		roundRobinCount: 0,
		port:            port,
		servers:         servers,
	}
}

func (s *simpleServer) Address() string { return s.addr }

func (s *simpleServer) IsAlive() bool { return true }

func (s *simpleServer) Serve(rw http.ResponseWriter, r *http.Request) { // Fixed: Removed pointer
	s.proxy.ServeHTTP(rw, r)
}

func (lb *LoadBalancer) getNextServer() Server {
	var server Server
	for {
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
		lb.roundRobinCount++
		if server.IsAlive() {
			break
		}
	}
	return server
}

func (lb *LoadBalancer) serveProxy(rw http.ResponseWriter, r *http.Request) { // Fixed: Removed pointer
	server := lb.getNextServer() // Ensure only one call
	fmt.Println("Forwarding request to:", server.Address())
	server.Serve(rw, r)
}

func main() {
	servers := []Server{
		newSimpleServer("http://www.yahoo.com"),
		newSimpleServer("http://duckduckgo.com"),
		newSimpleServer("http://www.google.com"),
	}

	lb := NewLoadBalancer("8080", servers)

	handleRedirect := func(rw http.ResponseWriter, r *http.Request) { // Fixed: Correct handler signature
		lb.serveProxy(rw, r)
	}

	http.HandleFunc("/", handleRedirect)

	fmt.Println("Listening on port 8080")
	err := http.ListenAndServe(":"+lb.port, nil)
	handleErr(err)
}
