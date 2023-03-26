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

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(res)
	})

	counter := 0

	r.HandleFunc("/get/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := vars["id"]

		value, err := cache.Get(key, func(key string) (value any, ttlSecs int, err error) {
			// sleeps mimics operation slowness
			time.Sleep(1 * time.Second)
			fmt.Println("read from db")
			res, err := conn.Get(key)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return nil, 0, fmt.Errorf("something wrong happened when searching: %v", err)
			}
			counter++
			return fmt.Sprintf("%s%v", res.Value, counter), ttlCache, nil
		})

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("something wrong happened when getting from cache: %v", err)))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(value)

	})

	srv := http.Server{
		Addr:    "localhost:8080",
		Handler: r,
	}

	// Set up a channel for handling Ctrl-C, etc
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			select {
			case sig := <-sigchan:
				fmt.Printf("Caught signal %v: terminating\n", sig)
				srv.Close()
				return
			}
		}
	}()

	fmt.Println(srv.ListenAndServe())
}
