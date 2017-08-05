package pgstorage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func DoQuery(ctx context.Context, db *sql.DB, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("Error running query=%s Err=%v", query, err)
	}

	results := make([]map[string]interface{}, 0)

	// If there aren't any rows, we return a nil result
	for rows.Next() {
		// Get the list of column names
		cols, _ := rows.Columns()
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		data := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			data[colName] = *val
		}

		results = append(results, data)
	}
	return results, nil
}

// Normalize field names. This takes a string such as "(data ->> 'created'::text)"
// and converts it to "data.created"
func normalizeFieldName(in string) string {
	if in[0] != '(' || in[len(in)-1] != ')' {
		return in
	}
	in = in[1 : len(in)-1]

	var output string

	for _, part := range strings.Split(in, " ") {
		if sepIdx := strings.Index(part, "'::"); sepIdx > -1 {
			part = part[1:sepIdx]
		}
		if part == "->>" {
			output += "."
		} else {
			output += part
		}
	}

	return output
}
