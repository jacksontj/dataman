package pgstorage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"strings"

	"github.com/jacksontj/dataman/datamantype"
	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/stream"
	"github.com/jacksontj/dataman/stream/local"
)

func DoQuery(ctx context.Context, db *sql.DB, query string, colAddrs []ColAddr, args ...interface{}) ([]record.Record, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("Error running query: Err=%v query=%s ", err, query)
	}

	results := make([]record.Record, 0)

	// Get the list of column names
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	columns := make([]interface{}, len(cols))
	columnPointers := make([]interface{}, len(cols))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}

	// If there aren't any rows, we return a nil result
	for rows.Next() {
		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			rows.Close()
			return nil, err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		data := make(record.Record)
		skipN := 0
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			if colAddrs != nil {
				if colAddrs[i].skipN > 0 {
					if *val != true {
						skipN = colAddrs[i].skipN
					} else {
						skipN = 0
					}
				} else {
					if skipN <= 0 {
						data.Set(colAddrs[i].key, *val)
					} else {
						skipN--
					}
				}
			} else {
				data[colName] = *val
			}
		}
		results = append(results, data)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func DoStreamQuery(ctx context.Context, db *sql.DB, query string, colAddrs []ColAddr, args ...interface{}) (stream.ClientStream, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("Error running query: Err=%v query=%s ", err, query)
	}

	resultsChan := make(chan record.Record, 100)
	errorChan := make(chan error, 1)

	serverStream := local.NewServerStream(ctx, resultsChan, errorChan)
	clientStream := local.NewClientStream(ctx, resultsChan, errorChan)

	// TODO: without goroutine?
	go func() {
		defer serverStream.Close()
		// Get the list of column names
		cols, err := rows.Columns()
		if err != nil {
			serverStream.SendError(err)
			return
		}
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// If there aren't any rows, we return a nil result
		for rows.Next() {
			// Scan the result into the column pointers...
			if err := rows.Scan(columnPointers...); err != nil {
				rows.Close()
				serverStream.SendError(err)
				return
			}

			// Create our map, and retrieve the value for each column from the pointers slice,
			// storing it in the map with the name of the column as the key.
			data := make(record.Record)
			skipN := 0
			for i, colName := range cols {
				val := columnPointers[i].(*interface{})
				if colAddrs != nil {
					if colAddrs[i].skipN > 0 {
						// if we didn't find the key in the selector, then we skipN
						// this accounts for nil and false return types
						if *val != true {
							skipN = colAddrs[i].skipN
						} else {
							skipN = 0
						}
					} else {
						if skipN <= 0 {
							data.Set(colAddrs[i].key, *val)
						} else {
							skipN--
						}
					}
				} else {
					data[colName] = *val
				}
			}
			serverStream.SendResult(data)
		}

		if err := rows.Err(); err != nil {
			serverStream.SendError(err)
			return
		}
	}()

	return clientStream, nil
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

// TODO: remove?
func serializeValue(v interface{}) (string, error) {
	switch vTyped := v.(type) {
	case time.Time:
		return fmt.Sprintf("'%v'", vTyped.Format(datamantype.DateTimeFormatStr)), nil
	case map[string]interface{}:
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("'%s'", string(b)), nil
	case map[string]string:
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("'%s'", string(b)), nil
	default:
		return fmt.Sprintf("'%v'", v), nil
	}
}

// TODO: remove?
func serializeValueBuilder(builder *strings.Builder, v interface{}) error {
	switch vTyped := v.(type) {
	case time.Time:
		fmt.Fprintf(builder, "'%v'", vTyped.Format(datamantype.DateTimeFormatStr))
		return nil
	case map[string]interface{}:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		fmt.Fprintf(builder, "'%s'", string(b))
		return nil
	case map[string]string:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		fmt.Fprintf(builder, "'%s'", string(b))
		return nil
	default:
		fmt.Fprintf(builder, "'%v'", v)
		return nil
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
		fieldChain := path[1:]
		return path[0] + "->'" + strings.Join(fieldChain[:len(fieldChain)-1], "'->'") + "'->>'" + path[len(path)-1] + "'"
	}
}

// TODO: remove? or consolidate?
// When we want to do existence checks ( top->'level'->'key' ? 'subkey' we can't use the
// ->> selector since it will return "text" (seemingly the actual value) whereas -> returns
// a map-like object with which we can do selection and ? checks on.
func collectionFieldParentToSelector(path []string) string {
	switch len(path) {
	case 1:
		return path[0]
	case 2:
		return path[0] + "->'" + path[1] + "'"
	default:
		fieldChain := path[1:]
		return path[0] + "->'" + strings.Join(fieldChain[:len(fieldChain)-1], "'->'") + "'->'" + path[len(path)-1] + "'"
	}
}

// ColAddr is a list of addresses of columns
type ColAddr struct {
	key []string
	// Number of columns this is a "selector" for. This is used for jsonb columns
	// so we can differentiate between nil meaning the value in the json is null
	// and the field not existing in the JSON
	// is this a `?` selector telling us whether or not to skip the next one
	skipN int
}

// selectFields returns a SELECT string and the corresponding ColAddr
func selectFields(fields []string) (string, []ColAddr) {
	// TODO: remove?
	// If no projection, then just return all
	if fields == nil {
		return "*", nil
	}

	fieldSelectors := make([]string, 0, len(fields))
	cAddrs := make([]ColAddr, 0, len(fields))
	for _, field := range fields {
		fieldParts := strings.Split(field, ".")
		if len(fieldParts) > 1 {
			cAddrs = append(cAddrs, ColAddr{skipN: 1})
			fieldSelectors = append(fieldSelectors, collectionFieldParentToSelector(fieldParts[:len(fieldParts)-1])+" ? '"+fieldParts[len(fieldParts)-1]+"'")
		}
		cAddrs = append(cAddrs, ColAddr{
			key: fieldParts,
		})
		fieldSelectors = append(fieldSelectors, collectionFieldToSelector(fieldParts))

	}

	return strings.Join(fieldSelectors, ","), cAddrs
}
