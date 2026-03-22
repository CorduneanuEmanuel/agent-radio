# 📻 AI Radio DJ

An autonomous internet radio station powered by local AI. Music plays continuously through an Icecast stream, with an LLM-driven Director agent selecting the next track and a DJ agent writing live banter that is synthesised into voice and mixed in before each transition — all without human intervention.

---

## How It Works

1. **A listener opens the webpage** — an audio stream begins immediately via Icecast
2. **45 seconds before a song ends**, the Go backend wakes up and starts the transition pipeline
3. **The Director agent (Qwen3)** queries the song library and picks the next track based on BPM, mood, and play history
4. **The DJ agent (Llama 3.3)** writes a short voiceover introducing the next song
5. **Chatterbox TTS** converts the text to a `.wav` file
6. **Liquidsoap** ducks the music, plays the DJ voice, then crossfades into the next track
7. The browser UI updates in real time via SSE, showing the current song and what the AI is thinking

If any step in the pipeline is too slow or fails, the system falls back to a direct crossfade — listeners never hear silence.

---

## Architecture

```
Browser (HTMX + SSE)
    │
    ▼
Go Backend ──────────────────────────────────────────┐
    │                                                 │
    ├── os/exec ──► Audio Analyser (aubio + librosa)  │
    │                      │                          │
    │                      ▼                          │
    │                   SQLite                        │
    │                                                 │
    ├── HTTP ────► Ollama (Qwen3)   Director Agent    │
    ├── HTTP ────► Ollama (Llama 3.3) DJ Agent        │
    ├── HTTP ────► Chatterbox TTS  :4123              │
    │                                                 │
    └── Unix Socket ──► Liquidsoap ──► Icecast :8000  │
                                           │          │
                                           ▼          │
                                     <audio> tag ◄────┘
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go |
| Database | SQLite |
| Audio Analysis | aubio (BPM/key), librosa (mood) |
| LLM Inference | Ollama — Qwen3 + Llama 3.3 |
| Text-to-Speech | Chatterbox TTS |
| Stream Engine | Liquidsoap |
| Stream Server | Icecast2 |
| Frontend | HTMX + SSE |

---

---

## User Stories

**As a Listener**

- As a listener, I want to open a webpage and immediately hear music playing, so that I don't have to configure anything to enjoy the station.
  - Given the Icecast stream is running, when a listener opens the page, audio begins playing within 3 seconds without any interaction
  - The page does not require a plugin, download, or account to play audio
  - If the stream is offline, a clear message is shown rather than a broken player
- As a listener, I want to hear a DJ voice introduce the next song before it plays, so that the experience feels like a real radio station rather than a playlist.
- As a listener, I want the music transitions to sound smooth, so that there are no jarring cuts or moments of silence between songs.
- As a listener, I want to see what song is currently playing on the page, so that I can look up tracks I like.
- As a listener, I want to see what the AI is "thinking" in real time (e.g. "Director: picking next track…"), so that the AI curation feels transparent and interesting.
- As a listener, I want the stream to recover automatically if my connection drops briefly, so that I don't have to manually refresh the page.

**As a Station Admin**

- As a station admin, I want to upload an MP3 file through the website, so that new tracks are added to the rotation without me touching the server directly.
  - Given a valid MP3 file, when the admin submits the upload form, a success message appears within 5 seconds
  - The file is saved to the server and appears as available in the song library
  - If the file is not an MP3 or exceeds size limits, an error message is shown and no file is saved
  - The admin does not need SSH or server access to complete the upload
- As a station admin, I want uploaded songs to be automatically analysed for BPM, key, and mood, so that I don't have to tag tracks manually.
- As a station admin, I want the DJ to never play the same song twice in a short window, so that the station doesn't feel repetitive to regular listeners.
  - Given a library of N songs, no song is repeated within the last 10 plays regardless of what the AI selects
  - The exclusion is enforced at the SQLite query level before candidates are sent to the Director agent
  - If the library has fewer than 10 songs, the window shrinks proportionally rather than blocking all playback
- As a station admin, I want the AI to pick songs that flow well together in tempo and mood, so that the station has a consistent energy rather than random jumps.
- As a station admin, I want to be notified if an uploaded file is corrupt or unreadable, so that bad files don't silently sit in the library unused.
- As a station admin, I want the station to keep playing music even if the AI agent is slow or fails, so that listeners never experience dead air.
  - Given Ollama is unresponsive or returns a malformed response, the current song continues playing without interruption
  - The system automatically falls back to a direct crossfade into a SQLite-selected song with no DJ voice
  - The fallback completes within the 45 second window so no silence occurs
  - An error is logged internally but the listener sees and hears nothing unusual

**As the System**

- As the system, I need to begin preparing the next transition 45 seconds before the current song ends, so that the DJ voice and next song are ready before they are needed.
- As the system, I need to fall back to a direct crossfade if TTS generation exceeds the time budget, so that a slow Chatterbox response never causes a gap in the stream.
  - Given Chatterbox has not returned a valid WAV within 40 seconds of being called, the system abandons the TTS request
  - Liquidsoap crossfades directly into the next song without DJ audio
  - The transition is still smooth — no silence, no abrupt cut
  - The failed TTS attempt is logged with a timestamp for debugging
- As the system, I need to validate the WAV file returned by Chatterbox before passing it to Liquidsoap, so that a malformed audio file doesn't crash the playout engine.
- As the system, I need to log every song play to the database with a timestamp, so that the Director agent can make history-aware decisions and avoid repetition.

---

## Team

| Module | Responsibility |
|---|---|
| Frontend | Browser UI, SSE display, upload form |
| Librarian | Audio analysis, SQLite schema, song queries |
| Backend | Go orchestrator, timing engine, all service clients |
| Voice & AI | Ollama prompts, Chatterbox setup, agent validation |
| Sound Tech | Liquidsoap script, Icecast config, crossfade logic |

---

## License

MIT
