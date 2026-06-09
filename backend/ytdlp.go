package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type youtubeDownloadRequest struct {
	URL   string `json:"url"`
	Query string `json:"query"`
}

type youtubeDownloadResponse struct {
	Message  string                 `json:"message"`
	Path     string                 `json:"path"`
	Analysis map[string]interface{} `json:"analysis,omitempty"`
}

func registerYouTubeDownloadHandler(db *sql.DB) {
	http.HandleFunc("/download-youtube", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method needs to be POST", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "could not read request", http.StatusBadRequest)
			return
		}

		request := youtubeDownloadRequest{
			URL:   strings.TrimSpace(r.FormValue("url")),
			Query: strings.TrimSpace(r.FormValue("query")),
		}

		if request.URL == "" && request.Query == "" {
			_ = json.NewDecoder(r.Body).Decode(&request)
			request.URL = strings.TrimSpace(request.URL)
			request.Query = strings.TrimSpace(request.Query)
		}

		input := strings.TrimSpace(request.URL)
		if input == "" {
			input = strings.TrimSpace(request.Query)
		}

		if input == "" {
			http.Error(w, "missing youtube url or search query", http.StatusBadRequest)
			return
		}

		source, err := resolveYouTubeSource(input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		downloadedPath, err := downloadYouTubeAudio(source)
		if err != nil {
			http.Error(w, fmt.Sprintf("youtube download failed: %v", err), http.StatusInternalServerError)
			return
		}

		result, err := analyzeAudio(downloadedPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("analysis failed: %v", err), http.StatusInternalServerError)
			return
		}

		if status, ok := result["status"].(string); !ok || status != "success" {
			http.Error(w, "analysis returned invalid status", http.StatusInternalServerError)
			return
		}

		if err := saveToDB(db, downloadedPath, result); err != nil {
			http.Error(w, fmt.Sprintf("database save failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(youtubeDownloadResponse{
			Message:  "youtube track downloaded and analyzed successfully",
			Path:     downloadedPath,
			Analysis: result,
		})
	})
}

func resolveYouTubeSource(input string) (string, error) {
	parsedURL, err := url.Parse(input)
	if err == nil && parsedURL.Scheme != "" && parsedURL.Host != "" {
		lowerHost := strings.ToLower(parsedURL.Host)
		if !strings.Contains(lowerHost, "youtube.com") && !strings.Contains(lowerHost, "youtu.be") {
			return "", fmt.Errorf("url must point to youtube")
		}

		return input, nil
	}

	if input == "" {
		return "", fmt.Errorf("missing youtube url or search query")
	}

	return "ytsearch1:" + input, nil
}

func downloadYouTubeAudio(source string) (string, error) {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return "", fmt.Errorf("yt-dlp not found in PATH")
	}

	targetMusicDir := musicDir

	if err := os.MkdirAll(targetMusicDir, 0o755); err != nil {
		return "", err
	}

	cmd := exec.Command(
		"yt-dlp",
		"--no-playlist",
		"--quiet",
		"--no-warnings",
		"--restrict-filenames",
		"-x",
		"--audio-format",
		"mp3",
		"--audio-quality",
		"0",
		"-o",
		filepath.Join(targetMusicDir, "%(title).200s [%(id)s].%(ext)s"),
		"--print",
		"after_move:filepath",
		source,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, strings.TrimSpace(string(output)))
	}

	filePath := lastNonEmptyLine(string(output))
	if filePath == "" {
		return "", fmt.Errorf("yt-dlp completed but did not return a file path")
	}

	if !filepath.IsAbs(filePath) {
		filePath = filepath.Clean(filepath.Join(targetMusicDir, filePath))
	}

	if _, err := os.Stat(filePath); err != nil {
		return "", err
	}

	return filePath, nil
}

func lastNonEmptyLine(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}

	return ""
}