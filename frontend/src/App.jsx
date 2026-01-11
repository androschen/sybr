// CRITICAL: This log should appear if App.jsx loads
console.log('🔥🔥🔥 App.jsx FILE LOADED')

import React, { useState, useEffect, useRef, useCallback } from 'react'
import './App.css'
import WindowMonitor from './components/WindowMonitor'
import HistoryLog from './components/HistoryLog'
import AutoStartSettings from './components/AutoStartSettings'
import BlocklistSettings from './components/BlocklistSettings'
import WarningModal from './components/WarningModal'
import { EventsOn } from './wailsjs/runtime/runtime'

// We'll use window.go.main.App directly - that's what Wails provides
// The generated bindings are just wrappers, but window.go.main.App is the source
console.log('📦 Will use window.go.main.App for backend calls')

function App() {
  const renderTime = new Date().toISOString()
  console.log(`[${renderTime}] 🚀🚀🚀 App component rendering...`)
  
  const [currentWindow, setCurrentWindow] = useState(null)
  const [history, setHistory] = useState([])
  const [autoStartEnabled, setAutoStartEnabled] = useState(false)
  const historyRef = useRef([])
  const lastWindowRef = useRef(null)
  
  // Log state changes
  useEffect(() => {
    console.log('📊 State changed - currentWindow:', currentWindow)
  }, [currentWindow])
  
  useEffect(() => {
    console.log('📊 State changed - history length:', history.length)
  }, [history])
  
  useEffect(() => {
    console.log('📊 State changed - autoStartEnabled:', autoStartEnabled)
  }, [autoStartEnabled])
  
  // Debug: Log window object
  useEffect(() => {
    const checkTime = new Date().toISOString()
    console.log(`[${checkTime}] 🔍 Window object check:`, {
      window: typeof window !== 'undefined',
      windowGo: typeof window.go !== 'undefined',
      windowGoMain: typeof window.go?.main !== 'undefined',
      windowGoMainApp: typeof window.go?.main?.App !== 'undefined',
      windowRuntime: typeof window.runtime !== 'undefined',
      windowRuntimeEventsOn: typeof window.runtime?.EventsOn,
    })
  }, [])

  // Add to history function - use useCallback to ensure it's stable
  const addToHistory = useCallback((windowInfo) => {
    console.log('📝 addToHistory called with:', windowInfo)
    if (!windowInfo) {
      console.log('❌ addToHistory: windowInfo is null/undefined, skipping')
      return
    }
    
    // Skip if same window
    if (lastWindowRef.current && 
        lastWindowRef.current.title === windowInfo.title && 
        lastWindowRef.current.exe === windowInfo.exe) {
      console.log('⏭️ addToHistory: Same window, skipping duplicate:', {
        current: lastWindowRef.current,
        new: windowInfo
      })
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
    
    console.log('✅ addToHistory: Adding entry to history:', entry)
    lastWindowRef.current = windowInfo
    historyRef.current = [entry, ...historyRef.current].slice(0, 200)
    console.log(`📊 History updated: ${historyRef.current.length} entries`)
    setHistory([...historyRef.current])
    console.log('✅ History state updated')
  }, [])

  useEffect(() => {
    let unsubscribe = null
    let pollInterval = null
    let isInitialized = false

    // Function to update window info
    const updateWindow = (windowInfo) => {
      const timestamp = new Date().toISOString()
      console.log(`[${timestamp}] 🔄 updateWindow called with:`, windowInfo)
      
      if (!windowInfo) {
        console.log(`[${timestamp}] ❌ updateWindow: windowInfo is null/undefined`)
        return
      }
      
      // Check if window info is valid (has at least title or exe)
      if (!windowInfo.title && !windowInfo.exe) {
        console.log(`[${timestamp}] ⚠️ updateWindow: windowInfo has no title or exe:`, windowInfo)
        return
      }

      // Only update if changed
      const current = lastWindowRef.current
      if (current && 
          current.title === windowInfo.title && 
          current.exe === windowInfo.exe) {
        console.log(`[${timestamp}] ⏭️ updateWindow: No change detected, skipping`)
        return // No change
      }

      console.log(`[${timestamp}] ✅ updateWindow: Window changed!`, {
        from: current,
        to: windowInfo
      })
      console.log(`[${timestamp}] 📤 Setting currentWindow state...`)
      setCurrentWindow(windowInfo)
      console.log(`[${timestamp}] 📤 Calling addToHistory...`)
      addToHistory(windowInfo)
      console.log(`[${timestamp}] ✅ updateWindow completed`)
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
      console.log('🎧 Setting up event listener...')
      try {
        if (typeof EventsOn === 'function') {
          console.log('✓ EventsOn function found (imported)')
          unsubscribe = EventsOn('window-changed', (windowInfo) => {
            const timestamp = new Date().toISOString()
            console.log(`[${timestamp}] 📡📡📡 EVENT RECEIVED (EventsOn):`, windowInfo)
            console.log(`[${timestamp}] 📡 Event data type:`, typeof windowInfo)
            console.log(`[${timestamp}] 📡 Event data keys:`, windowInfo ? Object.keys(windowInfo) : 'null')
            updateWindow(windowInfo)
          })
          console.log('✅ Event listener registered successfully (EventsOn import)')
        } else if (window.runtime && typeof window.runtime.EventsOn === 'function') {
          console.log('✓ window.runtime.EventsOn function found')
          unsubscribe = window.runtime.EventsOn('window-changed', (windowInfo) => {
            const timestamp = new Date().toISOString()
            console.log(`[${timestamp}] 📡📡📡 EVENT RECEIVED (window.runtime):`, windowInfo)
            console.log(`[${timestamp}] 📡 Event data type:`, typeof windowInfo)
            console.log(`[${timestamp}] 📡 Event data keys:`, windowInfo ? Object.keys(windowInfo) : 'null')
            updateWindow(windowInfo)
          })
          console.log('✅ Event listener registered successfully (window.runtime)')
        } else {
          console.warn('⚠️ EventsOn not available, using polling only')
          console.warn('⚠️ EventsOn type:', typeof EventsOn)
          console.warn('⚠️ window.runtime.EventsOn type:', typeof window.runtime?.EventsOn)
        }
      } catch (err) {
        console.error('❌ Error setting up event listener:', err)
        console.error('❌ Error stack:', err.stack)
      }

      // Get initial window immediately
      const fetchInitialWindow = async () => {
        try {
          console.log('🔍 Fetching initial window...')
          // Use window.go.main.App directly
          let windowInfo
          if (window.go?.main?.App?.GetCurrentWindow) {
            console.log('✓ Using window.go.main.App.GetCurrentWindow')
            windowInfo = await window.go.main.App.GetCurrentWindow()
          } else {
            console.error('❌ GetCurrentWindow not available - window.go:', window.go)
            return
          }
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
      console.log('⏰ Setting up polling interval (1 second)...')
      let pollCount = 0
      pollInterval = setInterval(async () => {
        pollCount++
        const timestamp = new Date().toISOString()
        console.log(`[${timestamp}] 🔄 Poll #${pollCount}: Fetching current window...`)
        try {
          let windowInfo
          if (window.go?.main?.App?.GetCurrentWindow) {
            console.log(`[${timestamp}] 📞 Calling window.go.main.App.GetCurrentWindow...`)
            windowInfo = await window.go.main.App.GetCurrentWindow()
            console.log(`[${timestamp}] 📦 GetCurrentWindow result:`, windowInfo)
          } else {
            console.warn(`[${timestamp}] ⚠️ GetCurrentWindow not available`)
            return
          }
          if (windowInfo && (windowInfo.title || windowInfo.exe)) {
            console.log(`[${timestamp}] ✅ Poll #${pollCount}: Valid window info, calling updateWindow`)
            updateWindow(windowInfo)
          } else {
            console.log(`[${timestamp}] ⏭️ Poll #${pollCount}: Invalid or empty window info, skipping`)
          }
        } catch (err) {
          console.error(`[${timestamp}] ❌ Poll #${pollCount} error:`, err)
        }
      }, 1000)
      console.log('✅ Polling started (1s interval)')

      // Check auto-start status
      try {
        let enabled
        if (window.go?.main?.App?.IsAutoStartEnabled) {
          enabled = await window.go.main.App.IsAutoStartEnabled()
        }
        if (enabled !== undefined) {
          setAutoStartEnabled(enabled)
          console.log('✓ Auto-start status:', enabled)
        }
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
    console.log('🗑️ clearHistory called')
    if (window.confirm('Are you sure you want to clear the history?')) {
      console.log('✅ User confirmed, clearing history...')
      historyRef.current = []
      setHistory([])
      console.log('✅ History cleared')
    } else {
      console.log('❌ User cancelled history clear')
    }
  }

  const handleEnableAutoStart = async () => {
    console.log('🔧 handleEnableAutoStart called')
    try {
      if (window.go?.main?.App?.EnableAutoStart) {
        console.log('📞 Calling window.go.main.App.EnableAutoStart...')
        await window.go.main.App.EnableAutoStart()
      } else {
        console.error('❌ EnableAutoStart not available')
        alert('Wails bindings not available')
        return
      }
      console.log('✅ Auto-start enabled, updating state...')
      setAutoStartEnabled(true)
      console.log('✅ Auto-start state updated')
    } catch (err) {
      console.error('❌ Error enabling auto-start:', err)
      alert('Failed to enable auto-start: ' + err)
    }
  }

  const handleDisableAutoStart = async () => {
    console.log('🔧 handleDisableAutoStart called')
    try {
      if (window.go?.main?.App?.DisableAutoStart) {
        console.log('📞 Calling window.go.main.App.DisableAutoStart...')
        await window.go.main.App.DisableAutoStart()
      } else {
        console.error('❌ DisableAutoStart not available')
        alert('Wails bindings not available')
        return
      }
      console.log('✅ Auto-start disabled, updating state...')
      setAutoStartEnabled(false)
      console.log('✅ Auto-start state updated')
    } catch (err) {
      console.error('❌ Error disabling auto-start:', err)
      alert('Failed to disable auto-start: ' + err)
    }
  }

  return (
    <div className="app">
      <div className="container">
        <header className="app-header">
          <h1>Window Monitor</h1>
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

          <div className="card">
            <BlocklistSettings />
          </div>

          <div className="card card-full">
            <HistoryLog 
              history={history} 
              onClear={clearHistory}
            />
          </div>
        </div>

        <WarningModal />
      </div>
    </div>
  )
}

export default App

