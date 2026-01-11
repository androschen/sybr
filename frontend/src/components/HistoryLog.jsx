import React, { useEffect, useRef } from 'react'
import './HistoryLog.css'

function HistoryLog({ history, onClear }) {
  const scrollRef = useRef(null)
  const prevHistoryLength = useRef(0)

  useEffect(() => {
    // Auto-scroll to top when new entries are added
    if (history.length > prevHistoryLength.current && scrollRef.current) {
      scrollRef.current.scrollTop = 0
    }
    prevHistoryLength.current = history.length
  }, [history.length])

  return (
    <div className="history-log">
      <div className="history-header">
        <div>
          <h2>Window Change History</h2>
          <span className="history-count">({history.length} entries)</span>
        </div>
        {history.length > 0 && (
          <button onClick={onClear} className="btn btn-clear">
            Clear History
          </button>
        )}
      </div>
      
      <div className="history-container" ref={scrollRef}>
        {history.length === 0 ? (
          <div className="history-empty">
            <p>No window changes recorded yet.</p>
            <p className="history-empty-sub">Switch between windows to see live updates here.</p>
          </div>
        ) : (
          <div className="history-list">
            {history.map((entry, index) => (
              <div
                key={entry.id}
                className={`history-item ${index === 0 ? 'history-item-new' : ''}`}
                style={{ opacity: Math.max(0.5, 1 - index * 0.01) }}
              >
                <div className="history-time">
                  <span className="history-date">{entry.date}</span>
                  <span className="history-time-value">{entry.time}</span>
                </div>
                {/* Show terminal-style line if available, otherwise show formatted */}
                {entry.terminalLine ? (
                  <div className="history-terminal-line">{entry.terminalLine}</div>
                ) : (
                  <>
                    <div className="history-window-title">{entry.title || 'Unknown'}</div>
                    <div className="history-window-exe">{entry.exe || '-'}</div>
                  </>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export default HistoryLog

