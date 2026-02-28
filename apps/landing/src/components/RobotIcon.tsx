interface RobotIconProps {
  className?: string;
  size?: number;
}

export function RobotIcon({ className = "", size = 48 }: RobotIconProps) {
  return (
    <svg 
      width={size} 
      height={size} 
      viewBox="0 0 48 48" 
      fill="none" 
      xmlns="http://www.w3.org/2000/svg"
      className={className}
    >
      {/* Robot head (CRT monitor) */}
      <rect x="12" y="8" width="24" height="18" rx="2" fill="currentColor" stroke="currentColor" strokeWidth="1.5"/>
      <rect x="14" y="10" width="20" height="14" rx="1" fill="none" stroke="currentColor" strokeWidth="1"/>
      
      {/* Screen glow/eyes */}
      <circle cx="18" cy="17" r="2" fill="currentColor" opacity="0.8"/>
      <circle cx="30" cy="17" r="2" fill="currentColor" opacity="0.8"/>
      
      {/* Screen line */}
      <line x1="20" y1="21" x2="28" y2="21" stroke="currentColor" strokeWidth="1"/>
      
      {/* Antenna */}
      <line x1="20" y1="8" x2="18" y2="4" stroke="currentColor" strokeWidth="2"/>
      <line x1="28" y1="8" x2="30" y2="4" stroke="currentColor" strokeWidth="2"/>
      <circle cx="18" cy="4" r="1" fill="currentColor"/>
      <circle cx="30" cy="4" r="1" fill="currentColor"/>
      
      {/* Body */}
      <rect x="16" y="26" width="16" height="12" rx="2" fill="none" stroke="currentColor" strokeWidth="1.5"/>
      
      {/* Body details */}
      <rect x="18" y="28" width="4" height="3" rx="0.5" fill="currentColor" opacity="0.6"/>
      <rect x="26" y="28" width="4" height="3" rx="0.5" fill="currentColor" opacity="0.6"/>
      <line x1="20" y1="34" x2="28" y2="34" stroke="currentColor" strokeWidth="1"/>
      
      {/* Tracks/base */}
      <ellipse cx="24" cy="41" rx="12" ry="3" fill="none" stroke="currentColor" strokeWidth="1.5"/>
      <ellipse cx="24" cy="41" rx="8" ry="2" fill="currentColor" opacity="0.3"/>
      
      {/* Track details */}
      <circle cx="18" cy="41" r="1" fill="currentColor"/>
      <circle cx="24" cy="41" r="1" fill="currentColor"/>
      <circle cx="30" cy="41" r="1" fill="currentColor"/>
    </svg>
  );
}

export function RobotChain({ className = "", size = 32 }: RobotIconProps) {
  return (
    <svg 
      width={size * 3} 
      height={size} 
      viewBox="0 0 96 32" 
      fill="none" 
      xmlns="http://www.w3.org/2000/svg"
      className={className}
    >
      {/* Robot 1 */}
      <g transform="translate(0, 0)">
        <rect x="4" y="6" width="12" height="9" rx="1" fill="currentColor" opacity="0.8"/>
        <rect x="5" y="7" width="10" height="7" rx="0.5" fill="none" stroke="currentColor" strokeWidth="0.5"/>
        <circle cx="7" cy="10" r="1" fill="currentColor"/>
        <circle cx="13" cy="10" r="1" fill="currentColor"/>
        <rect x="6" y="15" width="8" height="6" rx="1" fill="none" stroke="currentColor" strokeWidth="0.8"/>
        <ellipse cx="10" cy="24" rx="6" ry="2" fill="none" stroke="currentColor" strokeWidth="0.8"/>
      </g>
      
      {/* Connection line 1 */}
      <line x1="16" y1="16" x2="24" y2="16" stroke="currentColor" strokeWidth="2" opacity="0.6"/>
      
      {/* Robot 2 */}
      <g transform="translate(24, 0)">
        <rect x="4" y="6" width="12" height="9" rx="1" fill="currentColor" opacity="0.8"/>
        <rect x="5" y="7" width="10" height="7" rx="0.5" fill="none" stroke="currentColor" strokeWidth="0.5"/>
        <circle cx="7" cy="10" r="1" fill="currentColor"/>
        <circle cx="13" cy="10" r="1" fill="currentColor"/>
        <rect x="6" y="15" width="8" height="6" rx="1" fill="none" stroke="currentColor" strokeWidth="0.8"/>
        <ellipse cx="10" cy="24" rx="6" ry="2" fill="none" stroke="currentColor" strokeWidth="0.8"/>
      </g>
      
      {/* Connection line 2 */}
      <line x1="40" y1="16" x2="48" y2="16" stroke="currentColor" strokeWidth="2" opacity="0.6"/>
      
      {/* Robot 3 */}
      <g transform="translate(48, 0)">
        <rect x="4" y="6" width="12" height="9" rx="1" fill="currentColor" opacity="0.8"/>
        <rect x="5" y="7" width="10" height="7" rx="0.5" fill="none" stroke="currentColor" strokeWidth="0.5"/>
        <circle cx="7" cy="10" r="1" fill="currentColor"/>
        <circle cx="13" cy="10" r="1" fill="currentColor"/>
        <rect x="6" y="15" width="8" height="6" rx="1" fill="none" stroke="currentColor" strokeWidth="0.8"/>
        <ellipse cx="10" cy="24" rx="6" ry="2" fill="none" stroke="currentColor" strokeWidth="0.8"/>
      </g>
    </svg>
  );
}