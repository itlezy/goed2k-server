package ed2ksrv

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// normalizeDisplayText keeps valid UTF-8 as-is and falls back to GB18030 for legacy Chinese clients.
func normalizeDisplayText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if utf8.ValidString(trimmed) {
		return trimmed
	}
	decoded, err := simplifiedchinese.GB18030.NewDecoder().Bytes([]byte(trimmed))
	if err != nil {
		return trimmed
	}
	decodedText := strings.TrimSpace(string(decoded))
	if decodedText == "" {
		return trimmed
	}
	return decodedText
}
