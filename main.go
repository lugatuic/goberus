package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/lugatuic/goberus/config"
	"github.com/lugatuic/goberus/ldaps"
)

func main() {
	var cfg *config.Config
	var err error
	cfg, err = config.LoadFromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	var client *ldaps.Client
	client, err = ldaps.NewClient(cfg)
	if err != nil {
		log.Fatalf("ldaps client init: %v", err)
	}

	http.HandleFunc("/v1/member", func(w http.ResponseWriter, r *http.Request) {
		// minimal handler: GET /v1/member?username=foo or ?username=foo@domain
		var username string = r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "missing username parameter", http.StatusBadRequest)
			return
		}

		// use a background ctx with deadline
		var ctxTimeout context.Context
		var cancel context.CancelFunc
		ctxTimeout, cancel = context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		var info *ldaps.MemberInfo
		info, err = client.GetMemberInfo(ctxTimeout, username)
		if err != nil {
			http.Error(w, "search error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err = json.NewEncoder(w).Encode(info); err != nil {
			log.Printf("encode error: %v", err)
		}
	})

	log.Printf("listening on %s", cfg.BindAddr)
	if err = http.ListenAndServe(cfg.BindAddr, nil); err != nil {
		log.Fatalf("http server: %v", err)
	}
}
