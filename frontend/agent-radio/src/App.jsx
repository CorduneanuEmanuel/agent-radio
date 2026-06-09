import { useState, useEffect, useRef } from "react";

/* ─── Inline styles that can't be done with Tailwind alone ─── */
const globalStyles = `
  @import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&family=Inter:wght@300;400;600;800&display=swap');

  body { font-family: 'Inter', sans-serif; background-color: #09090b; color: #f4f4f5; min-height: 100vh; margin: 0; }
  .mono { font-family: 'JetBrains Mono', monospace; }

  .scanline {
    background:
      linear-gradient(to bottom, rgba(18,16,16,0) 50%, rgba(0,0,0,0.25) 50%),
      linear-gradient(90deg, rgba(255,0,0,0.06), rgba(0,255,0,0.02), rgba(0,0,255,0.06));
    background-size: 100% 2px, 3px 100%;
  }

  @keyframes pulse-cyan {
    0%   { box-shadow: 0 0 0 0   rgba(34,211,238,0.4); }
    70%  { box-shadow: 0 0 0 10px rgba(34,211,238,0);   }
    100% { box-shadow: 0 0 0 0   rgba(34,211,238,0);   }
  }
  .ai-pulse { animation: pulse-cyan 2s infinite; }

  @keyframes bounce { 0%,100% { transform: translateY(0); } 50% { transform: translateY(-6px); } }
  .animate-bounce { animation: bounce 1s infinite; }

  @keyframes spin-once { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
  .note-spin-trigger:hover .note-spin-icon {
    animation: spin-once 0.6s ease-in-out 1;
    display: inline-block;
    transform-origin: center;
  }

  @keyframes success-pop {
    0%   { opacity: 0; transform: scale(0.8) translate(-50%,-50%); }
    50%  { opacity: 1; transform: scale(1)   translate(-50%,-50%); }
    100% { opacity: 0; transform: scale(1.1) translate(-50%,-50%); }
  }
  .success-message {
    animation: success-pop 2s ease-out forwards;
    position: fixed;
    top: 50%; left: 50%;
    transform: translate(-50%,-50%);
    z-index: 50;
  }
`;

/* ─── SpinningNote ─── */
function SpinningNote() {
  return (
    <div className="note-spin-trigger" style={{ cursor: "pointer" }}>
      <div
        style={{
          width: 80, height: 80,
          background: "linear-gradient(135deg,#2563eb,#312e81)",
          borderRadius: 16,
          display: "flex", alignItems: "center", justifyContent: "center",
          boxShadow: "0 8px 24px rgba(30,58,138,0.2)",
        }}
      >
        <span className="animate-bounce note-spin-icon" style={{ fontSize: 30 }}>📻</span>
      </div>
    </div>
  );
}

/* ─── SuccessMessage ─── */
function SuccessMessage({ isVisible, onClose }) {
  useEffect(() => {
    if (!isVisible) return;
    const t = setTimeout(onClose, 2000);
    return () => clearTimeout(t);
  }, [isVisible, onClose]);

  if (!isVisible) return null;
  return (
    <div
      className="success-message"
      style={{
        background: "#059669", color: "#fff",
        padding: "16px 32px", borderRadius: 16,
        fontWeight: 700, fontSize: 18,
        display: "flex", alignItems: "center", gap: 12,
      }}
    >
      <span>✓</span> Fișier adăugat cu succes!
    </div>
  );
}

/* ─── UploadProgress bar ─── */
function ProgressBar({ percent }) {
  return (
    <div style={{ width: "100%", background: "#27272a", borderRadius: 9999, height: 4, overflow: "hidden" }}>
      <div
        style={{
          height: 4, width: `${percent}%`,
          background: "#3b82f6",
          transition: "width 0.3s ease",
          borderRadius: 9999,
        }}
      />
    </div>
  );
}

/* ─── Upload Panel ─── */
function UploadPanel({ onUploadSuccess }) {
  const [fileName, setFileName] = useState("");
  const [progress, setProgress] = useState(0);
  const fileRef = useRef(null);

  const handleFileChange = (e) => {
    if (e.target.files[0]) setFileName(e.target.files[0].name);
  };

  const handleUpload = () => {
    const file = fileRef.current?.files[0];
    if (!file) return;

    const formData = new FormData();
    formData.append("file", file);

    const xhr = new XMLHttpRequest();
    xhr.open("POST", "http://167.172.171.185:8080/upload");

    xhr.upload.addEventListener("progress", (e) => {
      if (e.lengthComputable) setProgress((e.loaded / e.total) * 100);
    });

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        onUploadSuccess(xhr.responseText);
        setFileName("");
        setProgress(0);
        if (fileRef.current) fileRef.current.value = "";
      }
    };

    xhr.send(formData);
  };

  return (
    <div
      style={{
        background: "#18181b", border: "1px solid #27272a",
        borderRadius: 24, padding: 8,
      }}
    >
      <label
        htmlFor="file-input"
        style={{
          display: "block",
          border: "2px dashed #3f3f46",
          padding: "48px 16px",
          borderRadius: "1.4rem",
          textAlign: "center",
          cursor: "pointer",
          transition: "all 0.2s",
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.borderColor = "rgba(59,130,246,0.5)";
          e.currentTarget.style.background = "rgba(59,130,246,0.05)";
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.borderColor = "#3f3f46";
          e.currentTarget.style.background = "transparent";
        }}
      >
        <input
          type="file"
          id="file-input"
          name="file"
          ref={fileRef}
          onChange={handleFileChange}
          style={{ display: "none" }}
        />
        <div style={{ display: "flex", flexDirection: "column", alignItems: "center", gap: 16 }}>
          <div
            style={{
              background: "#27272a", width: 64, height: 64,
              borderRadius: "50%",
              display: "flex", alignItems: "center", justifyContent: "center",
            }}
          >
            <svg width="32" height="32" fill="none" stroke="#60a5fa" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 4v16m8-8H4" />
            </svg>
          </div>
          <div>
            <p style={{ fontWeight: 600, color: "#e4e4e7", marginBottom: 4 }}>Adaugă melodii noi</p>
            <p style={{ fontSize: 12, color: "#71717a", textTransform: "uppercase", letterSpacing: "0.05em" }}>
              Essentia local analysis
            </p>
          </div>
          {fileName && (
            <p className="mono" style={{ color: "#60a5fa", fontSize: 12, fontStyle: "italic" }}>
              {fileName}
            </p>
          )}
        </div>
      </label>

      <div style={{ padding: "8px 16px" }}>
        <ProgressBar percent={progress} />
      </div>

      <button
        onClick={handleUpload}
        style={{
          width: "100%", marginTop: 8,
          background: "#2563eb", color: "#fff",
          fontWeight: 700, padding: "16px",
          borderRadius: 16, border: "none",
          cursor: "pointer", transition: "background 0.2s",
          fontSize: 14, letterSpacing: "0.05em",
        }}
        onMouseEnter={(e) => (e.currentTarget.style.background = "#3b82f6")}
        onMouseLeave={(e) => (e.currentTarget.style.background = "#2563eb")}
        onMouseDown={(e) => (e.currentTarget.style.transform = "scale(0.98)")}
        onMouseUp={(e) => (e.currentTarget.style.transform = "scale(1)")}
      >
        LANSEAZĂ ANALIZA
      </button>
    </div>
  );
}

/* ─── Stats / Metadata Panel ─── */
function StatsDisplay({ html }) {
  return (
    <div
      style={{
        background: "#18181b", border: "1px solid #27272a",
        borderRadius: 24, padding: 24, minHeight: 300,
        position: "relative", overflow: "hidden",
      }}
    >
      <h3
        style={{
          fontSize: 12, fontWeight: 700, color: "#71717a",
          textTransform: "uppercase", letterSpacing: "0.2em", marginBottom: 24,
        }}
      >
        Local Metadata (SQLite)
      </h3>

      {html ? (
        <div dangerouslySetInnerHTML={{ __html: html }} />
      ) : (
        /* SSE-driven live metadata */
        <MetadataSSE />
      )}

      {/* decorative glow */}
      <div
        style={{
          position: "absolute", bottom: -80, right: -80,
          width: 256, height: 256,
          background: "rgba(37,99,235,0.05)",
          borderRadius: "50%", filter: "blur(80px)",
          pointerEvents: "none",
        }}
      />
    </div>
  );
}

/* SSE listener for metadata events */
function MetadataSSE() {
  const [content, setContent] = useState("");

  useEffect(() => {
    const es = new EventSource("http://167.172.171.185:8080/events");
    es.addEventListener("metadata", (e) => setContent(e.data));
    return () => es.close();
  }, []);

  if (!content) {
    return (
      <div
        style={{
          display: "flex", flexDirection: "column",
          alignItems: "center", justifyContent: "center",
          height: 192, textAlign: "center", color: "#71717a", fontStyle: "italic",
        }}
      >
        <svg width="48" height="48" fill="none" stroke="currentColor" viewBox="0 0 24 24"
          style={{ opacity: 0.2, marginBottom: 8 }}
        >
          <path d="M9 19V6l12-3v13M9 19c0 1.1-.9 2-2 2s-2-.9-2-2 .9-2 2-2 2 .9 2 2zm12-3c0 1.1-.9 2-2 2s-2-.9-2-2 .9-2 2-2 2 .9 2 2z" />
        </svg>
        Așteptare date...
      </div>
    );
  }

  return <div dangerouslySetInnerHTML={{ __html: content }} />;
}

/* ─── AI Thoughts terminal panel ─── */
function AIThoughtsPanel() {
  const [thoughts, setThoughts] = useState(
    "> Initializing agents...\n> Waiting for SSE stream from server..."
  );
  const scrollRef = useRef(null);

  useEffect(() => {
    const es = new EventSource("http://167.172.171.185:8080/events");
    es.addEventListener("thoughts", (e) => {
      setThoughts(e.data);
    });
    return () => es.close();
  }, []);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [thoughts]);

  return (
    <div
      style={{
        background: "#000", border: "1px solid #27272a",
        borderRadius: 24, overflow: "hidden",
        boxShadow: "0 25px 50px rgba(6,78,59,0.05)",
      }}
    >
      {/* terminal header */}
      <div
        style={{
          background: "rgba(24,24,27,0.8)", padding: "12px 20px",
          borderBottom: "1px solid #27272a",
          display: "flex", alignItems: "center", justifyContent: "space-between",
        }}
      >
        <span
          className="mono"
          style={{
            fontSize: 12, fontWeight: 700, textTransform: "uppercase",
            letterSpacing: "0.15em", color: "#10b981"
          }}
        >
          Ollama Thinking Process
        </span>
        <div style={{ display: "flex", gap: 6 }}>
          <div style={{ width: 10, height: 10, borderRadius: "50%", background: "rgba(239,68,68,0.2)" }} />
          <div style={{ width: 10, height: 10, borderRadius: "50%", background: "rgba(234,179,8,0.2)" }} />
          <div style={{ width: 10, height: 10, borderRadius: "50%", background: "rgba(16,185,129,0.2)" }} />
        </div>
      </div>

      {/* scrollable output */}
      <div
        ref={scrollRef}
        className="scanline"
        style={{ padding: 24, height: 320, overflowY: "auto", position: "relative" }}
      >
        <pre
          className="mono"
          style={{
            fontSize: 14, lineHeight: 1.6,
            color: "rgba(52,211,153,0.9)",
            whiteSpace: "pre-wrap", margin: 0,
          }}
        >
          {thoughts}
        </pre>
      </div>
    </div>
  );
}

/* ─── App ─── */
export default function App() {
  const [showSuccess, setShowSuccess] = useState(false);
  const [uploadedHtml, setUploadedHtml] = useState("");

  const handleUploadSuccess = (responseHtml) => {
    setUploadedHtml(responseHtml);
    setShowSuccess(true);
  };

  return (
    <>
      {/* inject global CSS */}
      <style>{globalStyles}</style>

      {/* Header */}
      <header
        style={{
          borderBottom: "1px solid #27272a", padding: "16px",
          marginBottom: 32,
          background: "rgba(24,24,27,0.5)",
          backdropFilter: "blur(12px)",
          position: "sticky", top: 0, zIndex: 50,
        }}
      >
        <div
          style={{
            maxWidth: 1280, margin: "0 auto",
            display: "flex", justifyContent: "space-between", alignItems: "center",
          }}
        >
          <h1
            style={{
              fontSize: 24, fontWeight: 900, fontStyle: "italic",
              letterSpacing: "-0.02em",
              background: "linear-gradient(90deg,#60a5fa,#34d399)",
              WebkitBackgroundClip: "text", WebkitTextFillColor: "transparent",
              margin: 0,
            }}
          >
            RADIO AI LOCAL{" "}
            <span
              className="mono"
              style={{
                fontSize: 12, fontWeight: 400, fontStyle: "normal",
                color: "#71717a", WebkitTextFillColor: "#71717a",
                marginLeft: 8, textTransform: "uppercase", letterSpacing: "0.05em",
              }}
            >
              v1.0 Beta
            </span>
          </h1>
          <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
            <span
              className="ai-pulse"
              style={{ display: "block", width: 12, height: 12, borderRadius: "50%", background: "#10b981" }}
            />
            <span
              style={{
                fontSize: 12, fontWeight: 700, color: "#a1a1aa",
                textTransform: "uppercase", letterSpacing: "0.15em",
              }}
            >
              Ollama Online
            </span>
          </div>
        </div>
      </header>

      {/* Main */}
      <main
        style={{
          width: "80%", margin: "0 auto", padding: "0 16px 48px",
          display: "grid",
          gridTemplateColumns: "repeat(12, 1fr)",
          gap: 32,
        }}
      >
        {/* Left column */}
        <section style={{ gridColumn: "span 7", display: "flex", flexDirection: "column", gap: 24 }}>
          {/* On Air card */}
          <div
            style={{
              background: "#18181b", border: "1px solid #27272a",
              borderRadius: 24, padding: 32,
              boxShadow: "0 25px 50px rgba(0,0,0,0.5)",
              overflow: "hidden", position: "relative",
            }}
          >
            {/* watermark icon */}
            <div style={{ position: "absolute", top: 0, right: 0, padding: 16, opacity: 0.1, pointerEvents: "none" }}>
              <svg width="100" height="100" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z" />
              </svg>
            </div>

            <h2
              style={{
                fontSize: 12, fontWeight: 700, color: "#71717a",
                textTransform: "uppercase", letterSpacing: "0.2em", marginBottom: 16,
              }}
            >
              On Air / Icecast Stream
            </h2>

            <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
              <div style={{ display: "flex", alignItems: "center", gap: 24 }}>
                <SpinningNote />
                <div>
                  <h3 style={{ fontSize: 24, fontWeight: 700, margin: 0 }}>AI Curator Session</h3>
                  <p className="mono" style={{ color: "#a1a1aa", fontSize: 14, margin: 0 }}>
                    Host: Chatterbox TTS
                  </p>
                </div>
              </div>

              <audio
                controls
                style={{ width: "100%", height: 48, borderRadius: 8, filter: "brightness(0.9) contrast(1.25)" }}
              >
                <source src="http://167.172.171.185:8000/radio" type="audio/mpeg" />
              </audio>
            </div>
          </div>

          {/* AI Thoughts terminal */}
          <AIThoughtsPanel />
        </section>

        {/* Right column */}
        <section style={{ gridColumn: "span 5", display: "flex", flexDirection: "column", gap: 24 }}>
          <UploadPanel onUploadSuccess={handleUploadSuccess} />
          <StatsDisplay html={uploadedHtml} />
        </section>
      </main>

      {/* Footer */}
      <footer
        style={{
          maxWidth: 1280, margin: "0 auto", padding: "32px 16px",
          borderTop: "1px solid #18181b",
          textAlign: "center", color: "#52525b", fontSize: 12,
        }}
      >
        Proiectat cu HTMX + Tailwind + Ollama | Procesare locală în Python
      </footer>

      {/* Global success overlay */}
      <SuccessMessage isVisible={showSuccess} onClose={() => setShowSuccess(false)} />
    </>
  );
}
