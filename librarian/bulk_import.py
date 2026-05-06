import os
import json
import sqlite3
import sys
from pathlib import Path

try:
    from .analizator import preprocess_and_analyze
except ImportError:
    from analizator import preprocess_and_analyze

BASE_DIR = Path(__file__).resolve().parent

def save_to_db(data, db_path="radio_library.db"):
    """Salvează datele extrase în baza de date SQLite."""
    conn = sqlite3.connect(str(db_path))
    cursor = conn.cursor()
    
    try:
        # T2-06: Inserăm metadatele în tabelul 'songs'[cite: 2]
        cursor.execute('''
            INSERT INTO songs (filepath, title, artist, duration, bpm, music_key, 
                               energy, brightness, danceability, mood_label)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        ''', (
            data.get('filepath', 'Unknown Path'), 
            data['file_info']['title'],
            data['file_info']['artist'],
            data['file_info']['duration_sec'],
            data['audio_features']['bpm'],
            data.get('audio_features', {}).get('key', 'C'),
            data['audio_features']['energy'],
            data['audio_features']['brightness'],
            data['audio_features']['danceability'],
            data['audio_features']['mood_label']
        ))
        conn.commit()
        print(f"✅ Salvat cu succes în DB: {data['file_info']['title']}")
    except sqlite3.IntegrityError:
        print(f"⚠️ Piesa există deja în bază: {data['file_info']['title']}")
    except Exception as e:
        print(f"❌ Eroare la salvarea în DB: {str(e)}")
    finally:
        conn.close()

def import_folder(folder_path, db_path="radio_library.db"):
    """Parcurge folderul și analizează fiecare MP3."""
    if not os.path.isdir(folder_path):
        print(f"Eroare: Folderul '{folder_path}' nu există.")
        return

    mp3_files = [f for f in os.listdir(folder_path) if f.lower().endswith('.mp3')]
    print(f"S-au găsit {len(mp3_files)} fișiere MP3. Începem importul...\n")

    for fisier in mp3_files:
        cale_completa = os.path.join(folder_path, fisier)
        print(f"Analizăm: {fisier}...")
        
        # Apelăm analizatorul tău (T2-04)[cite: 2]
        rezultat_json = preprocess_and_analyze(cale_completa)
        date_procesate = json.loads(rezultat_json)
        
        if date_procesate.get("status") == "success":
            date_procesate['filepath'] = cale_completa
            save_to_db(date_procesate, db_path)
        else:
            print(f"⏭️ Sărit peste {fisier}: {date_procesate.get('message')}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Folosire: python bulk_import.py <cale_catre_folderul_cu_muzica>")
        sys.exit(1)
    
    folder = sys.argv[1]
    
    # Ne asigurăm că baza de date și tabelele există (T2-01)[cite: 2]
    db = sqlite3.connect(str(BASE_DIR / "radio_library.db"))
    if (BASE_DIR / "schema.sql").exists():
        with open(BASE_DIR / "schema.sql", "r", encoding="utf-8") as f:
            db.executescript(f.read())
    db.close()

    # Pornim importul
    import_folder(folder, db_path=str(BASE_DIR / "radio_library.db"))


