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
	"sync"
	"time"

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

var (
	voteYes      int
	voteNo       int
	voteMutex    sync.Mutex
	songStart    time.Time
	songDuration float64
	currentState = Idle
)

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
		result, err := analyzeAudio("/music/" + header.Filename)
		if err != nil {
			fmt.Println("analysis error:", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "error"})
			return
		}
		saveToDB(db, "/music/"+header.Filename, result)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	http.HandleFunc("/track-started", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("TRACK STARTED CALLED", r.Method)
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var metadata map[string]interface{}
		json.NewDecoder(r.Body).Decode(&metadata)
		fmt.Println("Track started:", metadata)
		fp, _ := metadata["filename"].(string)
		voteMutex.Lock()
		voteYes = 0
		voteNo = 0
		voteMutex.Unlock()
		var duration float64
		err := db.QueryRow("SELECT duration FROM songs WHERE filepath = ?", fp).Scan(&duration)
		if err != nil {
			fmt.Println("song not found in db:", err)
			w.WriteHeader(http.StatusOK)
			return
		}
		songStart = time.Now()
		songDuration = duration
		setState(Playing)
		go func() {
			half := time.Duration(duration/2) * time.Second
			time.Sleep(half)
			fmt.Println("Halfway - locking votes in 10s")
			time.Sleep(10 * time.Second)
			voteMutex.Lock()
			result := "like"
			if voteNo > voteYes {
				result = "dislike"
			}
			voteMutex.Unlock()
			fmt.Println("Vote result:", result)
			remaining := time.Duration(duration)*time.Second - time.Since(songStart) - 45*time.Second
			if remaining > 0 {
				time.Sleep(remaining)
			}
			setState(Selecting)
			fmt.Println("Triggering pipeline, vote was:", result)
		}()
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		fmt.Fprintf(w, "data: {\"status\": \"connected\"}\n\n")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	})

	http.Handle("/", http.FileServer(http.Dir("/root/agent-radio/frontend/agent-radio/dist")))

	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}
