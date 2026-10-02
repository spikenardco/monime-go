package monime

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// USSDOTPStatus is the state of a USSD OTP session.
type USSDOTPStatus string

const (
	USSDOTPPending  USSDOTPStatus = "pending"
	USSDOTPVerified USSDOTPStatus = "verified"
	USSDOTPExpired  USSDOTPStatus = "expired"
)

// USSDOTP is a USSD-based phone verification session.
type USSDOTP struct {
	ID                    string        `json:"id"`
	Status                USSDOTPStatus `json:"status"`
	DialCode              string        `json:"dialCode"`
	AuthorizedPhoneNumber string        `json:"authorizedPhoneNumber"`
	VerificationMessage   *string       `json:"verificationMessage,omitempty"`
	CreateTime            time.Time     `json:"createTime"`
	ExpireTime            time.Time     `json:"expireTime"`
	Metadata              Metadata      `json:"metadata,omitempty"`
}

// CreateUSSDOTPInput contains fields accepted when creating a USSD OTP session.
type CreateUSSDOTPInput struct {
	AuthorizedPhoneNumber string   `json:"authorizedPhoneNumber"`
	VerificationMessage   string   `json:"verificationMessage,omitempty"`
	Duration              string   `json:"duration,omitempty"`
	Metadata              Metadata `json:"metadata,omitempty"`
}

// USSDOTPListParams controls USSD OTP pagination.
type USSDOTPListParams struct {
	Limit int
	After string
}

// CreateUSSDOTP creates a phone-bound USSD OTP session.
func (c *Client) CreateUSSDOTP(
	ctx context.Context,
	idempotencyKey string,
	input CreateUSSDOTPInput,
) (*USSDOTP, *Response, error) {
	if input.AuthorizedPhoneNumber == "" {
		return nil, nil, newValidationError("AuthorizedPhoneNumber", "must be non-empty")
	}
	var otp USSDOTP
	response, err := c.do(ctx, operation{
		method:         http.MethodPost,
		path:           "/ussd-otps",
		input:          input,
		output:         &otp,
		idempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &otp, &response, nil
}

// GetUSSDOTP returns a USSD OTP session by ID.
func (c *Client) GetUSSDOTP(ctx context.Context, id string) (*USSDOTP, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	var otp USSDOTP
	response, err := c.do(ctx, operation{
		method: http.MethodGet,
		path:   "/ussd-otps/" + url.PathEscape(id),
		output: &otp,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &otp, &response, nil
}

// ListUSSDOTPs returns a page of USSD OTP sessions.
func (c *Client) ListUSSDOTPs(
	ctx context.Context,
	params USSDOTPListParams,
) (*Page[USSDOTP], *Response, error) {
	page, response, err := doList[USSDOTP](c, ctx, "/ussd-otps", params.values())
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &page, &response, nil
}

// DeleteUSSDOTP deletes a USSD OTP session.
func (c *Client) DeleteUSSDOTP(ctx context.Context, id string) (*Response, error) {
	if id == "" {
		return nil, newValidationError("ID", "must be non-empty")
	}
	response, err := c.do(ctx, operation{
		method:  http.MethodDelete,
		path:    "/ussd-otps/" + url.PathEscape(id),
		emptyOK: true,
	})
	return responseOrNil(response), err
}

func (p USSDOTPListParams) values() url.Values {
	values := url.Values{}
	if p.Limit != 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.After != "" {
		values.Set("after", p.After)
	}
	return values
}
