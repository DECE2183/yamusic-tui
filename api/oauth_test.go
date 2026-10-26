package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func withOAuthServer(t *testing.T, handler http.Handler) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	old := oauthBaseURL
	oauthBaseURL = server.URL + "/"
	t.Cleanup(func() { oauthBaseURL = old })
}

func TestRequestDeviceCode(t *testing.T) {
	withOAuthServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/device/code" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		if got := r.Form.Get("client_id"); got != yaOauthClientID {
			t.Errorf("client_id = %q, want %q", got, yaOauthClientID)
		}
		if got := r.Form.Get("device_id"); got != "device1234" {
			t.Errorf("device_id = %q", got)
		}
		if got := r.Form.Get("device_name"); got != "yamusic-tui" {
			t.Errorf("device_name = %q", got)
		}
		if got := r.Form.Get("scope"); got != yaOauthMusicScope {
			t.Errorf("scope = %q, want %q", got, yaOauthMusicScope)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"device_code":"dc","user_code":"uc","verification_url":"https://ya.ru/device","expires_in":300,"interval":5}`))
	}))

	code, err := RequestDeviceCode(context.Background(), "device1234", "yamusic-tui")
	if err != nil {
		t.Fatalf("RequestDeviceCode: %v", err)
	}
	if code.DeviceCode != "dc" || code.UserCode != "uc" {
		t.Errorf("unexpected code: %+v", code)
	}
	if code.VerificationURL != "https://ya.ru/device" {
		t.Errorf("verification url = %q", code.VerificationURL)
	}
	if code.Interval != 5 || code.ExpiresIn != 300 {
		t.Errorf("interval/expires = %d/%d", code.Interval, code.ExpiresIn)
	}
}

func TestRequestDeviceCodeDefaults(t *testing.T) {
	withOAuthServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"device_code":"dc","user_code":"uc"}`))
	}))

	code, err := RequestDeviceCode(context.Background(), "device1234", "yamusic-tui")
	if err != nil {
		t.Fatalf("RequestDeviceCode: %v", err)
	}
	if code.VerificationURL != "https://ya.ru/device" {
		t.Errorf("verification url = %q", code.VerificationURL)
	}
	if code.Interval <= 0 || code.ExpiresIn <= 0 {
		t.Errorf("interval/expires not defaulted: %+v", code)
	}
}

func TestRequestDeviceCodeError(t *testing.T) {
	withOAuthServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_client","error_description":"bad client"}`))
	}))

	_, err := RequestDeviceCode(context.Background(), "device1234", "yamusic-tui")
	if err == nil || !strings.Contains(err.Error(), "invalid_client") {
		t.Fatalf("want invalid_client error, got %v", err)
	}
}

func TestPollDeviceToken(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		body    string
		want    string
		wantErr error
	}{
		{
			name:   "token",
			status: http.StatusOK,
			body:   `{"access_token":"y0_token","token_type":"bearer"}`,
			want:   "y0_token",
		},
		{
			name:    "pending",
			status:  http.StatusBadRequest,
			body:    `{"error":"authorization_pending"}`,
			wantErr: ErrAuthorizationPending,
		},
		{
			name:    "slow down",
			status:  http.StatusBadRequest,
			body:    `{"error":"slow_down"}`,
			wantErr: ErrSlowDown,
		},
		{
			name:    "expired",
			status:  http.StatusBadRequest,
			body:    `{"error":"expired_token","error_description":"Code has expired"}`,
			wantErr: ErrExpiredToken,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withOAuthServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/token" {
					t.Errorf("unexpected path %q", r.URL.Path)
				}
				if err := r.ParseForm(); err != nil {
					t.Errorf("parse form: %v", err)
				}
				if got := r.Form.Get("grant_type"); got != "device_code" {
					t.Errorf("grant_type = %q", got)
				}
				if got := r.Form.Get("code"); got != "dc" {
					t.Errorf("code = %q", got)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))

			token, err := PollDeviceToken(context.Background(), "dc")
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("want %v, got %v", tc.wantErr, err)
				}
				return
			}
			if tc.want != "" {
				if err != nil {
					t.Fatalf("PollDeviceToken: %v", err)
				}
				if token != tc.want {
					t.Errorf("token = %q, want %q", token, tc.want)
				}
				return
			}
		})
	}
}

func TestRequestDeviceCodeImplausibleValues(t *testing.T) {
	withOAuthServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"device_code":"dc","user_code":"uc","verification_url":"file:///etc/passwd","expires_in":1099511627776}`))
	}))

	code, err := RequestDeviceCode(context.Background(), "device1234", "yamusic-tui")
	if err != nil {
		t.Fatalf("RequestDeviceCode: %v", err)
	}
	if code.VerificationURL != "https://ya.ru/device" {
		t.Errorf("unsafe verification url kept: %q", code.VerificationURL)
	}
	if code.ExpiresIn != 300 {
		t.Errorf("implausible expires_in kept: %d", code.ExpiresIn)
	}
}

func TestIsRetryable(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   bool
	}{
		{"pending", http.StatusBadRequest, `{"error":"authorization_pending"}`, true},
		{"slow down", http.StatusBadRequest, `{"error":"slow_down"}`, true},
		{"rate limit", http.StatusTooManyRequests, `{"error":"rate_limit"}`, true},
		{"server error", http.StatusInternalServerError, `{"error":"internal"}`, true},
		{"html 502", http.StatusBadGateway, `<html>`, true},
		{"invalid client", http.StatusBadRequest, `{"error":"invalid_client"}`, false},
		{"access denied", http.StatusBadRequest, `{"error":"access_denied"}`, false},
		{"expired", http.StatusBadRequest, `{"error":"expired_token"}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsRetryable(oauthResponseError([]byte(tc.body), tc.status)); got != tc.want {
				t.Errorf("IsRetryable = %v, want %v", got, tc.want)
			}
		})
	}

	if !IsRetryable(ErrMalformedResponse) {
		t.Error("malformed response must be retryable")
	}
	if !IsRetryable(context.DeadlineExceeded) {
		t.Error("request timeout must be retryable")
	}
	if IsRetryable(context.Canceled) {
		t.Error("cancellation must not be retryable")
	}
}
