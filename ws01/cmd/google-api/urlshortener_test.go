package googleapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/oauth2"
)

func TestThreeLeggedOAuthFlow(t *testing.T) {
	// ---- Fake OAuth Server ----
	mux := http.NewServeMux()

	// /authorize endpoint
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		// 基本檢查
		if q.Get("response_type") != "code" {
			t.Fatalf("unexpected response_type: %s", q.Get("response_type"))
		}

		redirectURI := q.Get("redirect_uri")
		if redirectURI == "" {
			t.Fatal("missing redirect_uri")
		}

		// 模擬 OAuth server 發 code
		u, _ := url.Parse(redirectURI)
		values := u.Query()
		values.Set("code", "test-auth-code")
		u.RawQuery = values.Encode()

		http.Redirect(w, r, u.String(), http.StatusFound)
	})

	// /token endpoint
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()

		if r.Form.Get("grant_type") != "authorization_code" {
			t.Fatalf("unexpected grant_type: %s", r.Form.Get("grant_type"))
		}

		if r.Form.Get("code") != "test-auth-code" {
			t.Fatalf("unexpected code: %s", r.Form.Get("code"))
		}

		resp := map[string]interface{}{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	oauthServer := httptest.NewServer(mux)
	defer oauthServer.Close()

	// ---- OAuth Client Config ----
	conf := &oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Scopes:       []string{"read", "write"},
		RedirectURL:  "http://client.example.com/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthServer.URL + "/authorize",
			TokenURL: oauthServer.URL + "/token",
		},
	}

	ctx := context.Background()

	// ---- Step 1: 產生授權 URL ----
	authURL := conf.AuthCodeURL("state-xyz", oauth2.AccessTypeOffline)
	if authURL == "" {
		t.Fatal("empty auth url")
	}

	// ---- Step 2: 模擬 user 瀏覽 authURL（跟隨 redirect 拿 code）----
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(authURL)
	if err != nil {
		t.Fatalf("authorize request failed: %v", err)
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if location == "" {
		t.Fatal("missing redirect location")
	}

	u, _ := url.Parse(location)
	code := u.Query().Get("code")
	if code != "test-auth-code" {
		t.Fatalf("unexpected auth code: %s", code)
	}

	// ---- Step 3: 用 code 換 access token ----
	token, err := conf.Exchange(ctx, code)
	if err != nil {
		t.Fatalf("token exchange failed: %v", err)
	}

	if token.AccessToken != "test-access-token" {
		t.Fatalf("unexpected access token: %s", token.AccessToken)
	}
}
