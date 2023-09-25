package lesson3

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type Request int8

const (
	replace Request = 1
	get     Request = 0
)

type Transaction struct {
	r Request
	b []byte
}

type TransactionHandler struct {
	journal  [][]byte
	snapshot []byte
	queue    chan Transaction
	storage  *Storage
}
type Storage struct {
	body   []byte
	body_m sync.Mutex
}

var logger = log.Default()

func (h *TransactionHandler) GetBody() []byte {
	h.storage.body_m.Lock()
	defer h.storage.body_m.Unlock()
	return h.storage.body
}

func (h *TransactionHandler) UpdateBody(newBody []byte) {
	h.storage.body_m.Lock()
	defer h.storage.body_m.Unlock()
	h.storage.body = newBody
}

func (h *TransactionHandler) HandleTransactions() {
	for nxt := range h.queue {
		if nxt.r == replace {
			h.journal = append(h.journal, nxt.b)
			h.UpdateBody(nxt.b)
			logger.Printf("transaction %d with '%s'", nxt.r, string(nxt.b))
		}
	}
}
func (h *TransactionHandler) LogSnapshot(now time.Time) {
	logger.Printf("snapshot: %s at %d:%d:%d", string(h.snapshot), now.Day(), now.Hour(), now.Minute())
	for a, b := range h.journal {
		logger.Printf("%d %s", a, string(b))
	}
}
func (h *TransactionHandler) StartSnapshoting() {
	freq := time.Second * 60
	timer := time.NewTimer(freq)
	for now := range timer.C {
		b := h.GetBody()
		h.snapshot = b
		h.LogSnapshot(now)
		h.journal = nil
		timer.Reset(freq)
	}
}

func (h *TransactionHandler) SubmitTransaction(t Transaction) {
	h.queue <- t
}

func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	b := h.GetBody()

	logger.Printf("/get: %s", b)

	if _, err := w.Write(b); err != nil {
		logger.Printf("getHandlerError: %s", err.Error())
	}

	h.SubmitTransaction(Transaction{get, b})
}

func (h *TransactionHandler) Replace(w http.ResponseWriter, r *http.Request) {

	logger.Print("/replace: ")

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		logger.Printf("with error: %s", err.Error())
	} else {
		logger.Printf("with %s", string(b))
	}

	h.SubmitTransaction(Transaction{replace, b})

	w.WriteHeader(200)
}

func RunServer() {
	server := http.NewServeMux()
	storage := Storage{}
	handler := TransactionHandler{queue: make(chan Transaction, 100), storage: &storage}

	server.HandleFunc("/replace", handler.Replace)
	server.HandleFunc("/get", handler.Get)

	logger.Println("listening: localhost:8080")

	go handler.HandleTransactions()
	go handler.StartSnapshoting()

	if err := http.ListenAndServe("localhost:8080", server); err != http.ErrServerClosed {
		logger.Printf("listenAndServeError: %s", err.Error())
	}

}
