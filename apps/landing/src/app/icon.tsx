import { ImageResponse } from 'next/og';

export const size = {
  width: 32,
  height: 32,
};

export const contentType = 'image/png';

export default function Icon() {
  return new ImageResponse(
    (
      <div
        style={{
          fontSize: 24,
          background: '#F5F3EF',
          width: '100%',
          height: '100%',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: '#D97548',
          fontFamily: 'monospace',
        }}
      >
        <svg
          width="28"
          height="28"
          viewBox="0 0 32 32"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          {/* Robot head (CRT monitor) */}
          <rect x="8" y="6" width="16" height="12" rx="1.5" fill="#D97548" stroke="#D97548" strokeWidth="1"/>
          <rect x="9" y="7" width="14" height="10" rx="1" fill="none" stroke="#D97548" strokeWidth="0.5"/>
          
          {/* Screen glow/eyes */}
          <circle cx="12" cy="12" r="1.5" fill="#FFB84D"/>
          <circle cx="20" cy="12" r="1.5" fill="#FFB84D"/>
          
          {/* Screen line */}
          <line x1="13" y1="15" x2="19" y2="15" stroke="#D97548" strokeWidth="0.5"/>
          
          {/* Antenna */}
          <line x1="13" y1="6" x2="11" y2="3" stroke="#D97548" strokeWidth="1.5"/>
          <line x1="19" y1="6" x2="21" y2="3" stroke="#D97548" strokeWidth="1.5"/>
          <circle cx="11" cy="3" r="0.75" fill="#FFB84D"/>
          <circle cx="21" cy="3" r="0.75" fill="#FFB84D"/>
          
          {/* Body */}
          <rect x="11" y="18" width="10" height="8" rx="1.5" fill="none" stroke="#D97548" strokeWidth="1"/>
          
          {/* Tracks/base */}
          <ellipse cx="16" cy="27" rx="8" ry="2" fill="none" stroke="#D97548" strokeWidth="1"/>
        </svg>
      </div>
    ),
    {
      ...size,
    }
  );
}