package main

// Code inspired by: https://medium.com/@mlowicki/http-s-proxy-in-golang-in-less-than-100-lines-of-code-6a51c2f2c38c

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"time"
	"strings"
	"sync/atomic"
)
var failureInjected uint64 = 0

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)

	log.Printf("Opening tunnel with %s\n", r.Host)

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	if injectFailure(w, r) {
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}
func injectFailure(w http.ResponseWriter, r *http.Request)  bool {
	atomic.AddUint64(&failureInjected, 1)
	failures := atomic.LoadUint64(&failureInjected)
	if failures <= 10 {
		if strings.Contains(r.Host, "www.facebook.com") {
			log.Println("Injecting latency")
			time.Sleep(5 * time.Second)
		}
		http.Error(w, "Simulated error", http.StatusServiceUnavailable)
		return true
	}
	return false
}
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}
func handleHTTP(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
func main() {
	var pemPath string
	flag.StringVar(&pemPath, "pem", "server.crt", "path to pem file")
	var keyPath string
	flag.StringVar(&keyPath, "key", "server.key", "path to key file")
	var proto string
	flag.StringVar(&proto, "proto", "https", "Proxy protocol (http or https)")
	flag.Parse()
	if proto != "http" && proto != "https" {
		log.Fatal("Protocol must be either http or https")
	}

	log.Println("Warning up the engine")

	server := &http.Server{
		Addr: ":8888",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				log.Println("Handling Tunneling")
				handleTunneling(w, r)
			} else {
				log.Println("Handling Http")
				handleHTTP(w, r)
			}
		}),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	if proto == "http" {
		log.Println("Starting http server")
		log.Fatal(server.ListenAndServe())
	} else {
		// for this to work the client needs to trust the certificate and to support a proxy over https.
		// For example java clients always speak to the proxy over http.
		// Golang clients instead since go 1.10 can speak to a proxy over https:
		// see https://medium.com/@mlowicki/https-proxies-support-in-go-1-10-b956fb501d6b
		log.Println("Starting https server")
		log.Fatal(server.ListenAndServeTLS(pemPath, keyPath))
	}
}