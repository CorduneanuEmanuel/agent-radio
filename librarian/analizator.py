import librosa
import numpy as np
import sys
import json
import warnings
from tinytag import TinyTag

# Ignorăm avertismentele matematice pentru un output JSON curat
warnings.filterwarnings('ignore')

def preprocess_and_analyze(file_path):
    try:
        # 1. PREPROCESARE RAPIDĂ: Metadate ID3 și Validare
        tag = TinyTag.get(file_path)
        duration = tag.duration if tag.duration else 0.0
        
        # T2-09: Gestionare caz limită - fișier prea scurt
        if duration < 10.0:
            return json.dumps({"status": "error", "message": "Fisier prea scurt pentru analiza."})

        # 2. ÎNCĂRCARE ȘI PREPROCESARE AUDIO (Librosa)
        # Resantionăm la 22050 Hz (standard) și convertim în mono pentru viteză
        y, sr = librosa.load(file_path, sr=22050, mono=True, duration=45.0)

        # PREPROCESARE: Eliminăm tăcerea (Silence Removal) de la început/sfârșit
        # Astfel, BPM-ul și energia nu vor fi afectate de secunde de liniște [cite: 39]
        y_trimmed, _ = librosa.effects.trim(y, top_db=20)

        # PREPROCESARE: Normalizare (Aducem volumul la un nivel standard)
        if len(y_trimmed) > 0:
            y_norm = librosa.util.normalize(y_trimmed)
        else:
            y_norm = y # Fallback dacă trim a eșuat

        # 3. EXTRAGEREA METADATELOR BRUTE (Feature Extraction)
        
        # BPM (Tempo)
        tempo, _ = librosa.beat.beat_track(y=y_norm, sr=sr)
        bpm = float(tempo[0]) if isinstance(tempo, np.ndarray) else float(tempo)

        # Energy (RMS) - Indică intensitatea piesei
        rms = librosa.feature.rms(y=y_norm)
        energy = float(np.mean(rms))

        # Spectral Centroid (Brightness) - Indică dacă sunetul e "luminos" sau "bass-y"
        centroid = librosa.feature.spectral_centroid(y=y_norm, sr=sr)
        brightness = float(np.mean(centroid))

        # Danceability Proxy (bazat pe onset strength) [cite: 8]
        onset_env = librosa.onset.onset_strength(y=y_norm, sr=sr)
        danceability = float(np.mean(onset_env))

        # 4. LOGICA PENTRU AGENTUL AI (Mood Labeling)
        # Aceste etichete ajută Ollama să facă recomandări [cite: 23, 35]
        mood = "neutral"
        if bpm > 125 and energy > 0.15:
            mood = "energetic"
        elif bpm < 90 and energy < 0.1:
            mood = "chill"
        elif energy > 0.2:
            mood = "aggressive/heavy"

        # 5. CONSTRUCȚIA REZULTATULUI PENTRU GO/SQLITE
        rezultat = {
            "status": "success",
            "file_info": {
                "title": tag.title if tag.title else "Unknown",
                "artist": tag.artist if tag.artist else "Unknown",
                "duration_sec": round(duration, 2)
            },
            "audio_features": {
                "bpm": round(bpm, 1),
                "energy": round(energy, 4),
                "brightness": round(brightness, 2),
                "danceability": round(danceability, 2),
                "mood_label": mood
            }
        }
        return json.dumps(rezultat)

    except Exception as e:
        # T2-09: Gestionare fișiere corupte [cite: 36, 69]
        return json.dumps({"status": "error", "message": str(e)})

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"status": "error", "message": "Cale fisier lipsa."}))
        sys.exit(1)
    
    print(preprocess_and_analyze(sys.argv[1]))