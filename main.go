package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

const (
	// SafeSearch VIPs provided directly by Google, Bing, and YouTube
	GoogleSafeSearchIP  = "216.239.38.120" // forcesafesearch.google.com
	BingSafeSearchIP    = "204.79.197.220" // strict.bing.com
	YouTubeRestrictIP   = "216.239.38.119" // restrict.youtube.com
	UpstreamParentalDNS = "1.1.1.3:53"     // Cloudflare Family DNS
	AllowlistFile       = "allowlist.txt"
)

type WalledGardenEngine struct {
	mu        sync.RWMutex
	allowlist map[string]struct{}
}

func NewWalledGardenEngine() *WalledGardenEngine {
	engine := &WalledGardenEngine{
		allowlist: make(map[string]struct{}),
	}

	// Default essential base allowlist (Safe search, CDNs, Education)
	defaultTrusted := []string{
		// Google Essentials (Search & Assets only)
		"google.com",
		"gstatic.com",
		"googleusercontent.com",
		"googleapis.com",

		// Educational Platforms
		"wikipedia.org",
		"wikimedia.org",
		"khanacademy.org",
		"kastatic.org",
		"pbskids.org",
		"scratch.mit.edu",
		"code.org",
		"duolingo.com",
		"nationalgeographic.com",
		"nasa.gov",

		// Learning Management & Schooling
		"classroom.google.com",
		"canvaslms.com",
		"blackboard.com",
	}

	for _, domain := range defaultTrusted {
		engine.allowlist[strings.ToLower(domain)] = struct{}{}
	}
	return engine
}

// Load custom allowlist from disk if available
func (wge *WalledGardenEngine) LoadAllowlistFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("[INFO] %s not found. Using default built-in allowlist.", path)
		return
	}
	defer file.Close()

	wge.mu.Lock()
	defer wge.mu.Unlock()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		domain := strings.ToLower(line)
		wge.allowlist[domain] = struct{}{}
		count++
	}

	// Fixes the linter warning: check if scanning stopped due to an error
	if err := scanner.Err(); err != nil {
		log.Printf("[ERROR] Error reading %s: %v", path, err)
		return
	}

	log.Printf("[ALLOWLIST] Loaded %d additional domains from %s", count, path)
}

func (wge *WalledGardenEngine) IsAllowed(domain string) bool {
	wge.mu.RLock()
	defer wge.mu.RUnlock()

	domain = strings.ToLower(strings.TrimSuffix(domain, "."))
	parts := strings.Split(domain, ".")

	// Walk up root domains (e.g. static.khanacademy.org -> khanacademy.org)
	for i := 0; i < len(parts)-1; i++ {
		sub := strings.Join(parts[i:], ".")
		if _, ok := wge.allowlist[sub]; ok {
			return true
		}
	}
	return false
}

func startDNSResolver(engine *WalledGardenEngine) {
	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		if len(r.Question) == 0 {
			w.WriteMsg(msg)
			return
		}

		q := r.Question[0]
		domain := strings.ToLower(strings.TrimSuffix(q.Name, "."))

		// 1. Strict Walled-Garden Drop: If not explicitly allowed, reject.
		if !engine.IsAllowed(domain) {
			log.Printf("[DROPPED - NOT ALLOWED] %s", domain)
			msg.Rcode = dns.RcodeNameError // NXDOMAIN
			w.WriteMsg(msg)
			return
		}

		// 2. SafeSearch Enforcement on permitted domains
		if q.Qtype == dns.TypeA {
			if strings.Contains(domain, "google.") {
				rr, _ := dns.NewRR(fmt.Sprintf("%s 300 IN A %s", q.Name, GoogleSafeSearchIP))
				msg.Answer = append(msg.Answer, rr)
				w.WriteMsg(msg)
				log.Printf("[SAFESEARCH APPLIED] %s", domain)
				return
			}
			if strings.Contains(domain, "bing.com") {
				rr, _ := dns.NewRR(fmt.Sprintf("%s 300 IN A %s", q.Name, BingSafeSearchIP))
				msg.Answer = append(msg.Answer, rr)
				w.WriteMsg(msg)
				log.Printf("[SAFESEARCH APPLIED] %s", domain)
				return
			}
			if strings.Contains(domain, "youtube.com") {
				rr, _ := dns.NewRR(fmt.Sprintf("%s 300 IN A %s", q.Name, YouTubeRestrictIP))
				msg.Answer = append(msg.Answer, rr)
				w.WriteMsg(msg)
				log.Printf("[YOUTUBE RESTRICTED APPLIED] %s", domain)
				return
			}
		}

		// 3. Forward validated queries to upstream DNS
		c := new(dns.Client)
		c.Timeout = 2 * time.Second
		resp, _, err := c.Exchange(r, UpstreamParentalDNS)
		if err != nil {
			dns.HandleFailed(w, r)
			return
		}
		w.WriteMsg(resp)
	})

	server := &dns.Server{Addr: "127.0.0.1:53", Net: "udp"}
	log.Println("[RUNNING] Walled Garden Active on 127.0.0.1:53")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("[FATAL] Port 53 bind error: %v. (Must run as Administrator/Root)", err)
	}
}

func main() {
	engine := NewWalledGardenEngine()
	engine.LoadAllowlistFile(AllowlistFile)
	startDNSResolver(engine)
}
