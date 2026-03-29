package tool

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"
)

const turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type turnstileVerifyResponse struct {
	Success bool `json:"success"`
}

func VerifyTurnstile(ctx context.Context, secret, token, ip string) (bool, error) {
	return VerifyTurnstileWithKey(ctx, secret, token, ip, "")
}

func VerifyTurnstileWithKey(ctx context.Context, secret, token, ip, idempotencyKey string) (bool, error) {
	if secret == "" || token == "" {
		return false, nil
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("secret", secret)
	_ = writer.WriteField("response", token)
	_ = writer.WriteField("remoteip", ip)
	if idempotencyKey != "" {
		_ = writer.WriteField("idempotency_key", idempotencyKey)
	}
	_ = writer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, turnstileVerifyURL, body)
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result turnstileVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Success, nil
}

func GenerateTurnstileIdempotencyKey() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("turnstile-%d", time.Now().UnixNano())
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:])
}
