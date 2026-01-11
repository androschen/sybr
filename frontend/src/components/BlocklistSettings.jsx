import React, { useState, useEffect } from 'react'
import './BlocklistSettings.css'

function BlocklistSettings() {
  console.log('🎨 BlocklistSettings component rendering')
  
  const [blocklist, setBlocklist] = useState([])
  const [newAppName, setNewAppName] = useState('')
  const [newDisplayName, setNewDisplayName] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  // Load blocklist on mount
  useEffect(() => {
    console.log('🔄 BlocklistSettings useEffect - loading blocklist on mount')
    loadBlocklist()
  }, [])

  const loadBlocklist = async () => {
    try {
      console.log('🔍 Loading blocklist...')
      if (window.go?.main?.App?.GetBlocklist) {
        console.log('📞 Calling GetBlocklist...')
        const apps = await window.go.main.App.GetBlocklist()
        console.log('✅ GetBlocklist result:', apps)
        console.log('✅ GetBlocklist result type:', typeof apps, Array.isArray(apps))
        console.log('✅ GetBlocklist result length:', apps ? apps.length : 0)
        
        if (apps && Array.isArray(apps)) {
          setBlocklist(apps)
          console.log('✅ Blocklist state updated with', apps.length, 'apps')
        } else {
          console.warn('⚠️ GetBlocklist returned non-array:', apps)
          setBlocklist([])
        }
      } else {
        console.error('❌ GetBlocklist not available')
        console.log('Available methods:', window.go?.main?.App ? Object.keys(window.go.main.App) : [])
        setError('GetBlocklist method not available')
      }
    } catch (err) {
      console.error('❌ Error loading blocklist:', err)
      setError('Failed to load blocklist: ' + err.message)
    }
  }

  const handleAdd = async () => {
    console.log('🚀 handleAdd called!', {
      newAppName: newAppName,
      newDisplayName: newDisplayName,
      trimmed: newAppName.trim()
    })

    if (!newAppName.trim()) {
      console.log('❌ No app name provided')
      setError('Please enter an executable name')
      return
    }

    console.log('✅ App name valid, setting loading state...')
    setLoading(true)
    setError('')
    
    // Safety timeout - reset loading after 10 seconds no matter what
    const safetyTimeout = setTimeout(() => {
      console.error('⚠️ Safety timeout triggered - resetting loading state')
      setLoading(false)
      setError('Operation timed out. Please try again.')
    }, 10000)

    try {
      console.log('🔍 Checking bindings...', {
        hasGo: typeof window.go !== 'undefined',
        hasMain: typeof window.go?.main !== 'undefined',
        hasApp: typeof window.go?.main?.App !== 'undefined',
        hasAddToBlocklist: typeof window.go?.main?.App?.AddToBlocklist === 'function',
        availableMethods: window.go?.main?.App ? Object.keys(window.go.main.App) : []
      })

      if (window.go?.main?.App?.AddToBlocklist) {
        console.log('📞 Calling AddToBlocklist with:', newAppName.trim(), newDisplayName.trim())
        
        // Store the app name before clearing
        const addedName = newAppName.trim()
        const addedDisplayName = newDisplayName.trim()
        
        // Call the method with timeout protection
        let addResult
        try {
          const addPromise = window.go.main.App.AddToBlocklist(addedName, addedDisplayName)
          const timeoutPromise = new Promise((_, reject) => 
            setTimeout(() => reject(new Error('AddToBlocklist timed out after 5 seconds')), 5000)
          )
          
          addResult = await Promise.race([addPromise, timeoutPromise])
          console.log('✅ AddToBlocklist completed, result:', addResult)
        } catch (err) {
          console.error('❌ AddToBlocklist error:', err)
          console.error('❌ Error stack:', err.stack)
          throw err // Re-throw to be caught by outer catch
        }
        
        // Clear inputs immediately after successful add
        setNewAppName('')
        setNewDisplayName('')
        
        // Small delay to ensure file is written, then reload
        console.log('⏳ Waiting 200ms before reloading...')
        await new Promise(resolve => setTimeout(resolve, 200))
        
        // Reload the blocklist
        console.log('🔄 Reloading blocklist...')
        try {
          await loadBlocklist()
          console.log('✅ Blocklist reloaded after adding:', addedName)
        } catch (err) {
          console.error('❌ Error reloading blocklist:', err)
          // Don't throw - app was added successfully, just refresh failed
          setError('App added but failed to refresh list. Please refresh the page.')
        }
        
        console.log('✅ All done, setting loading to false')
        clearTimeout(safetyTimeout)
        setLoading(false)
        console.log('✅ Loading state reset')
      } else {
        console.error('❌ AddToBlocklist not available')
        clearTimeout(safetyTimeout)
        setError('Wails bindings not available. Please restart the app.')
        setLoading(false)
      }
    } catch (err) {
      console.error('❌ Error adding to blocklist:', err)
      console.error('❌ Error details:', {
        message: err.message,
        stack: err.stack,
        name: err.name
      })
      clearTimeout(safetyTimeout)
      setError(err.message || 'Failed to add app to blocklist')
      setLoading(false)
      console.log('✅ Loading set to false after error')
    }
  }

  const handleRemove = async (executableName) => {
    setLoading(true)
    setError('')

    try {
      if (window.go?.main?.App?.RemoveFromBlocklist) {
        await window.go.main.App.RemoveFromBlocklist(executableName)
        await loadBlocklist()
      } else {
        setError('Wails bindings not available')
      }
    } catch (err) {
      console.error('Error removing from blocklist:', err)
      setError(err.message || 'Failed to remove app from blocklist')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="blocklist-settings">
      <h2>Focus Blocker</h2>
      <p className="blocklist-description">
        Add apps to block. You'll receive a warning when these apps are opened.
      </p>

      {error && (
        <div className="blocklist-error">
          {error}
        </div>
      )}

      <div className="blocklist-add">
        <div className="blocklist-input-group">
          <input
            type="text"
            placeholder="Executable name (e.g., chrome.exe)"
            value={newAppName}
            onChange={(e) => {
              console.log('📝 Input changed:', e.target.value)
              setNewAppName(e.target.value)
            }}
            onKeyPress={(e) => {
              if (e.key === 'Enter') {
                console.log('⌨️ Enter key pressed in executable input')
                handleAdd()
              }
            }}
            className="blocklist-input"
            disabled={loading}
          />
          <input
            type="text"
            placeholder="Display name (optional)"
            value={newDisplayName}
            onChange={(e) => {
              console.log('📝 Display name changed:', e.target.value)
              setNewDisplayName(e.target.value)
            }}
            onKeyPress={(e) => {
              if (e.key === 'Enter') {
                console.log('⌨️ Enter key pressed in display name input')
                handleAdd()
              }
            }}
            className="blocklist-input"
            disabled={loading}
          />
          <button
            onClick={(e) => {
              console.log('🔘 Add button clicked!', {
                newAppName: newAppName,
                loading: loading,
                hasHandler: typeof handleAdd === 'function'
              })
              e.preventDefault()
              handleAdd()
            }}
            className="btn btn-primary"
            disabled={loading || !newAppName.trim()}
          >
            {loading ? 'Adding...' : 'Add'}
          </button>
        </div>
      </div>

      <div className="blocklist-list">
        <h3>Blocked Apps ({blocklist.length})</h3>
        {blocklist.length === 0 ? (
          <div className="blocklist-empty">
            <p>No apps blocked yet.</p>
            <p className="blocklist-empty-sub">Add an executable name above to get started.</p>
          </div>
        ) : (
          <div className="blocklist-items">
            {blocklist.map((app, index) => {
              // Handle both camelCase (from JSON) and potential other formats
              const executableName = app.executableName || app.ExecutableName || ''
              const displayName = app.displayName || app.DisplayName || executableName
              
              return (
                <div key={index} className="blocklist-item">
                  <div className="blocklist-item-info">
                    <div className="blocklist-item-name">{displayName}</div>
                    <div className="blocklist-item-exe">{executableName}</div>
                  </div>
                  <button
                    onClick={() => handleRemove(executableName)}
                    className="btn btn-danger btn-small"
                    disabled={loading}
                  >
                    Remove
                  </button>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}

export default BlocklistSettings
