package git

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/config"
	internalgit "github.com/MinhTuLeHoang/minhthetus-cli/internal/git"
	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	patchInc bool
	minorInc bool
	majorInc bool
	message  string
	reconfig bool
	pickMode bool
	prodMode bool
)

var TagDevStgCmd = &cobra.Command{
	Use:   "tag-dev-stg",
	Short: "Dynamically manage, validate, create, and independently push deployment tags.",
	Long: `Calculates the next semantic version across repository-specific tag rules (sandbox or production),
validates all target tags against both local and remote before creating any tag, and creates/pushes each tag sequentially.`,
	Example: `minhthetus-cli git tag-dev-stg
minhthetus-cli git tag-dev-stg -P -m "Hotfix release"
minhthetus-cli git tag-dev-stg -p -P
minhthetus-cli git tag-dev-stg --prod -N
minhthetus-cli git tag-dev-stg --reconfig`,
	Args: cobra.NoArgs,
	Annotations: map[string]string{
		"title": "Git Tag Dev/Stg",
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"-P", "--patch",
			"-N", "--minor",
			"-M", "--major",
			"-m", "--message",
			"-c", "--reconfig",
			"-p", "--pick",
			"--prod",
			"-h", "--help",
		}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		runTagDevStg()
	},
}

func init() {
	TagDevStgCmd.Flags().BoolVarP(&patchInc, "patch", "P", false, "Increment the patch version (e.g. 1.0.0 -> 1.0.1)")
	TagDevStgCmd.Flags().BoolVarP(&minorInc, "minor", "N", false, "Increment the minor version (e.g. 1.0.0 -> 1.1.0)")
	TagDevStgCmd.Flags().BoolVarP(&majorInc, "major", "M", false, "Increment the major version (e.g. 1.0.0 -> 2.0.0)")
	TagDevStgCmd.Flags().StringVarP(&message, "message", "m", "", "Provide a custom tag message")
	TagDevStgCmd.Flags().BoolVarP(&reconfig, "reconfig", "c", false, "Reconfigure the repository's tag prefix rule")
	TagDevStgCmd.Flags().BoolVarP(&pickMode, "pick", "p", false, "Pick a subset of prefixes from the saved rule for this run only")
	TagDevStgCmd.Flags().BoolVar(&prodMode, "prod", false, "Use the repository's production tag rule (minhthetus.tag.prod)")

	TagDevStgCmd.MarkFlagsMutuallyExclusive("patch", "minor", "major")

	TagDevStgCmd.RegisterFlagCompletionFunc("message", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		templates := []string{
			"Release v",
			"dev release: ",
			"staging release: ",
			"hotfix release: ",
			"production release: ",
		}
		return templates, cobra.ShellCompDirectiveNoFileComp
	})
}

func runTagDevStg() {
	gitClient := internalgit.DefaultClient()

	// 1. Pre-flight repository & HEAD checks
	if !gitClient.IsRepo() {
		fmt.Println(ui.ErrorMessage("Current directory is not a Git repository."))
		os.Exit(1)
	}
	if !gitClient.HasCommits() {
		fmt.Println(ui.ErrorMessage("Repository has no commits yet (unborn HEAD). Create an initial commit first."))
		os.Exit(1)
	}
	currentBranch, err := gitClient.CurrentBranch()
	if err != nil {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Failed to determine current branch: %v", err)))
		os.Exit(1)
	}

	envBadge := ui.CyanStyle().Bold(true).Render(EnvLabel(prodMode))
	if prodMode {
		envBadge = ui.PurpleStyle().Bold(true).Render(EnvLabel(prodMode))
	}
	fmt.Printf("%s (%s)\n", ui.InfoMessage("Starting tag workflow"), envBadge)

	// 2. Check origin remote & fetch latest tags
	hasOrigin := gitClient.HasRemote("origin")
	if hasOrigin {
		fetchErr := ui.RunWithSpinner("Fetching latest tags from origin...", func() error {
			return gitClient.FetchTags()
		})
		if fetchErr != nil {
			fmt.Println(ui.WarningMessage(fmt.Sprintf("Could not fetch tags from origin: %v", fetchErr)))
		}
	} else {
		fmt.Println(ui.WarningMessage("No 'origin' remote configured. Running in local-only mode."))
	}

	// 3. Read all local tags
	localTags, err := gitClient.ListTags("")
	if err != nil {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Failed to list local tags: %v", err)))
		os.Exit(1)
	}

	// 4. Resolve or configure repository tag prefix rule
	configKey := ConfigKeyForEnv(prodMode)
	rawRule, err := gitClient.GetLocalConfig(configKey)
	if err != nil {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Failed to read local git config (%s): %v", configKey, err)))
		os.Exit(1)
	}

	rulePrefixes := ParsePrefixRule(rawRule)
	if !reconfig && len(rulePrefixes) > 0 {
		if vErr := ValidatePrefixRule(rulePrefixes); vErr != nil {
			fmt.Println(ui.ErrorMessage(fmt.Sprintf("Invalid tag prefix rule in local git config (%s=%q): %v. Run with -c/--reconfig to update it.", configKey, rawRule, vErr)))
			os.Exit(1)
		}
	}
	if reconfig || len(rulePrefixes) == 0 {
		configured, ok, cfgErr := configureRepoRule(gitClient, configKey, localTags, rulePrefixes, prodMode)
		if cfgErr != nil {
			if errors.Is(cfgErr, ui.ErrNonInteractiveTTY) {
				fmt.Println(ui.ErrorMessage(fmt.Sprintf("Cannot configure %s tag rule interactively without a terminal (TTY). Configure it via `git config --local %s <prefixes>` or run in an interactive terminal.", EnvLabel(prodMode), configKey)))
			} else {
				fmt.Println(ui.ErrorMessage(fmt.Sprintf("Failed to configure tag rule: %v", cfgErr)))
			}
			os.Exit(1)
		}
		if !ok {
			return
		}
		rulePrefixes = configured
	}

	// 5. Compute base version across all prefixes in the active rule
	details, baseVer := ComputeBaseVersion(localTags, rulePrefixes)

	formattedRule := make([]string, 0, len(rulePrefixes))
	for _, p := range rulePrefixes {
		formattedRule = append(formattedRule, p+"*")
	}

	fmt.Println("")
	fmt.Printf("%s %s [%s] %s\n",
		ui.BlueStyle().Render(strings.TrimSpace(ui.InfoIcon)),
		ui.BoldStyle.Render("Project Rule:"),
		ui.CyanStyle().Render(strings.Join(formattedRule, ", ")),
		ui.GrayStyle().Render("(use -c/--reconfig to change)"),
	)
	for _, d := range details {
		paddedPrefix := ui.CyanStyle().Render(fmt.Sprintf("%-12s", d.Prefix+"*:"))
		if d.Found {
			fmt.Printf("%s Latest %s %s (%s)\n",
				ui.TagIcon,
				paddedPrefix,
				ui.YellowStyle().Render(d.LatestTag),
				ui.GrayStyle().Render(d.Version.String()),
			)
		} else {
			fmt.Printf("%s Latest %s %s\n",
				ui.TagIcon,
				paddedPrefix,
				ui.GrayStyle().Render("None"),
			)
		}
	}
	fmt.Printf("%s %s %s\n",
		ui.BlueStyle().Render(strings.TrimSpace(ui.InfoIcon)),
		ui.BoldStyle.Render("Base version:"),
		ui.GreenStyle().Bold(true).Render(baseVer.String()),
	)

	// 6. Apply --pick subset selection (this run only, never modifies .git/config)
	activePrefixes := rulePrefixes
	if pickMode {
		if len(rulePrefixes) == 1 {
			fmt.Println(ui.InfoMessage(fmt.Sprintf("Active rule has only 1 prefix (%s); using it directly.", rulePrefixes[0]+"*")))
		} else {
			pickOpts := make([]string, 0, len(rulePrefixes))
			for _, p := range rulePrefixes {
				pickOpts = append(pickOpts, p+"*")
			}
			fmt.Println("")
			pickedFlags, pickErr := ui.MultiChoose("Select prefixes to tag for this run:", pickOpts, make([]bool, len(pickOpts)), 0)
			if pickErr != nil {
				if errors.Is(pickErr, ui.ErrSelectionCanceled) {
					fmt.Println(ui.WarningMessage("Prefix selection cancelled."))
					return
				}
				if errors.Is(pickErr, ui.ErrNonInteractiveTTY) {
					fmt.Println(ui.ErrorMessage("Cannot use -p/--pick interactively without a terminal (TTY). Omit -p/--pick to tag all configured prefixes or run in an interactive terminal."))
				} else {
					fmt.Println(ui.ErrorMessage(fmt.Sprintf("Failed to select prefixes: %v", pickErr)))
				}
				os.Exit(1)
			}
			subset := make([]string, 0, len(rulePrefixes))
			for i, checked := range pickedFlags {
				if checked {
					subset = append(subset, rulePrefixes[i])
				}
			}
			if len(subset) == 0 {
				fmt.Println(ui.WarningMessage("No prefixes selected for this run. Cancelling."))
				return
			}
			activePrefixes = subset
		}
	}

	// 7. Determine target version (flags or interactive bump + custom semver)
	newVer, incType, ok, verErr := resolveTargetVersion(baseVer)
	if verErr != nil {
		if errors.Is(verErr, ui.ErrNonInteractiveTTY) {
			fmt.Println(ui.ErrorMessage("Cannot prompt for version bump without a terminal (TTY). Pass -P/--patch, -N/--minor, or -M/--major."))
		} else {
			fmt.Println(ui.ErrorMessage(fmt.Sprintf("Failed to determine target version: %v", verErr)))
		}
		os.Exit(1)
	}
	if !ok {
		return
	}

	fmt.Println("")
	fmt.Printf("%s Incrementing version (%s)...\n", ui.HammerIcon, ui.BoldStyle.Render(incType))
	fmt.Printf("%s Target version: %s\n", ui.RocketIcon, ui.PurpleStyle().Bold(true).Render(newVer.String()))

	// 8. Resolve tag message
	tagMessage := strings.TrimSpace(message)
	isCustomMsg := tagMessage != ""
	if !isCustomMsg {
		entered, inputErr := ui.Input(fmt.Sprintf("Enter tag message (default: Release v%s)", newVer.String()), "")
		if errors.Is(inputErr, ui.ErrInputCanceled) {
			fmt.Println(ui.WarningMessage("Tagging cancelled."))
			return
		}
		if inputErr != nil {
			if errors.Is(inputErr, ui.ErrNonInteractiveTTY) {
				fmt.Println(ui.ErrorMessage("Cannot prompt for tag message without a terminal (TTY). Pass -m/--message \"<msg>\"."))
			} else {
				fmt.Println(ui.ErrorMessage(fmt.Sprintf("Failed to read tag message: %v", inputErr)))
			}
			os.Exit(1)
		}
		entered = strings.TrimSpace(entered)
		if entered != "" {
			tagMessage = entered
			isCustomMsg = true
		} else {
			tagMessage = "Release v" + newVer.String()
		}
	}

	// 9. Atomic pre-flight check across ALL target tags on BOTH local and remote
	targetTags := BuildTargetTags(activePrefixes, newVer)

	var remoteTags map[string]bool
	if hasOrigin {
		remoteErr := ui.RunWithSpinner("Checking remote tags on origin...", func() error {
			var listErr error
			remoteTags, listErr = gitClient.ListRemoteTags()
			return listErr
		})
		if remoteErr != nil {
			fmt.Println(ui.ErrorMessage(fmt.Sprintf("Pre-flight failed while querying origin remote: %v", remoteErr)))
			os.Exit(1)
		}
	}

	statuses := CheckTagAvailability(targetTags, gitClient.TagExistsLocally, remoteTags)
	if !HasAnyConflict(statuses) {
		fmt.Println(ui.SuccessMessage(fmt.Sprintf("Pre-flight check passed: %s", strings.Join(targetTags, ", "))))
	} else {
		fmt.Println("")
		fmt.Println(ui.ErrorMessage("Pre-flight check blocked! One or more target tags already exist:"))
		for _, s := range statuses {
			if s.HasConflict() {
				fmt.Printf("  %s %s — %s\n",
					ui.BulletIcon,
					ui.RedStyle().Bold(true).Render(s.Tag),
					ui.RedStyle().Render("exists on "+s.StatusLabel()),
				)
			} else {
				fmt.Printf("  %s %s — %s\n",
					ui.BulletIcon,
					ui.GreenStyle().Render(s.Tag),
					ui.GreenStyle().Render("available"),
				)
			}
		}

		missing := MissingTagsOnly(statuses)
		nextVer, nextTags, nextErr := NextAvailablePatchVersion(activePrefixes, newVer, gitClient.TagExistsLocally, remoteTags)

		recoveryOpts := make([]string, 0, 3)
		if nextErr == nil {
			recoveryOpts = append(recoveryOpts, fmt.Sprintf("Auto-increment patch (+1) -> %s (%s)", nextVer.String(), strings.Join(nextTags, ", ")))
		}
		if len(missing) > 0 {
			recoveryOpts = append(recoveryOpts, fmt.Sprintf("Create & push missing tags only (%s)", strings.Join(missing, ", ")))
		}
		recoveryOpts = append(recoveryOpts, "Cancel")

		fmt.Println("")
		recoveryChoice, chooseErr := ui.ChooseWithDefault("Resolve tag conflict:", recoveryOpts, 0)
		if chooseErr != nil {
			if errors.Is(chooseErr, ui.ErrNonInteractiveTTY) {
				fmt.Println(ui.ErrorMessage("Pre-flight tag conflict cannot be resolved interactively without a terminal (TTY). Resolve the conflicting tags or run in an interactive terminal."))
			} else {
				fmt.Println(ui.ErrorMessage(fmt.Sprintf("Failed to resolve tag conflict: %v", chooseErr)))
			}
			os.Exit(1)
		}
		if recoveryChoice == "" || recoveryChoice == "Cancel" {
			fmt.Println(ui.WarningMessage("Tagging cancelled. No tags were created."))
			return
		}

		if strings.HasPrefix(recoveryChoice, "Auto-increment") {
			newVer = nextVer
			targetTags = nextTags
			if !isCustomMsg {
				tagMessage = "Release v" + newVer.String()
			}
		} else if strings.HasPrefix(recoveryChoice, "Create & push missing") {
			targetTags = missing
		}
	}

	// 10. Sequential independent tag & push loop
	branchDisplay := currentBranch
	if currentBranch == "HEAD" {
		branchDisplay = "HEAD (detached)"
	}
	fmt.Printf("\n%s %s %s\n",
		ui.BlueStyle().Render(strings.TrimSpace(ui.InfoIcon)),
		ui.BoldStyle.Render("Current branch:"),
		ui.YellowStyle().Render(branchDisplay),
	)

	succeeded := make([]string, 0, len(targetTags))
	for i, tag := range targetTags {
		if err := gitClient.CreateAnnotatedTag(tag, "HEAD", tagMessage); err != nil {
			printPartialFailure(succeeded, tag, false, err, targetTags[i+1:])
			os.Exit(1)
		}

		if hasOrigin {
			pushErr := ui.RunWithSpinner(fmt.Sprintf("Pushing tag %s to origin...", tag), func() error {
				return gitClient.PushTag(tag)
			})
			if pushErr != nil {
				printPartialFailure(succeeded, tag, true, pushErr, targetTags[i+1:])
				os.Exit(1)
			}
			succeeded = append(succeeded, tag)
			fmt.Printf("%s %s %s\n",
				ui.GreenStyle().Render(ui.CheckIcon),
				ui.GreenStyle().Render("Created & Pushed:"),
				ui.BoldStyle.Render(tag),
			)
		} else {
			succeeded = append(succeeded, tag)
			fmt.Printf("%s %s %s\n",
				ui.GreenStyle().Render(ui.CheckIcon),
				ui.GreenStyle().Render("Created locally:"),
				ui.BoldStyle.Render(tag),
			)
		}
	}

	// 11. Final summary
	fmt.Println("")
	if hasOrigin {
		fmt.Println(ui.SuccessMessage(fmt.Sprintf("Successfully created and pushed %d tag(s) for version %s!", len(succeeded), newVer.String())))
	} else {
		fmt.Println(ui.SuccessMessage(fmt.Sprintf("Successfully created %d local tag(s) for version %s!", len(succeeded), newVer.String())))
	}
	for _, tag := range succeeded {
		fmt.Printf("  %s %s\n", ui.BlueStyle().Render(ui.TagIcon), ui.YellowStyle().Render(tag))
	}
}

func configureRepoRule(gitClient internalgit.Client, configKey string, localTags []string, existingRule []string, isProd bool) ([]string, bool, error) {
	validExisting := make([]string, 0, len(existingRule))
	for _, p := range existingRule {
		if ValidatePrefix(p) == nil {
			validExisting = append(validExisting, p)
		} else {
			fmt.Println(ui.WarningMessage(fmt.Sprintf("Ignoring invalid saved prefix %q during reconfiguration.", p)))
		}
	}

	detected := DiscoverTagPrefixes(localTags)
	presets, loadErr := config.LoadTagPresets()
	if loadErr != nil {
		fmt.Println(ui.WarningMessage(fmt.Sprintf("Could not read global tag presets (%v); using defaults.", loadErr)))
	}

	customLabel := fmt.Sprintf("%s Enter custom prefix...", ui.PlusIcon)
	menuItems := BuildDiscoveryMenuItems(detected, presets, validExisting, isProd, customLabel)

	labels := make([]string, 0, len(menuItems))
	preselected := make([]bool, 0, len(menuItems))
	for _, item := range menuItems {
		labels = append(labels, item.Label)
		preselected = append(preselected, item.Preselected)
	}

	var title string
	if len(validExisting) > 0 {
		title = fmt.Sprintf("%s Reconfigure %s tag prefixes:", ui.SearchIcon, EnvLabel(isProd))
	} else if len(detected) > 0 {
		title = fmt.Sprintf("%s No %s rule configured. Detected prefixes in repo:", ui.SearchIcon, EnvLabel(isProd))
	} else {
		title = fmt.Sprintf("%s No %s rule configured. Select tag prefixes:", ui.SearchIcon, EnvLabel(isProd))
	}

	fmt.Println("")
	selectedFlags, err := ui.MultiChoose(title, labels, preselected, 0)
	if err != nil {
		if errors.Is(err, ui.ErrSelectionCanceled) {
			fmt.Println(ui.WarningMessage("Tag rule configuration cancelled."))
			return nil, false, nil
		}
		return nil, false, err
	}

	chosenPrefixes := make([]string, 0, len(menuItems))
	wantsCustom := false
	for i, checked := range selectedFlags {
		if !checked {
			continue
		}
		if menuItems[i].IsCustom {
			wantsCustom = true
			continue
		}
		chosenPrefixes = append(chosenPrefixes, menuItems[i].Prefix)
	}

	if wantsCustom {
		customPrefixes, ok, customErr := promptCustomPrefixes()
		if customErr != nil {
			return nil, false, customErr
		}
		if !ok {
			if len(chosenPrefixes) == 0 {
				fmt.Println(ui.WarningMessage("Custom prefix input cancelled."))
				return nil, false, nil
			}
			fmt.Println(ui.WarningMessage("Custom prefix input cancelled; continuing with selected presets."))
		} else {
			_, learned, learnErr := config.LearnTagPresets(isProd, customPrefixes...)
			if learnErr != nil {
				fmt.Println(ui.WarningMessage(fmt.Sprintf("Could not persist learned presets: %v", learnErr)))
			} else if learned {
				fmt.Println(ui.InfoMessage(fmt.Sprintf("Learned new preset(s) into ~/.minhthetus-cli/%s: %s", config.TagPresetsFilename, strings.Join(customPrefixes, ", "))))
			}
			chosenPrefixes = append(chosenPrefixes, customPrefixes...)
		}
	}

	normalized := ParsePrefixRule(strings.Join(chosenPrefixes, ","))
	if len(normalized) == 0 {
		fmt.Println(ui.WarningMessage("No tag prefixes selected. Configuration cancelled."))
		return nil, false, nil
	}
	if vErr := ValidatePrefixRule(normalized); vErr != nil {
		return nil, false, vErr
	}

	ruleCSV := FormatPrefixRule(normalized)
	if err := gitClient.SetLocalConfig(configKey, ruleCSV); err != nil {
		return nil, false, err
	}

	fmt.Println(ui.SuccessMessage(fmt.Sprintf("Saved %s rule (%s = %s)", EnvLabel(isProd), configKey, ruleCSV)))
	return normalized, true, nil
}

func promptCustomPrefixes() ([]string, bool, error) {
	for {
		raw, err := ui.Input("Enter custom tag prefix(es), comma-separated (e.g. uat-v)", "")
		if errors.Is(err, ui.ErrInputCanceled) {
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}

		parsed := ParsePrefixRule(raw)
		if len(parsed) == 0 {
			fmt.Println(ui.ErrorMessage("Prefix cannot be empty. Press Esc to cancel."))
			continue
		}

		valid := true
		for _, p := range parsed {
			if vErr := ValidatePrefix(p); vErr != nil {
				fmt.Println(ui.ErrorMessage(vErr.Error()))
				valid = false
				break
			}
		}
		if valid {
			return parsed, true, nil
		}
	}
}

func resolveTargetVersion(baseVer SemVer) (SemVer, string, bool, error) {
	if patchInc {
		v, _ := baseVer.Bump("patch")
		return v, "patch", true, nil
	}
	if minorInc {
		v, _ := baseVer.Bump("minor")
		return v, "minor", true, nil
	}
	if majorInc {
		v, _ := baseVer.Bump("major")
		return v, "major", true, nil
	}

	majorVer, _ := baseVer.Bump("major")
	minorVer, _ := baseVer.Bump("minor")
	patchVer, _ := baseVer.Bump("patch")

	options := []string{
		fmt.Sprintf("major (%s)", majorVer.String()),
		fmt.Sprintf("minor (%s)", minorVer.String()),
		fmt.Sprintf("patch (%s)", patchVer.String()),
		"custom (enter version manually)",
	}

	fmt.Println("")
	choice, err := ui.ChooseWithDefault("Select version bump type:", options, 2)
	if err != nil {
		return SemVer{}, "", false, err
	}
	if choice == "" {
		fmt.Println(ui.WarningMessage("Version selection cancelled."))
		return SemVer{}, "", false, nil
	}

	switch {
	case strings.HasPrefix(choice, "major"):
		return majorVer, "major", true, nil
	case strings.HasPrefix(choice, "minor"):
		return minorVer, "minor", true, nil
	case strings.HasPrefix(choice, "patch"):
		return patchVer, "patch", true, nil
	default:
		for {
			raw, inputErr := ui.Input("Enter custom semantic version (e.g. 1.2.3 or v1.2.3)", "")
			if errors.Is(inputErr, ui.ErrInputCanceled) {
				fmt.Println(ui.WarningMessage("Custom version input cancelled."))
				return SemVer{}, "", false, nil
			}
			if inputErr != nil {
				return SemVer{}, "", false, inputErr
			}
			parsed, parseErr := ParseUserSemVer(raw)
			if parseErr != nil {
				fmt.Println(ui.ErrorMessage(parseErr.Error()))
				continue
			}
			return parsed, "custom", true, nil
		}
	}
}

func printPartialFailure(succeeded []string, failedTag string, createdLocally bool, err error, remaining []string) {
	fmt.Println("")
	if createdLocally {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Failed to push tag %s to origin: %v", failedTag, err)))
	} else {
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Failed to create tag %s: %v", failedTag, err)))
	}

	if len(succeeded) > 0 {
		fmt.Printf("  %s %s %s\n", ui.GreenStyle().Render(ui.CheckIcon), ui.BoldStyle.Render("Succeeded:"), strings.Join(succeeded, ", "))
	}
	if createdLocally {
		fmt.Printf("  %s %s %s (hint: retry with `git push origin refs/tags/%s` or delete with `git tag -d %s`)\n",
			ui.YellowStyle().Render(ui.WarningIcon),
			ui.BoldStyle.Render("Created locally only:"),
			ui.YellowStyle().Render(failedTag),
			failedTag,
			failedTag,
		)
	}
	if len(remaining) > 0 {
		fmt.Printf("  %s %s %s\n", ui.GrayStyle().Render(ui.BulletIcon), ui.BoldStyle.Render("Not created:"), strings.Join(remaining, ", "))
	}
}
