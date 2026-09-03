package monime

import (
	"context"
	"net/url"
	"strconv"
)

type MobileMoney struct {
	ProviderID string             `json:"providerId"`
	Name       string             `json:"name"`
	Country    string             `json:"country"`
	Status     ProviderStatus     `json:"status"`
	FeatureSet ProviderFeatureSet `json:"featureSet"`
	CreateTime string             `json:"createTime"`
	UpdateTime string             `json:"updateTime"`
}
type MobileMoneyListQuery struct {
	Country string
	Limit   int
	After   string
}
type MobileMoneyService struct{ client *Client }

func (s *MobileMoneyService) List(ctx context.Context, query MobileMoneyListQuery) (Page[MobileMoney], Response, error) {
	var items []MobileMoney
	response, err := s.client.do(ctx, "GET", "/momos", query.values(), nil, &items)
	return Page[MobileMoney]{Items: items}, response, err
}
func (s *MobileMoneyService) Get(ctx context.Context, providerID string) (MobileMoney, Response, error) {
	var provider MobileMoney
	response, err := s.client.do(ctx, "GET", "/momos/"+url.PathEscape(providerID), nil, nil, &provider)
	return provider, response, err
}
func (q MobileMoneyListQuery) values() url.Values {
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
