import React, { useState, useEffect, useRef, useCallback } from 'react'
import './App.css'
import WindowMonitor from './components/WindowMonitor'
import HistoryLog from './components/HistoryLog'
import AutoStartSettings from './components/AutoStartSettings'
import { EventsOn } from './wailsjs/runtime/runtime'

function App() {
  const [currentWindow, setCurrentWindow] = useState(null)
  const [history, setHistory] = useState([])
  const [autoStartEnabled, setAutoStartEnabled] = useState(false)
  const historyRef = useRef([])
  const lastWindowRef = useRef(null)

  // Add to history function - use useCallback to ensure it's stable
  const addToHistory = useCallback((windowInfo) => {
    if (!windowInfo) return
    
    // Skip if same window
    if (lastWindowRef.current && 
        lastWindowRef.current.title === windowInfo.title && 
        lastWindowRef.current.exe === windowInfo.exe) {
      return
    }
    
    const now = new Date()
    const entry = {
      ...windowInfo,
      time: now.toLocaleTimeString(),
      date: now.toLocaleDateString(),
      timestamp: now.getTime(),
      id: Date.now() + Math.random(),
      // Format like terminal output: "Active Window Changed: [exe] title"
      terminalLine: `Active Window Changed: [${windowInfo.exe || 'unknown'}] ${windowInfo.title || 'Unknown'}`
    }
    
    lastWindowRef.current = windowInfo
    historyRef.current = [entry, ...historyRef.current].slice(0, 200)
    setHistory([...historyRef.current])
  }, [])

  useEffect(() => {
    let unsubscribe = null
    let pollInterval = null
    let isInitialized = false

    // Function to update window info
    const updateWindow = (windowInfo) => {
      if (!windowInfo) {
        console.log('updateWindow called with null/undefined')
        return
      }
      
      // Check if window info is valid (has at least title or exe)
      if (!windowInfo.title && !windowInfo.exe) {
        console.log('updateWindow: windowInfo has no title or exe', windowInfo)
        return
      }

      // Only update if changed
      const current = lastWindowRef.current
      if (current && 
          current.title === windowInfo.title && 
          current.exe === windowInfo.exe) {
        return // No change
      }

      console.log('Updating window:', windowInfo)
      setCurrentWindow(windowInfo)
      addToHistory(windowInfo)
    }

    // Wait for Wails bindings to be available
    const init = async () => {
      if (isInitialized) return
      
      console.log('Initializing app...')
      
      // Check if bindings are available
      if (typeof window.go === 'undefined' || !window.go || !window.go.main || !window.go.main.App) {
        console.error('Wails bindings not found')
        return
      }

      console.log('✓ Bindings found')

      // Set up event listener for window changes
      try {
        if (typeof EventsOn === 'function') {
          unsubscribe = EventsOn('window-changed', (windowInfo) => {
            console.log('📡 Event received:', windowInfo)
            updateWindow(windowInfo)
          })
          console.log('✓ Event listener registered (EventsOn import)')
        } else if (window.runtime && typeof window.runtime.EventsOn === 'function') {
          unsubscribe = window.runtime.EventsOn('window-changed', (windowInfo) => {
            console.log('📡 Event received:', windowInfo)
            updateWindow(windowInfo)
          })
          console.log('✓ Event listener registered (window.runtime)')
        } else {
          console.warn('⚠ EventsOn not available, using polling only')
        }
      } catch (err) {
        console.error('❌ Error setting up event listener:', err)
      }

      // Get initial window immediately
      const fetchInitialWindow = async () => {
        try {
          console.log('🔍 Fetching initial window...')
          const windowInfo = await window.go.main.App.GetCurrentWindow()
          console.log('📦 Initial window result:', windowInfo)
          
          if (windowInfo && (windowInfo.title || windowInfo.exe)) {
            console.log('✓ Setting initial window')
            updateWindow(windowInfo)
          } else {
            console.warn('⚠ Initial window is empty, will retry...')
            // Retry after a short delay
            setTimeout(fetchInitialWindow, 500)
          }
        } catch (err) {
          console.error('❌ Error getting initial window:', err)
          // Retry after error
          setTimeout(fetchInitialWindow, 1000)
        }
      }
      
      fetchInitialWindow()

      // Set up polling (always active as fallback)
      pollInterval = setInterval(async () => {
        try {
          const windowInfo = await window.go.main.App.GetCurrentWindow()
          if (windowInfo && (windowInfo.title || windowInfo.exe)) {
            updateWindow(windowInfo)
          }
        } catch (err) {
          // Silently fail on polling errors
        }
      }, 1000)
      console.log('✓ Polling started (1s interval)')

      // Check auto-start status
      try {
        const enabled = await window.go.main.App.IsAutoStartEnabled()
        setAutoStartEnabled(enabled)
        console.log('✓ Auto-start status:', enabled)
      } catch (err) {
        console.error('❌ Error checking auto-start:', err)
      }

      isInitialized = true
    }

    // Try to initialize, retry if bindings not ready
    let attempts = 0
    const maxAttempts = 50
    const tryInit = () => {
      attempts++
      console.log(`[${attempts}/${maxAttempts}] Checking for bindings...`, {
        window: typeof window !== 'undefined',
        windowGo: typeof window.go !== 'undefined',
        windowGoMain: typeof window.go?.main !== 'undefined',
        windowGoMainApp: typeof window.go?.main?.App !== 'undefined',
        windowRuntime: typeof window.runtime !== 'undefined',
        windowRuntimeEventsOn: typeof window.runtime?.EventsOn !== 'function'
      })
      
      if (typeof window.go !== 'undefined' && window.go && window.go.main && window.go.main.App) {
        console.log('✅ Bindings found!', window.go.main.App)
        init()
      } else if (attempts < maxAttempts) {
        setTimeout(tryInit, 200)
      } else {
        console.error('❌ Wails bindings not found after', maxAttempts, 'attempts')
        console.error('Available on window:', Object.keys(window).filter(k => k.includes('go') || k.includes('runtime')))
      }
    }
    
    // Start checking after a short delay to let Wails inject bindings
    setTimeout(tryInit, 100)

    // Cleanup
    return () => {
      if (unsubscribe && typeof unsubscribe === 'function') {
        unsubscribe()
      }
      if (pollInterval) {
        clearInterval(pollInterval)
      }
    }
  }, [addToHistory])

  const clearHistory = () => {
    if (window.confirm('Are you sure you want to clear the history?')) {
      historyRef.current = []
      setHistory([])
    }
  }

  const handleEnableAutoStart = async () => {
    if (!window.go || !window.go.main || !window.go.main.App) {
      alert('Wails bindings not available')
      return
    }
    try {
      await window.go.main.App.EnableAutoStart()
      setAutoStartEnabled(true)
    } catch (err) {
      alert('Failed to enable auto-start: ' + err)
    }
  }

  const handleDisableAutoStart = async () => {
    if (!window.go || !window.go.main || !window.go.main.App) {
      alert('Wails bindings not available')
      return
    }
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

