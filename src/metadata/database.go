package metadata

type Database struct {
	Name  string
	Store *DataStore
	//TombstoneMap map[int]*DataStore
	Tables map[string]*Table
}
