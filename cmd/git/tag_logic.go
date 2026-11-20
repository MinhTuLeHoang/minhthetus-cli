package git

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/config"
)

const (
	// LocalConfigSandboxKey stores the comma-separated sandbox tag prefixes in .git/config.
	LocalConfigSandboxKey = "minhthetus.tag.sandbox"
	// LocalConfigProdKey stores the comma-separated production tag prefixes in .git/config.
	LocalConfigProdKey = "minhthetus.tag.prod"

	maxPatchIncrementSearch = 10000
)

var (
	strictSemVerRe   = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
	tagDiscoveryRe   = regexp.MustCompile(`^([a-zA-Z0-9_-]*?v?)(\d+\.\d+\.\d+)$`)
	validTagPrefixRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
)

// SemVer represents a numeric Major.Minor.Patch semantic version.
type SemVer struct {
	Major int
	Minor int
	Patch int
}

// ParseSemVer parses a strict "X.Y.Z" semantic version string.
func ParseSemVer(s string) (SemVer, error) {
	m := strictSemVerRe.FindStringSubmatch(s)
	if m == nil {
		return SemVer{}, fmt.Errorf("invalid semantic version %q: expected X.Y.Z", s)
	}
	major, err := strconv.Atoi(m[1])
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid major version in %q: %w", s, err)
	}
	minor, err := strconv.Atoi(m[2])
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid minor version in %q: %w", s, err)
	}
	patch, err := strconv.Atoi(m[3])
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid patch version in %q: %w", s, err)
	}
	return SemVer{Major: major, Minor: minor, Patch: patch}, nil
}

// ParseUserSemVer parses user-entered version input, trimming whitespace and stripping an optional leading 'v'/'V'.
func ParseUserSemVer(s string) (SemVer, error) {
	cleaned := strings.TrimSpace(s)
	cleaned = strings.TrimPrefix(cleaned, "v")
	cleaned = strings.TrimPrefix(cleaned, "V")
	return ParseSemVer(cleaned)
}

// String returns the canonical "Major.Minor.Patch" format.
func (v SemVer) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare compares v with other: returns -1 if v < other, 0 if equal, +1 if v > other.
func (v SemVer) Compare(other SemVer) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	return 0
}

// Bump increments the semantic version according to bumpType ("patch", "minor", or "major").
func (v SemVer) Bump(bumpType string) (SemVer, error) {
	switch bumpType {
	case "patch":
		return SemVer{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}, nil
	case "minor":
		return SemVer{Major: v.Major, Minor: v.Minor + 1, Patch: 0}, nil
	case "major":
		return SemVer{Major: v.Major + 1, Minor: 0, Patch: 0}, nil
	default:
		return SemVer{}, fmt.Errorf("unsupported bump type %q", bumpType)
	}
}

// DetectedPrefix represents a tag prefix discovered in repository tag history along with its latest tag.
type DetectedPrefix struct {
	Prefix    string
	LatestTag string
	Version   SemVer
}

// DiscoverTagPrefixes scans tag names using `^([a-zA-Z0-9_-]*?v?)(\d+\.\d+\.\d+)$`,
// skips bare numeric tags (empty prefix), and returns unique prefixes sorted alphabetically with their highest SemVer.
func DiscoverTagPrefixes(tags []string) []DetectedPrefix {
	byPrefix := make(map[string]DetectedPrefix)
	for _, raw := range tags {
		tag := strings.TrimSpace(raw)
		m := tagDiscoveryRe.FindStringSubmatch(tag)
		if m == nil {
			continue
		}
		prefix := m[1]
		if prefix == "" {
			// Skip bare semver tags like "1.0.0" to avoid invalid empty rules.
			continue
		}
		ver, err := ParseSemVer(m[2])
		if err != nil {
			continue
		}
		existing, ok := byPrefix[prefix]
		if !ok || ver.Compare(existing.Version) > 0 {
			byPrefix[prefix] = DetectedPrefix{
				Prefix:    prefix,
				LatestTag: tag,
				Version:   ver,
			}
		}
	}

	result := make([]DetectedPrefix, 0, len(byPrefix))
	for _, item := range byPrefix {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Prefix < result[j].Prefix
	})
	return result
}

// FindLatestTagForPrefix finds the highest semantic version tag strictly matching `^<QuoteMeta(prefix)>(\d+\.\d+\.\d+)$`.
func FindLatestTagForPrefix(tags []string, prefix string) (string, SemVer, bool) {
	if prefix == "" {
		return "", SemVer{}, false
	}
	re := regexp.MustCompile("^" + regexp.QuoteMeta(prefix) + `(\d+\.\d+\.\d+)$`)

	var bestTag string
	var bestVer SemVer
	found := false

	for _, raw := range tags {
		tag := strings.TrimSpace(raw)
		m := re.FindStringSubmatch(tag)
		if m == nil {
			continue
		}
		ver, err := ParseSemVer(m[1])
		if err != nil {
			continue
		}
		if !found || ver.Compare(bestVer) > 0 {
			bestTag = tag
			bestVer = ver
			found = true
		}
	}

	return bestTag, bestVer, found
}

// PrefixLatest holds the latest tag resolution result for a single rule prefix.
type PrefixLatest struct {
	Prefix    string
	LatestTag string
	Version   SemVer
	Found     bool
}

// ComputeBaseVersion inspects tags across all prefixes in the active rule and returns per-prefix details plus the max SemVer (or 0.0.0).
func ComputeBaseVersion(tags []string, prefixes []string) ([]PrefixLatest, SemVer) {
	details := make([]PrefixLatest, 0, len(prefixes))
	base := SemVer{}
	hasAny := false

	for _, prefix := range prefixes {
		latestTag, ver, found := FindLatestTagForPrefix(tags, prefix)
		details = append(details, PrefixLatest{
			Prefix:    prefix,
			LatestTag: latestTag,
			Version:   ver,
			Found:     found,
		})
		if found && (!hasAny || ver.Compare(base) > 0) {
			base = ver
			hasAny = true
		}
	}

	return details, base
}

// ValidatePrefix ensures a tag prefix starts with an alphanumeric character and contains only `[a-zA-Z0-9_-]`.
func ValidatePrefix(p string) error {
	if p == "" {
		return fmt.Errorf("prefix cannot be empty")
	}
	if !validTagPrefixRe.MatchString(p) {
		return fmt.Errorf("invalid prefix %q: must start with a letter or digit and use only letters, digits, '-' or '_'", p)
	}
	return nil
}

// ValidatePrefixRule validates every prefix in a parsed rule slice.
func ValidatePrefixRule(prefixes []string) error {
	for _, p := range prefixes {
		if err := ValidatePrefix(p); err != nil {
			return err
		}
	}
	return nil
}

// ParsePrefixRule splits a comma-separated rule string, trims whitespace, strips trailing '*',
// removes empty segments, and deduplicates while preserving order.
func ParsePrefixRule(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	seen := make(map[string]bool, len(parts))
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		cleaned := strings.TrimSpace(part)
		cleaned = strings.TrimRight(cleaned, "*")
		cleaned = strings.TrimSpace(cleaned)
		if cleaned == "" || seen[cleaned] {
			continue
		}
		seen[cleaned] = true
		result = append(result, cleaned)
	}
	return result
}

// FormatPrefixRule normalizes and joins prefixes into a comma-separated `.git/config` value.
func FormatPrefixRule(prefixes []string) string {
	return strings.Join(ParsePrefixRule(strings.Join(prefixes, ",")), ",")
}

// BuildTargetTags produces `<prefix><version>` for every prefix in order.
func BuildTargetTags(prefixes []string, ver SemVer) []string {
	vStr := ver.String()
	tags := make([]string, 0, len(prefixes))
	for _, p := range prefixes {
		tags = append(tags, p+vStr)
	}
	return tags
}

// IsProdHeuristicPrefix reports whether a discovered prefix looks like a production tag prefix.
func IsProdHeuristicPrefix(prefix string) bool {
	lower := strings.ToLower(prefix)
	return prefix == "v" ||
		prefix == "release-v" ||
		strings.Contains(lower, "prod") ||
		strings.Contains(lower, "release")
}

// ShouldPreselectDetected returns whether a detected prefix should be pre-selected during initial discovery.
func ShouldPreselectDetected(prefix string, isProd bool) bool {
	if isProd {
		return IsProdHeuristicPrefix(prefix)
	}
	return !IsProdHeuristicPrefix(prefix)
}

// DiscoveryMenuItem represents one selectable row in the Smart Auto-Discovery multi-select menu.
type DiscoveryMenuItem struct {
	Prefix      string
	Label       string
	Preselected bool
	IsCustom    bool
}

// BuildDiscoveryMenuItems constructs the ordered MultiChoose rows:
// 1. Detected prefixes from repository tag history (top, pre-selected according to existingRule or env heuristic)
// 2. Global presets from git-tag-presets.json not already listed (active environment first)
// 3. Custom prefix entry at the bottom
func BuildDiscoveryMenuItems(detected []DetectedPrefix, presets config.TagPresets, existingRule []string, isProd bool, customLabel string) []DiscoveryMenuItem {
	existingSet := make(map[string]bool, len(existingRule))
	for _, p := range existingRule {
		existingSet[p] = true
	}
	hasExisting := len(existingRule) > 0

	seen := make(map[string]bool)
	items := make([]DiscoveryMenuItem, 0, len(detected)+len(presets.Sandbox)+len(presets.Prod)+1)

	for _, d := range detected {
		if seen[d.Prefix] {
			continue
		}
		seen[d.Prefix] = true
		preselected := ShouldPreselectDetected(d.Prefix, isProd)
		if hasExisting {
			preselected = existingSet[d.Prefix]
		}
		items = append(items, DiscoveryMenuItem{
			Prefix:      d.Prefix,
			Label:       fmt.Sprintf("%s*  (Detected in repo - latest: %s)", d.Prefix, d.LatestTag),
			Preselected: preselected,
		})
	}

	appendPresets := func(list []string, kind string) {
		for _, p := range list {
			if p == "" || seen[p] {
				continue
			}
			seen[p] = true
			items = append(items, DiscoveryMenuItem{
				Prefix:      p,
				Label:       fmt.Sprintf("%s*  (%s)", p, kind),
				Preselected: hasExisting && existingSet[p],
			})
		}
	}

	if isProd {
		appendPresets(presets.Prod, "Preset Prod")
		appendPresets(presets.Sandbox, "Preset Sandbox")
	} else {
		appendPresets(presets.Sandbox, "Preset Sandbox")
		appendPresets(presets.Prod, "Preset Prod")
	}

	// If an existing rule contained custom prefixes not in detected or global presets, include them too.
	for _, p := range existingRule {
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		items = append(items, DiscoveryMenuItem{
			Prefix:      p,
			Label:       fmt.Sprintf("%s*  (Saved in repo rule)", p),
			Preselected: true,
		})
	}

	items = append(items, DiscoveryMenuItem{
		Label:    customLabel,
		IsCustom: true,
	})

	return items
}

// TagStatus captures local and remote existence for a single target tag during pre-flight check.
type TagStatus struct {
	Tag          string
	ExistsLocal  bool
	ExistsRemote bool
}

// HasConflict reports whether the tag already exists locally or on the remote.
func (s TagStatus) HasConflict() bool {
	return s.ExistsLocal || s.ExistsRemote
}

// StatusLabel returns a human-readable conflict/availability label.
func (s TagStatus) StatusLabel() string {
	switch {
	case s.ExistsLocal && s.ExistsRemote:
		return "Local & Remote"
	case s.ExistsLocal:
		return "Local"
	case s.ExistsRemote:
		return "Remote"
	default:
		return "Available"
	}
}

// CheckTagAvailability checks all targetTags against local and remote tag sets.
func CheckTagAvailability(targetTags []string, localExists func(string) bool, remoteTags map[string]bool) []TagStatus {
	statuses := make([]TagStatus, 0, len(targetTags))
	for _, tag := range targetTags {
		local := false
		if localExists != nil {
			local = localExists(tag)
		}
		remote := false
		if remoteTags != nil {
			remote = remoteTags[tag]
		}
		statuses = append(statuses, TagStatus{
			Tag:          tag,
			ExistsLocal:  local,
			ExistsRemote: remote,
		})
	}
	return statuses
}

// HasAnyConflict reports whether any TagStatus in statuses has a local or remote conflict.
func HasAnyConflict(statuses []TagStatus) bool {
	for _, s := range statuses {
		if s.HasConflict() {
			return true
		}
	}
	return false
}

// MissingTagsOnly returns the subset of tags that exist on neither local nor remote.
func MissingTagsOnly(statuses []TagStatus) []string {
	missing := make([]string, 0, len(statuses))
	for _, s := range statuses {
		if !s.HasConflict() {
			missing = append(missing, s.Tag)
		}
	}
	return missing
}

// NextAvailablePatchVersion increments Patch starting from start+1 until all generated target tags
// are absent from both local and remote repositories.
func NextAvailablePatchVersion(prefixes []string, start SemVer, localExists func(string) bool, remoteTags map[string]bool) (SemVer, []string, error) {
	if len(prefixes) == 0 {
		return SemVer{}, nil, fmt.Errorf("no prefixes provided")
	}
	candidate, _ := start.Bump("patch")
	for i := 0; i < maxPatchIncrementSearch; i++ {
		tags := BuildTargetTags(prefixes, candidate)
		statuses := CheckTagAvailability(tags, localExists, remoteTags)
		if !HasAnyConflict(statuses) {
			return candidate, tags, nil
		}
		candidate, _ = candidate.Bump("patch")
	}
	return SemVer{}, nil, fmt.Errorf("exceeded %d patch increments searching for available tag version", maxPatchIncrementSearch)
}

// ConfigKeyForEnv returns the .git/config key for the selected environment profile.
func ConfigKeyForEnv(isProd bool) string {
	if isProd {
		return LocalConfigProdKey
	}
	return LocalConfigSandboxKey
}

// EnvLabel returns the display label for the selected environment profile.
func EnvLabel(isProd bool) string {
	if isProd {
		return "PRODUCTION"
	}
	return "SANDBOX"
}
