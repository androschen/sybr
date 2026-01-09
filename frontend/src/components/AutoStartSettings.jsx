import React from 'react'
import './AutoStartSettings.css'

function AutoStartSettings({ enabled, onEnable, onDisable }) {
  return (
    <div className="autostart-settings">
      <h2>Auto-Start Settings</h2>
      <div className="autostart-status">
        <div className={`status-indicator ${enabled ? 'active' : 'inactive'}`}>
          <span className="status-dot"></span>
          <span>Auto-Start: {enabled ? 'Enabled' : 'Disabled'}</span>
        </div>
      </div>
      <div className="autostart-controls">
        {enabled ? (
          <button onClick={onDisable} className="btn btn-danger">
            Disable Auto-Start
          </button>
        ) : (
          <button onClick={onEnable} className="btn btn-primary">
            Enable Auto-Start
          </button>
        )}
      </div>
      <p className="autostart-info">
        When enabled, the app will automatically start when Windows boots.
      </p>
    </div>
  )
}

export default AutoStartSettings

