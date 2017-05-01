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
	"sync/atomic"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
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
	meta     atomic.Value
}

func (s *Storage) Init(metaFunc metadata.MetaFunc, c map[string]interface{}) error {
	var err error

	if val, ok := c["pg_string"]; ok {
		s.config = &StorageConfig{val.(string)}
	} else {
		return fmt.Errorf("Invalid config")
	}

	// TODO: pass in a database name for the metadata store locally
	s.db, err = sql.Open("postgres", s.config.pgStringForDB("dataman_storage"))
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

	selectQuery := fmt.Sprintf("SELECT * FROM %s.%s WHERE _id=%v", args["shard_instance"].(string), args["collection"], args["_id"])
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

func (s *Storage) Set(args query.QueryArgs) *query.Result {
	record := args["record"]
	if id, ok := record.(map[string]interface{})["_id"]; ok {
		args["filter"] = map[string]interface{}{"_id": id}
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
		field, ok := collection.FieldMap[fieldName]
		if !ok {
			result.Error = fmt.Sprintf("Field %s doesn't exist in %v.%v out of %v", fieldName, args["db"], args["collection"], collection.FieldMap)
			return result
		}

		fieldHeaders = append(fieldHeaders, "\""+fieldName+"\"")
		switch field.Type {
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

	// TODO: re-add
	// insertQuery := fmt.Sprintf("INSERT INTO public.%s (_created, %s) VALUES ('now', %s) RETURNING *", args["collection"], strings.Join(fieldHeaders, ","), strings.Join(fieldValues, ","))
	insertQuery := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s) RETURNING *", args["shard_instance"].(string), args["collection"], strings.Join(fieldHeaders, ","), strings.Join(fieldValues, ","))
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
		field, ok := collection.FieldMap[fieldName]
		if !ok {
			result.Error = fmt.Sprintf("Fuekd %s doesn't exist in %v.%v", fieldName, args["db"], args["collection"])
			return result
		}

		fieldHeaders = append(fieldHeaders, "\""+fieldName+"\"")
		switch field.Type {
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

	setClause := ""
	for i, header := range fieldHeaders {
		setClause += header + "=" + fieldValues[i]
		if i+1 < len(fieldHeaders) {
			setClause += ", "
		}
	}

	// TODO: move to some method
	filterData := args["filter"].(map[string]interface{})
	filterHeaders := make([]string, 0, len(filterData))
	filterValues := make([]string, 0, len(filterData))

	for filterName, filterValue := range filterData {
		if strings.HasPrefix(filterName, "_") {
			filterHeaders = append(filterHeaders, "\""+filterName+"\"")
			filterValues = append(filterValues, fmt.Sprintf("%v", filterValue))
			continue
		}
		field, ok := collection.FieldMap[filterName]
		if !ok {
			result.Error = fmt.Sprintf("Field2 %s doesn't exist in %v.%v", filterName, args["db"], args["collection"])
			return result
		}

		filterHeaders = append(filterHeaders, "\""+filterName+"\"")
		switch field.Type {
		case metadata.Document:
			fieldJson, err := json.Marshal(filterValue)
			if err != nil {
				result.Error = err.Error()
				return result
			}
			filterValues = append(filterValues, "'"+string(fieldJson)+"'")
		case metadata.Text:
			fallthrough
		case metadata.String:
			filterValues = append(filterValues, fmt.Sprintf("'%v'", filterValue))
		default:
			filterValues = append(filterValues, fmt.Sprintf("%v", filterValue))
		}
	}

	whereClause := ""
	for i, header := range filterHeaders {
		whereClause += header + "=" + filterValues[i]
		if i+1 < len(filterHeaders) {
			whereClause += ", "
		}
	}
	updateQuery := fmt.Sprintf("UPDATE %s.%s SET _updated='now',%s WHERE %s RETURNING *", args["shard_instance"].(string), args["collection"], setClause, whereClause)

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

	whereClause := ""

	if filter, ok := args["filter"]; ok {
		meta := s.GetMeta()
		collection, err := meta.GetCollection(args["db"].(string), args["shard_instance"].(string), args["collection"].(string))
		if err != nil {
			result.Error = err.Error()
			return result
		}

		// TODO: move to some method
		filterData := filter.(map[string]interface{})
		filterHeaders := make([]string, 0, len(filterData))
		filterValues := make([]string, 0, len(filterData))

		for filterName, filterValue := range filterData {
			if strings.HasPrefix(filterName, "_") {
				filterHeaders = append(filterHeaders, "\""+filterName+"\"")
				filterValues = append(filterValues, fmt.Sprintf("%v", filterValue))
				continue
			}
			field, ok := collection.FieldMap[filterName]
			if !ok {
				result.Error = fmt.Sprintf("Field %s doesn't exist in %v.%v", filterName, args["db"], args["collection"])
				return result
			}

			filterHeaders = append(filterHeaders, "\""+filterName+"\"")
			switch field.Type {
			case metadata.Document:
				fieldJson, err := json.Marshal(filterValue)
				if err != nil {
					result.Error = err.Error()
					return result
				}
				filterValues = append(filterValues, "'"+string(fieldJson)+"'")
			case metadata.Text:
				fallthrough
			case metadata.String:
				filterValues = append(filterValues, fmt.Sprintf("'%v'", filterValue))
			default:
				filterValues = append(filterValues, fmt.Sprintf("%v", filterValue))
			}
		}

		whereClause += ","
		for i, header := range filterHeaders {
			whereClause += header + "=" + filterValues[i]
			if i+1 < len(filterHeaders) {
				whereClause += ", "
			}
		}
	}

	sqlQuery := fmt.Sprintf("DELETE FROM %s.%s WHERE _id=%v%s RETURNING *", args["shard_instance"].(string), args["collection"], args["_id"], whereClause)
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
	sqlQuery := fmt.Sprintf("SELECT * FROM %s.%s", args["shard_instance"].(string), args["collection"])

	if _, ok := args["filter"]; ok && args["filter"] != nil {
		recordData := args["filter"].(map[string]interface{})
		meta := s.GetMeta()
		collection, err := meta.GetCollection(args["db"].(string), args["shard_instance"].(string), args["collection"].(string))
		if err != nil {
			result.Error = err.Error()
			return result
		}

		whereParts := make([]string, 0)
		for fieldName, fieldValue := range recordData {
			if strings.HasPrefix(fieldName, "_") {
				whereParts = append(whereParts, fmt.Sprintf(" %s=%v", fieldName, fieldValue))
				continue
			}
			field, ok := collection.FieldMap[fieldName]
			if !ok {
				result.Error = fmt.Sprintf("Field %s doesn't exist in %v.%v", fieldName, args["db"], args["collection"])
				return result
			}

			switch field.Type {
			case metadata.Document:
				// TODO: recurse and add many
				for innerName, innerValue := range fieldValue.(map[string]interface{}) {
					whereParts = append(whereParts, fmt.Sprintf(" \"%s\"->>'%s'='%v'", fieldName, innerName, innerValue))
				}
			case metadata.Text:
				fallthrough
			case metadata.String:
				whereParts = append(whereParts, fmt.Sprintf(" \"%s\"='%v'", fieldName, fieldValue))
			default:
				whereParts = append(whereParts, fmt.Sprintf(" \"%s\"=%v", fieldName, fieldValue))
			}
		}
		if len(whereParts) > 0 {
			sqlQuery += " WHERE " + strings.Join(whereParts, " AND ")
		}
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
			if field, ok := collection.FieldMap[k]; ok && v != nil {
				switch field.Type {
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
