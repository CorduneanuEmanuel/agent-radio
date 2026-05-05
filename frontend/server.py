from flask import Flask, request, render_template_string

app = Flask(__name__)

# Citim fișierul tău HTML
def get_radio_html():
    with open("Radio.html", "r") as f:
        return f.read()

@app.route("/")
def index():
    return get_radio_html()

# Aceasta este ruta care lipsea (cea care dădea 404)
@app.route("/analyze", methods=["POST"])
def analyze():
    # Preluăm calea fișierului trimisă prin HTMX
    file_path = request.form.get("file_path")
    
    print(f"DEBUG: Am primit calea: {file_path}")
    
    # Aici vei integra Essentia ulterior. 
    # Momentan trimitem un răspuns înapoi ca să vezi că merge.
    return f"""
    <div style="background: #1e293b; padding: 15px; border-radius: 8px; border-left: 4px solid #3b82f6;">
        <h4 style="color: #60a5fa; margin: 0;">Analiză în curs...</h4>
        <p style="color: #cbd5e1; font-size: 0.9em;">Se procesează: <strong>{file_path}</strong></p>
        <p style="color: #94a3b8; font-style: italic;">Status: Se apelează Essentia C++ Extractor</p>
    </div>
    """

if __name__ == "__main__":
    print("Serverul pornește pe http://localhost:5000")
    app.run(debug=True, port=5000)