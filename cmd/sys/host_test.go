package sys

import (
	"reflect"
	"testing"
)

func TestParseAddHostInput_MultiHostMultiLine(t *testing.T) {
	input := `10.40.81.10 zlpdev-ratelimiter.zalopay.vn zlpdev-ratelimit-api.zalopay.vn
10.30.94.60 zlpstg-ratelimit-api.zalopay.vn zlpstg-ratelimit.zalopay.vn
10.30.94.60 ratelimit.zalopay.vn ratelimit-api.zalopay.vn`

	groups, unassignedIPs, unassignedDomains := parseAddHostInput(input)

	if len(unassignedIPs) != 0 {
		t.Fatalf("expected 0 unassigned IPs, got: %v", unassignedIPs)
	}
	if len(unassignedDomains) != 0 {
		t.Fatalf("expected 0 unassigned domains, got: %v", unassignedDomains)
	}
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got: %d", len(groups))
	}

	expected := []HostInputGroup{
		{
			IP:      "10.40.81.10",
			Domains: []string{"zlpdev-ratelimiter.zalopay.vn", "zlpdev-ratelimit-api.zalopay.vn"},
		},
		{
			IP:      "10.30.94.60",
			Domains: []string{"zlpstg-ratelimit-api.zalopay.vn", "zlpstg-ratelimit.zalopay.vn"},
		},
		{
			IP:      "10.30.94.60",
			Domains: []string{"ratelimit.zalopay.vn", "ratelimit-api.zalopay.vn"},
		},
	}

	if !reflect.DeepEqual(groups, expected) {
		t.Errorf("groups mismatch.\nGot: %+v\nWant: %+v", groups, expected)
	}
}

func TestParseAddHostInput_MultiHostSingleLineFlattened(t *testing.T) {
	// In terminal paste when newlines are converted to spaces by single-line input
	input := "10.40.81.10 zlpdev-ratelimiter.zalopay.vn zlpdev-ratelimit-api.zalopay.vn 10.30.94.60 zlpstg-ratelimit-api.zalopay.vn zlpstg-ratelimit.zalopay.vn 10.30.94.60 ratelimit.zalopay.vn ratelimit-api.zalopay.vn"

	groups, unassignedIPs, unassignedDomains := parseAddHostInput(input)

	if len(unassignedIPs) != 0 || len(unassignedDomains) != 0 {
		t.Fatalf("expected no unassigned items, got IPs: %v, doms: %v", unassignedIPs, unassignedDomains)
	}
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got: %d", len(groups))
	}

	if groups[0].IP != "10.40.81.10" || len(groups[0].Domains) != 2 {
		t.Errorf("group 0 mismatch: %+v", groups[0])
	}
	if groups[1].IP != "10.30.94.60" || len(groups[1].Domains) != 2 {
		t.Errorf("group 1 mismatch: %+v", groups[1])
	}
	if groups[2].IP != "10.30.94.60" || len(groups[2].Domains) != 2 {
		t.Errorf("group 2 mismatch: %+v", groups[2])
	}
}

func TestParseAddHostInput_WithCommentsAndURLs(t *testing.T) {
	input := `# Header comment
127.0.0.1 https://dev.local/ http://test.local # trailing comment
192.168.1.1 router.internal`

	groups, _, _ := parseAddHostInput(input)

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got: %d", len(groups))
	}

	expected := []HostInputGroup{
		{
			IP:      "127.0.0.1",
			Domains: []string{"dev.local", "test.local"},
		},
		{
			IP:      "192.168.1.1",
			Domains: []string{"router.internal"},
		},
	}

	if !reflect.DeepEqual(groups, expected) {
		t.Errorf("groups mismatch.\nGot: %+v\nWant: %+v", groups, expected)
	}
}

func TestParseAddHostInput_SingleDomainFirstThenIP(t *testing.T) {
	input := "example.com 1.2.3.4"

	groups, unassignedIPs, unassignedDomains := parseAddHostInput(input)

	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(unassignedIPs) != 0 || len(unassignedDomains) != 0 {
		t.Fatalf("expected 0 unassigned items")
	}
	if groups[0].IP != "1.2.3.4" || len(groups[0].Domains) != 1 || groups[0].Domains[0] != "example.com" {
		t.Errorf("unexpected group: %+v", groups[0])
	}
}

func TestParseAddHostInput_OnlyIP(t *testing.T) {
	input := "192.168.1.100"

	groups, unassignedIPs, unassignedDomains := parseAddHostInput(input)

	if len(groups) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(groups))
	}
	if len(unassignedIPs) != 1 || unassignedIPs[0] != "192.168.1.100" {
		t.Errorf("unexpected unassignedIPs: %v", unassignedIPs)
	}
	if len(unassignedDomains) != 0 {
		t.Errorf("expected 0 unassigned domains, got %v", unassignedDomains)
	}
}

func TestParseAddHostInput_OnlyDomains(t *testing.T) {
	input := "foo.local bar.local"

	groups, unassignedIPs, unassignedDomains := parseAddHostInput(input)

	if len(groups) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(groups))
	}
	if len(unassignedIPs) != 0 {
		t.Errorf("expected 0 unassigned IPs, got %v", unassignedIPs)
	}
	if len(unassignedDomains) != 2 || unassignedDomains[0] != "foo.local" || unassignedDomains[1] != "bar.local" {
		t.Errorf("unexpected unassignedDomains: %v", unassignedDomains)
	}
}
