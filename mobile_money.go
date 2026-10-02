package monime

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// MobileMoney is a supported mobile-money provider.
type MobileMoney struct {
	ProviderID string             `json:"providerId"`
	Name       string             `json:"name"`
	Country    string             `json:"country"`
	Status     ProviderStatus     `json:"status"`
	FeatureSet ProviderFeatureSet `json:"featureSet"`
	CreateTime time.Time          `json:"createTime"`
	UpdateTime *time.Time         `json:"updateTime,omitempty"`
}

// MobileMoneyListParams filters a mobile-money provider list request.
type MobileMoneyListParams struct {
	Country string
	Limit   int
	After   string
}

// ListMobileMoneyProviders returns mobile-money providers in a country.
func (c *Client) ListMobileMoneyProviders(
	ctx context.Context,
	params MobileMoneyListParams,
) (*Page[MobileMoney], *Response, error) {
	if params.Country == "" {
		return nil, nil, newValidationError("Country", "must be non-empty")
	}
	page, response, err := doList[MobileMoney](c, ctx, "/momos", params.values())
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &page, &response, nil
}

// GetMobileMoneyProvider returns a mobile-money provider by provider ID.
func (c *Client) GetMobileMoneyProvider(
	ctx context.Context,
	providerID string,
) (*MobileMoney, *Response, error) {
	if providerID == "" {
		return nil, nil, newValidationError("ProviderID", "must be non-empty")
	}
	var provider MobileMoney
	response, err := c.do(ctx, operation{
		method: http.MethodGet,
		path:   "/momos/" + url.PathEscape(providerID),
		output: &provider,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &provider, &response, nil
}

func (p MobileMoneyListParams) values() url.Values {
	values := url.Values{"country": {p.Country}}
	if p.Limit != 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.After != "" {
		values.Set("after", p.After)
	}
	return values
}
