package monime

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"
)

const testWebhookSecret = "test-secret-key-at-least-32-chars!!"

var testWebhookBody = []byte(
	`{"apiVersion":"caph.2025-08-23",` +
		`"event":{"id":"evt_123","name":"payment.completed",` +
		`"timestamp":"2026-09-22T00:00:00Z"},` +
		`"object":{"id":"pay_123","type":"payment"},` +
		`"data":{"id":"pay_123"}}`,
)

func signWebhook(
	t *testing.T,
	body []byte,
	secret string,
	timestamp int64,
) string {
	t.Helper()
	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(mac, "%d_%s", timestamp, body)
	return fmt.Sprintf("t=%d,v1=%s", timestamp, base64.StdEncoding.EncodeToString(mac.Sum(nil)))
}

func TestVerifyWebhookSignature(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()
	valid := signWebhook(t, testWebhookBody, testWebhookSecret, now)
	sig := valid[len(fmt.Sprintf("t=%d,v1=", now)):]

	tests := []struct {
		name    string
		body    []byte
		header  string
		secret  string
		want    WebhookVerificationReason
		wantErr bool
	}{
		{
			name:   "accepts signed delivery",
			body:   testWebhookBody,
			header: valid,
			secret: testWebhookSecret,
		},
		{
			name:   "rejects modified body",
			body:   []byte(`{"tampered":true}`),
			header: valid,
			secret: testWebhookSecret,
			want:   WebhookReasonSignatureMismatch,
		},
		{
			name:   "rejects duplicated header fields",
			body:   testWebhookBody,
			header: fmt.Sprintf("v1=%s,v1=%s", sig, sig),
			secret: testWebhookSecret,
			want:   WebhookReasonSignatureHeaderInvalid,
		},
		{
			name:   "rejects truncated signature",
			body:   testWebhookBody,
			header: fmt.Sprintf("t=%d,v1=%s", now, sig[:len(sig)-1]),
			secret: testWebhookSecret,
			want:   WebhookReasonSignatureHeaderInvalid,
		},
		{
			name:   "rejects malformed header",
			body:   testWebhookBody,
			header: "not-a-signature",
			secret: testWebhookSecret,
			want:   WebhookReasonSignatureHeaderInvalid,
		},
		{
			name:   "rejects stale timestamp",
			body:   testWebhookBody,
			header: signWebhook(t, testWebhookBody, testWebhookSecret, now-600),
			secret: testWebhookSecret,
			want:   WebhookReasonTimestampOutsideTolerance,
		},
		{
			name:   "rejects invalid payload",
			body:   []byte(`not json`),
			header: signWebhook(t, []byte(`not json`), testWebhookSecret, now),
			secret: testWebhookSecret,
			want:   WebhookReasonPayloadInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			event, err := VerifyWebhookSignature(tt.body, tt.header, tt.secret)
			if tt.want == "" {
				if err != nil {
					t.Fatalf("VerifyWebhookSignature returned error: %v", err)
				}
				if event.Event.Name != "payment.completed" {
					t.Fatalf("unexpected event name: %q", event.Event.Name)
				}
				return
			}
			var verifyErr *WebhookVerificationError
			if !errors.As(err, &verifyErr) || verifyErr.Reason != tt.want {
				t.Fatalf("expected %s, got: %v", tt.want, err)
			}
		})
	}
}

func TestWebhookServiceVerifyRequiresSecret(t *testing.T) {
	t.Parallel()

	client, err := New(Config{SpaceID: "spc-test", AccessToken: "tok-test"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	now := time.Now().Unix()
	header := signWebhook(t, testWebhookBody, testWebhookSecret, now)
	if _, err = client.Webhooks().Verify(testWebhookBody, header); err == nil {
		t.Fatal("expected error without configured secret")
	}
}
