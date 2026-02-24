package bot

import (
	"log"
	"net/http"
	"os"
)

func StartHealthServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})

		addr := "0.0.0.0:" + port
		log.Printf("Health server listening on %s", addr)
		_ = http.ListenAndServe(addr, mux)
	}()
}
