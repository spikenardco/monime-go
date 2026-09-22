package monime

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DefaultWebhookTolerance is the maximum accepted age of a webhook delivery.
const DefaultWebhookTolerance = 300 * time.Second

// WebhookVerificationReason identifies why verification failed.
type WebhookVerificationReason string

const (
	WebhookReasonSignatureHeaderInvalid    WebhookVerificationReason = "signature_header_invalid"
	WebhookReasonTimestampOutsideTolerance WebhookVerificationReason = "timestamp_outside_tolerance"
	WebhookReasonSignatureMismatch         WebhookVerificationReason = "signature_mismatch"
	WebhookReasonPayloadInvalid            WebhookVerificationReason = "payload_invalid"
)

// WebhookVerificationError is returned when a webhook delivery cannot be
// authenticated or decoded. Inspect Reason instead of parsing the message.
type WebhookVerificationError struct {
	Reason WebhookVerificationReason
}

func (e *WebhookVerificationError) Error() string {
	return fmt.Sprintf("monime: invalid webhook (%s)", e.Reason)
}

// VerifyWebhookSignature authenticates a Monime webhook delivery using the
// HS256 shared secret and returns the decoded event.
//
// The rawBody must be the exact bytes received from Monime; parsing and
// re-serializing JSON before verification changes the signed payload.
// The signature header has the form "t=<unix-seconds>,v1=<base64-hmac>"
// where the HMAC-SHA256 is computed over "<timestamp>_<raw-body>".
//
// Only HS256 is supported: Monime has not published the delivery details
// needed to verify ES256 signatures, and the official HMAC verification
// guide (https://docs.monime.io/guide/webhook/hmac-verification) is
// unpublished. The wire protocol is mirrored from the monimejs reference
// implementation; the secret length rule (32-256 characters) comes from the
// versioned OpenAPI verificationMethod schema.
func VerifyWebhookSignature(rawBody []byte, signatureHeader, secret string) (WebhookEvent, error) {
	if len(secret) < 32 || len(secret) > 256 {
		return WebhookEvent{}, fmt.Errorf("monime: webhook secret must contain 32 to 256 characters")
	}

	timestampText, signature, err := parseSignatureHeader(signatureHeader)
	if err != nil {
		return WebhookEvent{}, err
	}

	timestamp, err := strconv.ParseInt(timestampText, 10, 64)
	if err != nil {
		return WebhookEvent{}, &WebhookVerificationError{Reason: WebhookReasonSignatureHeaderInvalid}
	}
	now := time.Now().Unix()
	diff := now - timestamp
	if diff < 0 {
		diff = -diff
	}
	if diff > int64(DefaultWebhookTolerance/time.Second) {
		return WebhookEvent{}, &WebhookVerificationError{Reason: WebhookReasonTimestampOutsideTolerance}
	}

	signed := make([]byte, 0, len(timestampText)+1+len(rawBody))
	signed = append(signed, timestampText...)
	signed = append(signed, '_')
	signed = append(signed, rawBody...)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(signed)
	if !hmac.Equal(mac.Sum(nil), signature) {
		return WebhookEvent{}, &WebhookVerificationError{Reason: WebhookReasonSignatureMismatch}
	}

	event, err := ParseWebhookEvent(rawBody)
	if err != nil {
		return WebhookEvent{}, &WebhookVerificationError{Reason: WebhookReasonPayloadInvalid}
	}
	return event, nil
}

func parseSignatureHeader(header string) (string, []byte, error) {
	invalid := &WebhookVerificationError{Reason: WebhookReasonSignatureHeaderInvalid}
	parts := strings.Split(header, ",")
	if len(parts) != 2 {
		return "", nil, invalid
	}
	fields := make(map[string]string, 2)
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return "", nil, invalid
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "t" && key != "v1" {
			return "", nil, invalid
		}
		fields[key] = value
	}
	timestampText, ok := fields["t"]
	signatureText, ok2 := fields["v1"]
	if !ok || !ok2 || !isCanonicalTimestamp(timestampText) {
		return "", nil, invalid
	}
	signature, err := base64.StdEncoding.DecodeString(signatureText)
	if err != nil || len(signature) != sha256.Size || base64.StdEncoding.EncodeToString(signature) != signatureText {
		return "", nil, invalid
	}
	return timestampText, signature, nil
}

func isCanonicalTimestamp(s string) bool {
	if s == "" {
		return false
	}
	if s == "0" {
		return true
	}
	if s[0] < '1' || s[0] > '9' {
		return false
	}
	for _, c := range s[1:] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
