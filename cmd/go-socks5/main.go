package main

import (
	"bytes"
	"encoding/json"
	"github.com/Chise1/go-socks5"
	"io"
	"net/http"
	"os"
)

func main() {
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}
	var configInfo socks5.ConfigInfo
	open, err := os.Open("conf/config.json")
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(open).Decode(&configInfo)
	if err != nil {
		panic(err)
	}
	err = socks5.Init(configInfo.Socks)
	if err != nil {
		panic(err)
	}
	go func() {
		// Create SOCKS5 proxy on localhost port 8000
		if err := server.ListenAndServe("tcp", configInfo.Addr); err != nil {
			panic(err)
		}
	}()

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var newConfigInfo []socks5.Socks
			err := json.NewDecoder(r.Body).Decode(&newConfigInfo)
			if err != nil {
				http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
				return
			}
			err = socks5.Init(newConfigInfo)
			if err != nil {
				http.Error(w, "Failed to init socks servers", http.StatusBadRequest)
				return
			}
			configInfo.Socks = newConfigInfo
			marshal, _ := json.Marshal(configInfo)
			os.WriteFile("conf/config.json", marshal, 0644)
			response := map[string]string{
				"status":  "success",
				"message": "JSON received and processed successfully",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else if r.Method == "GET" {
			file, err := os.Open("conf/index.html")
			if err != nil {
				return
			}
			all, err := io.ReadAll(file)
			if err != nil {
				return
			}
			marshal, _ := json.Marshal(configInfo.Socks)
			all = bytes.Replace(all, []byte("{value}"), marshal, -1)
			w.Write(all)
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/html")
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	http.ListenAndServe(":8080", nil)
}
