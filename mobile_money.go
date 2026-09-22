package monime

import (
	"context"
	"net/url"
	"strconv"
)

// MobileMoney is a supported mobile money provider.
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

func (s *MobileMoneyService) List(
	ctx context.Context,
	query MobileMoneyListQuery,
) (Page[MobileMoney], Response, error) {
	return doList[MobileMoney](s.client, ctx, "/momos", query.values())
}

func (s *MobileMoneyService) Get(ctx context.Context, providerID string) (MobileMoney, Response, error) {
	if providerID == "" {
		return MobileMoney{}, Response{}, newValidationError("ProviderID", "must be non-empty")
	}
	var provider MobileMoney
	response, err := s.client.do(ctx, operation{
		method: "GET",
		path:   "/momos/" + url.PathEscape(providerID),
		output: &provider,
	})
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
