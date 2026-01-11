// CRITICAL: This log should appear FIRST if the file loads
console.log('🔥🔥🔥 main.jsx FILE LOADED - This is the FIRST log')

import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'

// Import Wails runtime - this is automatically injected by Wails
// The bindings will be available on window.go.main.App

console.log('🎬 main.jsx: Starting React app...')
console.log('🔍 Window check:', {
  hasWindow: typeof window !== 'undefined',
  hasGo: typeof window?.go !== 'undefined',
  hasRuntime: typeof window?.runtime !== 'undefined',
})

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)

console.log('✅ React app rendered')
