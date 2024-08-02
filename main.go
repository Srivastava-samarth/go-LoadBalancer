package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface{
	Address() string
	isAlive()  bool
	Serve(w http.ResponseWriter,r *http.Request)
}

type simpleServer struct{
	address string
	proxy *httputil.ReverseProxy
}

type loadBalancer struct{
	port            string
	roundRobinCount int
	servers         []Server
}

func newLoadBalancer(port string,servers []Server) *loadBalancer{
	return &loadBalancer{
		port:port,
		roundRobinCount: 0,
		servers: servers,
	}
}

func newSimpleServer(address string) *simpleServer{
	serverUrl,err := url.Parse(address)
	handleErr(err);

	return &simpleServer{
		address: address,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func handleErr(err error){
	if err!=nil{
		fmt.Printf("error : %v\n",err);
		os.Exit(1);
	}
}

func (s *simpleServer) Address() string{
	return s.address;
}

func (s *simpleServer) isAlive() bool{return true;}

func (s *simpleServer) Serve(w http.ResponseWriter,r *http.Request){
	s.proxy.ServeHTTP(w,r);
}

func (lb *loadBalancer) getNextAvailableServer() Server{
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.isAlive(){
		lb.roundRobinCount++;
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++;
	return server;
}

func (lb *loadBalancer) getServerProxy(w http.ResponseWriter,r *http.Request){
	targetServer := lb.getNextAvailableServer();
	fmt.Printf("Forwarding request to address %q\n",targetServer.Address())
	targetServer.Serve(w,r);
}

func main(){
	servers := []Server{
		 newSimpleServer("http://www.facebook.com"),
		 newSimpleServer("http://www.youtube.com"),
		 newSimpleServer("http://www.leetcode.com"),
	}
	lb := newLoadBalancer("8000",servers);
	handleRedirect := func(w http.ResponseWriter,r *http.Request){
		lb.getServerProxy(w,r);
	}
	http.HandleFunc("/",handleRedirect);

	fmt.Printf("Server is running at port %s\n",lb.port)
	http.ListenAndServe(":" + lb.port,nil);
}