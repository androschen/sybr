# Window Monitor Frontend

React-based frontend for the Window Monitor application.

## Setup

1. Make sure you have Node.js and npm installed
2. Install dependencies:
   ```bash
   npm install
   ```

## Development

Run the development server:
```bash
npm run dev
```

## Build

Build for production:
```bash
npm run build
```

The built files will be in the `dist/` directory, which Wails will use.

## Wails Integration

This React app integrates with Wails through:
- `window.go.main.App` - Go backend methods
- `window.runtime` - Wails runtime functions (EventsOn, etc.)

Make sure to run `wails build` after making changes to generate the TypeScript bindings.

