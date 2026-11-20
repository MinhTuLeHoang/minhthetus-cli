package sys

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/config"
	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

const (
	aliasConfigFile       = "aliases.json"
	aliasDefaultJSON      = `{"aliases":[],"applied_suggestions":[]}`
	aliasKindPresetMaster = "preset:master"
	aliasScrollThreshold  = 15

	menuViewSearchAliases = "View / Search Aliases"
	menuAddAlias          = "Add Alias"
	menuEditAlias         = "Edit Alias"
	menuDeleteAlias       = "Delete Alias"
	menuSuggestions       = "Suggestions"
	menuExit              = "Exit"
)

var aliasNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

// AliasEntry represents a single shell command alias managed by minhthetus-cli.
type AliasEntry struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Kind    string `json:"kind,omitempty"`
}

// AliasConfig persists managed aliases and sticky applied suggestion IDs in ~/.minhthetus-cli/aliases.json.
type AliasConfig struct {
	Aliases            []AliasEntry `json:"aliases"`
	AppliedSuggestions []string     `json:"applied_suggestions"`
}

// AliasSuggestion represents a built-in recommended alias preset.
type AliasSuggestion struct {
	ID          string
	Name        string
	Description string
	Command     string
	Kind        string
}

// AliasCmd represents the sys alias command.
var AliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage shell command aliases",
	Long: `Interactively view, search, add, edit, and delete shell command aliases,
and apply built-in suggestions across your configured shell profiles.`,
	Example: "minhthetus-cli sys alias",
	Args:    cobra.NoArgs,
	Annotations: map[string]string{
		"title": "System Alias Manager",
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"-h", "--help"}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		runAliasCLI()
	},
}

func defaultSuggestions() []AliasSuggestion {
	return []AliasSuggestion{
		{
			ID:          "master",
			Name:        "master",
			Description: "Checkout origin HEAD branch, fetch origin & pull origin",
			Command:     "git checkout <origin-HEAD> && git fetch origin && git pull origin <branch>",
			Kind:        aliasKindPresetMaster,
		},
		{
			ID:          "pull",
			Name:        "pull",
			Description: "git pull origin",
			Command:     "git pull origin",
		},
		{
			ID:          "c",
			Name:        "c",
			Description: "clear",
			Command:     "clear",
		},
	}
}

func loadAliasConfig() (AliasConfig, error) {
	var cfg AliasConfig
	if err := config.ReadFileWithDefault(aliasConfigFile, []byte(aliasDefaultJSON), &cfg); err != nil {
		return AliasConfig{}, err
	}
	if cfg.Aliases == nil {
		cfg.Aliases = []AliasEntry{}
	}
	if cfg.AppliedSuggestions == nil {
		cfg.AppliedSuggestions = []string{}
	}
	return cfg, nil
}

func saveAliasConfig(cfg AliasConfig) error {
	if cfg.Aliases == nil {
		cfg.Aliases = []AliasEntry{}
	}
	if cfg.AppliedSuggestions == nil {
		cfg.AppliedSuggestions = []string{}
	}
	return config.WriteFile(aliasConfigFile, cfg)
}

func validateAliasName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("alias name cannot be empty")
	}
	if strings.EqualFold(name, "minhthetus-cli") {
		return fmt.Errorf("alias name %q is reserved", name)
	}
	if !aliasNamePattern.MatchString(name) {
		return fmt.Errorf("invalid alias name %q: must start with a letter or underscore and contain only letters, digits, '_' or '-'", name)
	}
	return nil
}

func validateAliasCommand(command string) error {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return fmt.Errorf("alias command cannot be empty")
	}
	if strings.ContainsAny(trimmed, "\r\n") {
		return fmt.Errorf("alias command must be a single line")
	}
	return nil
}

func findAliasIndex(aliases []AliasEntry, name string) int {
	for i, a := range aliases {
		if strings.EqualFold(a.Name, name) {
			return i
		}
	}
	return -1
}

func sortedAliases(aliases []AliasEntry) []AliasEntry {
	out := make([]AliasEntry, len(aliases))
	copy(out, aliases)
	sort.Slice(out, func(i, j int) bool {
		li := strings.ToLower(out[i].Name)
		lj := strings.ToLower(out[j].Name)
		if li != lj {
			return li < lj
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func filterAliases(aliases []AliasEntry, query string) []AliasEntry {
	q := strings.ToLower(strings.TrimSpace(query))
	sorted := sortedAliases(aliases)
	if q == "" {
		return sorted
	}
	var matches []AliasEntry
	for _, a := range sorted {
		if strings.Contains(strings.ToLower(a.Name), q) || strings.Contains(strings.ToLower(a.Command), q) {
			matches = append(matches, a)
		}
	}
	return matches
}

func remainingSuggestions(applied []string) []AliasSuggestion {
	appliedSet := make(map[string]bool, len(applied))
	for _, id := range applied {
		appliedSet[strings.ToLower(strings.TrimSpace(id))] = true
	}

	var remaining []AliasSuggestion
	for _, s := range defaultSuggestions() {
		if !appliedSet[strings.ToLower(s.ID)] {
			remaining = append(remaining, s)
		}
	}
	return remaining
}

func mainMenuOptions(applied []string) []string {
	options := []string{
		menuViewSearchAliases,
		menuAddAlias,
		menuEditAlias,
		menuDeleteAlias,
	}
	if len(remainingSuggestions(applied)) > 0 {
		options = append(options, menuSuggestions)
	}
	options = append(options, menuExit)
	return options
}

func markSuggestionApplied(cfg *AliasConfig, id string) {
	for _, existing := range cfg.AppliedSuggestions {
		if strings.EqualFold(existing, id) {
			return
		}
	}
	cfg.AppliedSuggestions = append(cfg.AppliedSuggestions, id)
}

func upsertAliasEntry(cfg *AliasConfig, entry AliasEntry) {
	if idx := findAliasIndex(cfg.Aliases, entry.Name); idx >= 0 {
		cfg.Aliases[idx] = entry
		return
	}
	cfg.Aliases = append(cfg.Aliases, entry)
}

func editAliasEntry(cfg *AliasConfig, index int, newName, newCommand string) error {
	if index < 0 || index >= len(cfg.Aliases) {
		return fmt.Errorf("alias index %d out of range", index)
	}
	newName = strings.TrimSpace(newName)
	newCommand = strings.TrimSpace(newCommand)
	if err := validateAliasName(newName); err != nil {
		return err
	}
	if err := validateAliasCommand(newCommand); err != nil {
		return err
	}
	if conflict := findAliasIndex(cfg.Aliases, newName); conflict >= 0 && conflict != index {
		return fmt.Errorf("alias %q already exists", cfg.Aliases[conflict].Name)
	}

	updated := cfg.Aliases[index]
	if newName != updated.Name || newCommand != updated.Command {
		// Clear preset Kind if the user modifies the command; keep Kind only when command stays intact on rename.
		if newCommand != updated.Command {
			updated.Kind = ""
		}
	}
	updated.Name = newName
	updated.Command = newCommand
	cfg.Aliases[index] = updated
	return nil
}

func deleteAliasEntry(cfg *AliasConfig, index int) (AliasEntry, error) {
	if index < 0 || index >= len(cfg.Aliases) {
		return AliasEntry{}, fmt.Errorf("alias index %d out of range", index)
	}
	removed := cfg.Aliases[index]
	cfg.Aliases = append(cfg.Aliases[:index], cfg.Aliases[index+1:]...)
	// Note: AppliedSuggestions intentionally remains sticky after deletion.
	return removed, nil
}

func runAliasCLI() {
	mutated := false
	var lastSynced []ProfileTarget

	for {
		cfg, err := loadAliasConfig()
		if err != nil {
			fmt.Println(ui.ErrorMessage(fmt.Sprintf("Failed to load aliases config: %v", err)))
			return
		}

		options := mainMenuOptions(cfg.AppliedSuggestions)
		choice, err := ui.Choose("Select Alias Management Action", options)
		if err != nil {
			fmt.Println(ui.ErrorMessage(fmt.Sprintf("UI error: %v", err)))
			return
		}

		if choice == menuExit || choice == "" {
			break
		}

		switch choice {
		case menuViewSearchAliases:
			handleViewSearchAliases(cfg)
		case menuAddAlias:
			if targets, changed := handleAddAlias(&cfg); changed {
				mutated = true
				lastSynced = targets
			}
		case menuEditAlias:
			if targets, changed := handleEditAlias(&cfg); changed {
				mutated = true
				lastSynced = targets
			}
		case menuDeleteAlias:
			if targets, changed := handleDeleteAlias(&cfg); changed {
				mutated = true
				lastSynced = targets
			}
		case menuSuggestions:
			if targets, changed := handleSuggestions(&cfg); changed {
				mutated = true
				lastSynced = targets
			}
		}
		fmt.Println()
	}

	if mutated {
		printReloadHint(lastSynced)
	}
}

func handleViewSearchAliases(cfg AliasConfig) {
	if len(cfg.Aliases) == 0 {
		fmt.Println(ui.InfoMessage("No managed aliases configured yet."))
		fmt.Println(ui.GrayStyle().Render("  Use 'Add Alias' or 'Suggestions' to create your first alias."))
		return
	}

	query, err := ui.Input("Search aliases by name or command (leave empty to view all)", "")
	if err != nil {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Input failed: %v", err)))
		return
	}

	matches := filterAliases(cfg.Aliases, query)
	if len(matches) == 0 {
		fmt.Println(ui.InfoMessage(fmt.Sprintf("No aliases matching %q.", strings.TrimSpace(query))))
		return
	}

	var b strings.Builder
	b.WriteString(ui.BoldStyle.Render(fmt.Sprintf("  %-18s  %s", "NAME", "COMMAND")) + "\n")
	b.WriteString(ui.GrayStyle().Render("  "+strings.Repeat("─", 60)) + "\n")
	for _, a := range matches {
		nameCol := ui.CyanStyle().Render(fmt.Sprintf("%-18s", a.Name))
		b.WriteString(fmt.Sprintf("  %s  %s\n", nameCol, a.Command))
	}

	targets := resolveCurrentProfileTargets()
	if len(targets) > 0 {
		b.WriteString("\n" + ui.GrayStyle().Render("Synced profiles:") + "\n")
		for _, t := range targets {
			b.WriteString(ui.GrayStyle().Render(fmt.Sprintf("  %s %s", ui.BulletIcon, t.Path)) + "\n")
		}
	}

	if len(matches) > aliasScrollThreshold {
		if err := ui.ScrollView(fmt.Sprintf("Managed Shell Aliases (%d)", len(matches)), b.String()); err != nil {
			fmt.Println(ui.ErrorMessage(fmt.Sprintf("Failed to display aliases: %v", err)))
			return
		}
		fmt.Println(ui.InfoMessage(fmt.Sprintf("Displayed %d alias(es).", len(matches))))
		return
	}

	fmt.Println(ui.BoldStyle.Render(fmt.Sprintf("Managed Shell Aliases (%d):", len(matches))))
	fmt.Print(b.String())
}

func persistAndSync(cfg AliasConfig) ([]ProfileTarget, error) {
	if err := saveAliasConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to save aliases config: %w", err)
	}
	targets, err := syncAllProfiles(cfg.Aliases)
	if err != nil {
		return targets, fmt.Errorf("saved config, but failed to sync shell profile(s): %w", err)
	}
	return targets, nil
}

func printSyncedProfiles(targets []ProfileTarget) {
	for _, t := range targets {
		fmt.Println(ui.GrayStyle().Render(fmt.Sprintf("  Updated %s", t.Path)))
	}
}

func handleAddAlias(cfg *AliasConfig) ([]ProfileTarget, bool) {
	name, err := ui.Input("Enter alias name (e.g. gs, ll, pull; leave empty to cancel)", "")
	if err != nil {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Input failed: %v", err)))
		return nil, false
	}
	name = strings.TrimSpace(name)
	if name == "" {
		fmt.Println(ui.InfoMessage("Add alias cancelled."))
		return nil, false
	}
	if err := validateAliasName(name); err != nil {
		fmt.Println(ui.ErrorMessage(err.Error()))
		return nil, false
	}

	if idx := findAliasIndex(cfg.Aliases, name); idx >= 0 {
		existing := cfg.Aliases[idx]
		confirmed, err := ui.Confirm(
			fmt.Sprintf("Alias '%s' already exists (%s). Overwrite it?", existing.Name, existing.Command),
			0,
			false,
		)
		if err != nil || !confirmed {
			fmt.Println(ui.InfoMessage("Add alias cancelled."))
			return nil, false
		}
	}

	command, err := ui.Input(fmt.Sprintf("Enter command for alias '%s'", name), "")
	if err != nil {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Input failed: %v", err)))
		return nil, false
	}
	command = strings.TrimSpace(command)
	if command == "" {
		fmt.Println(ui.WarningMessage("Alias command cannot be empty. Cancelled."))
		return nil, false
	}
	if err := validateAliasCommand(command); err != nil {
		fmt.Println(ui.ErrorMessage(err.Error()))
		return nil, false
	}

	upsertAliasEntry(cfg, AliasEntry{
		Name:    name,
		Command: command,
	})

	targets, err := persistAndSync(*cfg)
	if err != nil {
		fmt.Println(ui.ErrorMessage(err.Error()))
		return targets, false
	}

	fmt.Println(ui.SuccessMessage(fmt.Sprintf("Added alias '%s' -> '%s'.", name, command)))
	printSyncedProfiles(targets)
	return targets, true
}

func handleEditAlias(cfg *AliasConfig) ([]ProfileTarget, bool) {
	if len(cfg.Aliases) == 0 {
		fmt.Println(ui.WarningMessage("No aliases available to edit."))
		return nil, false
	}

	sorted := sortedAliases(cfg.Aliases)
	options := make([]string, 0, len(sorted)+1)
	for _, a := range sorted {
		options = append(options, fmt.Sprintf("%-16s -> %s", a.Name, a.Command))
	}
	options = append(options, "Cancel")

	choice, err := ui.Choose("Select alias to edit", options)
	if err != nil {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("UI error: %v", err)))
		return nil, false
	}
	if choice == "" || choice == "Cancel" {
		fmt.Println(ui.InfoMessage("Edit cancelled."))
		return nil, false
	}

	selectedIdx := -1
	for i, opt := range options[:len(sorted)] {
		if opt == choice {
			selectedIdx = i
			break
		}
	}
	if selectedIdx < 0 {
		fmt.Println(ui.InfoMessage("Edit cancelled."))
		return nil, false
	}

	target := sorted[selectedIdx]
	origIdx := findAliasIndex(cfg.Aliases, target.Name)
	if origIdx < 0 {
		fmt.Println(ui.ErrorMessage("Selected alias could not be located."))
		return nil, false
	}

	newName, err := ui.Input(fmt.Sprintf("Edit alias name (current: %s)", target.Name), target.Name)
	if err != nil {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Input failed: %v", err)))
		return nil, false
	}
	newName = strings.TrimSpace(newName)
	if newName == "" {
		fmt.Println(ui.InfoMessage("Edit cancelled."))
		return nil, false
	}
	if err := validateAliasName(newName); err != nil {
		fmt.Println(ui.ErrorMessage(err.Error()))
		return nil, false
	}
	if conflict := findAliasIndex(cfg.Aliases, newName); conflict >= 0 && conflict != origIdx {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Alias '%s' already exists.", cfg.Aliases[conflict].Name)))
		return nil, false
	}

	newCommand, err := ui.Input(fmt.Sprintf("Edit command for alias '%s'", newName), target.Command)
	if err != nil {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Input failed: %v", err)))
		return nil, false
	}
	newCommand = strings.TrimSpace(newCommand)
	if newCommand == "" {
		fmt.Println(ui.InfoMessage("Edit cancelled."))
		return nil, false
	}

	if err := editAliasEntry(cfg, origIdx, newName, newCommand); err != nil {
		fmt.Println(ui.ErrorMessage(err.Error()))
		return nil, false
	}

	targets, err := persistAndSync(*cfg)
	if err != nil {
		fmt.Println(ui.ErrorMessage(err.Error()))
		return targets, false
	}

	fmt.Println(ui.SuccessMessage(fmt.Sprintf("Updated alias '%s' -> '%s'.", newName, newCommand)))
	printSyncedProfiles(targets)
	return targets, true
}

func handleDeleteAlias(cfg *AliasConfig) ([]ProfileTarget, bool) {
	if len(cfg.Aliases) == 0 {
		fmt.Println(ui.WarningMessage("No aliases available to delete."))
		return nil, false
	}

	sorted := sortedAliases(cfg.Aliases)
	options := make([]string, 0, len(sorted)+1)
	for _, a := range sorted {
		options = append(options, fmt.Sprintf("%-16s -> %s", a.Name, a.Command))
	}
	options = append(options, "Cancel")

	choice, err := ui.Choose("Select alias to delete", options)
	if err != nil {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("UI error: %v", err)))
		return nil, false
	}
	if choice == "" || choice == "Cancel" {
		fmt.Println(ui.InfoMessage("Delete cancelled."))
		return nil, false
	}

	selectedIdx := -1
	for i, opt := range options[:len(sorted)] {
		if opt == choice {
			selectedIdx = i
			break
		}
	}
	if selectedIdx < 0 {
		fmt.Println(ui.InfoMessage("Delete cancelled."))
		return nil, false
	}

	target := sorted[selectedIdx]
	confirmed, err := ui.Confirm(
		fmt.Sprintf("Are you sure you want to delete alias '%s' (%s)?", target.Name, target.Command),
		0,
		false,
	)
	if err != nil || !confirmed {
		fmt.Println(ui.InfoMessage("Delete cancelled."))
		return nil, false
	}

	origIdx := findAliasIndex(cfg.Aliases, target.Name)
	if origIdx < 0 {
		fmt.Println(ui.ErrorMessage("Selected alias could not be located."))
		return nil, false
	}

	if _, err := deleteAliasEntry(cfg, origIdx); err != nil {
		fmt.Println(ui.ErrorMessage(err.Error()))
		return nil, false
	}

	targets, err := persistAndSync(*cfg)
	if err != nil {
		fmt.Println(ui.ErrorMessage(err.Error()))
		return targets, false
	}

	fmt.Println(ui.SuccessMessage(fmt.Sprintf("Deleted alias '%s'.", target.Name)))
	printSyncedProfiles(targets)
	return targets, true
}

func handleSuggestions(cfg *AliasConfig) ([]ProfileTarget, bool) {
	remaining := remainingSuggestions(cfg.AppliedSuggestions)
	if len(remaining) == 0 {
		return nil, false
	}

	options := make([]string, 0, len(remaining))
	for _, s := range remaining {
		options = append(options, fmt.Sprintf("%-8s -> %s", s.Name, s.Description))
	}

	selections, err := ui.MultiChoose("Select suggested aliases to apply:", options, make([]bool, len(options)), 0)
	if err != nil {
		fmt.Println(ui.InfoMessage("Suggestions cancelled."))
		return nil, false
	}

	var chosen []AliasSuggestion
	for i, selected := range selections {
		if selected && i < len(remaining) {
			chosen = append(chosen, remaining[i])
		}
	}
	if len(chosen) == 0 {
		fmt.Println(ui.InfoMessage("No suggestions selected."))
		return nil, false
	}

	appliedCount := 0
	for _, s := range chosen {
		if idx := findAliasIndex(cfg.Aliases, s.Name); idx >= 0 {
			existing := cfg.Aliases[idx]
			confirmed, err := ui.Confirm(
				fmt.Sprintf("Alias '%s' already exists (%s). Overwrite with suggested command?", existing.Name, existing.Command),
				0,
				false,
			)
			if err != nil || !confirmed {
				fmt.Println(ui.InfoMessage(fmt.Sprintf("Skipped suggestion '%s'.", s.Name)))
				continue
			}
		}

		upsertAliasEntry(cfg, AliasEntry{
			Name:    s.Name,
			Command: s.Command,
			Kind:    s.Kind,
		})
		markSuggestionApplied(cfg, s.ID)
		appliedCount++
	}

	if appliedCount == 0 {
		fmt.Println(ui.InfoMessage("No suggestions applied."))
		return nil, false
	}

	targets, err := persistAndSync(*cfg)
	if err != nil {
		fmt.Println(ui.ErrorMessage(err.Error()))
		return targets, false
	}

	fmt.Println(ui.SuccessMessage(fmt.Sprintf("Applied %d suggested alias(es).", appliedCount)))
	printSyncedProfiles(targets)

	if len(remainingSuggestions(cfg.AppliedSuggestions)) == 0 {
		fmt.Println(ui.InfoMessage("All suggestions have been applied. 'Suggestions' is now hidden from the menu."))
	}
	return targets, true
}
