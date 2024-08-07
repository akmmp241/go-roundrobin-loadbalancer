package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Server interface {
	Address() string
	IsAlive() bool
	Serve(w http.ResponseWriter, r *http.Request)
}

type SimpleServer struct {
	Addr  string
	proxy *httputil.ReverseProxy
}

func NewSimpleServer(addr string) *SimpleServer {
	serverUrl, err := url.Parse(addr)
	HandleError(err)

	return &SimpleServer{
		Addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func (s *SimpleServer) Address() string {
	return s.Addr
}

func (s *SimpleServer) IsAlive() bool {
	return true
}

func (s *SimpleServer) Serve(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

type LoadBalancer struct {
	Port            string
	RoundRobinCount int
	Servers         []Server
}

func NewLoadBalancer(servers []Server, port string) *LoadBalancer {
	return &LoadBalancer{
		Servers:         servers,
		Port:            port,
		RoundRobinCount: 0,
	}
}

func (l *LoadBalancer) GetNextAvailableServer() Server {
	server := l.Servers[l.RoundRobinCount%len(l.Servers)]

	for !server.IsAlive() {
		l.RoundRobinCount++
		server = l.Servers[l.RoundRobinCount%len(l.Servers)]
	}

	l.RoundRobinCount++
	log.Printf("round robin server %d", l.RoundRobinCount)
	return server
}

func (l *LoadBalancer) ServeProxy(w http.ResponseWriter, r *http.Request) {
	target := l.GetNextAvailableServer()
	log.Printf("forwarding to %s", target.Address())
	target.Serve(w, r)
}

func main() {
	servers := []Server{
		NewSimpleServer("http://localhost:4001"),
		NewSimpleServer("http://localhost:4002"),
	}

	lb := NewLoadBalancer(servers, "4000")
	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		lb.ServeProxy(w, r)
	}

	router := http.NewServeMux()

	router.HandleFunc("/", handleRedirect)

	server := http.Server{
		Addr:    ":" + lb.Port,
		Handler: router,
	}

	log.Printf("serving requests at 'localhost:%v'\n", lb.Port)
	log.Fatal(server.ListenAndServe())
}

func HandleError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		panic(err)
	}
}
