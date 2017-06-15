package main

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/client/http"
	"github.com/jacksontj/dataman/src/query"
)

func main() {
	client, err := datamanhttp.NewHTTPDatamanClient("http://127.0.0.1:8080/v1/")
	if err != nil {
		logrus.Fatalf("Error NewHTTPDatamanClient: %v", err)
	}

	ret := client.DoQuery(
		map[query.QueryType]query.QueryArgs{
			query.Filter: map[string]interface{}{
				"db":         "example_forum",
				"collection": "user",
				"filter":     map[string]interface{}{},
			},
		},
	)

	fmt.Println(ret)

}
