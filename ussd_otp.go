package monime

import (
	"context"
	"net/url"
	"strconv"
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
	CreateTime            string        `json:"createTime"`
	ExpireTime            string        `json:"expireTime"`
	Metadata              Metadata      `json:"metadata,omitempty"`
}

type CreateUSSDOTPInput struct {
	AuthorizedPhoneNumber string   `json:"authorizedPhoneNumber"`
	VerificationMessage   string   `json:"verificationMessage,omitempty"`
	Duration              string   `json:"duration,omitempty"`
	Metadata              Metadata `json:"metadata,omitempty"`
}

type USSDOTPListQuery struct {
	Limit                 int
	After                 string
	Status                USSDOTPStatus
	AuthorizedPhoneNumber string
}

type USSDOTPService struct{ client *Client }

func (s *USSDOTPService) Create(ctx context.Context, input CreateUSSDOTPInput) (USSDOTP, Response, error) {
	if input.AuthorizedPhoneNumber == "" {
		return USSDOTP{}, Response{}, newValidationError("AuthorizedPhoneNumber", "must be non-empty")
	}
	var result USSDOTP
	response, err := s.client.do(ctx, operation{
		method: "POST",
		path:   "/ussd-otps",
		input:  input,
		output: &result,
	})
	return result, response, err
}

func (s *USSDOTPService) Get(ctx context.Context, id string) (USSDOTP, Response, error) {
	if id == "" {
		return USSDOTP{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result USSDOTP
	response, err := s.client.do(ctx, operation{
		method: "GET",
		path:   "/ussd-otps/" + url.PathEscape(id),
		output: &result,
	})
	return result, response, err
}

func (s *USSDOTPService) List(ctx context.Context, query USSDOTPListQuery) (Page[USSDOTP], Response, error) {
	return doList[USSDOTP](s.client, ctx, "/ussd-otps", query.values())
}

func (s *USSDOTPService) Delete(ctx context.Context, id string) (Response, error) {
	if id == "" {
		return Response{}, newValidationError("ID", "must be non-empty")
	}
	return s.client.do(ctx, operation{
		method: "DELETE",
		path:   "/ussd-otps/" + url.PathEscape(id),
	})
}

func (q USSDOTPListQuery) values() url.Values {
	values := url.Values{}
	if q.Limit != 0 {
		values.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.After != "" {
		values.Set("after", q.After)
	}
	if q.Status != "" {
		values.Set("status", string(q.Status))
	}
	if q.AuthorizedPhoneNumber != "" {
		values.Set("authorizedPhoneNumber", q.AuthorizedPhoneNumber)
	}
	return values
}
