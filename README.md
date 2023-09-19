# Crucialy important: this repo is sponsored by VK!

# big-data-course

Repo with notes and homeworks for big data technologies course

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

var body []byte
var snapshot []byte
var journal [][]byte
var queue = make(chan []byte)

func replaceHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}
	queue <- b
	// body = b

	w.WriteHeader(200)
}
func getHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(body)
}
func main() {
	http.HandleFunc("/replace", replaceHandler)

	http.HandleFunc("/get", getHandler)
	fmt.Println("listening on localhost:8080")
	http.ListenAndServe("localhost:8080", nil)
}

// curl -v -I "content-type: application/json" -d "{'lolkek': 'cheburek'}" localhost:8080/reset
// curl -v localhost:8080/get
