package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sync"
	"time"
)

var (
	supersetURL = getenv("SUPERSET_URL", "http://localhost:8088")
	adminUser   = getenv("SUPERSET_ADMIN_USER", "admin")
	adminPass   = getenv("SUPERSET_ADMIN_PASS", "admin")
	dashboardID = getenv("GUEST_DASHBOARD_ID", "")
	port        = getenv("PORT", "3000")

	mu          sync.Mutex
	accessToken string
	csrfToken   string
	tokenExpiry time.Time
	httpClient  *http.Client
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func init() {
	jar, _ := cookiejar.New(nil)
	httpClient = &http.Client{Timeout: 15 * time.Second, Jar: jar}
}

// login fetches access + CSRF tokens from Superset and caches them.
func login() error {
	mu.Lock()
	defer mu.Unlock()

	if time.Now().Before(tokenExpiry) {
		return nil
	}

	// Step 1: get access token
	body, _ := json.Marshal(map[string]string{
		"username": adminUser,
		"password": adminPass,
		"provider": "db",
		"refresh":  "true",
	})
	resp, err := httpClient.Post(supersetURL+"/api/v1/security/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("superset login: %w", err)
	}
	defer resp.Body.Close()

	var loginResult struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResult); err != nil || loginResult.AccessToken == "" {
		return fmt.Errorf("superset login failed (status %d)", resp.StatusCode)
	}
	accessToken = loginResult.AccessToken

	// Step 2: get CSRF token (required for POST endpoints)
	req, _ := http.NewRequest(http.MethodGet, supersetURL+"/api/v1/security/csrf_token/", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Referer", supersetURL)
	csrfResp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("csrf token request: %w", err)
	}
	defer csrfResp.Body.Close()

	var csrfResult struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(csrfResp.Body).Decode(&csrfResult); err != nil || csrfResult.Result == "" {
		return fmt.Errorf("csrf token fetch failed (status %d)", csrfResp.StatusCode)
	}
	csrfToken = csrfResult.Result
	tokenExpiry = time.Now().Add(4 * time.Minute)

	log.Println("Superset login OK, CSRF token acquired")
	return nil
}

func guestTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := login(); err != nil {
		log.Printf("login error: %v", err)
		http.Error(w, "failed to authenticate with Superset", http.StatusBadGateway)
		return
	}

	id := dashboardID
	if id == "" {
		id = r.URL.Query().Get("dashboard_id")
	}
	if id == "" {
		http.Error(w, "GUEST_DASHBOARD_ID not configured", http.StatusInternalServerError)
		return
	}

	payload, _ := json.Marshal(map[string]any{
		"user": map[string]string{
			"username":   "guest",
			"first_name": "Guest",
			"last_name":  "User",
		},
		"resources": []map[string]string{
			{"type": "dashboard", "id": id},
		},
		"rls": []any{},
	})

	req, _ := http.NewRequest(http.MethodPost, supersetURL+"/api/v1/security/guest_token/", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-CSRFToken", csrfToken)
	req.Header.Set("Referer", supersetURL)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("guest token error: %v", err)
		http.Error(w, "guest token request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("guest token response status=%d body=%s", resp.StatusCode, string(body))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

var allowedOrigins = func() map[string]bool {
	origins := map[string]bool{
		"http://localhost:5173": true,
		"http://localhost:5174": true,
	}
	if o := os.Getenv("ALLOWED_ORIGIN"); o != "" {
		origins[o] = true
	}
	return origins
}()

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/guest-token", guestTokenHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	addr := ":" + port
	log.Printf("Go backend listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, corsMiddleware(mux)))
}
