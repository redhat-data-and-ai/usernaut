package logger

import "strings"

// MaskEmail masks an email address for safe logging.
// "user@example.com" becomes "us***@example.com".
// If the input has no @ sign it masks the string directly.
func MaskEmail(email string) string {
	if email == "" {
		return ""
	}

	at := strings.LastIndex(email, "@")
	if at < 0 {
		return maskLocal(email)
	}

	return maskLocal(email[:at]) + email[at:]
}

func maskLocal(local string) string {
	if len(local) <= 2 {
		return local + "***"
	}
	return local[:2] + "***"
}
