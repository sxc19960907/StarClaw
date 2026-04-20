package audit

import (
	"regexp"
)

// redactPatterns compiled once at package init
var redactPatterns []*regexp.Regexp

func init() {
	patterns := []string{
		// AWS access key IDs
		`AKIA[0-9A-Z]{16}`,
		// JWT tokens (header.payload.signature format)
		`eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*`,
		// sk- style API keys (OpenAI, Stripe, etc.)
		`sk-[a-zA-Z0-9]{20,}`,
		// key- style API keys
		`key-[a-zA-Z0-9]{20,}`,
		// Bearer tokens
		`Bearer\s+[A-Za-z0-9_\-\.]+`,
		// PEM content markers
		`-----BEGIN[A-Z\s]*-----`,
		`-----END[A-Z\s]*-----`,
		// Env var assignments with secret-like names
		`(?i)[A-Z_]*(?:KEY|SECRET|TOKEN|PASSWORD)\s*=\s*\S+`,
		// GitHub tokens (ghp_, gho_, ghs_, etc.)
		`gh[pousr]_[A-Za-z0-9]{36,}`,
		// Generic API key patterns
		`(?i)api[_-]?key["\s]*[:=]["\s]*[A-Za-z0-9]{16,}`,
	}

	for _, p := range patterns {
		if re, err := regexp.Compile(p); err == nil {
			redactPatterns = append(redactPatterns, re)
		}
	}
}

// RedactSecrets replaces known secret patterns with [REDACTED]
func RedactSecrets(text string) string {
	result := text
	for _, re := range redactPatterns {
		result = re.ReplaceAllString(result, "[REDACTED]")
	}
	return result
}
