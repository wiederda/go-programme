package main

import (
	"context"
	"os"
	"strings"
)

type DNSProviderConfig struct {
	Authentication string
	Required       []DNSProviderField
	Optional       []DNSProviderField
	SupportsTXT    bool
	SupportsDelete bool
}

type DNSProviderField struct {
	Name        string
	Description string
	Required    bool
}

type DNSProvider interface {
	Name() string
	Description() string
	Config() DNSProviderConfig

	Present(ctx context.Context, fqdn string, value string) error
	Cleanup(ctx context.Context, fqdn string, value string) error
}

type IPv64Provider struct {
	Token string
}

func NewIPv64Provider(token string) *IPv64Provider {
	return &IPv64Provider{
		Token: token,
	}
}

func (p *IPv64Provider) Name() string {
	return "ipv64"
}

func (p *IPv64Provider) Description() string {
	return "IPv64 DNS API"
}

func (p *IPv64Provider) Config() DNSProviderConfig {
	return DNSProviderConfig{
		Authentication: "API Token",
		Required: []DNSProviderField{
			{
				Name:        "token",
				Description: "API Token",
				Required:    true,
			},
		},
		SupportsTXT:    true,
		SupportsDelete: true,
	}
}

func (p *IPv64Provider) Present(ctx context.Context, fqdn string, value string) error {
	domain, prefix, err := IPv64SplitFQDN(fqdn)
	if err != nil {
		return err
	}

	return p.addRecord(ctx, domain, prefix, "TXT", value)
}

func (p *IPv64Provider) Cleanup(ctx context.Context, fqdn string, value string) error {
	domain, prefix, err := IPv64SplitFQDN(fqdn)
	if err != nil {
		return err
	}

	return p.deleteRecord(ctx, domain, prefix, "TXT", value)
}

var dnsProviders = map[string]DNSProvider{
	"ipv64": NewIPv64Provider(""),
}

func GetDNSProvider(name string) (DNSProvider, bool) {
	provider, ok := dnsProviders[name]

	if !ok {
		return nil, false
	}

	switch p := provider.(type) {
	case *IPv64Provider:
		p.Token = GetIPv64Token()
	}

	return provider, true
}

func GetIPv64Token() string {
	return strings.TrimSpace(os.Getenv("IPV64_TOKEN"))
}
