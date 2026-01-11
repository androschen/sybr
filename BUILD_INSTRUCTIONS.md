# Build Instructions

## The Problem
Wails is serving the placeholder `frontend/dist/index.html` because the React app hasn't been built yet.

## Solution: Build the Frontend

You need to build the React app so it creates the proper files in `frontend/dist/`.

### Step 1: Install Dependencies (if not done)
```bash
cd frontend
npm install
```

### Step 2: Build the Frontend
```bash
npm run build
```

This will:
- Compile the React app
- Create optimized bundles
- Generate `frontend/dist/index.html` with proper script tags
- Create all the JS and CSS files in `frontend/dist/assets/`

### Step 3: Run Wails
```bash
cd ..
wails dev
```

After building, the React app will be embedded and you should see:
- ✅ All the console logs we added
- ✅ The React UI loading
- ✅ Window monitoring working

## Note
After building, you'll need to rebuild (`npm run build`) every time you change the frontend code. This is not ideal for development, but it will work until we fix the Wails dev server proxying issue.
