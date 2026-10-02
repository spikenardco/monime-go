package monime

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// Country describes a country supported by Monime.
type Country struct {
	Code                string          `json:"code"`
	Name                string          `json:"name"`
	Currency            CountryCurrency `json:"currency"`
	SupportedCurrencies []Currency      `json:"supportedCurrencies,omitempty"`
}

// CountryCurrency describes a country's default currency and minor unit.
type CountryCurrency struct {
	Code       Currency `json:"code"`
	Unit       string   `json:"unit"`
	UnitLength int      `json:"unitLength"`
}

// CountryListParams filters a country list request.
type CountryListParams struct {
	Limit int
	After string
}

// ListCountries returns supported countries.
func (c *Client) ListCountries(
	ctx context.Context,
	params CountryListParams,
) (*Page[Country], *Response, error) {
	page, response, err := doList[Country](c, ctx, "/countries", params.values())
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &page, &response, nil
}

// GetCountry returns a country by ISO code.
func (c *Client) GetCountry(ctx context.Context, countryCode string) (*Country, *Response, error) {
	if countryCode == "" {
		return nil, nil, newValidationError("CountryCode", "must be non-empty")
	}
	var country Country
	response, err := c.do(ctx, operation{
		method: http.MethodGet,
		path:   "/countries/" + url.PathEscape(countryCode),
		output: &country,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &country, &response, nil
}

func (p CountryListParams) values() url.Values {
	values := url.Values{}
	if p.Limit != 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.After != "" {
		values.Set("after", p.After)
	}
	return values
}

func responseOrNil(response Response) *Response {
	if response == (Response{}) {
		return nil
	}
	return &response
}
