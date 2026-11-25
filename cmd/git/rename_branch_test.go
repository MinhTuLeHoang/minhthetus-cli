package git

import "testing"

func TestFormatBranchName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "converts single spaces to hyphens",
			input:    "feature/my new branch",
			expected: "feature/my-new-branch",
		},
		{
			name:     "trims surrounding spaces and collapses multiple spaces",
			input:    "  hotfix/fix   login bug  ",
			expected: "hotfix/fix-login-bug",
		},
		{
			name:     "preserves uppercase and slashes",
			input:    "feature/PCFBANK-9404 add button",
			expected: "feature/PCFBANK-9404-add-button",
		},
		{
			name:     "empty or whitespace only returns empty string",
			input:    "   \t  ",
			expected: "",
		},
		{
			name:     "no spaces remains unchanged",
			input:    "improve/git-rename-branch",
			expected: "improve/git-rename-branch",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatBranchName(tc.input); got != tc.expected {
				t.Errorf("formatBranchName(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
