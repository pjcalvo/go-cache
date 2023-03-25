package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pjcalvo/go-cache/cache"
	"github.com/pjcalvo/go-cache/connection"
	"github.com/pjcalvo/go-cache/message"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const ttlCache = 15

func main() {

	cache := cache.InitCache()

	conn, err := connection.Init()
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/new/{value}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		value := vars["value"]

		id := uuid.New()
		res := connection.Response{
			Id:    id.String(),
			Value: value,
		}

		err := conn.Add(res)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("something wrong happened when inserting: %v", err)))
			return
		}

		json.NewEncoder(w).Encode(res)
		w.WriteHeader(http.StatusAccepted)
	})

	prod, err := message.InitProducer()
	if err != nil {
		panic(err)
	}

	r.HandleFunc("/get/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := vars["id"]

		value, err := cache.Get(key, func(key string) (value any, ttlSecs int, err error) {
			time.Sleep(10 * time.Second)
			res, err := conn.Get(key)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return nil, 0, fmt.Errorf("something wrong happened when searching: %v", err)
			}
			err = prod.Send(message.Message{MessageID: key})
			if err != nil {
				return nil, 0, fmt.Errorf("something wrong happened when writing message: %v", err)
			}
			return res.Value, ttlCache, nil
		})

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("something wrong happened when getting from cache: %v", err)))
			return
		}

		json.NewEncoder(w).Encode(value)
		w.WriteHeader(http.StatusAccepted)

	})

	srv := http.Server{
		Addr:    "localhost:8080",
		Handler: r,
	}

	// start kafka consumer
	cons, err := message.InitConsumer()
	// Set up a channel for handling Ctrl-C, etc
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			select {
			case sig := <-sigchan:
				fmt.Printf("Caught signal %v: terminating\n", sig)
				srv.Close()
				prod.Close()
				cons.Close()
				return
			default:
				cons.Read()
			}
		}
	}()

	fmt.Println(srv.ListenAndServe())
}
