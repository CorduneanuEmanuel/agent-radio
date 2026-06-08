import React, { useEffect, useState } from 'react';

export const SuccessMessage = ({ isVisible, onClose }) => {
  useEffect(() => {
    if (isVisible) {
      const timer = setTimeout(onClose, 2000);
      return () => clearTimeout(timer);
    }
  }, [isVisible, onClose]);

  if (!isVisible) return null;

  return (
    <div className="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 bg-emerald-600 text-white px-8 py-4 rounded-2xl font-bold text-lg success-message flex items-center gap-3 z-50">
      <span>✓</span> Fișier adăugat cu succes!
    </div>
  );
};

export default SuccessMessage;