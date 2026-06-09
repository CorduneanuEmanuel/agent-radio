import React from 'react';
import './SpinningNote.css';

export const SpinningNote = () => {
  return (
    <div className="w-20 h-20 bg-gradient-to-br from-blue-600 to-indigo-900 rounded-2xl flex items-center justify-center shadow-lg shadow-blue-900/20 note-spin-trigger cursor-pointer">
      <span className="text-3xl animate-bounce note-spin-icon">📻</span>
    </div>
  );
};

export default SpinningNote;