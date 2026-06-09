package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	_ "github.com/mattn/go-sqlite3"
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

func analyzeAudio(filePath string) (map[string]interface{}, error) {
	cmd := exec.Command("python3", "/root/agent-radio/librarian/analizator.py", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func callOllama(prompt string, model string) (string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	})

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	return result["response"].(string), nil
}

func saveToDB(db *sql.DB, filePath string, metadata map[string]interface{}) error {
	fileInfo := metadata["file_info"].(map[string]interface{})
	audioFeatures := metadata["audio_features"].(map[string]interface{})

	_, err := db.Exec(`
		INSERT OR IGNORE INTO songs 
		(filepath, title, artist, duration, bpm, energy, brightness, danceability, mood_label)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filePath,
		fileInfo["title"],
		fileInfo["artist"],
		fileInfo["duration_sec"],
		audioFeatures["bpm"],
		audioFeatures["energy"],
		audioFeatures["brightness"],
		audioFeatures["danceability"],
		audioFeatures["mood_label"],
	)
	return err
}

func main() {
	db, err := sql.Open("sqlite3", "/root/agent-radio/librarian/radio_library.db")
	if err != nil {
		fmt.Println("could not open db:", err)
		return
	}
	defer db.Close()

	http.Handle("/", http.FileServer(http.Dir("/root/agent-radio/frontend/agent-radio/dist")))

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

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

		go func() {
			result, err := analyzeAudio("/music/" + header.Filename)
			if err != nil {
				fmt.Println("analysis error:", err)
				return
			}
			saveToDB(db, "/music/"+header.Filename, result)
		}()

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
