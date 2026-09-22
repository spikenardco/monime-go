package monime

import (
	"encoding/json"
	"testing"
)

func TestChannelDecodesProviderDetails(t *testing.T) {
	var channel Channel
	if err := json.Unmarshal([]byte(`{
		"type":"momo",
		"provider":"m17",
		"reference":"txn-123",
		"phoneNumber":"23276000000",
		"fingerprint":"fingerprint-123",
		"metadata":{"source":"checkout"}
	}`), &channel); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if channel.Type != ChannelTypeMomo {
		t.Fatalf("channel type = %q, want %q", channel.Type, ChannelTypeMomo)
	}
	if channel.Provider == nil || *channel.Provider != "m17" {
		t.Fatalf("channel provider = %v, want m17", channel.Provider)
	}
	if channel.Reference == nil || *channel.Reference != "txn-123" {
		t.Fatalf("channel reference = %v, want txn-123", channel.Reference)
	}
	if channel.PhoneNumber == nil || *channel.PhoneNumber != "23276000000" {
		t.Fatalf("channel phone number = %v, want 23276000000", channel.PhoneNumber)
	}
	if channel.Fingerprint == nil || *channel.Fingerprint != "fingerprint-123" {
		t.Fatalf("channel fingerprint = %v, want fingerprint-123", channel.Fingerprint)
	}
	if channel.Metadata["source"] != "checkout" {
		t.Fatalf("channel metadata = %#v, want checkout source", channel.Metadata)
	}
}

func TestChannelMarshalsCardDetails(t *testing.T) {
	scheme := "visa"
	last4 := "4242"
	channel := Channel{
		Type:   ChannelTypeCard,
		Scheme: &scheme,
		Last4:  &last4,
	}

	encoded, err := json.Marshal(channel)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	want := `{"type":"card","scheme":"visa","last4":"4242"}`
	if string(encoded) != want {
		t.Fatalf("json.Marshal = %s, want %s", encoded, want)
	}
}
