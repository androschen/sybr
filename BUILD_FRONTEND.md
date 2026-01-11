# Building Frontend for Wails

Since Wails dev mode proxying isn't working, we need to build the frontend:

## Quick Build Command

```bash
cd frontend
npm run build
```

This will create the built files in `frontend/dist/` that Wails can embed.

## Then Run Wails

```bash
wails dev
```

**Note:** After building, you'll need to rebuild the frontend every time you make changes to see them. This is not ideal for development, but it will work.

## For True Dev Mode (Hot Reload)

We need to fix the Wails dev server proxying issue. The problem is that `wails dev` should automatically start Vite and proxy to it, but it's not working.
