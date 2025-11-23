package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/lugatuic/goberus/config"
	"github.com/lugatuic/goberus/ldaps"
	"go.uber.org/zap"
)

func main() {
	// initialize structured logger early so we can log config errors
	logger, lerr := zap.NewProduction()
	if lerr != nil {
		panic("failed to initialize logger: " + lerr.Error())
	}
	defer logger.Sync()

	var cfg *config.Config
	var err error
	cfg, err = config.LoadFromEnv()
	if err != nil {
		logger.Fatal("config load failed", zap.Error(err))
	}

	var client *ldaps.Client
	client, err = ldaps.NewClient(cfg, logger)
	if err != nil {
		logger.Fatal("ldaps client init failed", zap.Error(err))
	}

	http.HandleFunc("/v1/member", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
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
				logger.Error("encode response failed", zap.Error(err))
			}

		case http.MethodPost:
			// Create a new user. Expect JSON body matching ldaps.UserInfo
			defer r.Body.Close()
			var u ldaps.UserInfo
			if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
				logger.Warn("invalid json in add user request", zap.Error(err))
				http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
				return
			}
			if u.Username == "" || u.Password == "" {
				logger.Warn("missing username or password in add user request")
				http.Error(w, "missing username or password", http.StatusBadRequest)
				return
			}

			var ctxTimeout context.Context
			var cancel context.CancelFunc
			ctxTimeout, cancel = context.WithTimeout(r.Context(), 15*time.Second)
			defer cancel()

			if err := client.AddUser(ctxTimeout, &u); err != nil {
				logger.Error("add user failed", zap.Error(err), zap.String("username", u.Username))
				http.Error(w, "add user error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "created"})

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	logger.Info("listening", zap.String("addr", cfg.BindAddr))
	if err = http.ListenAndServe(cfg.BindAddr, nil); err != nil {
		logger.Fatal("http server failed", zap.Error(err))
	}
}
