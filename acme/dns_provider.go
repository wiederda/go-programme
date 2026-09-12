package main

import "context"

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

type IPv64Provider struct{}

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
	return nil
}

func (p *IPv64Provider) Cleanup(ctx context.Context, fqdn string, value string) error {
	return nil
}

var dnsProviders = map[string]DNSProvider{
	"ipv64": &IPv64Provider{},
}

func GetDNSProvider(name string) (DNSProvider, bool) {
	provider, ok := dnsProviders[name]
	return provider, ok
}
