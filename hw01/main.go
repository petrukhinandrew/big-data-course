package main

import (
	"io/ioutil"
	"net/http"
)

var body []byte

func main() {
	http.HandleFunc("/replace", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		body = b
		w.WriteHeader(200)
	})
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
	})

	http.ListenAndServe("localhost:8080", nil)
}

// curl -v -I "content-type: application/json" -d "{'lolkek': 'cheburek'}" localhost:8080/reset
// curl -v localhost:8080/get
