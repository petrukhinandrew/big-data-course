package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

var logger = log.Default()
var body []byte

func replaceHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Printf("replaceHandlerError: %s", err.Error())
	}

	body = b
	logger.Printf("Accepted replace with %s", string(body))

	w.WriteHeader(200)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	logger.Printf("Accepted get, response is %s", string(body))

	_, err := w.Write(body)

	if err != nil {
		logger.Printf("getHandlerError: %s", err.Error())
	}

}

func main() {
	http.HandleFunc("/replace", replaceHandler)

	http.HandleFunc("/get", getHandler)

	logger.Println("listening on localhost:8080")

	err := http.ListenAndServe("localhost:8080", nil)
	if err != http.ErrServerClosed {
		log.Printf("serveError: %s", err.Error())
	}
}

// curl -v -I "content-type: application/json" -d "{'lolkek': 'cheburek'}" localhost:8080/replace
// curl -v localhost:8080/get
