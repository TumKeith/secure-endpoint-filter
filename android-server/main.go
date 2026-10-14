package main

import (
	"bufio"
	"crypto/tls"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

const (
	CertFile            = "/certs/fullchain.pem"
	KeyFile             = "/certs/privkey.pem"
	UpstreamParentalDNS = "1.1.1.3:53"
	AllowlistFile       = "allowlist.txt"
)

type GuardEngine struct {
	mu        sync.RWMutex
	allowlist map[string]struct{}
}

func NewGuardEngine() *GuardEngine {
	ge := &GuardEngine{allowlist: make(map[string]struct{})}
	trusted := []string{
		"google.com", "gstatic.com", "googleusercontent.com", "googleapis.com",
		"wikipedia.org", "wikimedia.org", "khanacademy.org", "kastatic.org",
		"pbskids.org", "scratch.mit.edu", "code.org", "duolingo.com",
		"nationalgeographic.com", "nasa.gov",
	}
	for _, domain := range trusted {
		ge.allowlist[strings.ToLower(domain)] = struct{}{}
	}
	return ge
}

func (ge *GuardEngine) LoadAllowlist(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("[INFO] %s not found, using default allowlist", path)
		return
	}
	defer file.Close()

	ge.mu.Lock()
	defer ge.mu.Unlock()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ge.allowlist[strings.ToLower(line)] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("[ERROR] Reading %s: %v", path, err)
	}
}

func (ge *GuardEngine) IsAllowed(domain string) bool {
	ge.mu.RLock()
	defer ge.mu.RUnlock()

	domain = strings.ToLower(strings.TrimSuffix(domain, "."))
	parts := strings.Split(domain, ".")
	for i := 0; i < len(parts)-1; i++ {
		sub := strings.Join(parts[i:], ".")
		if _, ok := ge.allowlist[sub]; ok {
			return true
		}
	}
	return false
}

func handleDoT(engine *GuardEngine) dns.HandlerFunc {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		if len(r.Question) == 0 {
			w.WriteMsg(msg)
			return
		}

		q := r.Question[0]
		domain := strings.ToLower(strings.TrimSuffix(q.Name, "."))

		if !engine.IsAllowed(domain) {
			msg.Rcode = dns.RcodeNameError
			w.WriteMsg(msg)
			return
		}

		c := new(dns.Client)
		c.Timeout = 2 * time.Second
		resp, _, err := c.Exchange(r, UpstreamParentalDNS)
		if err != nil {
			dns.HandleFailed(w, r)
			return
		}
		w.WriteMsg(resp)
	}
}

func main() {
	engine := NewGuardEngine()
	engine.LoadAllowlist(AllowlistFile)

	cert, err := tls.LoadX509KeyPair(CertFile, KeyFile)
	if err != nil {
		log.Printf("[WARN] Running in development mode without TLS certs: %v", err)
	}

	var tlsConfig *tls.Config
	if err == nil {
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
	}

	server := &dns.Server{
		Addr:      ":853",
		Net:       "tcp-tls",
		TLSConfig: tlsConfig,
		Handler:   dns.HandlerFunc(handleDoT(engine)),
	}

	log.Println("[RUNNING] Android DNS-over-TLS (DoT) Engine listening on :853")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server listener terminated: %v", err)
	}
}
