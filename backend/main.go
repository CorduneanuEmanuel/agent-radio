package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)


type State int

const (
	Idle State = iota
	Selecting
	Generating
	Ready
	Playing
)

var currentState = Idle

func setState(s State) {
	currentState = s
	states := map[State]string{
		Idle:       "idle",
		Selecting:  "selecting",
		Generating: "generating",
		Ready:      "ready",
		Playing:    "playing",
	}
	fmt.Printf("State: %s\n", states[s])
}


func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Go server is running")
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    if r.Method == http.MethodOptions {
        return
    }

    if r.Method != http.MethodPost {
        http.Error(w, "Method needs to be POST", http.StatusMethodNotAllowed)
        return
    }

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "could not read file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    dst, err := os.Create("/music/" + header.Filename)
    if err != nil {
        http.Error(w, "could not save file", http.StatusInternalServerError)
        return
    }
    defer dst.Close()
    io.Copy(dst, file)

    fmt.Fprintln(w, "upload successful")
})

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		fmt.Fprintf(w, "data: {\"status\": \"connected\"}\n\n")
		w.(http.Flusher).Flush()

		<-r.Context().Done()
	})

	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}
