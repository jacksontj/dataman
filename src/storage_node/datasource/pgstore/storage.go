package pgstorage

// TODO: real escaping of the various queries (sql injection is bad ;) )
// TODO: look into codegen or something for queries (terribly inefficient right now)

/*
This is a storagenode using postgres as a json document store

Metadata about the storage node will be stored in a database called _dataman.storagenode

*/

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
	"github.com/jacksontj/dataman/src/storage_node/metadata/filter"
	_ "github.com/lib/pq"
)

type StorageConfig struct {
	// How to connect to postgres
	PGString string `yaml:"pg_string"`
}

func (c *StorageConfig) pgStringForDB(name string) string {
	return c.PGString + fmt.Sprintf(" database=%s", name)
}

type Storage struct {
	config *StorageConfig
	// Connection to main db?
	db *sql.DB

	// TODO: lazily load these, maybe even pool them
	dbMap map[string]*sql.DB

	metaFunc metadata.MetaFunc
}

func (s *Storage) Init(metaFunc metadata.MetaFunc, c map[string]interface{}) error {
	var err error

	if val, ok := c["pg_string"]; ok {
		s.config = &StorageConfig{val.(string)}
	} else {
		return fmt.Errorf("Invalid config")
	}

	// TODO: pass in a database name for the metadata store locally
	// TODO: don't require the name to be set here-- because people might use their own metadata
	s.db, err = sql.Open("postgres", s.config.pgStringForDB(""))
	if err != nil {
		return err
	}

	s.dbMap = make(map[string]*sql.DB)

	s.metaFunc = metaFunc

	// TODO: ensure that the metadata store exists (and the schema is correct)
	return nil
}

// TODO: some TTL thing instead of this
func (s *Storage) getDB(name string) *sql.DB {
	if db, ok := s.dbMap[name]; ok {
		return db
	} else {
		dbConn, err := sql.Open("postgres", s.config.pgStringForDB(name))
		if err != nil {
			// TODO: not this, this is not okay :/
			fmt.Printf("Err opening postgres conn: %v", err)
			return nil
		}
		s.dbMap[name] = dbConn
		return dbConn
	}
}

// TODO: remove
func (s *Storage) GetMeta() *metadata.Meta {
	return s.metaFunc()
}

// Do a single item get
func (s *Storage) Get(args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	// TODO: figure out how to do cross-db queries? Seems that most golang drivers
	// don't support it (new in postgres 7.3)

	selectQuery := fmt.Sprintf("SELECT * FROM \"%s\".%s WHERE _id=%v", args["shard_instance"].(string), args["collection"], args["_id"])
	var err error
	result.Return, err = DoQuery(s.getDB(args["db"].(string)), selectQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	s.normalizeResult(args, result)

	// TODO: error if there is more than one result
	return result
}

// TODO: move this up to the storage node layer (since this is datasource agnostic)
func (s *Storage) Set(args query.QueryArgs) *query.Result {
	record := args["record"]
	if id, ok := record.(map[string]interface{})["_id"]; ok {
		args["filter"] = map[string]interface{}{"_id": []interface{}{filter.Equal, id}}
		delete(record.(map[string]interface{}), "_id")

		return s.Update(args)
	} else {
		return s.Insert(args)
	}
}

func (s *Storage) Insert(args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	meta := s.GetMeta()
	collection, err := meta.GetCollection(args["db"].(string), args["shard_instance"].(string), args["collection"].(string))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	recordData := args["record"].(map[string]interface{})
	fieldHeaders := make([]string, 0, len(recordData))
	fieldValues := make([]string, 0, len(recordData))

	for fieldName, fieldValue := range recordData {
		field, ok := collection.Fields[fieldName]
		if !ok {
			result.Error = fmt.Sprintf("Field %s doesn't exist in %v.%v out of %v", fieldName, args["db"], args["collection"], collection.Fields)
			return result
		}

		fieldHeaders = append(fieldHeaders, "\""+fieldName+"\"")
		switch fieldValue.(type) {
		case nil:
			fieldValues = append(fieldValues, "null")
		default:
			switch field.FieldType.DatamanType {
			case metadata.Document:
				fieldJson, err := json.Marshal(fieldValue)
				if err != nil {
					result.Error = err.Error()
					return result
				}
				fieldValues = append(fieldValues, "'"+string(fieldJson)+"'")
			case metadata.Text:
				fallthrough
			case metadata.String:
				fieldValues = append(fieldValues, fmt.Sprintf("'%v'", fieldValue))
			default:
				fieldValues = append(fieldValues, fmt.Sprintf("%v", fieldValue))
			}
		}
	}

	// TODO: re-add
	// insertQuery := fmt.Sprintf("INSERT INTO public.%s (_created, %s) VALUES ('now', %s) RETURNING *", args["collection"], strings.Join(fieldHeaders, ","), strings.Join(fieldValues, ","))
	insertQuery := fmt.Sprintf("INSERT INTO \"%s\".%s (%s) VALUES (%s) RETURNING *", args["shard_instance"].(string), args["collection"], strings.Join(fieldHeaders, ","), strings.Join(fieldValues, ","))
	result.Return, err = DoQuery(s.getDB(args["db"].(string)), insertQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// TODO: add metadata back to the result
	return result
}

func (s *Storage) Update(args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	meta := s.GetMeta()
	collection, err := meta.GetCollection(args["db"].(string), args["shard_instance"].(string), args["collection"].(string))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	recordData := args["record"].(map[string]interface{})
	fieldHeaders := make([]string, 0, len(recordData))
	fieldValues := make([]string, 0, len(recordData))

	for fieldName, fieldValue := range recordData {
		field, ok := collection.Fields[fieldName]
		if !ok {
			result.Error = fmt.Sprintf("CollectionField %s doesn't exist in %v.%v", fieldName, args["db"], args["collection"])
			return result
		}

		fieldHeaders = append(fieldHeaders, "\""+fieldName+"\"")
		switch fieldValue.(type) {
		case nil:
			fieldValues = append(fieldValues, "null")
		default:
			switch field.FieldType.DatamanType {
			case metadata.Document:
				fieldJson, err := json.Marshal(fieldValue)
				if err != nil {
					result.Error = err.Error()
					return result
				}
				fieldValues = append(fieldValues, "'"+string(fieldJson)+"'")
			case metadata.Text:
				fallthrough
			case metadata.String:
				fieldValues = append(fieldValues, fmt.Sprintf("'%v'", fieldValue))
			default:
				fieldValues = append(fieldValues, fmt.Sprintf("%v", fieldValue))
			}
		}
	}

	setClause := ""
	for i, header := range fieldHeaders {
		setClause += header + "=" + fieldValues[i]
		if i+1 < len(fieldHeaders) {
			setClause += ", "
		}
	}

	whereClause, err := s.filterToWhere(args)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	//updateQuery := fmt.Sprintf("UPDATE \"%s\".%s SET _updated='now',%s WHERE %s RETURNING *", args["shard_instance"].(string), args["collection"], setClause, whereClause)
	updateQuery := fmt.Sprintf("UPDATE \"%s\".%s SET %s WHERE %s RETURNING *", args["shard_instance"].(string), args["collection"], setClause, whereClause)

	result.Return, err = DoQuery(s.getDB(args["db"].(string)), updateQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	s.normalizeResult(args, result)

	// TODO: add metadata back to the result
	return result
}

func (s *Storage) Delete(args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	whereClause, err := s.filterToWhere(args)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	sqlQuery := fmt.Sprintf("DELETE FROM \"%s\".%s WHERE _id=%v%s RETURNING *", args["shard_instance"].(string), args["collection"], args["_id"], whereClause)
	rows, err := DoQuery(s.getDB(args["db"].(string)), sqlQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Return = rows
	s.normalizeResult(args, result)
	return result

}

func (s *Storage) Filter(args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	// TODO: figure out how to do cross-db queries? Seems that most golang drivers
	// don't support it (new in postgres 7.3)
	sqlQuery := fmt.Sprintf("SELECT * FROM \"%s\".%s", args["shard_instance"].(string), args["collection"])

	whereClause, err := s.filterToWhere(args)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if whereClause != "" {
		sqlQuery += " WHERE " + whereClause
	}

	rows, err := DoQuery(s.getDB(args["db"].(string)), sqlQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Return = rows
	s.normalizeResult(args, result)

	return result
}

func (s *Storage) normalizeResult(args query.QueryArgs, result *query.Result) {

	// TODO: better -- we need to convert "documents" into actual structure (instead of just json strings)
	meta := s.GetMeta()
	collection, err := meta.GetCollection(args["db"].(string), args["shard_instance"].(string), args["collection"].(string))
	if err != nil {
		result.Error = err.Error()
		return
	}
	for _, row := range result.Return {
		for k, v := range row {
			if field, ok := collection.Fields[k]; ok && v != nil {
				switch field.FieldType.DatamanType {
				case metadata.Document:
					var tmp map[string]interface{}
					json.Unmarshal(v.([]byte), &tmp)
					row[k] = tmp
				default:
					continue
				}
			}
		}
	}
}

func filterTypeToComparator(f filter.FilterType) string {
	switch f {
	case filter.In:
		return "IN"
	case filter.NotIn:
		return "NOT IN"
	default:
		return string(f)
	}
}

// Take a filter map and return the "where" section (without the actual WHERE statement) for the given filter
// This takes a map of filter which would look something like this:
//
//	{"_id": ["=", 100]}
//
//	{"count": ["<", 100], "foo.bar.baz": [">", 10000]}
//
func (s *Storage) filterToWhere(args map[string]interface{}) (string, error) {
	whereClause := ""
	if rawFilter, ok := args["filter"]; ok && rawFilter != nil {
		filterData := rawFilter.(map[string]interface{})
		meta := s.GetMeta()
		collection, err := meta.GetCollection(args["db"].(string), args["shard_instance"].(string), args["collection"].(string))
		if err != nil {
			return "", err
		}

		whereParts := make([]string, 0)
		for rawFieldName, fieldFilterRaw := range filterData {
			fieldNameParts := strings.Split(rawFieldName, ".")

			field, ok := collection.Fields[fieldNameParts[0]]
			if !ok {
				return "", fmt.Errorf("Field %s doesn't exist in %v.%v", fieldNameParts[0], args["db"], args["collection"])
			}

			fieldName := `"` + fieldNameParts[0] + `"`

			if len(fieldNameParts) > 1 {
				var ok bool
				for _, fieldNamePart := range fieldNameParts[1:] {
					fieldName += "->>'" + fieldNamePart + "'"
					if field == nil {
						field, ok = collection.Fields[fieldNameParts[0]]
						if !ok {
							return "", fmt.Errorf("Field %s doesn't exist in %v.%v", fieldName, args["db"], args["collection"])
						}
					} else {
						field, ok = field.SubFields[fieldNamePart]
						if !ok {
							return "", fmt.Errorf("SubField %s doesn't exist in %v.%v: %v", fieldName, args["db"], args["collection"], field.SubFields)
						}
					}
				}
			}

			fieldFilter, ok := fieldFilterRaw.([]interface{})
			if !ok {
				return "", fmt.Errorf(`"filter" must be a list not %v`, fieldFilterRaw)
			}

			filterType := filter.FilterType(fieldFilter[0].(string))

			fieldValue := fieldFilter[1]

			switch fieldValue.(type) {
			// SQL treats nulls completely differently-- so we need to do that
			case nil:
				var comparator string
				// Note: sql kinda sucks for NIL types-- so we have to special case this
				switch filterType {
				case filter.Equal:
					comparator = "IS"
				case filter.NotEqual:
					comparator = "IS NOT"
				default:
					comparator = filterTypeToComparator(filterType)
				}
				whereParts = append(whereParts, fmt.Sprintf(" \"%s\" %s NULL", fieldName, comparator))
			default:
				switch field.FieldType.DatamanType {
				case metadata.Document:
					// TODO: recurse and add many
					for innerName, innerValue := range fieldValue.(map[string]interface{}) {
						whereParts = append(whereParts, fmt.Sprintf(" %s->>'%s'%s'%v'", fieldName, innerName, filterTypeToComparator(filterType), innerValue))
					}
				case metadata.Text:
					fallthrough
				case metadata.String:
					whereParts = append(whereParts, fmt.Sprintf(" %s%s'%v'", fieldName, filterTypeToComparator(filterType), fieldValue))
				default:
					// TODO: better? Really in postgres once you have an object values are always going to be treated as text-- so we want to do so
					// This is just cheating assuming that any depth is an object-- but we'll need to do better once we support arrays etc.
					if len(fieldNameParts) > 1 {
						whereParts = append(whereParts, fmt.Sprintf(" %s%s'%v'", fieldName, filterTypeToComparator(filterType), fieldValue))
					} else {
						whereParts = append(whereParts, fmt.Sprintf(" %s%s%v", fieldName, filterTypeToComparator(filterType), fieldValue))
					}
				}
			}

		}
		if len(whereParts) > 0 {
			whereClause += strings.Join(whereParts, " AND ")
		}
	}
	return whereClause, nil
}
