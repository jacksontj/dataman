package main

import (
	"encoding/json"
	"fmt"
)

type Query string
type QueryArgs map[string]interface{}

const (
	Filter Query = "filter"
)

func main() {
	a := []byte(`{"filter": {"table": "user", "fields": {"id": ["=", 5]}}}`)

	var tmp map[Query]QueryArgs

	if err := json.Unmarshal(a, &tmp); err != nil {
		panic(err)
	}
	fmt.Println(tmp)

}
