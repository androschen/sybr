import React from 'react'
import './WindowMonitor.css'

function WindowMonitor({ window }) {
  return (
    <div className="window-monitor">
      <h2>Current Active Window</h2>
      <div className="window-display">
        {window ? (
          <>
            <div className="window-title">{window.title || 'Unknown'}</div>
            <div className="window-exe">{window.exe || '-'}</div>
            <div className="status-indicator active">
              <span className="status-dot"></span>
              Monitoring Active
            </div>
          </>
        ) : (
          <div className="window-loading">Waiting for window change...</div>
        )}
      </div>
    </div>
  )
}

export default WindowMonitor

