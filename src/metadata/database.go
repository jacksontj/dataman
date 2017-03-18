package metadata

func NewDatabase(name string) *Database {
	return &Database{
		Name:   name,
		Tables: make(map[string]*Table),
	}
}

type Database struct {
	Name  string     `json:"name"`
	Store *DataStore `json:"store,omitempty"`
	//TombstoneMap map[int]*DataStore
	Tables map[string]*Table `json:"tables"`
}

func (d *Database) ListTables() []string {
	tables := make([]string, 0, len(d.Tables))
	for name, _ := range d.Tables {
		tables = append(tables, name)
	}
	return tables
}
