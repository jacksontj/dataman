This is the daemon for router nodes


# Test of a simple get
[jacksontj@localhost ~]$ curl -sd '[{"get": {"db": "test", "table": "user", "id": 5}}]' localhost:8080/v1/data/raw | json_reformat
[
    {
        "return": [
            {
                "id": 5,
                "name": "5"
            }
        ],
        "meta": {
            "datasource": "postgres"
        }
    }
]
