package monime

import (
	"context"
	"net/url"
	"strconv"
)

type ProviderStatus struct {
	Active bool `json:"active"`
}

type ProviderFeature struct {
	CanPayTo         bool     `json:"canPayTo,omitempty"`
	CanPayFrom       bool     `json:"canPayFrom,omitempty"`
	CanVerifyAccount bool     `json:"canVerifyAccount,omitempty"`
	Schemes          []string `json:"schemes,omitempty"`
	Metadata         Metadata `json:"metadata,omitempty"`
}

type ProviderFeatureSet struct {
	Payout          ProviderFeature `json:"payout"`
	Payment         ProviderFeature `json:"payment"`
	KYCVerification ProviderFeature `json:"kycVerification"`
}

// Bank is a supported bank provider.
type Bank struct {
	ProviderID string             `json:"providerId"`
	Name       string             `json:"name"`
	Country    string             `json:"country"`
	Status     ProviderStatus     `json:"status"`
	FeatureSet ProviderFeatureSet `json:"featureSet"`
	CreateTime string             `json:"createTime"`
	UpdateTime string             `json:"updateTime"`
}

type BankListQuery struct {
	Country string
	Limit   int
	After   string
}

type BankService struct{ client *Client }

func (s *BankService) List(ctx context.Context, query BankListQuery) (Page[Bank], Response, error) {
	return doList[Bank](s.client, ctx, "/banks", query.values())
}

func (s *BankService) Get(ctx context.Context, providerID string) (Bank, Response, error) {
	if providerID == "" {
		return Bank{}, Response{}, newValidationError("ProviderID", "must be non-empty")
	}
	var bank Bank
	response, err := s.client.do(ctx, operation{
		method: "GET",
		path:   "/banks/" + url.PathEscape(providerID),
		output: &bank,
	})
	return bank, response, err
}

func (q BankListQuery) values() url.Values {
	values := url.Values{}
	if q.Country != "" {
		values.Set("country", q.Country)
	}
	if q.Limit != 0 {
		values.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.After != "" {
		values.Set("after", q.After)
	}
	return values
}
