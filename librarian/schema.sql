-- Tabelul principal pentru melodiile analizate (T2-01)
CREATE TABLE IF NOT EXISTS songs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filepath TEXT UNIQUE NOT NULL,
    title TEXT,
    artist TEXT,
    duration REAL,
    bpm REAL,
    music_key TEXT,
    energy REAL,
    brightness REAL,
    danceability REAL,
    mood_label TEXT,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Tabelul pentru istoricul redărilor (T2-08)
-- Ajută AI-ul să nu repete piese recent jucate
CREATE TABLE IF NOT EXISTS play_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    song_id INTEGER,
    played_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (song_id) REFERENCES songs(id)
);