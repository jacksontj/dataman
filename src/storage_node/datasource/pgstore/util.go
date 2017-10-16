package pgstorage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"strings"

	"github.com/jacksontj/dataman/src/datamantype"
)

func DoQuery(ctx context.Context, db *sql.DB, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("Error running query: Err=%v query=%s ", err, query)
	}

	results := make([]map[string]interface{}, 0)

	// Get the list of column names
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	// If there aren't any rows, we return a nil result
	for rows.Next() {
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

func serializeValue(t datamantype.DatamanType, v interface{}) (string, error) {
	switch t {
	case datamantype.DateTime:
		return fmt.Sprintf("'%v'", v.(time.Time).Format(datamantype.DateTimeFormatStr)), nil
	default:
		return fmt.Sprintf("'%v'", v), nil
	}
}

// Take a path to an object and convert it to postgres json addressing
func collectionFieldToSelector(path []string) string {
	switch len(path) {
	case 1:
		return path[0]
	case 2:
		return path[0] + "->>'" + path[1] + "'"
	default:
		fieldChain := path[1:len(path)]
		return path[0] + "->'" + strings.Join(fieldChain[:len(fieldChain)-1], "'->'") + "'->>'" + path[len(path)-1] + "'"
	}

}
