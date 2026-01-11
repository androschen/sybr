# How to Build and Run

## The Problem
Wails is serving a placeholder file because the React frontend hasn't been built yet.

## Solution: Build the Frontend

### Option 1: Use the Build Script (Recommended)

1. Open PowerShell in the `frontend` directory
2. Run:
   ```powershell
   ..\build-frontend.ps1
   ```

### Option 2: Manual Build

1. **Open a terminal/PowerShell**

2. **Navigate to the frontend directory:**
   ```powershell
   cd C:\Users\andro\Documents\Code\tools\sybr\frontend
   ```

3. **Install dependencies (if not done):**
   ```powershell
   npm install
   ```
   If `npm` is not found, you need to:
   - Install Node.js from https://nodejs.org/
   - Or add Node.js to your PATH

4. **Build the frontend:**
   ```powershell
   npm run build
   ```

5. **Verify the build:**
   - Check that `frontend/dist/index.html` exists
   - Check that `frontend/dist/assets/` contains JS and CSS files
   - The `index.html` should have `<script>` tags pointing to files in `assets/`

6. **Run Wails:**
   ```powershell
   cd ..
   wails dev
   ```

## What to Expect After Building

After building, when you run `wails dev`, you should see in the browser console:
- ✅ `🔥🔥🔥 main.jsx FILE LOADED`
- ✅ `🔥🔥🔥 App.jsx FILE LOADED`
- ✅ `🎬 main.jsx: Starting React app...`
- ✅ All the initialization logs
- ✅ The React UI should appear

## Troubleshooting

### npm not found
- Install Node.js from https://nodejs.org/
- Restart your terminal after installation
- Verify with: `npm --version`

### Build fails
- Check for error messages in the terminal
- Make sure all dependencies are installed: `npm install`
- Check `package.json` is correct

### Still seeing placeholder
- Make sure you built in the `frontend` directory
- Check that `frontend/dist/index.html` has script tags (not just the placeholder message)
- Verify `frontend/dist/assets/` has files
