# Build Frontend Script
# This script builds the React frontend for Wails

Write-Host "🔨 Building Frontend..." -ForegroundColor Cyan

# Check if we're in the right directory
if (-not (Test-Path "package.json")) {
    Write-Host "❌ Error: package.json not found. Make sure you're in the frontend directory." -ForegroundColor Red
    Write-Host "   Current directory: $(Get-Location)" -ForegroundColor Yellow
    exit 1
}

# Check for npm
$npmPath = Get-Command npm -ErrorAction SilentlyContinue
if (-not $npmPath) {
    Write-Host "❌ Error: npm not found in PATH" -ForegroundColor Red
    Write-Host "   Please install Node.js from https://nodejs.org/" -ForegroundColor Yellow
    Write-Host "   Or add npm to your PATH" -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ npm found: $($npmPath.Source)" -ForegroundColor Green

# Check if node_modules exists
if (-not (Test-Path "node_modules")) {
    Write-Host "📦 Installing dependencies..." -ForegroundColor Cyan
    npm install
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Error: npm install failed" -ForegroundColor Red
        exit 1
    }
    Write-Host "✓ Dependencies installed" -ForegroundColor Green
} else {
    Write-Host "✓ Dependencies already installed" -ForegroundColor Green
}

# Build the frontend
Write-Host "🔨 Building React app..." -ForegroundColor Cyan
npm run build

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Error: Build failed" -ForegroundColor Red
    exit 1
}

# Verify build output
Write-Host "`n✅ Build completed! Verifying output..." -ForegroundColor Green

$distIndex = "dist/index.html"
if (Test-Path $distIndex) {
    $content = Get-Content $distIndex -Raw
    if ($content -match '<script.*src=.*assets/.*\.js') {
        Write-Host "✓ dist/index.html contains built scripts" -ForegroundColor Green
    } else {
        Write-Host "⚠️  Warning: dist/index.html might not have proper script tags" -ForegroundColor Yellow
    }
} else {
    Write-Host "❌ Error: dist/index.html not found after build" -ForegroundColor Red
    exit 1
}

$assetFiles = Get-ChildItem "dist/assets" -ErrorAction SilentlyContinue
if ($assetFiles) {
    Write-Host "✓ Found $($assetFiles.Count) asset files in dist/assets/" -ForegroundColor Green
} else {
    Write-Host "⚠️  Warning: No asset files found in dist/assets/" -ForegroundColor Yellow
}

Write-Host "`n✅ Frontend build complete! You can now run 'wails dev'" -ForegroundColor Green
