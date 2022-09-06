package utils

import (
	"encoding/csv"
	"strings"
)

func ParseSearchQuery(query string) []string {
	query = strings.ToLower(query)
	r := csv.NewReader(strings.NewReader(query))
	r.Comma = ' ' // space
	fields, _ := r.Read()
	formatted := []string{}
	for _, field := range fields {
		formatted = append(formatted, strings.ReplaceAll(field, "\"", ""))
	}
	return formatted
}
