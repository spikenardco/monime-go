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
	CanPayTo         bool              `json:"canPayTo"`
	CanPayFrom       bool              `json:"canPayFrom"`
	CanVerifyAccount bool              `json:"canVerifyAccount"`
	Schemes          []string          `json:"schemes"`
	Metadata         map[string]string `json:"metadata"`
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
	var items []Bank
	response, err := s.client.do(ctx, "GET", "/banks", query.values(), nil, &items)
	return Page[Bank]{Items: items}, response, err
}
func (s *BankService) Get(ctx context.Context, providerID string) (Bank, Response, error) {
	var bank Bank
	response, err := s.client.do(ctx, "GET", "/banks/"+url.PathEscape(providerID), nil, nil, &bank)
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
