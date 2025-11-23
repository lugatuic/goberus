package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/lugatuic/goberus/handlers"
	"github.com/lugatuic/goberus/ldaps"
)

type UserClient interface {
	GetMemberInfo(ctx context.Context, username string) (*ldaps.MemberInfo, error)
	AddUser(ctx context.Context, u *ldaps.UserInfo) error
}

// HandleGetMember serves GET /v1/member.
func HandleGetMember(client UserClient, w http.ResponseWriter, r *http.Request) error {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "missing username parameter", http.StatusBadRequest)
		return nil
	}

	ctxTimeout, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	info, err := client.GetMemberInfo(ctxTimeout, username)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		return err
	}
	return nil
}

// HandleCreateMember serves POST /v1/member.
func HandleCreateMember(client UserClient, w http.ResponseWriter, r *http.Request) error {
	defer func() {
		if err := r.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close request body: %v\n", err)
		}
	}()
	var u ldaps.UserInfo
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return nil
	}
	if err := handlers.SanitizeUser(&u); err != nil {
		http.Error(w, "invalid input: "+err.Error(), http.StatusBadRequest)
		return nil
	}

	ctxTimeout, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if err := client.AddUser(ctxTimeout, &u); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "created"}); err != nil {
		return err
	}
	return nil
}
