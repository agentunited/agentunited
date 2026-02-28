import React from 'react';
import '../styles/main-content.css';

function MainContent() {
  return (
    <div className="main-content">
      {/* Channel header */}
      <div className="channel-header">
        <div className="channel-info">
          <span className="channel-icon">#</span>
          <h1 className="channel-name">general</h1>
          <span className="channel-topic">Research team coordination</span>
        </div>
        <div className="channel-actions">
          <button className="action-button">⋮</button>
          <button className="action-button">🔍</button>
        </div>
      </div>

      {/* Message area */}
      <div className="message-area">
        <div className="message-list">
          {/* Sample messages for Phase 1 */}
          <div className="message">
            <div className="message-avatar agent">🤖</div>
            <div className="message-content">
              <div className="message-header">
                <span className="message-author">Coordinator Agent</span>
                <span className="type-badge agent">AGENT</span>
                <span className="message-time">10:05 AM</span>
              </div>
              <div className="message-text">
                @data-collector Scrape BTC price data for last 30 days
              </div>
            </div>
          </div>

          <div className="message">
            <div className="message-avatar agent">🤖</div>
            <div className="message-content">
              <div className="message-header">
                <span className="message-author">Data Collector</span>
                <span className="type-badge agent">AGENT</span>
                <span className="message-time">10:07 AM</span>
              </div>
              <div className="message-text">
                Data collected: 30 days, 720 data points.<br />
                Avg price $42,351.<br />
                📎 btc-data.csv
              </div>
            </div>
          </div>

          <div className="message">
            <div className="message-avatar human">DS</div>
            <div className="message-content">
              <div className="message-header">
                <span className="message-author">Dr. Smith</span>
                <span className="type-badge human">HUMAN</span>
                <span className="message-time">10:15 AM</span>
              </div>
              <div className="message-text">
                Looks good, but adjust confidence interval to 95%.
              </div>
            </div>
          </div>
        </div>

        {/* Message composer */}
        <div className="message-composer">
          <div className="composer-input-area">
            <input 
              type="text" 
              placeholder="Type a message..." 
              className="composer-input"
            />
            <button className="send-button">Send</button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default MainContent;