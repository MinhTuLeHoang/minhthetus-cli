package sys

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

// HostEntry represents a single domain-IP mapping
type HostEntry struct {
	IP         string
	Domain     string
	Normalized string
	ParsedIP   net.IP
	IPPriority int
}

// HostCmd represents the sys host command
var HostCmd = &cobra.Command{
	Use:   "host",
	Short: "Manage the system host file (/etc/hosts)",
	Long: `Provides search, list, add, delete, and formatting/sorting operations
on the system host file (/etc/hosts). Formatting and writing back changes 
requires superuser (sudo) privileges.`,
	Run: func(cmd *cobra.Command, args []string) {
		runHostCLI()
	},
}

func runHostCLI() {
	for {
		// Read and parse the hosts file on each iteration to ensure we operate on the latest state
		headerComments, entries, err := parseHostsFile()
		if err != nil {
			fmt.Println(ui.ErrorMessage(fmt.Sprintf("Failed to parse hosts file: %v", err)))
			return
		}

		choice, err := ui.Choose("Select Host Management Action", []string{
			"Search Hosts",
			"Show All Hosts",
			"Add Host",
			"Delete Host",
			"Format & Sort hosts file",
			"Exit",
		})
		if err != nil {
			fmt.Println(ui.ErrorMessage(fmt.Sprintf("UI error: %v", err)))
			return
		}

		if choice == "Exit" || choice == "" {
			break
		}

		switch choice {
		case "Search Hosts":
			if err := searchEntries(entries); err != nil {
				fmt.Println(ui.ErrorMessage(fmt.Sprintf("Search failed: %v", err)))
			}
			return
		case "Show All Hosts":
			if err := showAllEntries(entries); err != nil {
				fmt.Println(ui.ErrorMessage(fmt.Sprintf("Show all failed: %v", err)))
			}
		case "Add Host":
			if err := addEntry(headerComments, entries); err != nil {
				fmt.Println(ui.ErrorMessage(fmt.Sprintf("Add failed: %v", err)))
			}
			return
		case "Delete Host":
			if err := deleteEntry(headerComments, entries); err != nil {
				fmt.Println(ui.ErrorMessage(fmt.Sprintf("Delete failed: %v", err)))
			}
			return
		case "Format & Sort hosts file":
			if err := formatHostsFile(headerComments, entries); err != nil {
				fmt.Println(ui.ErrorMessage(fmt.Sprintf("Format failed: %v", err)))
				return
			}
			fmt.Println("\n" + ui.BoldStyle.Render("Formatted /etc/hosts contents:") + "\n")
			catCmd := exec.Command("cat", "/etc/hosts")
			catCmd.Stdout = os.Stdout
			catCmd.Stderr = os.Stderr
			_ = catCmd.Run()
			return
		}
		fmt.Println()
	}
}

// parseHostsFile reads /etc/hosts, extracts header comments, and parses domain-IP pairs
func parseHostsFile() ([]string, []HostEntry, error) {
	file, err := os.Open("/etc/hosts")
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var headerComments []string
	var entries []HostEntry
	scanner := bufio.NewScanner(file)
	inHeader := true

	for scanner.Scan() {
		line := scanner.Text()

		if inHeader {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				headerComments = append(headerComments, line)
				continue
			} else {
				inHeader = false
			}
		}

		// Strip inline comments
		parseLine := line
		if idx := strings.Index(parseLine, "#"); idx >= 0 {
			parseLine = parseLine[:idx]
		}
		trimmed := strings.TrimSpace(parseLine)

		if trimmed == "" {
			continue
		}

		parts := strings.Fields(trimmed)
		if len(parts) < 2 {
			continue
		}

		ipStr := parts[0]
		parsedIP, _ := parseIP(ipStr)
		priority := getIPPriority(parsedIP, ipStr)

		for _, domain := range parts[1:] {
			entries = append(entries, HostEntry{
				IP:         ipStr,
				Domain:     domain,
				Normalized: normalizeDomain(domain),
				ParsedIP:   parsedIP,
				IPPriority: priority,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return headerComments, entries, nil
}

// parseIP attempts to parse an IP address, handling optional port syntax
func parseIP(ipStr string) (net.IP, bool) {
	if ip := net.ParseIP(ipStr); ip != nil {
		return ip, true
	}
	if host, _, err := net.SplitHostPort(ipStr); err == nil {
		if ip := net.ParseIP(host); ip != nil {
			return ip, true
		}
	}
	return nil, false
}

// getIPPriority maps an IP address to its priority category
func getIPPriority(ip net.IP, ipStr string) int {
	if ip == nil {
		return 5 // Invalid
	}
	// Loopback
	if ip.IsLoopback() || ip.Equal(net.ParseIP("::1")) || ip.Equal(net.ParseIP("127.0.0.1")) {
		return 1
	}
	// Broadcast / Link-local
	if ip.Equal(net.IPv4bcast) || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ipStr == "255.255.255.255" {
		return 2
	}
	// IPv4
	if ip.To4() != nil {
		return 3
	}
	// IPv6
	return 4
}

// normalizeDomain converts a domain to lowercase and strips common environment prefixes to group environments
func normalizeDomain(domain string) string {
	d := strings.ToLower(domain)
	prefixes := []string{
		"qc-", "qcc-", "qccs-", "stg-", "dev-", "prod-", "internal-", "new-",
		"qc", "qcc", "qccs", "stg", "dev", "prod",
	}
	for {
		found := false
		for _, p := range prefixes {
			if strings.HasPrefix(d, p) {
				d = strings.TrimPrefix(d, p)
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	return d
}

// compareIPs compares two IP addresses lexicographically using their 16-byte representation
func compareIPs(ipA, ipB net.IP) int {
	if ipA == nil && ipB == nil {
		return 0
	}
	if ipA == nil {
		return 1
	}
	if ipB == nil {
		return -1
	}

	bytesA := ipA.To16()
	bytesB := ipB.To16()
	for i := 0; i < 16; i++ {
		if bytesA[i] < bytesB[i] {
			return -1
		}
		if bytesA[i] > bytesB[i] {
			return 1
		}
	}
	return 0
}

// compareHostEntries determines relative ordering based on IP priority, IP byte value, normalized domain, and original domain
func compareHostEntries(a, b HostEntry) bool {
	if a.IPPriority != b.IPPriority {
		return a.IPPriority < b.IPPriority
	}

	if a.IPPriority != 5 {
		ipComp := compareIPs(a.ParsedIP, b.ParsedIP)
		if ipComp != 0 {
			return ipComp < 0
		}
	} else {
		if a.IP != b.IP {
			return a.IP < b.IP
		}
	}

	if a.Normalized != b.Normalized {
		return a.Normalized < b.Normalized
	}

	return a.Domain < b.Domain
}

// formatLine formats a line with clean alignment
func formatLine(ip, domain string) string {
	if len(ip) < 16 {
		return fmt.Sprintf("%-16s%s", ip, domain)
	}
	return fmt.Sprintf("%s\t%s", ip, domain)
}

// searchEntries queries parsed entries for exact matches and relevant results
func searchEntries(entries []HostEntry) error {
	query, err := ui.Input("Enter domain or IP to search", "")
	if err != nil {
		return err
	}
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		fmt.Println(ui.WarningMessage("Search query cannot be empty."))
		return nil
	}

	var exactMatches []HostEntry
	var relevantEntries []HostEntry

	exactKeys := make(map[string]bool)
	relevantKeys := make(map[string]bool)

	// Direct matches
	for _, entry := range entries {
		ipMatch := strings.Contains(strings.ToLower(entry.IP), query)
		domainMatch := strings.Contains(strings.ToLower(entry.Domain), query)
		if ipMatch || domainMatch {
			key := entry.IP + " " + entry.Domain
			exactMatches = append(exactMatches, entry)
			exactKeys[key] = true
		}
	}

	// Relevant results (same IP or same normalized domain)
	for _, exact := range exactMatches {
		for _, entry := range entries {
			key := entry.IP + " " + entry.Domain
			if exactKeys[key] || relevantKeys[key] {
				continue
			}
			sameIP := entry.IP == exact.IP
			sameNorm := entry.Normalized == exact.Normalized

			if sameIP || sameNorm {
				relevantEntries = append(relevantEntries, entry)
				relevantKeys[key] = true
			}
		}
	}

	if len(exactMatches) == 0 {
		fmt.Println(ui.InfoMessage("No matching entries found."))
		return nil
	}

	fmt.Println("\n" + ui.BoldStyle.Render("Exact Matches:"))
	for _, entry := range exactMatches {
		fmt.Printf("  %s\n", formatLine(entry.IP, entry.Domain))
	}

	if len(relevantEntries) > 0 {
		fmt.Println("\n" + ui.BoldStyle.Render("Relevant Results (same IP or normalized domain):"))
		for _, entry := range relevantEntries {
			fmt.Printf("  %s\n", formatLine(entry.IP, entry.Domain))
		}
	}
	fmt.Println()
	return nil
}

// showAllEntries lists all hosts, sorted cleanly according to spec
func showAllEntries(entries []HostEntry) error {
	if len(entries) == 0 {
		fmt.Println(ui.InfoMessage("No entries found in hosts file."))
		return nil
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return compareHostEntries(entries[i], entries[j])
	})

	fmt.Println("\n" + ui.BoldStyle.Render("Current Host Entries:"))
	var prevIP string
	for _, entry := range entries {
		if prevIP != "" && entry.IP != prevIP {
			fmt.Println()
		}
		fmt.Printf("  %s\n", formatLine(entry.IP, entry.Domain))
		prevIP = entry.IP
	}
	fmt.Println()
	return nil
}

// addEntry adds a domain-IP entry, prompting for missing details and checking duplicates
func addEntry(headerComments []string, entries []HostEntry) error {
	input, err := ui.Input("Enter IP and/or Domains (e.g., 127.0.0.1 google.com facebook.com)", "")
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Println(ui.ErrorMessage("Input cannot be empty."))
		return nil
	}

	parts := strings.Fields(input)
	var ipStr string
	var domains []string

	if len(parts) == 1 {
		val := parts[0]
		_, ok := parseIP(val)
		if ok {
			ipStr = val
			domainInput, err := ui.Input("Enter Domains for "+ipStr, "")
			if err != nil {
				return err
			}
			domains = strings.Fields(domainInput)
		} else {
			domains = []string{val}
			ipInput, err := ui.Input("Enter IP for "+val, "")
			if err != nil {
				return err
			}
			ipStr = strings.TrimSpace(ipInput)
		}
	} else if len(parts) >= 2 {
		_, ok := parseIP(parts[0])
		if ok {
			ipStr = parts[0]
			domains = parts[1:]
		} else {
			domains = parts
			ipInput, err := ui.Input(fmt.Sprintf("Enter IP for %d domains", len(domains)), "")
			if err != nil {
				return err
			}
			ipStr = strings.TrimSpace(ipInput)
		}
	}

	ipStr = strings.TrimSpace(ipStr)
	if ipStr == "" || len(domains) == 0 {
		fmt.Println(ui.ErrorMessage("Both IP and at least one Domain are required."))
		return nil
	}

	parsedIP, _ := parseIP(ipStr)

	var addedDomains []string
	var skippedDomains []string

	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			continue
		}

		duplicated := false
		for _, entry := range entries {
			if entry.IP == ipStr && strings.ToLower(entry.Domain) == strings.ToLower(domain) {
				duplicated = true
				break
			}
		}

		if duplicated {
			skippedDomains = append(skippedDomains, domain)
			fmt.Println(ui.WarningMessage(fmt.Sprintf("Host already exists (skipped): %s -> %s", ipStr, domain)))
		} else {
			addedDomains = append(addedDomains, domain)
			newEntry := HostEntry{
				IP:         ipStr,
				Domain:     domain,
				Normalized: normalizeDomain(domain),
				ParsedIP:   parsedIP,
				IPPriority: getIPPriority(parsedIP, ipStr),
			}
			entries = append(entries, newEntry)
		}
	}

	if len(addedDomains) > 0 {
		var sb strings.Builder
		sb.WriteString(ui.BoldStyle.Render("Preparing to add new host entries:") + "\n")
		header := ui.BoldStyle.Render(fmt.Sprintf("%-16s   %s", "IP", "DOMAIN"))
		sb.WriteString("  " + header + "\n")
		sb.WriteString("  " + ui.GrayStyle().Render(strings.Repeat("-", 50)) + "\n")
		for i, dom := range addedDomains {
			if i == 0 {
				sb.WriteString(fmt.Sprintf("  %-16s   %s\n", ipStr, dom))
			} else {
				sb.WriteString(fmt.Sprintf("  %-16s   %s\n", "", dom))
			}
		}
		sb.WriteString("\n" + ui.InfoMessage("Sudo privileges required. Please authenticate using your password or biometrics to continue."))

		err = writeHostsFile(headerComments, entries, sb.String())
		if err != nil {
			return err
		}
		fmt.Println(ui.SuccessMessage(fmt.Sprintf("Successfully added domains to %s: %s", ipStr, strings.Join(addedDomains, ", "))))
	} else if len(skippedDomains) > 0 {
		fmt.Println(ui.InfoMessage("No new entries were added (all input domains were duplicates)."))
	}

	return nil
}

// deleteEntry deletes an entry interactively
func deleteEntry(headerComments []string, entries []HostEntry) error {
	query, err := ui.Input("Enter domain or IP to filter entries for deletion", "")
	if err != nil {
		return err
	}
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		fmt.Println(ui.WarningMessage("Filter query cannot be empty."))
		return nil
	}

	var matches []HostEntry
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.IP), query) || strings.Contains(strings.ToLower(entry.Domain), query) {
			matches = append(matches, entry)
		}
	}

	if len(matches) == 0 {
		fmt.Println(ui.InfoMessage("No matching entries found to delete."))
		return nil
	}

	options := make([]string, len(matches))
	for i, entry := range matches {
		options[i] = fmt.Sprintf("%s -> %s", entry.IP, entry.Domain)
	}
	options = append(options, "Cancel")

	selected, err := ui.Choose("Select host entry to delete", options)
	if err != nil {
		return err
	}

	if selected == "Cancel" || selected == "" {
		fmt.Println(ui.InfoMessage("Deletion cancelled."))
		return nil
	}

	var targetIP, targetDomain string
	parts := strings.Split(selected, " -> ")
	if len(parts) == 2 {
		targetIP = parts[0]
		targetDomain = parts[1]
	}

	if targetIP == "" || targetDomain == "" {
		fmt.Println(ui.ErrorMessage("Could not parse selected entry."))
		return nil
	}

	confirm, err := ui.Confirm(fmt.Sprintf("Are you sure you want to delete %s %s?", targetIP, targetDomain), 0, false)
	if err != nil {
		return err
	}

	if !confirm {
		fmt.Println(ui.InfoMessage("Deletion cancelled."))
		return nil
	}

	var newEntries []HostEntry
	deleted := false
	for _, entry := range entries {
		if entry.IP == targetIP && entry.Domain == targetDomain {
			deleted = true
			continue
		}
		newEntries = append(newEntries, entry)
	}

	if !deleted {
		fmt.Println(ui.ErrorMessage("Selected entry not found in the original list."))
		return nil
	}

	var sb strings.Builder
	sb.WriteString(ui.BoldStyle.Render("Preparing to delete host entry:") + "\n")
	header := ui.BoldStyle.Render(fmt.Sprintf("%-16s   %s", "IP", "DOMAIN"))
	sb.WriteString("  " + header + "\n")
	sb.WriteString("  " + ui.GrayStyle().Render(strings.Repeat("-", 50)) + "\n")
	sb.WriteString(fmt.Sprintf("  %-16s   %s\n", targetIP, targetDomain))
	sb.WriteString("\n" + ui.InfoMessage("Sudo privileges required. Please authenticate using your password or biometrics to continue."))

	err = writeHostsFile(headerComments, newEntries, sb.String())
	if err != nil {
		return err
	}

	fmt.Println(ui.SuccessMessage(fmt.Sprintf("Successfully deleted entry: %s %s", targetIP, targetDomain)))
	return nil
}

// formatHostsFile sorts and reformats /etc/hosts
func formatHostsFile(headerComments []string, entries []HostEntry) error {
	sort.SliceStable(entries, func(i, j int) bool {
		return compareHostEntries(entries[i], entries[j])
	})

	preview := ui.BoldStyle.Render("Preparing to format and sort hosts file.") +
		"\n\n" + ui.InfoMessage("Sudo privileges required. Please authenticate using your password or biometrics to continue.")

	err := writeHostsFile(headerComments, entries, preview)
	if err != nil {
		return err
	}

	fmt.Println(ui.SuccessMessage("Successfully sorted and formatted hosts file!"))
	return nil
}

// writeHostsFile generates hosts output structure, writes to temp file, and copies via sudo cp
func writeHostsFile(headerComments []string, entries []HostEntry, previewMsg string) error {
	sort.SliceStable(entries, func(i, j int) bool {
		return compareHostEntries(entries[i], entries[j])
	})

	var sb strings.Builder
	for _, line := range headerComments {
		sb.WriteString(line + "\n")
	}

	var prevIP string
	for i, entry := range entries {
		if i > 0 && entry.IP != prevIP {
			sb.WriteString("\n")
		}
		sb.WriteString(formatLine(entry.IP, entry.Domain) + "\n")
		prevIP = entry.IP
	}

	tmpFile, err := os.CreateTemp("", "minhthetus-hosts")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tmpFile.Name()
	defer os.Remove(tempPath)

	if _, err := tmpFile.WriteString(sb.String()); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpFile.Close()

	if previewMsg != "" {
		fmt.Println()
		fmt.Println(previewMsg)
		fmt.Println()
	}

	cmd := exec.Command("sudo", "cp", tempPath, "/etc/hosts")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to overwrite /etc/hosts (requires sudo): %w", err)
	}

	return nil
}
