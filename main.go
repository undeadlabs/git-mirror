package main

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	// Parse config.
	if len(os.Args) != 2 {
		log.Fatal("please specify the path to a config file, an example config is available at https://github.com/beefsack/git-mirror/blob/master/example-config.toml")
	}
	cfg, repos, err := parseConfig(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(cfg.BasePath, 0755); err != nil {
		log.Fatalf("failed to create %s, %s", cfg.BasePath, err)
	}

	// Run background threads to keep mirrors up to date.
	var wg sync.WaitGroup
	for _, r := range repos {
		wg.Add(1)
		go func(r repo) {
			for {
				log.Printf("updating %s", r.Name)
				if err := mirror(cfg, r); err != nil {
					log.Printf("error updating %s, %s", r.Name, err)
				} else {
					if r.Mirror != "" {
						log.Printf("updated %s (pushed to %s)", r.Name, r.Mirror)
					} else {
						log.Printf("updated %s", r.Name)
					}
				}
				time.Sleep(r.Interval.Duration)
			}
			wg.Done()
		}(r)
	}

	// Run HTTP server to serve mirrors.
	if !cfg.NoServe {
		http.Handle("/", http.FileServer(http.Dir(cfg.BasePath)))
		log.Printf("starting web server on %s", cfg.ListenAddr)
		if err := http.ListenAndServe(cfg.ListenAddr, nil); err != nil {
			log.Fatalf("failed to start server, %s", err)
		}
	}

	wg.Wait()
}
