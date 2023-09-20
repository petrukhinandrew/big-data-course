package lesson3

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Request int8

const (
	replace Request = 1
	get             = 0
)

type Transaction struct {
	r Request
	b []byte
}

var logger = log.Default()
var body []byte
var queue = make(chan Transaction, 100)

var journal [][]byte
var snapshot []byte

func HandleTransactions(q chan Transaction) {
	for nxt := range q {
		if nxt.r == replace {
			journal = append(journal, nxt.b)
			body = nxt.b
			logger.Printf("transaction %d with '%s'", nxt.r, string(nxt.b))
		}
	}
}

func MakeSnapshot() {
	freq := time.Second * 60
	timer := time.NewTimer(freq)
	for tmp := range timer.C {

		snapshot = body
		logger.Printf("snapshot: %d:%d:%d", tmp.Day(), tmp.Hour(), tmp.Minute())
		for a, b := range journal {
			logger.Printf("%d %s", a, string(b))
		}
		journal = nil
		timer.Reset(freq)
	}
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	logger.Printf("/get: %s", body)

	_, err := w.Write(body)

	if err != nil {
		logger.Printf("getHandlerError: %s", err.Error())
	}

	queue <- Transaction{get, body}
}

func ReplaceHandler(w http.ResponseWriter, r *http.Request) {
	logger.Print("/replace: ")

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		logger.Printf("with error: %s", err.Error())
	} else {
		logger.Printf("with %s", string(b))
	}

	queue <- Transaction{replace, b}

	w.WriteHeader(200)
}

func RunServer() {
	server := http.NewServeMux()
	server.HandleFunc("/replace", ReplaceHandler)
	server.HandleFunc("/get", GetHandler)

	logger.Println("listening: localhost:8080")

	go HandleTransactions(queue)
	go MakeSnapshot()

	if err := http.ListenAndServe("localhost:8080", server); err != http.ErrServerClosed {
		logger.Printf("listenAndServeError: %s", err.Error())
	}

}
