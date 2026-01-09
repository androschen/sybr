import React, { useState, useEffect, useRef } from 'react'
import './App.css'
import WindowMonitor from './components/WindowMonitor'
import HistoryLog from './components/HistoryLog'
import AutoStartSettings from './components/AutoStartSettings'

function App() {
  const [currentWindow, setCurrentWindow] = useState(null)
  const [history, setHistory] = useState([])
  const [autoStartEnabled, setAutoStartEnabled] = useState(false)
  const historyRef = useRef([])

  useEffect(() => {
    // Initialize - get current window
    const init = async () => {
      try {
        const window = await window.go.main.App.GetCurrentWindow()
        if (window) {
          setCurrentWindow(window)
          addToHistory(window)
        }
      } catch (err) {
        console.error('Error getting current window:', err)
      }

      // Check auto-start status
      try {
        const enabled = await window.go.main.App.IsAutoStartEnabled()
        setAutoStartEnabled(enabled)
      } catch (err) {
        console.error('Error checking auto-start:', err)
      }
    }

    init()

    // Listen for window change events
    window.runtime.EventsOn('window-changed', (windowInfo) => {
      console.log('Window changed:', windowInfo)
      setCurrentWindow(windowInfo)
      addToHistory(windowInfo)
    })

    // Cleanup
    return () => {
      window.runtime.EventsOff('window-changed')
    }
  }, [])

  const addToHistory = (windowInfo) => {
    const now = new Date()
    const entry = {
      ...windowInfo,
      time: now.toLocaleTimeString(),
      date: now.toLocaleDateString(),
      timestamp: now.getTime(),
      id: Date.now() + Math.random()
    }
    
    historyRef.current = [entry, ...historyRef.current].slice(0, 100)
    setHistory([...historyRef.current])
  }

  const clearHistory = () => {
    if (window.confirm('Are you sure you want to clear the history?')) {
      historyRef.current = []
      setHistory([])
    }
  }

  const handleEnableAutoStart = async () => {
    try {
      await window.go.main.App.EnableAutoStart()
      setAutoStartEnabled(true)
    } catch (err) {
      alert('Failed to enable auto-start: ' + err)
    }
  }

  const handleDisableAutoStart = async () => {
    try {
      await window.go.main.App.DisableAutoStart()
      setAutoStartEnabled(false)
    } catch (err) {
      alert('Failed to disable auto-start: ' + err)
    }
  }

  return (
    <div className="app">
      <div className="container">
        <header className="app-header">
          <h1>🪟 Window Monitor</h1>
          <p className="subtitle">Real-time window tracking and monitoring</p>
        </header>

        <div className="content-grid">
          <div className="card">
            <WindowMonitor window={currentWindow} />
          </div>

          <div className="card">
            <AutoStartSettings
              enabled={autoStartEnabled}
              onEnable={handleEnableAutoStart}
              onDisable={handleDisableAutoStart}
            />
          </div>

          <div className="card card-full">
            <HistoryLog 
              history={history} 
              onClear={clearHistory}
            />
          </div>
        </div>
      </div>
    </div>
  )
}

export default App

