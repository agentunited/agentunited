import React, { useState } from 'react';
import Sidebar from './components/Sidebar';
import MainContent from './components/MainContent';
import TitleBar from './components/TitleBar';
import './App.css';

function App() {
  const [sidebarWidth, setSidebarWidth] = useState(240);

  return (
    <div className="app">
      <TitleBar />
      <div className="app-body">
        <Sidebar width={sidebarWidth} onWidthChange={setSidebarWidth} />
        <MainContent />
      </div>
    </div>
  );
}

export default App;