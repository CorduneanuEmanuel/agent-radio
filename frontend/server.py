import json
import os
import sqlite3
import sys
from pathlib import Path

from flask import Flask, request
from werkzeug.utils import secure_filename

BASE_DIR = Path(__file__).resolve().parent
ROOT_DIR = BASE_DIR.parent
LIBRARIAN_DIR = ROOT_DIR / "librarian"
UPLOAD_DIR = ROOT_DIR / "melodii"
DB_PATH = LIBRARIAN_DIR / "radio_library.db"
SCHEMA_PATH = LIBRARIAN_DIR / "schema.sql"

sys.path.insert(0, str(ROOT_DIR))

from librarian.analizator import preprocess_and_analyze
from librarian.bulk_import import save_to_db

app = Flask(__name__)


def init_db():
    if not SCHEMA_PATH.exists():
        return

    DB_PATH.parent.mkdir(parents=True, exist_ok=True)
    with sqlite3.connect(DB_PATH) as conn:
        with SCHEMA_PATH.open("r", encoding="utf-8") as schema_file:
            conn.executescript(schema_file.read())

# Citim fișierul tău HTML
def get_radio_html():
    with open(BASE_DIR / "Radio.html", "r", encoding="utf-8") as f:
        return f.read()

@app.route("/")
def index():
    return get_radio_html()

def render_status_card(title, message, tone="blue"):
    accent_map = {
        "blue": "#3b82f6",
        "emerald": "#10b981",
        "red": "#ef4444",
        "amber": "#f59e0b",
    }
    accent = accent_map.get(tone, accent_map["blue"])
    return f"""
    <div style="background: #1e293b; padding: 15px; border-radius: 12px; border-left: 4px solid {accent}; color: #e2e8f0;">
        <h4 style="color: {accent}; margin: 0 0 8px 0;">{title}</h4>
        <div style="font-size: 0.95em; line-height: 1.5;">{message}</div>
    </div>
    """


@app.route("/upload", methods=["POST"])
def upload():
    uploaded_file = request.files.get("file")

    if uploaded_file is None or uploaded_file.filename == "":
        return render_status_card("Lipsește fișierul", "Selectează un fișier audio înainte de analiză.", "red")

    filename = secure_filename(uploaded_file.filename)
    if not filename.lower().endswith(".mp3"):
        return render_status_card("Format invalid", "Momentan acceptăm doar fișiere .mp3.", "amber")

    UPLOAD_DIR.mkdir(parents=True, exist_ok=True)
    file_path = UPLOAD_DIR / filename
    uploaded_file.save(file_path)

    rezultat_json = preprocess_and_analyze(str(file_path))
    rezultat = json.loads(rezultat_json)

    if rezultat.get("status") != "success":
        return render_status_card(
            "Eroare la analiză",
            f"{filename}: {rezultat.get('message', 'necunoscut')}",
            "red",
        )

    rezultat["filepath"] = str(file_path)
    save_to_db(rezultat, db_path=str(DB_PATH))

    file_info = rezultat["file_info"]
    audio_features = rezultat["audio_features"]

    return f"""
    <div style="background: #0f172a; padding: 16px; border-radius: 12px; border: 1px solid #334155; color: #e2e8f0;">
        <h4 style="color: #22c55e; margin: 0 0 12px 0;">Analiză finalizată</h4>
        <p style="margin: 0 0 8px 0;">Fișier: <strong>{filename}</strong></p>
        <p style="margin: 0 0 8px 0;">Titlu: <strong>{file_info['title']}</strong> | Artist: <strong>{file_info['artist']}</strong></p>
        <p style="margin: 0 0 8px 0;">Durată: <strong>{file_info['duration_sec']} sec</strong></p>
        <p style="margin: 0 0 8px 0;">BPM: <strong>{audio_features['bpm']}</strong> | Energie: <strong>{audio_features['energy']}</strong></p>
        <p style="margin: 0;">Mood: <strong>{audio_features['mood_label']}</strong></p>
    </div>
    """


@app.route("/analyze", methods=["POST"])
def analyze():
    return upload()

if __name__ == "__main__":
    init_db()
    print("Serverul pornește pe http://localhost:5000")
    app.run(debug=True, port=5000)