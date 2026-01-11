import React, { useEffect, useState } from 'react'
import './WarningModal.css'
import { EventsOn } from '../wailsjs/runtime/runtime'

function WarningModal() {
  const [isVisible, setIsVisible] = useState(false)
  const [warningData, setWarningData] = useState(null)

  useEffect(() => {
    // Listen for warning-detected event
    const unsubscribe = EventsOn('warning-detected', (data) => {
      console.log('⚠️ Warning detected:', data)
      setWarningData(data)
      setIsVisible(true)
    })

    return () => {
      if (unsubscribe && typeof unsubscribe === 'function') {
        unsubscribe()
      }
    }
  }, [])

  const handleContinue = () => {
    setIsVisible(false)
    setWarningData(null)
  }

  const handleCloseApp = async () => {
    if (!warningData || !warningData.executableName) {
      setIsVisible(false)
      return
    }

    try {
      // For Phase 1, we'll just close the modal
      // In Phase 2, we can add actual app closing logic
      console.log('Would close app:', warningData.executableName)
      setIsVisible(false)
      setWarningData(null)
    } catch (err) {
      console.error('Error closing app:', err)
    }
  }

  if (!isVisible || !warningData) {
    return null
  }

  return (
    <div className="warning-modal-overlay">
      <div className="warning-modal">
        <div className="warning-modal-header">
          <h2>⚠️ Focus Warning</h2>
        </div>
        <div className="warning-modal-content">
          <p className="warning-message">
            You're about to open a blocked app:
          </p>
          <div className="warning-app-info">
            <div className="warning-app-name">
              {warningData.displayName || warningData.executableName}
            </div>
            <div className="warning-app-exe">
              {warningData.executableName}
            </div>
            {warningData.title && (
              <div className="warning-app-window">
                Window: {warningData.title}
              </div>
            )}
          </div>
          <p className="warning-question">
            Are you sure you want to continue?
          </p>
        </div>
        <div className="warning-modal-actions">
          <button
            onClick={handleContinue}
            className="btn btn-secondary"
          >
            Continue Anyway
          </button>
          <button
            onClick={handleCloseApp}
            className="btn btn-primary"
          >
            Close App
          </button>
        </div>
      </div>
    </div>
  )
}

export default WarningModal
