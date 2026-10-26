package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// oauthBaseURL is the Yandex OAuth service root. It is a variable so tests can
// point the device flow at a local server.
var oauthBaseURL = yaOauthServerURL

// DeviceCode is the payload returned by the device authorization endpoint.
type DeviceCode struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

var (
	// ErrAuthorizationPending means the user has not confirmed the login yet:
	// poll again after the interval returned with the device code.
	ErrAuthorizationPending = errors.New("authorization pending")
	// ErrSlowDown means the client polls the token endpoint too often.
	ErrSlowDown = errors.New("slow down")
	// ErrExpiredToken means the device code is no longer valid and a new one
	// has to be requested.
	ErrExpiredToken = errors.New("device code expired")
	// ErrMalformedResponse means the OAuth service returned something that is
	// not the expected JSON: a proxy page, an empty body or a truncated reply.
	ErrMalformedResponse = errors.New("malformed oauth response")
)

// HTTPStatusError is an OAuth response with an unexpected HTTP status.
type HTTPStatusError struct {
	Status int
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("oauth: unexpected status %d", e.Status)
}

type oauthError struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
	HTTPStatus  int    `json:"-"`
}

func (e *oauthError) Error() string {
	if e.Description != "" {
		return e.Code + ": " + e.Description
	}
	return e.Code
}

// RequestDeviceCode starts the OAuth device flow. The returned code has to be
// shown to the user: they open VerificationURL, enter UserCode and confirm.
func RequestDeviceCode(ctx context.Context, deviceID, deviceName string) (*DeviceCode, error) {
	body, status, err := oauthRequest(ctx, "device/code", url.Values{
		"client_id":   {yaOauthClientID},
		"device_id":   {deviceID},
		"device_name": {deviceName},
		"scope":       {yaOauthMusicScope},
	})
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, oauthResponseError(body, status)
	}

	var code DeviceCode
	if err := json.Unmarshal(body, &code); err != nil {
		return nil, fmt.Errorf("%w: device code: %v", ErrMalformedResponse, err)
	}
	if code.DeviceCode == "" || code.UserCode == "" {
		return nil, fmt.Errorf("%w: empty device code", ErrMalformedResponse)
	}
	if u, err := url.Parse(code.VerificationURL); err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		code.VerificationURL = "https://ya.ru/device"
	}
	if code.Interval <= 0 {
		code.Interval = 5
	}
	if code.ExpiresIn <= 0 || code.ExpiresIn > 24*60*60 {
		code.ExpiresIn = 300
	}

	return &code, nil
}

// PollDeviceToken performs a single token request. It returns the access token,
// ErrAuthorizationPending while the user has not confirmed yet, ErrSlowDown if
// polling should be less frequent, or any other OAuth error.
func PollDeviceToken(ctx context.Context, deviceCode string) (string, error) {
	body, status, err := oauthRequest(ctx, "token", url.Values{
		"grant_type":    {"device_code"},
		"code":          {deviceCode},
		"client_id":     {yaOauthClientID},
		"client_secret": {yaOauthClientSecret},
	})
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", oauthResponseError(body, status)
	}

	var token struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &token); err != nil {
		return "", fmt.Errorf("%w: access token: %v", ErrMalformedResponse, err)
	}
	if token.AccessToken == "" {
		return "", fmt.Errorf("%w: empty access token", ErrMalformedResponse)
	}

	return token.AccessToken, nil
}

func oauthRequest(ctx context.Context, path string, form url.Values) ([]byte, int, error) {
	endpoint, err := url.JoinPath(oauthBaseURL, path)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func oauthResponseError(body []byte, status int) error {
	var oauthErr oauthError
	if err := json.Unmarshal(body, &oauthErr); err == nil && oauthErr.Code != "" {
		switch oauthErr.Code {
		case "authorization_pending":
			return ErrAuthorizationPending
		case "slow_down":
			return ErrSlowDown
		case "expired_token":
			return ErrExpiredToken
		}
		oauthErr.HTTPStatus = status
		return &oauthErr
	}
	return &HTTPStatusError{Status: status}
}

// IsRetryable reports whether a poll may be repeated without asking the user
// to authorize again: a pending confirmation, a network problem, a rate limit
// or a temporary server failure.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrAuthorizationPending) || errors.Is(err, ErrSlowDown) {
		return true
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	var statusErr *HTTPStatusError
	if errors.As(err, &statusErr) {
		return statusErr.Status == http.StatusTooManyRequests || statusErr.Status >= 500
	}
	var oauthErr *oauthError
	if errors.As(err, &oauthErr) {
		return oauthErr.HTTPStatus == http.StatusTooManyRequests || oauthErr.HTTPStatus >= 500
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	if errors.Is(err, ErrMalformedResponse) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	return false
}
