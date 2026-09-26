package main

import (
	"io"
	"net/http"
)

func handleHealthz(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")

	res.WriteHeader(200)

	io.WriteString(res, "OK")
}
