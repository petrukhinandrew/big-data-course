package lesson2

import (
	"io/ioutil"
	"log"
	"net/http"
)

var logger = log.Default()
var body []byte

func ReplaceHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Printf("replaceHandlerError: %s", err.Error())
	}

	body = b
	logger.Printf("/replace: got: %s", string(body))

	w.WriteHeader(200)
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	logger.Printf("/get: response: %s", string(body))

	_, err := w.Write(body)

	if err != nil {
		logger.Printf("getHandlerError: %s", err.Error())
	}

}

func main() {
	http.HandleFunc("/replace", ReplaceHandler)

	http.HandleFunc("/get", GetHandler)

	logger.Println("listening: localhost:8080")

	err := http.ListenAndServe("localhost:8080", nil)
	if err != http.ErrServerClosed {
		log.Printf("listenAndServeError: %s", err.Error())
	}
}

// curl -v -I "content-type: application/json" -d "{'lolkek': 'cheburek'}" localhost:8080/replace
// curl -v localhost:8080/get
