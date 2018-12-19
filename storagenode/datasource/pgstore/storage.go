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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jacksontj/dataman/datamantype"
	"github.com/jacksontj/dataman/query"
	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/storagenode/metadata"
	"github.com/jacksontj/dataman/storagenode/metadata/aggregation"
	"github.com/jacksontj/dataman/storagenode/metadata/filter"
	"github.com/jacksontj/dataman/storagenode/metadata/recordop"
	_ "github.com/lib/pq"
	yaml "gopkg.in/yaml.v2"
)

type StorageConfig struct {
	// How to connect to postgres
	PGString        string         `yaml:"pg_string"`
	MaxIdleConns    *int           `yaml:"max_idle_conns"`
	MaxOpenConns    *int           `yaml:"max_open_conns"`
	ConnMaxLifetime *time.Duration `yaml:"conn_max_lifetime"`
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

	// TODO: better
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	s.config = &StorageConfig{}
	if err := yaml.Unmarshal(b, s.config); err != nil {
		return err
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

		// Apply options
		if s.config.MaxIdleConns != nil {
			dbConn.SetMaxIdleConns(*s.config.MaxIdleConns)
		}

		if s.config.MaxOpenConns != nil {
			dbConn.SetMaxOpenConns(*s.config.MaxOpenConns)
		}

		if s.config.ConnMaxLifetime != nil {
			dbConn.SetConnMaxLifetime(*s.config.ConnMaxLifetime)
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

	meta := s.GetMeta()
	collection, err := meta.GetCollection(args.DB, args.ShardInstance, args.Collection)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

	whereParts := make([]string, 0)
	for _, fieldName := range collection.PrimaryIndex.Fields {
		fieldNameParts := strings.Split(fieldName, ".")
		field := collection.GetField(fieldNameParts)
		if field == nil {
			result.Errors = []string{"pkey " + fieldName + " missing from meta? Shouldn't be possible"}
			return result
		}
		fieldValue, ok := args.PKey[fieldName]
		if !ok {
			result.Errors = []string{"missing " + fieldName + " from pkey"}
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
	selectFields, colAddr := selectFields(args.Fields)
	selectQuery := fmt.Sprintf("SELECT %s FROM \"%s\".%s WHERE %s",
		selectFields,
		args.ShardInstance,
		args.Collection,
		strings.Join(whereParts, " AND "),
	)
	result.Return, err = DoQuery(ctx, s.getDB(args.DB), selectQuery, colAddr)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}
	s.normalizeResult(collection, result)

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
	collection, err := meta.GetCollection(args.DB, args.ShardInstance, args.Collection)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

	// TODO: move to a separate method
	fieldHeaders := make([]string, 0, len(args.Record))
	fieldValues := make([]string, 0, len(args.Record))

	recordValues, err := s.recordOpDo(args, args.Record, collection)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

DEFAULT_LOOP:
	for fieldName, field := range collection.Fields {
		// Exclude primary keys from this null setting
		for _, indexFieldName := range collection.PrimaryIndex.Fields {
			if indexFieldName == fieldName {
				continue DEFAULT_LOOP
			}
		}

		// If the field is defined in the args
		if _, ok := args.Record[fieldName]; ok {
			continue
		}
		// If the field is part of the record_op
		if _, ok := recordValues[fieldName]; ok {
			continue
		}

		if field.NotNull {
			result.Errors = []string{fmt.Sprintf("Field %s doesn't allow null values", fieldName)}
			return result
		}

		fieldHeaders = append(fieldHeaders, "\""+fieldName+"\"")
		fieldValues = append(fieldValues, "null")
	}

	for fieldName, fieldValue := range args.Record {
		field, ok := collection.Fields[fieldName]
		if !ok {
			result.Errors = []string{fmt.Sprintf("Field %s doesn't exist in %v.%v out of %v", fieldName, args.DB, args.Collection, collection.Fields)}
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
					result.Errors = []string{err.Error()}
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

	if recordValues != nil {
		// Apply recordValues (assuming they exist
		for k, v := range recordValues {
			if _, ok := args.Record[k]; ok {
				result.Errors = []string{fmt.Sprintf("Already have value in record for %s can't also have in record_op", k)}
				return result
			}

			updatePairs = append(updatePairs, fmt.Sprintf("\"%s\"=%v", k, v))
		}
	}

	selectFields, colAddr := selectFields(args.Fields)
	upsertQuery := fmt.Sprintf(`INSERT INTO "%s".%s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s RETURNING %s`,
		args.ShardInstance,
		args.Collection,
		strings.Join(fieldHeaders, ","),
		strings.Join(fieldValues, ","),
		strings.Join(collection.PrimaryIndex.Fields, ","),
		strings.Join(updatePairs, ","),
		selectFields,
	)

	result.Return, err = DoQuery(ctx, s.getDB(args.DB), upsertQuery, colAddr)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}
	s.normalizeResult(collection, result)

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
	collection, err := meta.GetCollection(args.DB, args.ShardInstance, args.Collection)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

	fieldHeaders := make([]string, 0, len(args.Record))
	fieldValues := make([]string, 0, len(args.Record))

	for fieldName, fieldValue := range args.Record {
		field, ok := collection.Fields[fieldName]
		if !ok {
			result.Errors = []string{fmt.Sprintf("Field %s doesn't exist in %v.%v out of %v", fieldName, args.DB, args.Collection, collection.Fields)}
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
					result.Errors = []string{err.Error()}
					return result
				}
				fieldJson := buffer.Bytes()
				// TODO: switch from string escape of ' to using args from the sql driver
				fieldValues = append(fieldValues, "'"+strings.Replace(string(fieldJson), "'", `''`, -1)+"'")
			case datamantype.DateTime:
				fieldValues = append(fieldValues, fmt.Sprintf("'%v'", fieldValue.(time.Time).Format(datamantype.DateTimeFormatStr)))
			default:
				fieldValues = append(fieldValues, fmt.Sprintf("'%v'", fieldValue))
			}
		}
	}

	selectFields, colAddr := selectFields(args.Fields)
	insertQuery := fmt.Sprintf("INSERT INTO \"%s\".%s (%s) VALUES (%s) RETURNING %s",
		args.ShardInstance,
		args.Collection,
		strings.Join(fieldHeaders, ","),
		strings.Join(fieldValues, ","),
		selectFields,
	)
	result.Return, err = DoQuery(ctx, s.getDB(args.DB), insertQuery, colAddr)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}
	s.normalizeResult(collection, result)

	// TODO: add metadata back to the result
	return result
}

func (s *Storage) InsertMany(ctx context.Context, args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	meta := s.GetMeta()
	collection, err := meta.GetCollection(args.DB, args.ShardInstance, args.Collection)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

	fieldHeaders := make([]string, 0, len(args.Records[0]))
	for fieldName, _ := range args.Records[0] {
		fieldHeaders = append(fieldHeaders, fieldName)
	}
	// A list of the values placed in per record
	groupFieldValues := make([][]string, 0, len(args.Records))

	for _, rec := range args.Records {
		if len(rec) != len(fieldHeaders) {
			result.Errors = []string{"insert_many currently requires that all records have the same top-level fields"}
			return result
		}

		fieldValues := make([]string, 0, len(rec))
		for _, fieldName := range fieldHeaders {
			fieldValue, ok := rec[fieldName]
			if !ok {
				result.Errors = []string{"insert_many currently requires that all records have the same top-level fields"}
				return result
			}

			field, ok := collection.Fields[fieldName]
			if !ok {
				result.Errors = []string{fmt.Sprintf("Field %s doesn't exist in %v.%v out of %v", fieldName, args.DB, args.Collection, collection.Fields)}
				return result
			}
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
						result.Errors = []string{err.Error()}
						return result
					}
					fieldJson := buffer.Bytes()
					// TODO: switch from string escape of ' to using args from the sql driver
					fieldValues = append(fieldValues, "'"+strings.Replace(string(fieldJson), "'", `''`, -1)+"'")
				case datamantype.DateTime:
					fieldValues = append(fieldValues, fmt.Sprintf("'%v'", fieldValue.(time.Time).Format(datamantype.DateTimeFormatStr)))
				default:
					fieldValues = append(fieldValues, fmt.Sprintf("'%v'", fieldValue))
				}
			}
		}
		groupFieldValues = append(groupFieldValues, fieldValues)
	}

	valuesBuilder := strings.Builder{}
	for i, fieldValues := range groupFieldValues {
		if i > 0 {
			valuesBuilder.WriteString(",")
		}
		fmt.Fprintf(&valuesBuilder, " (%s)", strings.Join(fieldValues, ","))
	}

	selectFields, colAddr := selectFields(args.Fields)
	insertQuery := fmt.Sprintf("INSERT INTO \"%s\".%s (%s) VALUES %s RETURNING %s",
		args.ShardInstance,
		args.Collection,
		"\""+strings.Join(fieldHeaders, "\",\"")+"\"",
		valuesBuilder.String(),
		selectFields,
	)
	result.Return, err = DoQuery(ctx, s.getDB(args.DB), insertQuery, colAddr)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}
	s.normalizeResult(collection, result)

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
	collection, err := meta.GetCollection(args.DB, args.ShardInstance, args.Collection)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

	fieldHeaders := make([]string, 0, len(args.Record))
	fieldValues := make([]string, 0, len(args.Record))

	for fieldName, fieldValue := range args.Record {
		field, ok := collection.Fields[fieldName]
		if !ok {
			result.Errors = []string{fmt.Sprintf("CollectionField %s doesn't exist in %v.%v", fieldName, args.DB, args.Collection)}
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
					result.Errors = []string{err.Error()}
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

	recordValues, err := s.recordOpDo(args, args.Record, collection)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}
	if recordValues != nil {
		// Apply recordValues (assuming they exist
		for k, v := range recordValues {
			if _, ok := args.Record[k]; ok {
				result.Errors = []string{fmt.Sprintf("Already have value in record for %s can't also have in record_op", k)}
				return result
			}
			fieldHeaders = append(fieldHeaders, `"`+k+`"`)
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
		result.Errors = []string{err.Error()}
		return result
	}

	selectFields, colAddr := selectFields(args.Fields)
	updateQuery := fmt.Sprintf("UPDATE \"%s\".%s SET %s WHERE %s RETURNING %s",
		args.ShardInstance,
		args.Collection,
		setClause,
		whereClause,
		selectFields,
	)

	result.Return, err = DoQuery(ctx, s.getDB(args.DB), updateQuery, colAddr)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}
	s.normalizeResult(collection, result)

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

	if args.PKey == nil {
		result.Errors = []string{"pkey record required"}
		return result
	}

	meta := s.GetMeta()
	collection, err := meta.GetCollection(args.DB, args.ShardInstance, args.Collection)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

	whereParts := make([]string, 0)
	for _, fieldName := range collection.PrimaryIndex.Fields {
		fieldNameParts := strings.Split(fieldName, ".")
		field := collection.GetField(fieldNameParts)
		if field == nil {
			result.Errors = []string{"pkey " + fieldName + " missing from meta? Shouldn't be possible"}
			return result
		}
		fieldValue, ok := args.PKey[fieldName]
		if !ok {
			result.Errors = []string{"missing " + fieldName + " from pkey"}
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
		result.Errors = []string{err.Error()}
		return result
	}
	if whereClause != "" {
		whereClause = " AND " + whereClause
	}

	selectFields, colAddr := selectFields(args.Fields)
	sqlQuery := fmt.Sprintf("DELETE FROM \"%s\".%s WHERE %s%s RETURNING %s",
		args.ShardInstance,
		args.Collection,
		strings.Join(whereParts, ","),
		whereClause,
		selectFields,
	)
	rows, err := DoQuery(ctx, s.getDB(args.DB), sqlQuery, colAddr)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

	if len(rows) == 0 {
		result.Errors = []string{"Unable to find record with given pkey"}
		return result
	}

	result.Return = rows
	s.normalizeResult(collection, result)
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

	meta := s.GetMeta()
	collection, err := meta.GetCollection(args.DB, args.ShardInstance, args.Collection)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

	// TODO: figure out how to do cross-db queries? Seems that most golang drivers
	// don't support it (new in postgres 7.3)
	selectFields, colAddr := selectFields(args.Fields)
	queryBuilder := strings.Builder{}
	fmt.Fprintf(&queryBuilder, "SELECT %s FROM \"%s\".%s", selectFields, args.ShardInstance, args.Collection)

	if err := s.filterToWhereBuilder(&queryBuilder, args); err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

	if args.Sort != nil && len(args.Sort) > 0 {
		if args.Sort != nil {
			if args.SortReverse == nil {
				args.SortReverse = make([]bool, len(args.Sort))
				// TODO: better, seems heavy
				for i := range args.SortReverse {
					args.SortReverse[i] = false
				}
			}
		}

		fmt.Fprintf(&queryBuilder, " ORDER BY ")
		for i, sortKey := range args.Sort {
			if args.SortReverse[i] {
				fmt.Fprintf(&queryBuilder, `"`+sortKey+`" DESC NULLS LAST`)
			} else {
				fmt.Fprintf(&queryBuilder, `"`+sortKey+`" ASC NULLS FIRST`)
			}
		}
	}

	if args.Limit > 0 {
		fmt.Fprintf(&queryBuilder, fmt.Sprintf(" LIMIT %d", args.Limit))
	}

	if args.Offset > 0 {
		fmt.Fprintf(&queryBuilder, fmt.Sprintf(" OFFSET %d ROWS", args.Offset))
	}

	rows, err := DoQuery(ctx, s.getDB(args.DB), queryBuilder.String(), colAddr)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

	result.Return = rows
	s.normalizeResult(collection, result)

	return result
}

func (s *Storage) Aggregate(ctx context.Context, args query.QueryArgs) *query.Result {
	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	meta := s.GetMeta()
	collection, err := meta.GetCollection(args.DB, args.ShardInstance, args.Collection)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

	sFields := make([]string, 0, len(args.AggregationFields))
	colAddrs := make([]ColAddr, 0, len(args.AggregationFields))
	groupByFields := make([]string, 0)

	for fieldName, aggregationList := range args.AggregationFields {
		fieldParts := strings.Split(fieldName, ".")
		if aggregationList == nil || len(aggregationList) == 0 {
			if len(fieldParts) > 1 {
				colAddrs = append(colAddrs, ColAddr{skipN: 1})
				checker := collectionFieldToSelector(fieldParts[:len(fieldParts)-1]) + " ? '" + fieldParts[len(fieldParts)-1] + "'"
				sFields = append(sFields, checker)
				groupByFields = append(groupByFields, checker)
			}
			groupByFields = append(groupByFields, collectionFieldToSelector(fieldParts))
			sFields = append(sFields, collectionFieldToSelector(fieldParts))
			colAddrs = append(colAddrs, ColAddr{key: fieldParts})
		} else {
			lastFieldPart := fieldParts[len(fieldParts)-1]
			for _, aggregationType := range aggregationList {
				fieldParts[len(fieldParts)-1] = lastFieldPart
				// TODO: util function to map aggregationType to SQL strings
				switch aggregationType {
				case aggregation.Min, aggregation.Max, aggregation.Sum:
					sFields = append(sFields, fmt.Sprintf("%s((%s)::integer)", aggregationType, collectionFieldToSelector(fieldParts)))
				default:
					sFields = append(sFields, fmt.Sprintf("%s(%s)", aggregationType, collectionFieldToSelector(fieldParts)))
				}
				// TODO: avoid making an entire copy? If we don't then the colAddrs end up being
				// the same for all aggregationTypes in the aggregationList
				newFieldParts := make([]string, len(fieldParts))
				for i, fieldPart := range fieldParts {
					newFieldParts[i] = fieldPart
				}
				newFieldParts[len(fieldParts)-1] = lastFieldPart + "." + string(aggregationType)
				colAddrs = append(colAddrs, ColAddr{key: newFieldParts})
			}
		}

	}
	// TODO: actual conversion
	// TODO: figure out how to do cross-db queries? Seems that most golang drivers
	// don't support it (new in postgres 7.3)
	sqlQuery := fmt.Sprintf("SELECT %s FROM \"%s\".%s", strings.Join(sFields, ","), args.ShardInstance, args.Collection)
	whereClause, err := s.filterToWhere(args)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}
	if whereClause != "" {
		sqlQuery += " WHERE " + whereClause
	}

	if len(groupByFields) > 0 {
		sqlQuery += fmt.Sprintf(" GROUP BY (%s)", strings.Join(groupByFields, ","))
	}

	if args.Sort != nil && len(args.Sort) > 0 {
		if args.Sort != nil {
			if args.SortReverse == nil {
				args.SortReverse = make([]bool, len(args.Sort))
				// TODO: better, seems heavy
				for i := range args.SortReverse {
					args.SortReverse[i] = false
				}
			}
		}

		sqlQuery += " ORDER BY "
		for i, sortKey := range args.Sort {
			k := collectionFieldToSelector(strings.Split(sortKey, "."))
			if args.SortReverse[i] {
				sqlQuery += k + ` DESC NULLS LAST`
			} else {
				sqlQuery += k + ` ASC NULLS FIRST`
			}
		}
	}

	if args.Limit > 0 {
		sqlQuery += fmt.Sprintf(" LIMIT %d", args.Limit)
	}

	if args.Offset > 0 {
		sqlQuery += fmt.Sprintf(" OFFSET %d ROWS", args.Offset)
	}

	rows, err := DoQuery(ctx, s.getDB(args.DB), sqlQuery, colAddrs)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}
	result.Return = rows
	s.normalizeResult(collection, result)

	return result
}

// TODO: combine filter & filterStream query generation (they are literally a copy/paste up until the actual query execution)
func (s *Storage) FilterStream(ctx context.Context, args query.QueryArgs) *query.ResultStream {
	result := &query.ResultStream{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "postgres",
		},
	}

	meta := s.GetMeta()
	collection, err := meta.GetCollection(args.DB, args.ShardInstance, args.Collection)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

	// TODO: figure out how to do cross-db queries? Seems that most golang drivers
	// don't support it (new in postgres 7.3)
	selectFields, colAddr := selectFields(args.Fields)
	sqlQuery := fmt.Sprintf("SELECT %s FROM \"%s\".%s", selectFields, args.ShardInstance, args.Collection)

	whereClause, err := s.filterToWhere(args)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}
	if whereClause != "" {
		sqlQuery += " WHERE " + whereClause
	}

	if args.Sort != nil && len(args.Sort) > 0 {
		if args.Sort != nil {
			if args.SortReverse == nil {
				args.SortReverse = make([]bool, len(args.Sort))
				// TODO: better, seems heavy
				for i := range args.SortReverse {
					args.SortReverse[i] = false
				}
			}
		}

		sqlQuery += " ORDER BY "
		for i, sortKey := range args.Sort {
			if args.SortReverse[i] {
				sqlQuery += `"` + sortKey + `" DESC NULLS LAST`
			} else {
				sqlQuery += `"` + sortKey + `" ASC NULLS FIRST`
			}
		}
	}

	if args.Limit > 0 {
		sqlQuery += fmt.Sprintf(" LIMIT %d", args.Limit)
	}

	if args.Offset > 0 {
		sqlQuery += fmt.Sprintf(" OFFSET %d ROWS", args.Offset)
	}

	streamChan, err := DoStreamQuery(ctx, s.getDB(args.DB), sqlQuery, colAddr)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

	result.Stream = streamChan

	// Add transformation to normalize the various JSON fields
	result.AddTransformation(func(r record.Record) (record.Record, error) {
		return s.normalizeRecord(collection, r), nil
	})

	return result
}

func (s *Storage) normalizeResult(collection *metadata.Collection, result *query.Result) {
	for _, row := range result.Return {
		s.normalizeRecord(collection, row)
	}
}

func (s *Storage) normalizeRecord(collection *metadata.Collection, row record.Record) record.Record {
	for k, v := range row {
		if field, ok := collection.Fields[k]; ok && v != nil {
			switch field.FieldType.DatamanType {
			case datamantype.JSON:
				if byteSlice, ok := v.([]byte); ok {
					var tmp interface{}
					json.Unmarshal(byteSlice, &tmp)
					row[k] = tmp
				}
			case datamantype.Document:
				if byteSlice, ok := v.([]byte); ok {
					var tmp map[string]interface{}
					json.Unmarshal(byteSlice, &tmp)
					row[k] = tmp
				}
			default:
				continue
			}
		}
	}
	return row
}

func filterTypeToComparator(f filter.FilterType) string {
	switch f {
	case filter.In:
		return " IN "
	case filter.NotIn:
		return " NOT IN "
	case filter.RegexEqual:
		return "~"
	case filter.RegexNotEqual:
		return "!~"
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
func (s *Storage) filterToWhere(args query.QueryArgs) (string, error) {
	whereClause := ""
	if args.Filter != nil {
		meta := s.GetMeta()
		collection, err := meta.GetCollection(args.DB, args.ShardInstance, args.Collection)
		if err != nil {
			return "", err
		}

		switch args.Filter.(type) {
		case []interface{}, map[string]interface{}:
			whereClause, err = s.filterToWhereInner(collection, args.Filter)
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
		// If the filter is empty we should skip it. Instead of special handling it above
		// we'll just convert it to a filter which is always true (so its still a no-op)
		if len(filterData) == 0 {
			return "true", nil
		}
		whereParts := make([]string, 0, len(filterData))
		for rawFieldName, fieldFilterRaw := range filterData {
			if !collection.IsValidProjection(rawFieldName) {
				return "", errors.New("Invalid field in filter: " + rawFieldName)
			}

			var filterType filter.FilterType
			var fieldValue interface{}
			var err error

			switch fieldFilterTyped := fieldFilterRaw.(type) {
			case []interface{}:
				switch filterTyped := fieldFilterTyped[0].(type) {
				case filter.FilterType:
					filterType = filterTyped
				case string:
					filterType, err = filter.StringToFilterType(filterTyped)
					if err != nil {
						return "", err
					}
				default:
					return "", fmt.Errorf("Invalid filter type %v", filterTyped)
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
				whereParts = append(whereParts, " "+collectionFieldToSelector(strings.Split(rawFieldName, "."))+" "+comparator+" NULL")
			default:
				switch filterType {
				case filter.In, filter.NotIn:
					var items []string
					switch typedFieldValue := fieldValue.(type) {
					case []interface{}:
						items = make([]string, len(typedFieldValue))
						for i, rawItem := range typedFieldValue {
							if item, err := serializeValue(rawItem); err == nil {
								items[i] = item
							} else {
								return "", err
							}
						}
					case []string:
						items = make([]string, len(typedFieldValue))
						for i, rawItem := range typedFieldValue {
							items[i] = "'" + rawItem + "'"
						}
					default:
						return "", fmt.Errorf("Value of %s must be a list", filterType)
					}
					whereParts = append(whereParts, fmt.Sprintf(" %s%s%s", collectionFieldToSelector(strings.Split(rawFieldName, ".")), filterTypeToComparator(filterType), "("+strings.Join(items, ",")+")"))
					whereParts = append(whereParts, " "+collectionFieldToSelector(strings.Split(rawFieldName, "."))+filterTypeToComparator(filterType)+"("+strings.Join(items, ",")+")")
				default:
					normalizedFieldValue, err := serializeValue(fieldValue)
					if err != nil {
						return "", err
					}
					whereParts = append(whereParts, " "+collectionFieldToSelector(strings.Split(rawFieldName, "."))+filterTypeToComparator(filterType)+normalizedFieldValue)
				}
			}
		}
		return strings.Join(whereParts, " AND "), nil
	}
	// TODO: better error message
	return "", fmt.Errorf("Unknown where clause!")
}

// Take a filter map and return the "where" section (without the actual WHERE statement) for the given filter
// This takes a map of filter which would look something like this:
//
//	{"_id": ["=", 100]}
//
//	{"count": ["<", 100], "foo.bar.baz": [">", 10000]}
//
func (s *Storage) filterToWhereBuilder(queryBuilder *strings.Builder, args query.QueryArgs) error {
	if args.Filter != nil {
		meta := s.GetMeta()
		collection, err := meta.GetCollection(args.DB, args.ShardInstance, args.Collection)
		if err != nil {
			return err
		}

		queryBuilder.WriteString(" WHERE ")

		switch args.Filter.(type) {
		case []interface{}, map[string]interface{}:
			return s.filterToWhereInnerBuilder(queryBuilder, collection, args.Filter)
		default:
			return fmt.Errorf("Filters must have a map or a list at the top level")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO: refactor to be less... ugly
func (s *Storage) filterToWhereInnerBuilder(queryBuilder *strings.Builder, collection *metadata.Collection, f interface{}) error {
	switch filterData := f.(type) {
	// If this is simply an operator
	case string:
		switch strings.ToUpper(filterData) {
		// TODO: use them from the filter package
		case "AND":
			queryBuilder.WriteString(string(filter.And))
			return nil
		case "OR":
			queryBuilder.WriteString(string(filter.Or))
			return nil
		default:
			return fmt.Errorf("Invalid operator %s", filterData)
		}
	case []interface{}:
		if len(filterData) != 3 {
			return fmt.Errorf("where lists need to be A op B")
		}
		operatorRaw, ok := filterData[1].(string)
		if !ok {
			return fmt.Errorf("Operator must be a string")
		}
		upperOperator := strings.ToUpper(operatorRaw)
		var operator string
		switch upperOperator {
		case "AND":
			operator = upperOperator
		case "OR":
			operator = upperOperator
		default:
			return fmt.Errorf("Invalid operator %s", filterData)
		}

		queryBuilder.WriteString("(")

		if err := s.filterToWhereInnerBuilder(queryBuilder, collection, filterData[0]); err != nil {
			return err
		}

		queryBuilder.WriteString(" " + operator + " ")
		if err := s.filterToWhereInnerBuilder(queryBuilder, collection, filterData[2]); err != nil {
			return err
		}
		queryBuilder.WriteString(")")

		return nil

	case map[string]interface{}:
		// If the filter is empty we should skip it. Instead of special handling it above
		// we'll just convert it to a filter which is always true (so its still a no-op)
		if len(filterData) == 0 {
			queryBuilder.WriteString("true")
			return nil
		}

		whereCount := 0
		for rawFieldName, fieldFilterRaw := range filterData {
			if !collection.IsValidProjection(rawFieldName) {
				return errors.New("Invalid field in filter: " + rawFieldName)
			}

			var filterType filter.FilterType
			var fieldValue interface{}
			var err error

			switch fieldFilterTyped := fieldFilterRaw.(type) {
			case []interface{}:
				switch filterTyped := fieldFilterTyped[0].(type) {
				case filter.FilterType:
					filterType = filterTyped
				case string:
					filterType, err = filter.StringToFilterType(filterTyped)
					if err != nil {
						return err
					}
				default:
					return fmt.Errorf("Invalid filter type %v", filterTyped)
				}

				fieldValue = fieldFilterTyped[1]
			case []string:
				filterType, err = filter.StringToFilterType(fieldFilterTyped[0])
				if err != nil {
					return err
				}
				fieldValue = fieldFilterTyped[1]
			default:
				return fmt.Errorf(`filter must be a list`)
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
				if whereCount > 0 {
					queryBuilder.WriteString(" AND ")
				}
				queryBuilder.WriteString(" " + collectionFieldToSelector(strings.Split(rawFieldName, ".")) + " " + comparator + " NULL")
				whereCount++
			default:
				switch filterType {
				case filter.In, filter.NotIn:
					if whereCount > 0 {
						queryBuilder.WriteString(" AND ")
					}
					queryBuilder.WriteString(" " + collectionFieldToSelector(strings.Split(rawFieldName, ".")) + filterTypeToComparator(filterType) + "(")

					switch typedFieldValue := fieldValue.(type) {
					case []interface{}:
						for i, rawItem := range typedFieldValue {
							if i > 0 {
								queryBuilder.WriteString(",")
							}
							if err := serializeValueBuilder(queryBuilder, rawItem); err != nil {
								return err
							}
						}
					case []string:
						for i, rawItem := range typedFieldValue {
							if i > 0 {
								queryBuilder.WriteString(",")
							}
							queryBuilder.WriteString("'" + rawItem + "'")
						}

					default:
						return fmt.Errorf("Value of %s must be a list", filterType)
					}
					queryBuilder.WriteString(")")
					whereCount++
				default:
					if whereCount > 0 {
						queryBuilder.WriteString(" AND ")
					}
					queryBuilder.WriteString(" " + collectionFieldToSelector(strings.Split(rawFieldName, ".")) + filterTypeToComparator(filterType))
					if err := serializeValueBuilder(queryBuilder, fieldValue); err != nil {
						return err
					}
					whereCount++
				}
			}
		}
		return nil
	}
	// TODO: better error message
	return fmt.Errorf("Unknown where clause!")
}

func (s *Storage) recordOpDo(args query.QueryArgs, recordData map[string]interface{}, collection *metadata.Collection) (map[string]string, error) {
	// Return map of header -> value
	opValues := make(map[string]string)

	for fieldAddr, fieldOpList := range args.RecordOp {
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

			opValues[fieldAddr] = fmt.Sprintf("%s.%s %s %v", collection.Name, fieldAddr, opType, opValue)

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
					collection.Name+"."+collectionFieldToSelector(fieldAddrParts),
					opType,
					opValue,
				)
			} else {
				opValues[fieldAddrParts[0]] = fmt.Sprintf(jsonbSetTemplate,
					collection.Name+"."+fieldAddrParts[0],
					strings.Join(fieldAddrParts[1:], ","),
					collection.Name+"."+collectionFieldToSelector(fieldAddrParts),
					opType,
					opValue,
				)
			}
		}
	}
	return opValues, nil
}
