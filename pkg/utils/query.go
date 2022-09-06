package utils

import (
	"encoding/csv"
	"strings"
)

//ParseSearchQuery is responsible for splitting the query entered into appropriate search terms
// For example, if the query searched is: "blue eyed" man
// this function will return two terms, []strings{"blue eyed", "man"},
// meaning that terms between quotes are queried together.
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
