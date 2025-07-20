package storage

import (
	"fmt"
	"strings"
)

type SelectBuilder func(table string, columns []string, filters map[string]string) string

var BuildSelect SelectBuilder = func(table string, columns []string, filters map[string]string) string {
	cols := "*"
	if len(columns) > 0 {
		cols = strings.Join(columns, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", cols, table)

	if len(filters) > 0 {
		conditions := make([]string, 0, len(filters))
		for column, value := range filters {
			conditions = append(conditions, fmt.Sprintf("%s = '%s'", column, value))
		}
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += ";"
	return query
}
