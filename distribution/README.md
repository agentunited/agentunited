# Agent United Distribution

Distribution files for Agent United desktop apps.

---

## Homebrew (macOS)

### Installation

```bash
brew install --cask agent-united
```

### Cask Formula

Location: `homebrew/agent-united.rb`

**Publishing to Homebrew:**

1. Fork `homebrew/homebrew-cask`
2. Add `Casks/agent-united.rb` (copy from `homebrew/agent-united.rb`)
3. Submit pull request
4. Once merged, users can install via `brew install --cask agent-united`

**For development:**

```bash
# Test cask locally
brew install --cask homebrew/agent-united.rb

# Audit cask
brew audit --cask homebrew/agent-united.rb

# Uninstall
brew uninstall --cask agent-united
```

---

## Direct Download (macOS)

**URL:** `https://agentunited.ai/download/macos`

**Hosting:**
- Store `.dmg` files on CDN or GitHub Releases
- Serve via nginx/CloudFront
- Update `agentunited.ai/download/macos` redirect to latest version

**Example nginx config:**
```nginx
location /download/macos {
    return 302 https://github.com/agentunited/agentunited/releases/download/v0.1.0/Agent-United-0.1.0.dmg;
}
```

---

## Mac App Store (Phase 4)

**Requirements:**
- Apple Developer account ($99/year)
- App Store Connect setup
- Code signing certificate
- App review process (1-2 weeks)

**Steps:**
1. Enroll in Apple Developer Program
2. Create App Store Connect app
3. Configure app metadata (screenshots, description)
4. Build with `npm run build:mas` (Mac App Store target)
5. Upload with `altool` or Transporter.app
6. Submit for review

**Not implemented yet** - Phase 4 roadmap item.

---

## Windows / Linux (Phase 3)

**Not yet supported** - macOS priority for Phase 2.

**Future distribution channels:**
- Windows: Microsoft Store, Chocolatey, direct `.exe`
- Linux: Snap Store, Flatpak, `.deb`/`.rpm`, AppImage

---

## Auto-Updater

The macOS app uses `electron-updater` for automatic updates.

**Update server requirements:**
- `latest-mac.yml` (metadata file)
- `Agent-United-{version}-mac.zip` (update package)

**Hosting options:**
- GitHub Releases (free, automatic)
- S3 + CloudFront (full control)
- Custom update server

**Example GitHub Releases config:**

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: npm install
      - run: npm run build:mac
      - uses: softprops/action-gh-release@v1
        with:
          files: |
            build/Agent-United-*.dmg
            build/Agent-United-*-mac.zip
            build/latest-mac.yml
```

---

## Code Signing (macOS)

**Development (unsigned):**
```bash
npm run build:dmg
```

Users see "unidentified developer" warning. Right-click → Open to bypass.

**Production (signed + notarized):**

```bash
# Set environment variables
export APPLE_ID="your-apple-id@example.com"
export APPLE_ID_PASSWORD="app-specific-password"
export APPLE_TEAM_ID="XXXXXXXXXX"

# Build (electron-builder will sign + notarize automatically)
npm run build:mac
```

**Requirements:**
- Developer ID Application certificate (from Apple Developer)
- App-specific password (from appleid.apple.com)
- Team ID (from developer.apple.com)

**electron-builder config** (in `package.json`):
```json
{
  "build": {
    "mac": {
      "hardenedRuntime": true,
      "gatekeeperAssess": false,
      "entitlements": "entitlements.mac.plist"
    },
    "afterSign": "notarize.js"
  }
}
```

---

## Release Checklist

**Before each release:**

- [ ] Bump version in `package.json`
- [ ] Update `CHANGELOG.md`
- [ ] Build macOS app: `npm run build:mac`
- [ ] Test .dmg install manually
- [ ] Test auto-updater (if not first release)
- [ ] Upload to GitHub Releases
- [ ] Update Homebrew cask formula
- [ ] Update `agentunited.ai/download` redirect
- [ ] Announce release (Discord, Twitter, website)

---

## Distribution Metrics

**Track:**
- Download count (GitHub Releases API)
- Install count (Homebrew analytics)
- Update success rate (auto-updater telemetry)
- Platform distribution (macOS Intel vs Apple Silicon)

**GitHub Releases download count:**
```bash
curl https://api.github.com/repos/agentunited/agentunited/releases/latest | \
  jq '.assets[] | {name: .name, downloads: .download_count}'
```

---

## Support

- **Docs:** https://agentunited.ai/docs/installation
- **Issues:** https://github.com/agentunited/agentunited/issues
- **Discord:** https://discord.com/invite/agent-united
