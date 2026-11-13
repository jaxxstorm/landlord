//go:build integration

package floci

import (
	"net/url"
	"os"
	"regexp"
	"testing"
)

func TestStart(t *testing.T) {
	fixture := Start(t)

	endpoint, err := url.Parse(fixture.Endpoint)
	if err != nil {
		t.Fatalf("parse Floci endpoint: %s", err)
	}
	if endpoint.Scheme != "http" || endpoint.Host == "" {
		t.Fatalf("unexpected Floci endpoint: %q", fixture.Endpoint)
	}
	if fixture.Region == "" || fixture.AccessKeyID == "" || fixture.SecretAccessKey == "" {
		t.Fatal("Floci fixture must provide AWS connection details")
	}
	if !regexp.MustCompile(`^\d{12}$`).MatchString(fixture.AccessKeyID) {
		t.Fatalf("Floci account ID must be 12 digits: %q", fixture.AccessKeyID)
	}
	if !regexp.MustCompile(`^landlord-[a-f0-9]{32}$`).MatchString(fixture.ResourcePrefix) {
		t.Fatalf("unexpected Floci resource prefix: %q", fixture.ResourcePrefix)
	}

	fixture.SetEnvironment(t)
	if got := os.Getenv("AWS_ENDPOINT_URL"); got != fixture.Endpoint {
		t.Fatalf("AWS endpoint override = %q, want %q", got, fixture.Endpoint)
	}
	if got := os.Getenv("AWS_ACCESS_KEY_ID"); got != fixture.AccessKeyID {
		t.Fatalf("AWS account override = %q, want %q", got, fixture.AccessKeyID)
	}
}
