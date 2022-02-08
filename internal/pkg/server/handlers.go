package server

import (
	"fmt"
	"net/http"
)

func registerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/update/", updateHandler)
}

func updateHandler(_ http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)
}
