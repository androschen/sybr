# Project Flow & Architecture Explanation

## 🏗️ Architecture Overview

This is a **Wails v2** application that combines:
- **Go Backend**: Handles Windows API calls, window monitoring, system tray
- **React Frontend**: Modern UI built with React + Vite
- **Wails Bridge**: Connects Go and React, handles bindings and events

```
┌─────────────────────────────────────────────────────────────┐
│                    Wails Application                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐          ┌──────────────┐                 │
│  │   Go Backend │ ◄─────►  │React Frontend│                 │
│  │              │  Events  │              │                 │
│  │ - Window API │  &       │ - UI Display │                 │
│  │ - Systray    │  Bindings│ - State Mgmt │                 │
│  │ - Auto-start │          │ - Components │                 │
│  └──────────────┘          └──────────────┘                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 📁 Project Structure

```
sybr/
├── main.go              # Entry point, Wails setup, systray
├── app.go               # App struct - exposed to frontend
├── watcher.go           # Window monitoring logic
├── autostart.go         # Windows Registry auto-start
├── wails.json           # Wails configuration
├── go.mod               # Go dependencies
│
└── frontend/
    ├── src/
    │   ├── main.jsx     # React entry point
    │   ├── App.jsx      # Main React component
    │   ├── components/  # React components
    │   └── wailsjs/     # Auto-generated Wails bindings
    ├── dist/            # Built frontend (embedded in Go binary)
    └── package.json     # Node dependencies
```

---

## 🔄 Data Flow: How It Works

### 1. **Application Startup** (`main.go`)

```
1. main() function runs
   ├─ Creates App instance
   ├─ Creates WindowWatcher (with nil context initially)
   ├─ Starts systray in goroutine
   └─ Calls wails.Run() to start Wails app

2. Wails OnStartup() callback
   ├─ Receives live Wails context
   ├─ Calls app.OnStartup(ctx)
   └─ This sets context on watcher and starts monitoring
```

**Key Point**: The Wails context is only available in `OnStartup`. This is why we:
- Create watcher with `nil` context initially
- Set the context in `OnStartup` via `watcher.SetContext(ctx)`
- Start monitoring after context is set

### 2. **Window Monitoring** (`watcher.go`)

```
WindowWatcher.monitorLoop() runs every 1 second:
│
├─ Calls GetActiveWindow()
│  ├─ Uses Windows API: GetForegroundWindow()
│  ├─ Gets window title via GetWindowTextW()
│  └─ Gets process name via GetModuleBaseNameW()
│
├─ Compares with previous window
│  └─ If changed → Emit event
│
└─ runtime.EventsEmit(ctx, "window-changed", WindowInfo)
   └─ Sends data to frontend via Wails event system
```

**Key Point**: Events only work if `ctx` is set (from `OnStartup`). That's why we had the "dead context" issue before.

### 3. **Frontend Reception** (`App.jsx`)

```
React App initializes:
│
├─ Waits for Wails bindings (window.go.main.App)
│  └─ Retries up to 50 times (200ms intervals)
│
├─ Sets up event listener:
│  └─ EventsOn('window-changed', callback)
│
├─ Fetches initial window:
│  └─ window.go.main.App.GetCurrentWindow()
│
└─ Starts polling fallback (every 1 second)
   └─ Also calls GetCurrentWindow() as backup
```

**Key Point**: The frontend uses both:
- **Events** (real-time, preferred)
- **Polling** (fallback if events fail)

### 4. **State Management** (`App.jsx`)

```
When window changes:
│
├─ updateWindow(windowInfo) called
│  ├─ setCurrentWindow(windowInfo)  → Updates UI
│  └─ addToHistory(windowInfo)       → Adds to history log
│
└─ React re-renders:
   ├─ WindowMonitor component shows new window
   └─ HistoryLog component adds new entry
```

---

## 🔌 Wails Bindings: How Go ↔ React Communication Works

### **Method Bindings** (Go → React)

In `app.go`, methods on the `App` struct are automatically exposed:

```go
// app.go
func (a *App) GetCurrentWindow() (*WindowInfo, error) { ... }
func (a *App) EnableAutoStart() error { ... }
```

Wails generates JavaScript bindings in `frontend/wailsjs/go/main/App.js`:

```javascript
// Frontend can call:
await window.go.main.App.GetCurrentWindow()
await window.go.main.App.EnableAutoStart()
```

**How it works:**
1. Wails scans `App` struct methods
2. Generates TypeScript/JavaScript bindings
3. Injects `window.go.main.App` into frontend
4. Frontend calls methods → Wails bridges to Go → Returns result

### **Events** (Go → React)

```go
// watcher.go
runtime.EventsEmit(ctx, "window-changed", windowInfo)
```

```javascript
// App.jsx
EventsOn('window-changed', (windowInfo) => {
  updateWindow(windowInfo)
})
```

**How it works:**
1. Go emits event with `runtime.EventsEmit()`
2. Wails event system broadcasts to frontend
3. Frontend listener receives data
4. React updates state → UI re-renders

---

## 🎯 Key Components Explained

### **main.go** - Application Entry Point

**Responsibilities:**
- Initialize Wails app
- Start systray
- Configure asset server (embedded frontend or dev proxy)
- Handle lifecycle (OnStartup, OnShutdown)

**Key Code:**
```go
wails.Run(&options.App{
    Bind: []interface{}{app},  // Exposes App methods to frontend
    OnStartup: func(ctx context.Context) {
        app.OnStartup(ctx)  // Sets context, starts monitoring
    },
})
```

### **app.go** - Frontend Interface

**Responsibilities:**
- Bridge between frontend and backend
- Expose methods to React (GetCurrentWindow, EnableAutoStart, etc.)
- Manage Wails context
- Control window visibility

**Key Methods:**
- `GetCurrentWindow()` - Called by frontend to get current window
- `OnStartup(ctx)` - Sets watcher context, starts monitoring
- `EnableAutoStart()` / `DisableAutoStart()` - Manage auto-start

### **watcher.go** - Window Monitoring Engine

**Responsibilities:**
- Poll Windows API for active window
- Detect window changes
- Emit Wails events when window changes
- Thread-safe state management

**Key Methods:**
- `GetActiveWindow()` - Uses Windows API to get current window
- `StartMonitoring()` - Starts polling loop
- `monitorLoop()` - Runs every 1 second, emits events on change

**Windows API Calls:**
- `GetForegroundWindow()` - Gets active window handle
- `GetWindowTextW()` - Gets window title
- `GetModuleBaseNameW()` - Gets process executable name

### **autostart.go** - Windows Registry Management

**Responsibilities:**
- Enable/disable auto-start via Windows Registry
- Check auto-start status
- Uses `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`

### **Frontend Components**

**App.jsx** - Main orchestrator
- Manages state (currentWindow, history, autoStartEnabled)
- Sets up event listeners
- Handles initialization
- Polling fallback

**WindowMonitor.jsx** - Displays current window
- Shows window title and executable
- Status indicator

**HistoryLog.jsx** - Shows window change history
- Displays list of window changes
- Auto-scrolls to top
- Terminal-style formatting

**AutoStartSettings.jsx** - Auto-start controls
- Toggle button
- Status display

---

## 🚀 Build & Run Process

### **Development Mode**

```bash
# Terminal 1: Start Vite dev server
cd frontend
npm run dev
# Runs on http://localhost:34115

# Terminal 2: Start Wails (proxies to Vite)
cd ..
wails dev
# Wails detects dev mode and proxies to Vite
```

**What happens:**
1. Vite serves React app with hot-reload
2. Wails proxies requests to Vite
3. Go backend runs, generates bindings
4. Frontend connects to backend via bindings

### **Production Build**

```bash
# 1. Build frontend
cd frontend
npm install
npm run build
# Creates frontend/dist/ with compiled React app

# 2. Build Wails app
cd ..
wails build
# Embeds frontend/dist/ into Go binary
# Creates build/bin/sybr.exe
```

**What happens:**
1. `npm run build` compiles React → `frontend/dist/`
2. `go:embed all:frontend/dist` embeds dist into Go binary
3. `wails build` creates single executable with embedded frontend

---

## 🔧 How to Extend This Project

### **Add a New Backend Method**

1. Add method to `App` struct in `app.go`:
```go
func (a *App) MyNewMethod(param string) (string, error) {
    return "result", nil
}
```

2. Rebuild:
```bash
wails dev  # or wails build
```

3. Use in frontend:
```javascript
const result = await window.go.main.App.MyNewMethod("test")
```

### **Add a New Event**

1. Emit event in Go:
```go
runtime.EventsEmit(ctx, "my-event", data)
```

2. Listen in frontend:
```javascript
EventsOn('my-event', (data) => {
  console.log('Received:', data)
})
```

### **Add a New Frontend Component**

1. Create component in `frontend/src/components/`
2. Import and use in `App.jsx`
3. Add CSS file for styling

### **Modify Window Monitoring**

Edit `watcher.go`:
- Change polling interval (currently 1 second)
- Add more window information (e.g., window class, process ID)
- Filter certain windows (e.g., ignore system windows)

---

## 🐛 Common Issues & Solutions

### **Bindings Not Found**
- **Cause**: Frontend loads before Wails injects bindings
- **Solution**: Retry logic in `App.jsx` (already implemented)

### **Events Not Working**
- **Cause**: Context is nil or not set
- **Solution**: Ensure `watcher.SetContext(ctx)` is called in `OnStartup`

### **Frontend Blank in Dev Mode**
- **Cause**: Wails not proxying to Vite
- **Solution**: Ensure Vite is running on port 34115, check `wails.json` config

### **Build Fails**
- **Cause**: Frontend not built
- **Solution**: Run `npm run build` in `frontend/` directory first

---

## 📚 Key Concepts

### **Wails Context**
- Only available in `OnStartup` callback
- Required for `runtime.EventsEmit()` and `runtime.WindowShow()`
- Must be passed to components that need it

### **Goroutines**
- `systray.Run()` runs in separate goroutine
- `watcher.monitorLoop()` runs in separate goroutine
- Allows concurrent execution (systray + monitoring + UI)

### **Event System**
- One-way: Go → React
- Real-time updates
- Type-safe (Wails handles JSON serialization)

### **Embedding Frontend**
- `//go:embed all:frontend/dist` embeds files at compile time
- Creates single executable with no external dependencies
- Production builds use embedded files, dev uses Vite proxy

---

## 🎓 Next Steps

1. **Add Features:**
   - Window filtering (ignore certain apps)
   - Time tracking per window
   - Export history to file
   - Customizable polling interval

2. **Improve UI:**
   - Add charts/graphs
   - Search/filter history
   - Dark/light theme toggle

3. **Add Tests:**
   - Unit tests for `watcher.go`
   - Integration tests for Wails bindings

4. **Deploy:**
   - Create installer (e.g., using Inno Setup)
   - Add auto-update mechanism
   - Code signing for Windows

---

## 📖 Resources

- **Wails Docs**: https://wails.io/docs/
- **Wails v2 Guide**: https://wails.io/docs/gettingstarted/
- **Windows API**: https://docs.microsoft.com/en-us/windows/win32/api/
- **React Docs**: https://react.dev/
- **Vite Docs**: https://vitejs.dev/

---

**Good luck with your project! 🚀**
