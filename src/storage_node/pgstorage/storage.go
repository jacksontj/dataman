package pgstorage

import (
	"database/sql"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/query"
	_ "github.com/lib/pq"
)

type StorageConfig struct {
	// How to connect to postgres
	PGString string `yaml:"pg_string"`
}

type Storage struct {
	config *StorageConfig
	db     *sql.DB
}

func (s *Storage) Init(c map[string]interface{}) error {
	var err error

	if val, ok := c["pg_string"]; ok {
		s.config = &StorageConfig{val.(string)}
	} else {
		return fmt.Errorf("Invalid config")
	}

	s.db, err = sql.Open("postgres", s.config.PGString)
	if err != nil {
		return err
	}
	return nil
}

// TODO: find a nicer way to do this, this is a mess
func (s *Storage) doQuery(query string) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
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
	rows, err := s.doQuery(fmt.Sprintf("SELECT * FROM public.%s WHERE id=%v", args["table"], args["id"]))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// TODO: error if there is more than one result

	result.Return = rows
	return result
}

func (s *Storage) Set(query.QueryArgs) *query.Result {
	return nil
}

func (s *Storage) Delete(query.QueryArgs) *query.Result {
	return nil
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
	sqlQuery := fmt.Sprintf("SELECT * FROM public.%s WHERE", args["table"])

	// TODO: validate the query before running (right now if "fields" is missing this exits)
	// TODO: again without so much string concat
	for columnName, columnValue := range args["fields"].(map[string]interface{}) {
		logrus.Infof("%v %v", columnName, columnValue)
		switch typedValue := columnValue.(type) {
		// TODO: define what we want to do here -- not sure if we want to have "=" here,
		// and if we do, we might want to just be consistent with that markup
		// if the value is a list it is something like ["=", 5] (which is just defining a comparator)
		case []interface{}:
			logrus.Infof("not-yet-implemented list of thing %v", typedValue)
		case interface{}:
			sqlQuery = sqlQuery + fmt.Sprintf(" %s='%v'", columnName, columnValue)
		default:
			result.Error = fmt.Sprintf("Error parsing field %s", columnName)
			return result
		}
	}

	rows, err := s.doQuery(sqlQuery)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Return = rows
	return result

	return result
}
