package pgstorage

// TODO: real escaping of the various queries (sql injection is bad ;) )
// TODO: look into codegen or something for queries (terribly inefficient right now)

/*
This is a storagenode using postgres as a json document store

Metadata about the storage node will be stored in a database called _dataman.storagenode

*/

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jacksontj/dataman/src/datamantype"
	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
	"github.com/jacksontj/dataman/src/storage_node/metadata/filter"
	"github.com/jacksontj/dataman/src/storage_node/metadata/recordop"
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
func (s *Storage) Get(ctx context.Context, args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	// TODO: figure out how to do cross-db queries? Seems that most golang drivers
	// don't support it (new in postgres 7.3)

	rawPkeyRecord, ok := args["pkey"]
	if !ok {
		result.Error = "pkey record required"
		return result
	}
	pkeyRecord, ok := rawPkeyRecord.(map[string]interface{})
	if !ok {
		result.Error = "pkey must be a map[string]interface{}"
		return result
	}

	meta := s.GetMeta()
	collection, err := meta.GetCollection(args["db"].(string), args["shard_instance"].(string), args["collection"].(string))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	whereParts := make([]string, 0)
	for _, fieldName := range collection.PrimaryIndex.Fields {
		fieldNameParts := strings.Split(fieldName, ".")
		field := collection.GetField(fieldNameParts)
		if field == nil {
			result.Error = "pkey " + fieldName + " missing from meta? Shouldn't be possible"
			return result
		}
		fieldValue, ok := pkeyRecord[fieldName]
		if !ok {
			result.Error = "missing " + fieldName + " from pkey"
			return result
		}
		switch field.FieldType.DatamanType {
		case datamantype.Document:
			// TODO: recurse and add many
			for innerName, innerValue := range fieldValue.(map[string]interface{}) {
				whereParts = append(whereParts, fmt.Sprintf(" %s->>'%s'='%v'", fieldName, innerName, innerValue))
			}
		case datamantype.Text, datamantype.String:
			whereParts = append(whereParts, fmt.Sprintf(" %s='%v'", fieldName, fieldValue))
		default:
			// TODO: better? Really in postgres once you have an object values are always going to be treated as text-- so we want to do so
			// This is just cheating assuming that any depth is an object-- but we'll need to do better once we support arrays etc.
			if len(fieldNameParts) > 1 {
				whereParts = append(whereParts, fmt.Sprintf(" %s='%v'", fieldName, fieldValue))
			} else {
				whereParts = append(whereParts, fmt.Sprintf(" %s=%v", fieldName, fieldValue))
			}
		}
	}

	selectQuery := fmt.Sprintf("SELECT * FROM \"%s\".%s WHERE %s", args["shard_instance"].(string), args["collection"], strings.Join(whereParts, " AND "))
	result.Return, err = DoQuery(ctx, s.getDB(args["db"].(string)), selectQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	s.normalizeResult(args, result)

	// TODO: error if there is more than one result
	return result
}

// Set() is a special-case of "upsert" where we do the upsert on the primary key
func (s *Storage) Set(ctx context.Context, args query.QueryArgs) *query.Result {
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

	// TODO: move to a separate method
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
			case datamantype.JSON, datamantype.Document:
				// TODO: make util method?
				// workaround for https://stackoverflow.com/questions/28595664/how-to-stop-json-marshal-from-escaping-and
				buffer := &bytes.Buffer{}
				encoder := json.NewEncoder(buffer)
				encoder.SetEscapeHTML(false)
				err := encoder.Encode(fieldValue)
				if err != nil {
					result.Error = err.Error()
					return result
				}
				fieldJson := buffer.Bytes()
				// TODO: switch from string escape of ' to using args from the sql driver
				fieldValues = append(fieldValues, "'"+strings.Replace(string(fieldJson), "'", `''`, -1)+"'")
			case datamantype.DateTime:
				fieldValues = append(fieldValues, fmt.Sprintf("'%v'", fieldValue.(time.Time).Format(datamantype.DateTimeFormatStr)))
			case datamantype.Text, datamantype.String:
				fieldValues = append(fieldValues, fmt.Sprintf("'%v'", fieldValue))
			default:
				fieldValues = append(fieldValues, fmt.Sprintf("%v", fieldValue))
			}
		}
	}

	pkeyHeaders := make(map[string]struct{})
	for _, pkeyField := range collection.PrimaryIndex.Fields {
		pkeyHeaders[`"`+pkeyField+`"`] = struct{}{}
	}

	updatePairs := make([]string, 0, len(fieldHeaders))
	for j, header := range fieldHeaders {
		updatePairs = append(updatePairs, header+"="+fieldValues[j])
	}

	recordValues, err := s.recordOpDo(args, recordData, collection)
	if err != nil {
		fmt.Println("err?", err)
		result.Error = err.Error()
		return result
	}
	if recordValues != nil {
		// Apply recordValues (assuming they exist
		for k, v := range recordValues {
			fmt.Println("kv", k, v)
			if _, ok := recordData[k]; ok {
				result.Error = fmt.Sprintf("Already have value in record for %s can't also have in record_op", k)
				return result
			}

			updatePairs = append(updatePairs, fmt.Sprintf("\"%s\"=%s.%v", k, collection.Name, v))
		}
	}

	upsertQuery := fmt.Sprintf(`INSERT INTO "%s".%s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s RETURNING *`,
		args["shard_instance"].(string),
		args["collection"],
		strings.Join(fieldHeaders, ","),
		strings.Join(fieldValues, ","),
		strings.Join(collection.PrimaryIndex.Fields, ","),
		strings.Join(updatePairs, ","),
	)

	result.Return, err = DoQuery(ctx, s.getDB(args["db"].(string)), upsertQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	s.normalizeResult(args, result)

	// TODO: add metadata back to the result
	return result
}

func (s *Storage) Insert(ctx context.Context, args query.QueryArgs) *query.Result {
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
			case datamantype.JSON, datamantype.Document:
				// TODO: make util method?
				// workaround for https://stackoverflow.com/questions/28595664/how-to-stop-json-marshal-from-escaping-and
				buffer := &bytes.Buffer{}
				encoder := json.NewEncoder(buffer)
				encoder.SetEscapeHTML(false)
				err := encoder.Encode(fieldValue)
				if err != nil {
					result.Error = err.Error()
					return result
				}
				fieldJson := buffer.Bytes()
				// TODO: switch from string escape of ' to using args from the sql driver
				fieldValues = append(fieldValues, "'"+strings.Replace(string(fieldJson), "'", `''`, -1)+"'")
			case datamantype.DateTime:
				fieldValues = append(fieldValues, fmt.Sprintf("'%v'", fieldValue.(time.Time).Format(datamantype.DateTimeFormatStr)))
			case datamantype.Text, datamantype.String:
				fieldValues = append(fieldValues, fmt.Sprintf("'%v'", fieldValue))
			default:
				fieldValues = append(fieldValues, fmt.Sprintf("%v", fieldValue))
			}
		}
	}

	// TODO: re-add
	// insertQuery := fmt.Sprintf("INSERT INTO public.%s (_created, %s) VALUES ('now', %s) RETURNING *", args["collection"], strings.Join(fieldHeaders, ","), strings.Join(fieldValues, ","))
	insertQuery := fmt.Sprintf("INSERT INTO \"%s\".%s (%s) VALUES (%s) RETURNING *", args["shard_instance"].(string), args["collection"], strings.Join(fieldHeaders, ","), strings.Join(fieldValues, ","))
	result.Return, err = DoQuery(ctx, s.getDB(args["db"].(string)), insertQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	s.normalizeResult(args, result)

	// TODO: add metadata back to the result
	return result
}

func (s *Storage) Update(ctx context.Context, args query.QueryArgs) *query.Result {
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
			case datamantype.JSON, datamantype.Document:
				// TODO: make util method?
				// workaround for https://stackoverflow.com/questions/28595664/how-to-stop-json-marshal-from-escaping-and
				buffer := &bytes.Buffer{}
				encoder := json.NewEncoder(buffer)
				encoder.SetEscapeHTML(false)
				err := encoder.Encode(fieldValue)
				if err != nil {
					result.Error = err.Error()
					return result
				}
				fieldJson := buffer.Bytes()

				// TODO: switch from string escape of ' to using args from the sql driver
				fieldValues = append(fieldValues, "'"+strings.Replace(string(fieldJson), "'", `''`, -1)+"'")
			case datamantype.DateTime:
				fieldValues = append(fieldValues, fmt.Sprintf("'%v'", fieldValue.(time.Time).Format(datamantype.DateTimeFormatStr)))
			case datamantype.Text, datamantype.String:
				fieldValues = append(fieldValues, fmt.Sprintf("'%v'", fieldValue))
			default:
				fieldValues = append(fieldValues, fmt.Sprintf("%v", fieldValue))
			}
		}
	}

	recordValues, err := s.recordOpDo(args, recordData, collection)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if recordValues != nil {
		// Apply recordValues (assuming they exist
		for k, v := range recordValues {
			if _, ok := recordData[k]; ok {
				result.Error = fmt.Sprintf("Already have value in record for %s can't also have in record_op", k)
				return result
			}
			fieldHeaders = append(fieldHeaders, k)
			fieldValues = append(fieldValues, v)
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

	result.Return, err = DoQuery(ctx, s.getDB(args["db"].(string)), updateQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	s.normalizeResult(args, result)

	// TODO: add metadata back to the result
	return result
}

func (s *Storage) Delete(ctx context.Context, args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	rawPkeyRecord, ok := args["pkey"]
	if !ok {
		result.Error = "pkey record required"
		return result
	}
	pkeyRecord, ok := rawPkeyRecord.(map[string]interface{})
	if !ok {
		result.Error = "pkey must be a map[string]interface{}"
		return result
	}

	meta := s.GetMeta()
	collection, err := meta.GetCollection(args["db"].(string), args["shard_instance"].(string), args["collection"].(string))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	whereParts := make([]string, 0)
	for _, fieldName := range collection.PrimaryIndex.Fields {
		fieldNameParts := strings.Split(fieldName, ".")
		field := collection.GetField(fieldNameParts)
		if field == nil {
			result.Error = "pkey " + fieldName + " missing from meta? Shouldn't be possible"
			return result
		}
		fieldValue, ok := pkeyRecord[fieldName]
		if !ok {
			result.Error = "missing " + fieldName + " from pkey"
			return result
		}
		switch field.FieldType.DatamanType {
		case datamantype.Document:
			// TODO: recurse and add many
			for innerName, innerValue := range fieldValue.(map[string]interface{}) {
				whereParts = append(whereParts, fmt.Sprintf(" %s->>'%s'='%v'", fieldName, innerName, innerValue))
			}
		case datamantype.DateTime:
			whereParts = append(whereParts, fmt.Sprintf(" %s='%v'", fieldName, fieldValue.(time.Time).Format(datamantype.DateTimeFormatStr)))
		case datamantype.Text, datamantype.String:
			whereParts = append(whereParts, fmt.Sprintf(" %s='%v'", fieldName, fieldValue))
		default:
			// TODO: better? Really in postgres once you have an object values are always going to be treated as text-- so we want to do so
			// This is just cheating assuming that any depth is an object-- but we'll need to do better once we support arrays etc.
			if len(fieldNameParts) > 1 {
				whereParts = append(whereParts, fmt.Sprintf(" %s='%v'", fieldName, fieldValue))
			} else {
				whereParts = append(whereParts, fmt.Sprintf(" %s=%v", fieldName, fieldValue))
			}
		}
	}

	whereClause, err := s.filterToWhere(args)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if whereClause != "" {
		whereClause = " AND " + whereClause
	}

	sqlQuery := fmt.Sprintf("DELETE FROM \"%s\".%s WHERE %s%s RETURNING *", args["shard_instance"].(string), args["collection"], strings.Join(whereParts, ","), whereClause)
	rows, err := DoQuery(ctx, s.getDB(args["db"].(string)), sqlQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	if len(rows) == 0 {
		result.Error = "Unable to find record with given pkey"
		return result
	}

	result.Return = rows
	s.normalizeResult(args, result)
	return result

}

func (s *Storage) Filter(ctx context.Context, args query.QueryArgs) *query.Result {
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

	rows, err := DoQuery(ctx, s.getDB(args["db"].(string)), sqlQuery)
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
				case datamantype.JSON:
					var tmp interface{}
					json.Unmarshal(v.([]byte), &tmp)
					row[k] = tmp
				case datamantype.Document:
					var tmp map[string]interface{}
					json.Unmarshal(v.([]byte), &tmp)
					row[k] = tmp
				case datamantype.DateTime:
					row[k] = v.(time.Time).Format(datamantype.DateTimeFormatStr)
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
		return " IN "
	case filter.NotIn:
		return " NOT IN "
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
		meta := s.GetMeta()
		collection, err := meta.GetCollection(args["db"].(string), args["shard_instance"].(string), args["collection"].(string))
		if err != nil {
			return "", err
		}
		switch rawFilter.(type) {
		case []interface{}, map[string]interface{}:
			whereClause, err = s.filterToWhereInner(collection, rawFilter)
		default:
			return "", fmt.Errorf("Filters must have a map or a list at the top level")
		}
		if err != nil {
			return "", err
		}
	}
	return whereClause, nil
}

// TODO: refactor to be less... ugly
func (s *Storage) filterToWhereInner(collection *metadata.Collection, f interface{}) (string, error) {
	switch filterData := f.(type) {
	// If this is simply an operator
	case string:
		switch strings.ToUpper(filterData) {
		// TODO: use them from the filter package
		case "AND":
			return string(filter.And), nil
		case "OR":
			return string(filter.Or), nil
		default:
			return "", fmt.Errorf("Invalid operator %s", filterData)
		}
	case []interface{}:
		if len(filterData) != 3 {
			return "", fmt.Errorf("where lists need to be A op B")
		}
		operatorRaw, ok := filterData[1].(string)
		if !ok {
			return "", fmt.Errorf("Operator must be a string")
		}
		upperOperator := strings.ToUpper(operatorRaw)
		var operator string
		switch upperOperator {
		case "AND":
			operator = upperOperator
		case "OR":
			operator = upperOperator
		default:
			return "", fmt.Errorf("Invalid operator %s", filterData)
		}

		first, err := s.filterToWhereInner(collection, filterData[0])
		if err != nil {
			return "", err
		}
		last, err := s.filterToWhereInner(collection, filterData[2])
		if err != nil {
			return "", err
		}

		return "(" + first + " " + operator + " " + last + ")", nil

	case map[string]interface{}:
		whereParts := make([]string, 0)
		for rawFieldName, fieldFilterRaw := range filterData {
			fieldNameParts := strings.Split(rawFieldName, ".")

			field, ok := collection.Fields[fieldNameParts[0]]
			if !ok {
				return "", fmt.Errorf("Field %s doesn't exist in %s", fieldNameParts[0], collection.Name)
			}

			fieldName := `"` + fieldNameParts[0] + `"`

			if len(fieldNameParts) > 1 {
				var ok bool
				for _, fieldNamePart := range fieldNameParts[1:] {
					fieldName += "->>'" + fieldNamePart + "'"
					if field == nil {
						field, ok = collection.Fields[fieldNameParts[0]]
						if !ok {
							return "", fmt.Errorf("Field %s doesn't exist in %s", fieldName, collection.Name)
						}
					} else {
						subField, ok := field.SubFields[fieldNamePart]
						if !ok {
							return "", fmt.Errorf("SubField %s doesn't exist in %s: %v", fieldName, collection.Name, field.SubFields)
						}
						field = subField
					}
				}
			}

			var filterType filter.FilterType
			var fieldValue interface{}
			var err error

			switch fieldFilterTyped := fieldFilterRaw.(type) {
			case []interface{}:
				filterTypeString, ok := fieldFilterTyped[0].(string)
				if !ok {
					return "", fmt.Errorf("Invalid filter type %v", fieldFilterTyped[0])
				}
				filterType, err = filter.StringToFilterType(filterTypeString)
				if err != nil {
					return "", err
				}
				fieldValue = fieldFilterTyped[1]
			case []string:
				filterType, err = filter.StringToFilterType(fieldFilterTyped[0])
				if err != nil {
					return "", err
				}
				fieldValue = fieldFilterTyped[1]
			default:
				return "", fmt.Errorf(`filter must be a list`)
			}

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
				whereParts = append(whereParts, fmt.Sprintf(" %s %s NULL", fieldName, comparator))
			default:
				switch field.FieldType.DatamanType {
				case datamantype.Document:
					// TODO: recurse and add many
					for innerName, innerValue := range fieldValue.(map[string]interface{}) {
						whereParts = append(whereParts, fmt.Sprintf(" %s->>'%s'%s'%v'", fieldName, innerName, filterTypeToComparator(filterType), innerValue))
					}
				default:
					switch filterType {
					case filter.In, filter.NotIn:
						var items []string
						switch typedFieldValue := fieldValue.(type) {
						case []interface{}:
							items = make([]string, len(typedFieldValue))
							for i, rawItem := range typedFieldValue {
								if item, err := serializeValue(field.FieldType.DatamanType, rawItem); err == nil {
									items[i] = item
								} else {
									return "", err
								}
							}
						case []string:
							items = make([]string, len(typedFieldValue))
							for i, rawItem := range typedFieldValue {
								if item, err := serializeValue(field.FieldType.DatamanType, rawItem); err == nil {
									items[i] = item
								} else {
									return "", err
								}
							}
						default:
							return "", fmt.Errorf("Value of %s must be a list", filterType)
						}
						whereParts = append(whereParts, fmt.Sprintf(" %s%s%s", fieldName, filterTypeToComparator(filterType), "("+strings.Join(items, ",")+")"))
					default:
						if v, err := serializeValue(field.FieldType.DatamanType, fieldValue); err == nil {
							whereParts = append(whereParts, fmt.Sprintf(" %s%s%s", fieldName, filterTypeToComparator(filterType), v))
						} else {
							return "", err
						}
					}
				}
			}
		}
		return strings.Join(whereParts, " AND "), nil
	}
	// TODO: better error message
	return "", fmt.Errorf("Unknown where clause!")
}

func (s *Storage) recordOpDo(args map[string]interface{}, recordData map[string]interface{}, collection *metadata.Collection) (map[string]string, error) {
	recordOpRaw, ok := args["record_op"]
	if !ok {
		return nil, nil
	}
	recordOpMap, ok := recordOpRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("record_op must be a map[string]interface{}")
	}

	// Return map of header -> value
	opValues := make(map[string]string)

	for fieldAddr, fieldOpList := range recordOpMap {
		var opType recordop.RecordOp
		var opValue interface{}
		var err error

		switch fieldOpTyped := fieldOpList.(type) {
		case []interface{}:
			opTypeString, ok := fieldOpTyped[0].(string)
			if !ok {
				return nil, fmt.Errorf("Invalid op type %v", fieldOpTyped[0])
			}
			opType, err = recordop.StringToRecordOp(opTypeString)
			if err != nil {
				return nil, err
			}
			opValue = fieldOpTyped[1]
		case []string:
			opType, err = recordop.StringToRecordOp(fieldOpTyped[0])
			if err != nil {
				return nil, err
			}
			opValue = fieldOpTyped[1]
		default:
			return nil, fmt.Errorf("record_op must be a list")
		}

		fieldAddrParts := strings.Split(fieldAddr, ".")

		if len(fieldAddrParts) == 1 {
			// If the value is in recordData we don't allow it (for now)
			if _, ok := recordData[fieldAddr]; ok {
				return nil, fmt.Errorf("Already have value in record for %s can't also have in record_op", fieldAddr)
			}

			// If the field doesn't exist -- don't do it
			if collection.GetFieldByName(fieldAddr) == nil {
				return nil, fmt.Errorf("record_op field %s doesn't exist in collection", fieldAddr)
			}

			opValues[fieldAddr] = fmt.Sprintf("%s %s %v", fieldAddr, opType, opValue)

		} else {
			// If the value is in recordData we don't allow it (for now)
			if _, ok := recordData[fieldAddrParts[0]]; ok {
				return nil, fmt.Errorf("Already have value in record for %s can't also have in record_op", fieldAddr)
			}

			opField := collection.GetFieldByName(fieldAddr)
			// If the field doesn't exist -- don't do it
			if opField == nil {
				return nil, fmt.Errorf("record_op field %s doesn't exist in collection", fieldAddr)
			}

			jsonbSetTemplate := `jsonb_set(%s, '{%s}', (COALESCE(%s,'0')::int %s %v)::text::jsonb)`

			// If we already had an op to this top-level key, then we need to nest
			if baseValue, ok := opValues[fieldAddrParts[0]]; ok {
				opValues[fieldAddrParts[0]] = fmt.Sprintf(jsonbSetTemplate,
					baseValue,
					strings.Join(fieldAddrParts[1:], ","),
					collectionFieldToSelector(fieldAddrParts),
					opType,
					opValue,
				)
			} else {
				opValues[fieldAddrParts[0]] = fmt.Sprintf(jsonbSetTemplate,
					fieldAddrParts[0],
					strings.Join(fieldAddrParts[1:], ","),
					collectionFieldToSelector(fieldAddrParts),
					opType,
					opValue,
				)
			}
		}
	}
	return opValues, nil
}
