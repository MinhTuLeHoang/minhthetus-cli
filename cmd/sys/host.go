package sys

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

// getHostsFilePath returns the absolute path to the system hosts file depending on the OS
func getHostsFilePath() string {
	if runtime.GOOS == "windows" {
		systemRoot := os.Getenv("SystemRoot")
		if systemRoot == "" {
			systemRoot = `C:\Windows`
		}
		return filepath.Join(systemRoot, "System32", "drivers", "etc", "hosts")
	}
	return "/etc/hosts"
}

// privilegeMessage returns an OS-appropriate message regarding required permissions
func privilegeMessage() string {
	if runtime.GOOS == "windows" {
		return "Administrator privileges required. A UAC elevation prompt may appear to apply changes."
	}
	return "Sudo privileges required. Please authenticate using your password or biometrics to continue."
}

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
	Short: "Manage the system host file",
	Long: `Provides search, list, add, delete, and formatting/sorting operations
on the system host file. Formatting and writing back changes 
requires elevated (sudo or Administrator) privileges.`,
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
			hostsPath := getHostsFilePath()
			fmt.Println("\n" + ui.BoldStyle.Render(fmt.Sprintf("Formatted %s contents:", hostsPath)) + "\n")
			data, err := os.ReadFile(hostsPath)
			if err == nil {
				fmt.Print(string(data))
			}
			return
		}
		fmt.Println()
	}
}

// parseHostsFile reads the system hosts file, extracts header comments, and parses domain-IP pairs
func parseHostsFile() ([]string, []HostEntry, error) {
	file, err := os.Open(getHostsFilePath())
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

	var sb strings.Builder
	sb.WriteString(ui.BoldStyle.Render("Exact Matches:") + "\n")
	for _, entry := range exactMatches {
		sb.WriteString(fmt.Sprintf("  %s\n", formatLine(entry.IP, entry.Domain)))
	}

	if len(relevantEntries) > 0 {
		sb.WriteString("\n" + ui.BoldStyle.Render("Relevant Results (same IP or normalized domain):") + "\n")
		for _, entry := range relevantEntries {
			sb.WriteString(fmt.Sprintf("  %s\n", formatLine(entry.IP, entry.Domain)))
		}
	}

	// If there are relevant results or many matches, display in an interactive scroll view starting at the top
	if len(relevantEntries) > 0 || len(exactMatches) > 10 {
		title := fmt.Sprintf("Search Results for %q (%d exact, %d relevant)", query, len(exactMatches), len(relevantEntries))
		if err := ui.ScrollView(title, sb.String()); err != nil {
			fmt.Println(ui.ErrorMessage(fmt.Sprintf("Viewer error: %v", err)))
		}
	}

	// Print Exact Matches to terminal history so user retains key results upon exit
	fmt.Println("\n" + ui.BoldStyle.Render("Exact Matches:"))
	for _, entry := range exactMatches {
		fmt.Printf("  %s\n", formatLine(entry.IP, entry.Domain))
	}
	if len(relevantEntries) > 0 {
		fmt.Printf("\n  %s\n", ui.GrayStyle().Render(fmt.Sprintf("(%d relevant results were shown in scroll viewer)", len(relevantEntries))))
	}
	fmt.Println()
	return nil
}

// showAllEntries lists all hosts, sorted cleanly according to spec
func showAllEntries(entries []HostEntry) error {
	if len(entries) == 0 {
		hostsPath := getHostsFilePath()
		fmt.Println(ui.InfoMessage(fmt.Sprintf("No active host mappings found in %s.", hostsPath)))
		if runtime.GOOS == "windows" {
			fmt.Println(ui.GrayStyle().Render("  (By default on Windows, localhost entries are commented out with '#')"))
		}
		fmt.Println(ui.GrayStyle().Render("  Use 'Add Host' to add your first domain-IP mapping."))
		return nil
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return compareHostEntries(entries[i], entries[j])
	})

	if len(entries) > 20 {
		var sb strings.Builder
		sb.WriteString(ui.BoldStyle.Render("Current Host Entries:") + "\n\n")
		var prevIP string
		for _, entry := range entries {
			if prevIP != "" && entry.IP != prevIP {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("  %s\n", formatLine(entry.IP, entry.Domain)))
			prevIP = entry.IP
		}
		return ui.ScrollView(fmt.Sprintf("All Hosts (%d entries)", len(entries)), sb.String())
	}

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

// HostInputGroup represents an IP and its associated domains parsed from input
type HostInputGroup struct {
	IP      string
	Domains []string
}

// parseAddHostInput parses raw input text (single line, multi-line, or space-separated stream)
// into structured HostInputGroup items, along with any unassigned IPs or domains.
func parseAddHostInput(input string) (groups []HostInputGroup, unassignedIPs []string, unassignedDomains []string) {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		// Strip comments
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		var currentIP string
		var currentDomains []string

		for _, token := range fields {
			cleanToken := strings.Trim(token, `",';`)
			cleanToken = strings.TrimPrefix(cleanToken, "http://")
			cleanToken = strings.TrimPrefix(cleanToken, "https://")
			cleanToken = strings.TrimRight(cleanToken, "/")

			if cleanToken == "" {
				continue
			}

			if _, ok := parseIP(cleanToken); ok {
				if currentIP != "" {
					if len(currentDomains) > 0 {
						groups = append(groups, HostInputGroup{IP: currentIP, Domains: currentDomains})
						currentDomains = nil
					} else {
						unassignedIPs = append(unassignedIPs, currentIP)
					}
				}
				currentIP = cleanToken
			} else {
				if currentIP != "" {
					currentDomains = append(currentDomains, cleanToken)
				} else {
					unassignedDomains = append(unassignedDomains, cleanToken)
				}
			}
		}

		if currentIP != "" {
			if len(currentDomains) > 0 {
				groups = append(groups, HostInputGroup{IP: currentIP, Domains: currentDomains})
			} else {
				unassignedIPs = append(unassignedIPs, currentIP)
			}
		}
	}

	// Post-processing:
	// If no groups formed, but exactly 1 unassigned IP and some unassigned domains
	// (e.g. user typed "google.com 127.0.0.1" where domain was placed before IP)
	if len(groups) == 0 && len(unassignedIPs) == 1 && len(unassignedDomains) > 0 {
		groups = append(groups, HostInputGroup{
			IP:      unassignedIPs[0],
			Domains: unassignedDomains,
		})
		unassignedIPs = nil
		unassignedDomains = nil
	}

	// If exactly 1 group and leading unassigned domains (e.g. "foo.com 127.0.0.1 bar.com")
	if len(groups) == 1 && len(unassignedIPs) == 0 && len(unassignedDomains) > 0 {
		groups[0].Domains = append(unassignedDomains, groups[0].Domains...)
		unassignedDomains = nil
	}

	return groups, unassignedIPs, unassignedDomains
}

// addEntry adds domain-IP entries, supporting single or multi-host input, prompting for missing details and checking duplicates
func addEntry(headerComments []string, entries []HostEntry) error {
	input, err := ui.Input("Enter IP and/or Domains (e.g., 127.0.0.1 google.com, or paste multiple lines)", "")
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Println(ui.ErrorMessage("Input cannot be empty."))
		return nil
	}

	groups, unassignedIPs, unassignedDomains := parseAddHostInput(input)

	// If no groups found but only domains were entered, prompt for IP
	if len(groups) == 0 && len(unassignedIPs) == 0 && len(unassignedDomains) > 0 {
		var promptMsg string
		if len(unassignedDomains) == 1 {
			promptMsg = "Enter IP for " + unassignedDomains[0]
		} else {
			promptMsg = fmt.Sprintf("Enter IP for %d domains", len(unassignedDomains))
		}
		ipInput, err := ui.Input(promptMsg, "")
		if err != nil {
			return err
		}
		ipInput = strings.TrimSpace(ipInput)
		if ipInput == "" {
			fmt.Println(ui.ErrorMessage("IP cannot be empty."))
			return nil
		}
		if _, ok := parseIP(ipInput); !ok {
			fmt.Println(ui.ErrorMessage(fmt.Sprintf("Invalid IP address: %s", ipInput)))
			return nil
		}
		groups = append(groups, HostInputGroup{IP: ipInput, Domains: unassignedDomains})
		unassignedDomains = nil
	}

	// If there are unassigned IPs without domains, prompt for domains
	for _, ip := range unassignedIPs {
		domainInput, err := ui.Input("Enter Domains for "+ip, "")
		if err != nil {
			return err
		}
		doms := strings.Fields(domainInput)
		if len(doms) > 0 {
			groups = append(groups, HostInputGroup{IP: ip, Domains: doms})
		}
	}

	// If there are still unassigned domains left
	if len(unassignedDomains) > 0 {
		ipInput, err := ui.Input(fmt.Sprintf("Enter IP for unassigned domains (%s)", strings.Join(unassignedDomains, ", ")), "")
		if err != nil {
			return err
		}
		ipInput = strings.TrimSpace(ipInput)
		if ipInput != "" {
			if _, ok := parseIP(ipInput); ok {
				groups = append(groups, HostInputGroup{IP: ipInput, Domains: unassignedDomains})
			}
		}
	}

	if len(groups) == 0 {
		fmt.Println(ui.ErrorMessage("No valid host entries to add."))
		return nil
	}

	var addedEntries []HostEntry
	var skippedEntries []HostEntry

	for _, group := range groups {
		ipStr := group.IP
		parsedIP, _ := parseIP(ipStr)
		priority := getIPPriority(parsedIP, ipStr)

		for _, domain := range group.Domains {
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
				skippedEntries = append(skippedEntries, HostEntry{IP: ipStr, Domain: domain})
				fmt.Println(ui.WarningMessage(fmt.Sprintf("Host already exists (skipped): %s -> %s", ipStr, domain)))
			} else {
				newEntry := HostEntry{
					IP:         ipStr,
					Domain:     domain,
					Normalized: normalizeDomain(domain),
					ParsedIP:   parsedIP,
					IPPriority: priority,
				}
				addedEntries = append(addedEntries, newEntry)
				entries = append(entries, newEntry)
			}
		}
	}

	if len(addedEntries) > 0 {
		var sb strings.Builder
		sb.WriteString(ui.BoldStyle.Render("Preparing to add new host entries:") + "\n")
		header := ui.BoldStyle.Render(fmt.Sprintf("%-16s   %s", "IP", "DOMAIN"))
		sb.WriteString("  " + header + "\n")
		sb.WriteString("  " + ui.GrayStyle().Render(strings.Repeat("-", 50)) + "\n")
		var prevIP string
		for _, entry := range addedEntries {
			if entry.IP != prevIP {
				sb.WriteString(fmt.Sprintf("  %-16s   %s\n", entry.IP, entry.Domain))
			} else {
				sb.WriteString(fmt.Sprintf("  %-16s   %s\n", "", entry.Domain))
			}
			prevIP = entry.IP
		}
		sb.WriteString("\n" + ui.InfoMessage(privilegeMessage()))

		err = writeHostsFile(headerComments, entries, sb.String())
		if err != nil {
			return err
		}

		ipMap := make(map[string][]string)
		var uniqueIPs []string
		for _, entry := range addedEntries {
			if _, exists := ipMap[entry.IP]; !exists {
				uniqueIPs = append(uniqueIPs, entry.IP)
			}
			ipMap[entry.IP] = append(ipMap[entry.IP], entry.Domain)
		}

		if len(uniqueIPs) == 1 {
			ip := uniqueIPs[0]
			fmt.Println(ui.SuccessMessage(fmt.Sprintf("Successfully added domains to %s: %s", ip, strings.Join(ipMap[ip], ", "))))
		} else {
			fmt.Println(ui.SuccessMessage(fmt.Sprintf("Successfully added %d host entries across %d IPs:", len(addedEntries), len(uniqueIPs))))
			for _, ip := range uniqueIPs {
				fmt.Printf("  %s -> %s\n", ui.BoldStyle.Render(ip), strings.Join(ipMap[ip], ", "))
			}
		}
	} else if len(skippedEntries) > 0 {
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
	sb.WriteString("\n" + ui.InfoMessage(privilegeMessage()))

	err = writeHostsFile(headerComments, newEntries, sb.String())
	if err != nil {
		return err
	}

	fmt.Println(ui.SuccessMessage(fmt.Sprintf("Successfully deleted entry: %s %s", targetIP, targetDomain)))
	return nil
}

// formatHostsFile sorts and reformats the hosts file
func formatHostsFile(headerComments []string, entries []HostEntry) error {
	sort.SliceStable(entries, func(i, j int) bool {
		return compareHostEntries(entries[i], entries[j])
	})

	preview := ui.BoldStyle.Render("Preparing to format and sort hosts file.") +
		"\n\n" + ui.InfoMessage(privilegeMessage())

	err := writeHostsFile(headerComments, entries, preview)
	if err != nil {
		return err
	}

	fmt.Println(ui.SuccessMessage("Successfully sorted and formatted hosts file!"))
	return nil
}

// copyFile copies the contents from src to dst. On Windows, it clears any read-only attribute first.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		_ = os.Chmod(dst, 0666)
	}
	return os.WriteFile(dst, data, 0644)
}

// flushDNS flushes the system DNS cache so hosts file changes take effect immediately
func flushDNS() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("ipconfig", "/flushdns")
		if err := cmd.Run(); err == nil {
			fmt.Println(ui.SuccessMessage("Flushed DNS resolver cache (ipconfig /flushdns)."))
		}
	case "darwin":
		_ = exec.Command("dscacheutil", "-flushcache").Run()
		_ = exec.Command("sudo", "killall", "-HUP", "mDNSResponder").Run()
		fmt.Println(ui.SuccessMessage("Flushed macOS DNS cache."))
	case "linux":
		if err := exec.Command("resolvectl", "flush-caches").Run(); err != nil {
			_ = exec.Command("systemd-resolve", "--flush-caches").Run()
		}
		fmt.Println(ui.SuccessMessage("Flushed Linux DNS cache."))
	}
}

// writeHostsFile generates hosts output structure, writes to temp file, and overwrites the hosts file with elevated privileges
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

	hostsPath := getHostsFilePath()

	if runtime.GOOS == "windows" {
		// Attempt direct write/copy first (works if current prompt is already Administrator)
		if err := copyFile(tempPath, hostsPath); err != nil {
			// Otherwise, trigger UAC elevation prompt via PowerShell Start-Process
			psArgs := fmt.Sprintf(
				`$p = Start-Process cmd -ArgumentList @('/c', 'copy', '/y', '%s', '%s') -Verb RunAs -Wait -PassThru; exit $p.ExitCode`,
				strings.ReplaceAll(tempPath, "'", "''"),
				strings.ReplaceAll(hostsPath, "'", "''"),
			)
			cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psArgs)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to overwrite %s (requires Administrator privileges): %w", hostsPath, err)
			}
		}

		flushDNS()
		return nil
	}

	cmd := exec.Command("sudo", "cp", tempPath, hostsPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to overwrite %s (requires sudo): %w", hostsPath, err)
	}

	flushDNS()
	return nil
}
