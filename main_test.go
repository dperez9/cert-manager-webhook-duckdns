package main

import (
	"os"
	"testing"

	"github.com/cert-manager/cert-manager/test/acme/dns"
	"github.com/dperez9/cert-manager-webhook-duckdns/duckdns"
)

var (
	zone    = os.Getenv("TEST_ZONE_NAME")
	dnsname = os.Getenv("DNS_NAME")
)

func TestRunsSuite(t *testing.T) {
	fixture := dns.NewFixture(duckdns.NewSolver(),
		dns.SetBinariesPath("__main__/hack/bin"),
		dns.SetResolvedZone(zone),
		dns.SetDNSName(dnsname),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath("testdata/duckdns"),
	)

	fixture.RunConformance(t)
}
