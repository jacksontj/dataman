package main

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/client"
	"github.com/jacksontj/dataman/src/client/http"
	"github.com/jacksontj/dataman/src/query"
)

func doExamples(client datamanclient.DatamanClient) {
	ret, err := client.DoQuery(
		map[query.QueryType]query.QueryArgs{
			query.Filter: map[string]interface{}{
				"db":         "example_forum",
				"collection": "user",
				"filter":     map[string]interface{}{},
			},
		},
	)

	fmt.Println(ret, err)
}

func main() {

	client, err := datamanhttp.NewHTTPDatamanClient("http://127.0.0.1:8080/v1/")
	if err != nil {
		logrus.Fatalf("Error NewHTTPDatamanClient: %v", err)
	}
	doExamples(client)

}
