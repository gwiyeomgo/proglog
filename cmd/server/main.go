package main

import (
	"github.com/gwiyeomgo/proglog/internal/server"
	"log"
)

func main() {
	srv := server.NewHTTPServer(":8080")
	log.Fatal(srv.ListenAndServe())
}

/*

curl -X POST localhost:8080 -d \ '{"record":{"value":"REWRWERWEFSD"}}'
curl -X GET localhost:8080 -d \ '{"offset":0}'
*/
