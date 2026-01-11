# Install Node.js and npm

## Quick Installation

### Option 1: Download and Install (Recommended)

1. **Download Node.js:**
   - Go to: https://nodejs.org/
   - Download the **LTS version** (recommended)
   - Choose the Windows Installer (.msi) for your system (64-bit)

2. **Install Node.js:**
   - Run the downloaded installer
   - Follow the installation wizard
   - **Important:** Make sure to check "Add to PATH" during installation
   - Complete the installation

3. **Verify Installation:**
   - Open a **NEW** PowerShell or Command Prompt window
   - Run:
     ```powershell
     node --version
     npm --version
     ```
   - You should see version numbers

4. **Build the Frontend:**
   ```powershell
   cd C:\Users\andro\Documents\Code\tools\sybr\frontend
   npm install
   npm run build
   ```

5. **Run Wails:**
   ```powershell
   cd C:\Users\andro\Documents\Code\tools\sybr
   wails dev
   ```

### Option 2: Use Chocolatey (Requires Admin)

If you have administrator access, you can run:
```powershell
# Run PowerShell as Administrator, then:
choco install nodejs -y
```

Then restart your terminal and follow steps 4-5 above.

## After Installation

Once Node.js is installed:
1. **Close and reopen** your terminal/PowerShell
2. Navigate to the frontend directory
3. Run `npm install` (first time only)
4. Run `npm run build`
5. Run `wails dev`

The React app should then load properly!
