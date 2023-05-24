package server

import (
	"github.com/gorilla/mux"
	"net/http"
)

/*
gorilla/mux 으로 rest api 만들기
https://github.com/gorilla/mux
*/

func NewHTTPServer(addr string) *http.Server {
	//httpsrv := newHTTPServer()

	r := mux.NewRouter()

	//r.HandleFunc("/", httpsrv.handleProduce).Method("POST")
	//r.HandleFunc("/", httpsrv.handleConsuume).Method("GET")

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
