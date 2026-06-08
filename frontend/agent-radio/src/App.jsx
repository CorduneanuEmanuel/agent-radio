import React, { useState } from 'react';
import SuccessMessage from './components/SuccessMessage';
import SpinningNote from './components/SpinningNote';
import './App.css';

function App() {
  const [showSuccess, setShowSuccess] = useState(false);

  const handleUploadSuccess = () => {
    setShowSuccess(true);
    setTimeout(() => setShowSuccess(false), 2000);
  };

  return (
    <div className="bg-zinc-950 text-zinc-100 min-h-screen">
      <header className="border-b border-zinc-800 p-4 mb-8 bg-zinc-900/50 backdrop-blur-md sticky top-0 z-50">
        <div className="container mx-auto flex justify-between items-center">
          <h1 className="text-2xl font-black tracking-tighter italic bg-gradient-to-r from-blue-400 to-emerald-400 bg-clip-text text-transparent">
            RADIO AI LOCAL <span className="text-xs font-mono text-zinc-500 uppercase not-italic ml-2">v1.0 Beta</span>
          </h1>
          <div className="flex items-center gap-4">
            <span className="flex h-3 w-3 rounded-full bg-emerald-500 ai-pulse"></span>
            <span className="text-xs font-bold text-zinc-400 uppercase tracking-widest">Ollama Online</span>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 pb-12 grid grid-cols-1 lg:grid-cols-12 gap-8">
        <section className="lg:col-span-7 space-y-6">
          <div className="bg-zinc-900 border border-zinc-800 rounded-3xl p-8 shadow-2xl overflow-hidden relative group">
            <h2 className="text-zinc-500 text-xs font-bold uppercase tracking-[0.2em] mb-4">On Air / Icecast Stream</h2>
            <div className="flex flex-col gap-6">
              <div className="flex items-center gap-6">
                <SpinningNote />
                <div>
                  <h3 className="text-2xl font-bold">AI Curator Session</h3>
                  <p className="text-zinc-400 text-sm mono">Host: Chatterbox TTS</p>
                </div>
              </div>
              <audio controls className="w-full h-12 brightness-90 contrast-125 rounded-lg">
                <source src="http://167.172.171.185:8000/radio" type="audio/mpeg" />
              </audio>
            </div>
          </div>
        </section>
      </main>

      <SuccessMessage isVisible={showSuccess} onClose={() => setShowSuccess(false)} />
    </div>
  );
}

export default App;