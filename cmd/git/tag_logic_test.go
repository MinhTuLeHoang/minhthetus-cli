package git

import (
	"reflect"
	"strings"
	"testing"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/config"
	"github.com/spf13/cobra"
)

func TestParseAndCompareSemVer(t *testing.T) {
	v1, err := ParseSemVer("0.69.9")
	if err != nil {
		t.Fatalf("unexpected error parsing 0.69.9: %v", err)
	}
	v2, err := ParseSemVer("0.69.11")
	if err != nil {
		t.Fatalf("unexpected error parsing 0.69.11: %v", err)
	}
	if v1.Compare(v2) >= 0 {
		t.Fatalf("expected 0.69.9 < 0.69.11, got %d", v1.Compare(v2))
	}

	v3, _ := ParseSemVer("1.2.3")
	v4, _ := ParseSemVer("1.10.0")
	if v3.Compare(v4) >= 0 || v4.Compare(v3) <= 0 {
		t.Fatalf("expected 1.2.3 < 1.10.0")
	}
	if v3.Compare(v3) != 0 {
		t.Fatalf("expected 1.2.3 == 1.2.3")
	}

	for _, invalid := range []string{"", "1.2", "abc", "v1.2.3", "1.2.3-beta", "1.2.3.4"} {
		if _, err := ParseSemVer(invalid); err == nil {
			t.Errorf("expected ParseSemVer(%q) to fail", invalid)
		}
	}

	u1, err := ParseUserSemVer("  v1.2.3 ")
	if err != nil || u1.String() != "1.2.3" {
		t.Fatalf("ParseUserSemVer('  v1.2.3 ') = %v, %v; want 1.2.3", u1, err)
	}
	u2, err := ParseUserSemVer("V2.0.1")
	if err != nil || u2.String() != "2.0.1" {
		t.Fatalf("ParseUserSemVer('V2.0.1') = %v, %v; want 2.0.1", u2, err)
	}
}

func TestBumpSemVer(t *testing.T) {
	zero := SemVer{0, 0, 0}
	if p, err := zero.Bump("patch"); err != nil || p.String() != "0.0.1" {
		t.Fatalf("patch bump from 0.0.0 = %v, %v", p, err)
	}
	if m, err := zero.Bump("minor"); err != nil || m.String() != "0.1.0" {
		t.Fatalf("minor bump from 0.0.0 = %v, %v", m, err)
	}
	if mj, err := zero.Bump("major"); err != nil || mj.String() != "1.0.0" {
		t.Fatalf("major bump from 0.0.0 = %v, %v", mj, err)
	}

	base := SemVer{1, 2, 3}
	if p, _ := base.Bump("patch"); p.String() != "1.2.4" {
		t.Fatalf("got %s, want 1.2.4", p.String())
	}
	if m, _ := base.Bump("minor"); m.String() != "1.3.0" {
		t.Fatalf("got %s, want 1.3.0", m.String())
	}
	if mj, _ := base.Bump("major"); mj.String() != "2.0.0" {
		t.Fatalf("got %s, want 2.0.0", mj.String())
	}
	if _, err := base.Bump("unknown"); err == nil {
		t.Fatalf("expected error for unknown bump type")
	}
}

func TestDiscoverTagPrefixes(t *testing.T) {
	tags := []string{
		"qc-v1.0.0",
		"qc-v1.1.0",
		"stg-v1.1.0",
		"qc-cicd-v1.0.0",
		"release-v1.10.0",
		"release-v1.2.0",
		"v1.0.0",
		"v12.3.4",
		"app2-v1.2.3",
		"v2-qc-v1.0.0",
		"1.0.0",           // bare semver -> empty prefix skipped
		"qc-v1.0.0-beta",  // pre-release suffix -> ignored
		"release/v1.0.0",  // slash -> ignored
		"legacy-tag",      // non-semver -> ignored
	}

	got := DiscoverTagPrefixes(tags)
	want := map[string]string{
		"app2-v":    "app2-v1.2.3",
		"qc-cicd-v": "qc-cicd-v1.0.0",
		"qc-v":      "qc-v1.1.0",
		"release-v": "release-v1.10.0",
		"stg-v":     "stg-v1.1.0",
		"v":         "v12.3.4",
		"v2-qc-v":   "v2-qc-v1.0.0",
	}

	if len(got) != len(want) {
		t.Fatalf("DiscoverTagPrefixes returned %d prefixes (%+v), want %d", len(got), got, len(want))
	}
	for _, d := range got {
		if wantLatest, ok := want[d.Prefix]; !ok {
			t.Errorf("unexpected prefix %q", d.Prefix)
		} else if d.LatestTag != wantLatest {
			t.Errorf("prefix %q latest = %q, want %q", d.Prefix, d.LatestTag, wantLatest)
		}
	}
}

func TestFindLatestTagForPrefix_AnchoredDigitsInPrefix(t *testing.T) {
	tags := []string{"v1.0.0", "v2-qc-v1.0.0", "v2-qc-v1.2.0", "v-legacy", "v1.0.0-rc1"}

	tagV, verV, foundV := FindLatestTagForPrefix(tags, "v")
	if !foundV || tagV != "v1.0.0" || verV.String() != "1.0.0" {
		t.Fatalf("prefix 'v' got (%q, %v, %v), want ('v1.0.0', 1.0.0, true)", tagV, verV, foundV)
	}

	tagV2, verV2, foundV2 := FindLatestTagForPrefix(tags, "v2-qc-v")
	if !foundV2 || tagV2 != "v2-qc-v1.2.0" || verV2.String() != "1.2.0" {
		t.Fatalf("prefix 'v2-qc-v' got (%q, %v, %v), want ('v2-qc-v1.2.0', 1.2.0, true)", tagV2, verV2, foundV2)
	}
}

func TestComputeBaseVersion(t *testing.T) {
	tags := []string{"qc-v1.2.0", "stg-v1.1.5", "qc-cicd-v0.9.0", "v-legacy"}

	details, base := ComputeBaseVersion(tags, []string{"qc-v", "stg-v"})
	if len(details) != 2 || base.String() != "1.2.0" {
		t.Fatalf("got (%+v, %s), want base 1.2.0", details, base.String())
	}

	detailsSingle, baseSingle := ComputeBaseVersion(tags, []string{"qc-cicd-v"})
	if len(detailsSingle) != 1 || baseSingle.String() != "0.9.0" {
		t.Fatalf("got (%+v, %s), want base 0.9.0", detailsSingle, baseSingle.String())
	}

	_, baseEmpty := ComputeBaseVersion([]string{"v-legacy"}, []string{"qc-v", "stg-v"})
	if baseEmpty.String() != "0.0.0" {
		t.Fatalf("expected 0.0.0 fallback, got %s", baseEmpty.String())
	}
}

func TestRuleParsingAndValidation(t *testing.T) {
	got := ParsePrefixRule(" qc-v** , , stg-v* , qc-v ")
	want := []string{"qc-v", "stg-v"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParsePrefixRule = %v, want %v", got, want)
	}

	if formatted := FormatPrefixRule([]string{"qc-v**", "stg-v*", "qc-v"}); formatted != "qc-v,stg-v" {
		t.Fatalf("FormatPrefixRule = %q, want 'qc-v,stg-v'", formatted)
	}

	for _, valid := range []string{"qc-v", "stg_v", "qc-cicd-v", "v", "app2-v", "2fast-v"} {
		if err := ValidatePrefix(valid); err != nil {
			t.Errorf("ValidatePrefix(%q) unexpected error: %v", valid, err)
		}
	}
	for _, invalid := range []string{"", "-v", "-qc", "-", "_v", "bad tag", "qc-v*", "qc-v**", "release/v", "..", "v1.0"} {
		if err := ValidatePrefix(invalid); err == nil {
			t.Errorf("ValidatePrefix(%q) expected error, got nil", invalid)
		}
	}

	if err := ValidatePrefixRule([]string{"qc-v", "stg-v"}); err != nil {
		t.Errorf("ValidatePrefixRule valid rule unexpected error: %v", err)
	}
	for _, badRule := range [][]string{{"-v"}, {"qc-v", "-v"}, {"qc-v", "bad tag"}, {".."}} {
		if err := ValidatePrefixRule(badRule); err == nil {
			t.Errorf("ValidatePrefixRule(%v) expected error, got nil", badRule)
		}
	}
}

func TestBuildDiscoveryMenuItems(t *testing.T) {
	detected := []DetectedPrefix{
		{Prefix: "qc-cicd-v", LatestTag: "qc-cicd-v1.0.0", Version: SemVer{1, 0, 0}},
		{Prefix: "release-v", LatestTag: "release-v1.10.0", Version: SemVer{1, 10, 0}},
	}
	presets := config.DefaultTagPresets()

	// Initial sandbox discovery: qc-cicd-v pre-selected, release-v not pre-selected, Custom last
	items := BuildDiscoveryMenuItems(detected, presets, nil, false, "+ Custom")
	if len(items) < 4 {
		t.Fatalf("expected at least 4 items, got %d", len(items))
	}
	if items[0].Prefix != "qc-cicd-v" || !items[0].Preselected {
		t.Errorf("expected first item qc-cicd-v preselected, got %+v", items[0])
	}
	if items[1].Prefix != "release-v" || items[1].Preselected {
		t.Errorf("expected second item release-v NOT preselected for sandbox, got %+v", items[1])
	}
	if !items[len(items)-1].IsCustom {
		t.Errorf("expected last item to be Custom, got %+v", items[len(items)-1])
	}

	// Prod reconfig with existingRule = ["v"]
	prodItems := BuildDiscoveryMenuItems(detected, presets, []string{"v"}, true, "+ Custom")
	foundVSelected := false
	for _, it := range prodItems {
		if it.Prefix == "v" && it.Preselected {
			foundVSelected = true
		}
		if it.Prefix == "qc-cicd-v" && it.Preselected {
			t.Errorf("did not expect qc-cicd-v preselected when existingRule is [v]")
		}
	}
	if !foundVSelected {
		t.Errorf("expected existing rule 'v' to be preselected")
	}
}

func TestPreFlightConflictResolution(t *testing.T) {
	targetTags := []string{"qc-v1.1.0", "stg-v1.1.0", "uat-v1.1.0"}
	localSet := map[string]bool{"qc-v1.1.0": true}
	remoteSet := map[string]bool{"stg-v1.1.0": true}

	statuses := CheckTagAvailability(targetTags, func(s string) bool { return localSet[s] }, remoteSet)
	if !HasAnyConflict(statuses) {
		t.Fatalf("expected conflict")
	}
	if statuses[0].StatusLabel() != "Local" || statuses[1].StatusLabel() != "Remote" || statuses[2].StatusLabel() != "Available" {
		t.Fatalf("unexpected status labels: %+v", statuses)
	}
	missing := MissingTagsOnly(statuses)
	if !reflect.DeepEqual(missing, []string{"uat-v1.1.0"}) {
		t.Fatalf("MissingTagsOnly = %v, want [uat-v1.1.0]", missing)
	}

	// All conflict -> MissingTagsOnly must be empty
	allLocal := map[string]bool{"qc-v1.1.0": true, "stg-v1.1.0": true}
	allStatuses := CheckTagAvailability([]string{"qc-v1.1.0", "stg-v1.1.0"}, func(s string) bool { return allLocal[s] }, nil)
	if len(MissingTagsOnly(allStatuses)) != 0 {
		t.Fatalf("expected MissingTagsOnly to be empty when all tags exist")
	}

	// NextAvailablePatchVersion skips consecutive occupied versions across local & remote
	occupiedLocal := map[string]bool{"qc-v1.1.1": true, "stg-v1.1.2": true}
	occupiedRemote := map[string]bool{"stg-v1.1.1": true}
	nextVer, nextTags, err := NextAvailablePatchVersion(
		[]string{"qc-v", "stg-v"},
		SemVer{1, 1, 0},
		func(s string) bool { return occupiedLocal[s] },
		occupiedRemote,
	)
	if err != nil {
		t.Fatalf("NextAvailablePatchVersion unexpected error: %v", err)
	}
	if nextVer.String() != "1.1.3" {
		t.Fatalf("got version %s, want 1.1.3", nextVer.String())
	}
	if !reflect.DeepEqual(nextTags, []string{"qc-v1.1.3", "stg-v1.1.3"}) {
		t.Fatalf("got tags %v, want [qc-v1.1.3 stg-v1.1.3]", nextTags)
	}
}

func TestTagDevStgCobraFlagsAndCompletion(t *testing.T) {
	for _, flagName := range []string{"patch", "minor", "major", "message", "reconfig", "pick", "prod"} {
		if TagDevStgCmd.Flags().Lookup(flagName) == nil {
			t.Fatalf("expected flag --%s to be registered", flagName)
		}
	}

	// Verify mutual exclusion directly on TagDevStgCmd (-P and -N)
	if err := TagDevStgCmd.ValidateFlagGroups(); err != nil {
		t.Fatalf("expected clean default flag groups, got: %v", err)
	}
	if err := TagDevStgCmd.ParseFlags([]string{"-P", "-N"}); err != nil {
		t.Fatalf("unexpected ParseFlags error: %v", err)
	}
	t.Cleanup(func() {
		_ = TagDevStgCmd.Flags().Set("patch", "false")
		_ = TagDevStgCmd.Flags().Set("minor", "false")
		TagDevStgCmd.Flags().Lookup("patch").Changed = false
		TagDevStgCmd.Flags().Lookup("minor").Changed = false
	})
	if execErr := TagDevStgCmd.ValidateFlagGroups(); execErr == nil || !strings.Contains(execErr.Error(), "none of the others can be") {
		t.Fatalf("expected mutual exclusion error for -P -N, got: %v", execErr)
	}

	// Verify RegisterFlagCompletionFunc for message
	fn, ok := TagDevStgCmd.GetFlagCompletionFunc("message")
	if !ok || fn == nil {
		t.Fatalf("expected flag completion function for 'message'")
	}
	msgSuggestions, msgDir := fn(TagDevStgCmd, nil, "")
	if msgDir != cobra.ShellCompDirectiveNoFileComp || len(msgSuggestions) == 0 {
		t.Fatalf("unexpected message completions: %v, %v", msgSuggestions, msgDir)
	}

	completions, directive := TagDevStgCmd.ValidArgsFunction(TagDevStgCmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("expected ShellCompDirectiveNoFileComp")
	}
	compJoined := strings.Join(completions, ",")
	for _, expected := range []string{"--reconfig", "--pick", "--prod", "--patch", "--minor", "--major", "--message"} {
		if !strings.Contains(compJoined, expected) {
			t.Errorf("expected ValidArgsFunction completions to include %s, got %v", expected, completions)
		}
	}
}
