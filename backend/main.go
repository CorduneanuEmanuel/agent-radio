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
	"path/filepath"
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

var currentState = Idle

var (
	analyzerPath       = getEnv("ANALYZER_PATH", "../librarian/analizator.py")
	dbPath             = getEnv("RADIO_DB_PATH", "../librarian/radio_library.db")
	musicDir           = getEnv("MUSIC_DIR", "../music")
	frontendDistDir    = os.Getenv("FRONTEND_DIST_DIR")
	frontendFallbackDir = getEnv("FRONTEND_FALLBACK_DIR", "../frontend")
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func resolveFrontendDir() string {
	if info, err := os.Stat(frontendFallbackDir); err == nil && info.IsDir() {
		return frontendFallbackDir
	}
	if frontendDistDir != "" {
		if info, err := os.Stat(frontendDistDir); err == nil && info.IsDir() {
			return frontendDistDir
		}
	}
	if info, err := os.Stat("../frontend/agent-radio/dist"); err == nil && info.IsDir() {
		return "../frontend/agent-radio/dist"
	}
	return frontendDistDir
}

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
	cmd := exec.Command("python3", analyzerPath, filePath)
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
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Println("could not open db:", err)
		return
	}
	defer db.Close()

	if err := os.MkdirAll(musicDir, 0o755); err != nil {
		fmt.Println("could not create music dir:", err)
		return
	}

	frontendDir := resolveFrontendDir()
	fmt.Printf("Serving frontend from: %s\n", frontendDir)
	fmt.Printf("Using music dir: %s\n", musicDir)
	fmt.Printf("Using analyzer path: %s\n", analyzerPath)

	registerYouTubeDownloadHandler(db)

	http.Handle("/", http.FileServer(http.Dir(frontendDir)))

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

		safeFileName := filepath.Base(header.Filename)
		uploadedPath := filepath.Join(musicDir, safeFileName)

		dst, err := os.Create(uploadedPath)
		if err != nil {
			http.Error(w, "could not save file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)

		result, err := analyzeAudio(uploadedPath)
		if err != nil {
			fmt.Println("analysis error:", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "error"})
			return
		}
		saveToDB(db, uploadedPath, result)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	var (
		voteYes      int
		voteNo       int
		voteMutex    sync.Mutex
		songStart    time.Time
	)

	http.HandleFunc("/track-started", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}

		var metadata map[string]interface{}
		json.NewDecoder(r.Body).Decode(&metadata)

		// Get filepath from Liquidsoap metadata
		filepath, _ := metadata["filename"].(string)
		fmt.Println("Track started:", filepath)

		// Reset votes
		voteMutex.Lock()
		voteYes = 0
		voteNo = 0
		voteMutex.Unlock()

		// Look up duration in SQLite
		var duration float64
		err := db.QueryRow("SELECT duration FROM songs WHERE filepath = ?", filepath).Scan(&duration)
		if err != nil {
			fmt.Println("song not found in db:", err)
			w.WriteHeader(http.StatusOK)
			return
		}

		songStart = time.Now()
		setState(Playing)

		// Start timers in background
		go func() {
			// Halfway point - lock voting
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

			// Pipeline trigger at duration - 45s
			remaining := time.Duration(duration)*time.Second - time.Since(songStart) - 45*time.Second
			if remaining > 0 {
				time.Sleep(remaining)
			}

			setState(Selecting)
			fmt.Println("Triggering pipeline, vote was:", result)
			// TODO: call Ollama here
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

	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("server stopped:", err)
	}
}
