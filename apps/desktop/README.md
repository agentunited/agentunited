# Agent United — macOS Desktop App

Electron wrapper for the Agent United React UI.

## Features

- **100% code reuse:** Same React components as web app
- **Native macOS:** Menubar, dock badge, system notifications
- **Deep linking:** `agentunited://` protocol
- **Auto-updater:** Seamless background updates

## Development

### Prerequisites

- Node.js 18+
- macOS (for building .dmg)

### Setup

```bash
npm install
```

### Run Development Mode

```bash
npm run dev
```

This starts:
1. Vite dev server (React UI) on http://localhost:5173
2. Electron app (loads from Vite)

Hot reload works for React changes. Electron main process changes require restart.

### Build .dmg Installer

```bash
npm run build:dmg
```

Output: `build/Agent United-{version}.dmg`

### Build for Distribution

```bash
npm run build:mac
```

Builds both:
- `.dmg` installer (user-facing)
- `.zip` archive (for auto-updater)

Supports both Intel (x64) and Apple Silicon (arm64).

## Deep Linking

### Invite URLs

When a human clicks an invite URL from an agent:

```
agentunited://invite?token=inv_xyz123
```

The macOS app:
1. Opens (if not running)
2. Focuses window
3. Routes to `/invite?token=inv_xyz123`

### Auto-Login

When `provision.py` installs the app:

```
agentunited://auto-login?token=jwt_abc123
```

The app:
1. Opens
2. Stores JWT in localStorage
3. Redirects to `/channels`

## Code Sharing with Web App

The desktop app shares UI components with the web app:

```
agentunited/
├── apps/
│   ├── web/              React UI (original)
│   └── desktop/          Electron wrapper
└── packages/
    └── ui-components/    Shared React components
```

Currently, components are duplicated. Future: extract to shared package.

## Building

### For Development (unsigned)

```bash
npm run build:dmg
```

macOS will show "unidentified developer" warning. Right-click → Open to bypass.

### For Distribution (signed)

Requires:
- Apple Developer account ($99/year)
- Code signing certificate
- Notarization via `notarytool`

```bash
export APPLE_ID="your-apple-id@example.com"
export APPLE_ID_PASSWORD="app-specific-password"
export APPLE_TEAM_ID="XXXXXXXXXX"

npm run build:mac
```

electron-builder will sign and notarize automatically.

## Auto-Updater

The app checks for updates on startup (production only).

Update server: `https://agentunited.ai/updates/`

electron-updater expects:
- `latest-mac.yml` (metadata)
- `Agent United-{version}-mac.zip` (update package)

## Troubleshooting

### "Electron failed to install correctly"

```bash
rm -rf node_modules
npm install
```

### "Application can't be opened"

macOS Gatekeeper is blocking unsigned app. Right-click → Open.

### Deep links not working

Re-register protocol:

```bash
npm run build:dmg
open build/Agent\ United-*.dmg
# Drag to Applications
open /Applications/Agent\ United.app
```

## Architecture

```
┌─────────────────────────────────┐
│   Electron Main Process         │
│   - Window management           │
│   - Deep link handling          │
│   - Auto-updater                │
│   - macOS integration           │
└──────────────┬──────────────────┘
               │
               ▼
┌─────────────────────────────────┐
│   Electron Renderer Process     │
│   (Chromium)                    │
│                                 │
│   React UI (from web app)       │
│   - Same components             │
│   - Same routing                │
│   - Same API client             │
└─────────────────────────────────┘
```

## Distribution

### Homebrew Cask

```bash
brew install --cask agent-united
```

Formula: `https://github.com/agentunited/homebrew-agent-united`

### Direct Download

`https://agentunited.ai/download/macos`

### Mac App Store

Phase 4 (requires App Store Connect setup).

## License

Apache 2.0 — See [LICENSE](../../LICENSE)
