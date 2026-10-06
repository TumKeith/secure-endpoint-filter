package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	ProxyPort     = ":8080"
	AllowlistFile = "allowlist.txt"
)

type ProxyEngine struct {
	mu        sync.RWMutex
	allowlist map[string]struct{}
}

func NewProxyEngine() *ProxyEngine {
	pe := &ProxyEngine{
		allowlist: make(map[string]struct{}),
	}

	trusted := []string{
		"google.com", "gstatic.com", "googleusercontent.com", "googleapis.com",
		"wikipedia.org", "wikimedia.org", "khanacademy.org", "kastatic.org",
		"pbskids.org", "scratch.mit.edu", "code.org", "duolingo.com",
		"nationalgeographic.com", "nasa.gov", "classroom.google.com",
		"canvaslms.com", "blackboard.com",
	}

	for _, domain := range trusted {
		pe.allowlist[strings.ToLower(domain)] = struct{}{}
	}
	return pe
}

func (pe *ProxyEngine) LoadAllowlist(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("[INFO] %s not found, running with built-in allowlist", path)
		return
	}
	defer file.Close()

	pe.mu.Lock()
	defer pe.mu.Unlock()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pe.allowlist[strings.ToLower(line)] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[ERROR] Reading %s: %v", path, err)
	}
}

func (pe *ProxyEngine) IsAllowed(host string) bool {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	// Strip port if present (e.g., "wikipedia.org:443")
	h, _, err := net.SplitHostPort(host)
	if err == nil {
		host = h
	}
	host = strings.ToLower(strings.TrimSuffix(host, "."))

	parts := strings.Split(host, ".")
	for i := 0; i < len(parts)-1; i++ {
		sub := strings.Join(parts[i:], ".")
		if _, ok := pe.allowlist[sub]; ok {
			return true
		}
	}
	return false
}

func handleHTTP(w http.ResponseWriter, req *http.Request, engine *ProxyEngine) {
	if !engine.IsAllowed(req.Host) {
		http.Error(w, "Access Denied by Endpoint Policy", http.StatusForbidden)
		return
	}

	// HTTPS Tunneling (CONNECT method used by Chrome for all secure sites)
	if req.Method == http.MethodConnect {
		destConn, err := net.DialTimeout("tcp", req.Host, 5*time.Second)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			destConn.Close()
			return
		}

		clientConn, _, err := hijacker.Hijack()
		if err != nil {
			destConn.Close()
			return
		}

		go transfer(destConn, clientConn)
		go transfer(clientConn, destConn)
		return
	}

	// Standard plain HTTP
	req.RequestURI = ""
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func main() {
	engine := NewProxyEngine()
	engine.LoadAllowlist(AllowlistFile)

	server := &http.Server{
		Addr: "127.0.0.1:8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleHTTP(w, r, engine)
		}),
	}

	log.Println("[RUNNING] Endpoint Proxy Filter active on 127.0.0.1:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
