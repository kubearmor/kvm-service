package main

import (
	"fmt"
	"net/http"
)

const ServerPort = "3000"

func handler(res http.ResponseWriter, req *http.Request) {
	data := []byte(req.URL.Path[1:])
	res.Write(data)
}

func main() {
	fmt.Println("🚀 HTTP echo server listening on port " + ServerPort)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+ServerPort, nil)
}
