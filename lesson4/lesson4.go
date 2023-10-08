package lesson4

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var logger = log.Default()

var localTransactionCounter uint64 = 0
var ltcLock sync.Mutex

var source string = `"Petrukhin"`

var peers []string

var vclock = map[string]uint64{source: localTransactionCounter}

func updateVclock() {
	vclock[source] = localTransactionCounter
}

func getLTC() uint64 {
	ltcLock.Lock()
	defer ltcLock.Unlock()
	return localTransactionCounter
}

func incrementLTC() {
	ltcLock.Lock()
	defer ltcLock.Unlock()
	localTransactionCounter += 1
}

type Transaction struct {
	Source  string
	Id      uint64
	Payload string
}

type LolKek struct {
	Source  string
	Id      uint64
	Payload interface{}
}

func (t Transaction) String() string {
	return fmt.Sprintf(`{"Source": %s, "Id": %d, "Payload": %s}`, t.Source, t.Id, t.Payload)
}

func (t Transaction) Bytes() []byte {
	return []byte(t.String())
}

type TransactionManager struct {
	state []byte
	snap  []byte
	wal   []string
	lock  sync.Mutex
	queue chan Transaction
}

func (m *TransactionManager) CreateTransaction(patch string) Transaction {
	ltc := getLTC()
	return Transaction{source, ltc, patch}
}

func (m *TransactionManager) StartSnapshoting() {
	freq := time.Second * 60
	timer := time.NewTimer(freq)
	for now := range timer.C {
		logger.Printf("snapshot %s", now)
		m.lock.Lock()
		m.snap = m.state
		m.wal = nil
		m.lock.Unlock()
	}
}
func (m *TransactionManager) StartManaging() {
	for t := range m.queue {

		if t.Id < localTransactionCounter {
			continue
		}
		patch, err := jsonpatch.DecodePatch([]byte(t.Payload))

		if err != nil {
			logger.Printf("Apply transaction: %s", err)
			continue
		}
		logger.Printf("Apply transaction: %s", t.Payload)

		m.lock.Lock()
		initData := m.state
		m.lock.Unlock()

		patched, err := patch.ApplyWithOptions(initData, &jsonpatch.ApplyOptions{EnsurePathExistsOnAdd: true})
		if err != nil {
			logger.Printf("Apply transaction: %s", err)
			continue
		}
		m.lock.Lock()
		m.wal = append(m.wal, t.String())
		m.state = patched
		m.lock.Unlock()

		incrementLTC()
	}
}

var manager = TransactionManager{state: []byte("{}"), queue: make(chan Transaction, 100)}

type HttpHandler struct {
}

//go:embed statics/*
var statics embed.FS

func (h *HttpHandler) Test(w http.ResponseWriter, r *http.Request) {
	kek, err := fs.ReadFile(statics, "statics/index.html")
	if err != nil {
		logger.Fatalf("Test: %s", err)
	} else {
		logger.Println("Test: OK")
	}
	w.WriteHeader(http.StatusOK)
	w.Write(kek)
}

func (h *HttpHandler) Vclock(w http.ResponseWriter, r *http.Request) {
	updateVclock()

	resp, err := json.Marshal(vclock)

	if err != nil {
		logger.Printf("Vclock: %s", err)
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *HttpHandler) Replace(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Printf("Replace: %s", err)
		return
	}
	patch := string(data)
	logger.Printf("Replace: %s", patch)
	manager.queue <- manager.CreateTransaction(patch)
}

func (h *HttpHandler) Get(w http.ResponseWriter, r *http.Request) {
	manager.lock.Lock()
	resp := manager.state
	manager.lock.Unlock()
	logger.Printf("Get: %s", resp)

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(resp))
	if err != nil {
		logger.Printf("Get: %s", err)
	}
}

func (h *HttpHandler) Ws(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true, OriginPatterns: []string{"*"}})
	if err != nil {
		logger.Printf("WS: %s", err)
	}
	defer c.Close(websocket.StatusInternalError, "blabla")

	logger.Println("WS: OK")
	for _, transaction := range manager.wal {
		wsjson.Write(ctx, c, transaction)
	}
	c.Close(websocket.StatusNormalClosure, "")
}

func Dial() {
	for _, peer := range peers {
		go func(peer string) {
			time.Sleep(time.Second * 3)
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			c, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s/ws", peer), nil)
			if err != nil {
				logger.Printf("Dial: %s", err)
				return
			}

			defer c.Close(websocket.StatusInternalError, "blabla")

			logger.Printf("Dialing %s", peer)

			var v interface{}
			if err := wsjson.Read(ctx, c, &v); err != nil {
				logger.Printf("Dial ERR : %s ! %v", err, v)
				return
			}
			rawTransaction := fmt.Sprintf("%v", v)
			var transaction LolKek

			if err := json.Unmarshal([]byte(rawTransaction), &transaction); err != nil {
				logger.Printf("Dial ERR : %s ! %s", err, rawTransaction)
				return
			}

			payload, err := json.Marshal(transaction.Payload)
			if err != nil {
				logger.Printf("Dial ERR : %s", err)
				return
			}
			response := Transaction{transaction.Source, transaction.Id, string(payload)}
			logger.Printf("LOKKEK: %s", response.String())
			manager.queue <- response

		}(peer)
	}
}

func RunServer(port string) {
	if port == "8080" {
		// peers = []string{"localhost: 8081"}
	} else {
		peers = []string{"localhost:8080"}
	}
	server := http.NewServeMux()
	handler := HttpHandler{}

	server.HandleFunc("/test", handler.Test)
	server.HandleFunc("/vclock", handler.Vclock)
	server.HandleFunc("/replace", handler.Replace)
	server.HandleFunc("/get", handler.Get)
	server.HandleFunc("/ws", handler.Ws)
	logger.Printf("listening: localhost:%s", port)

	go manager.StartManaging()
	go manager.StartSnapshoting()
	go Dial()

	if err := http.ListenAndServe("localhost:"+port, server); err != http.ErrServerClosed {
		logger.Printf("listenAndServeError: %s", err.Error())
	}
}
