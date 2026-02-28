import React from 'react';

interface TitleBarProps {
  title?: string;
}

function TitleBar({ title = 'Agent United' }: TitleBarProps) {
  return (
    <div className="title-bar">
      <div className="title-bar-content">
        <span className="app-title">{title}</span>
        <div className="title-bar-controls">
          {/* Account and settings controls will be added here */}
        </div>
      </div>
    </div>
  );
}

export default TitleBar;