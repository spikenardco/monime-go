package monime

import (
	"context"
	"net/url"
	"strconv"
)

type USSDOTP struct {
	ID                    string            `json:"id"`
	Status                string            `json:"status"`
	DialCode              string            `json:"dialCode"`
	AuthorizedPhoneNumber string            `json:"authorizedPhoneNumber"`
	VerificationMessage   string            `json:"verificationMessage"`
	CreateTime            string            `json:"createTime"`
	ExpireTime            string            `json:"expireTime"`
	Metadata              map[string]string `json:"metadata"`
}

type CreateUSSDOTPInput struct {
	AuthorizedPhoneNumber string            `json:"authorizedPhoneNumber"`
	VerificationMessage   string            `json:"verificationMessage,omitempty"`
	Duration              string            `json:"duration,omitempty"`
	Metadata              map[string]string `json:"metadata,omitempty"`
}

type USSDOTPListQuery struct {
	Limit                                int
	After, Status, AuthorizedPhoneNumber string
}
type USSDOTPService struct{ client *Client }

func (s *USSDOTPService) Create(ctx context.Context, input CreateUSSDOTPInput) (USSDOTP, Response, error) {
	var result USSDOTP
	response, err := s.client.do(ctx, "POST", "/ussd-otps", nil, input, &result)
	return result, response, err
}
func (s *USSDOTPService) Get(ctx context.Context, id string) (USSDOTP, Response, error) {
	var result USSDOTP
	response, err := s.client.do(ctx, "GET", "/ussd-otps/"+url.PathEscape(id), nil, nil, &result)
	return result, response, err
}
func (s *USSDOTPService) List(ctx context.Context, query USSDOTPListQuery) (Page[USSDOTP], Response, error) {
	var items []USSDOTP
	response, err := s.client.do(ctx, "GET", "/ussd-otps", query.values(), nil, &items)
	return Page[USSDOTP]{Items: items}, response, err
}
func (s *USSDOTPService) Delete(ctx context.Context, id string) (Response, error) {
	return s.client.do(ctx, "DELETE", "/ussd-otps/"+url.PathEscape(id), nil, nil, nil)
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
		values.Set("status", q.Status)
	}
	if q.AuthorizedPhoneNumber != "" {
		values.Set("authorizedPhoneNumber", q.AuthorizedPhoneNumber)
	}
	return values
}
