package questions

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"unicode"
)

func normalizeText(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, value)
	return strings.Join(strings.Fields(value), " ")
}

func hashText(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func normalizeOptions(options []OptionInput) (string, string) {
	parts := make([]string, 0, len(options))
	for _, option := range options {
		parts = append(parts, normalizeText(strings.ToUpper(option.Key))+"="+normalizeText(option.Text))
	}
	return strings.Join(parts, "\n"), hashText(strings.Join(parts, "\n"))
}

func normalizeAnswer(answer string, options []OptionInput) string {
	lookup := make(map[string]string, len(options))
	for _, option := range options {
		key := strings.ToUpper(normalizeText(option.Key))
		lookup[key] = normalizeText(option.Text)
		lookup[strings.TrimSuffix(key, ".")] = normalizeText(option.Text)
	}
	parts := strings.Split(answer, "###")
	for index, part := range parts {
		value := normalizeText(part)
		key := strings.ToUpper(strings.TrimSpace(value))
		if mapped, ok := lookup[key]; ok {
			value = mapped
		}
		parts[index] = value
	}
	if len(parts) > 1 {
		sort.Strings(parts)
	}
	return strings.Join(parts, "###")
}
