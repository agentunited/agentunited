#!/usr/bin/env bash
# testflight-upload.sh — Build and upload iOS app to TestFlight
# Usage: ./scripts/testflight-upload.sh [build_number]
# If build_number is omitted, auto-increments from project.pbxproj
#
# Prerequisites (already set up on Mac mini):
#   ~/.appstoreconnect/private_keys/AuthKey_6L9Q32LK6L.p8
#   ~/Library/MobileDevice/Provisioning Profiles/AgentUnited_iOS_Distribution.mobileprovision
#   Distribution cert imported in default keychain

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
XCPROJECT="$REPO_ROOT/apps/ios/AgentUnited.xcodeproj"
SCHEME="AgentUnited"
ARCHIVE="/tmp/AgentUnited.xcarchive"
EXPORT_DIR="/tmp/AgentUnited-export"
EXPORT_PLIST="/tmp/ExportOptions.plist"

KEY_ID="6L9Q32LK6L"
ISSUER_ID="a7974bf7-5c01-4e02-8672-00fca87e5c39"
APP_ID="6760973893"
BETA_GROUP_ID="d1d10e26-9293-4fa0-8dec-58e9b2309580"
TEAM_ID="G8QPW9SH2T"
BUNDLE_ID="ai.agentunited.app"

# ── Step 1: Build number ──────────────────────────────────────────────────────
PBXPROJ="$XCPROJECT/project.pbxproj"
CURRENT=$(grep -m1 'CURRENT_PROJECT_VERSION' "$PBXPROJ" | grep -o '[0-9]*')
if [[ -n "${1:-}" ]]; then
  BUILD_NUMBER="$1"
else
  BUILD_NUMBER=$((CURRENT + 1))
fi

if [[ "$BUILD_NUMBER" -le "$CURRENT" ]]; then
  echo "❌ Build number $BUILD_NUMBER ≤ current $CURRENT. Pass a higher number or omit to auto-increment."
  exit 1
fi

echo "▶ Build number: $BUILD_NUMBER (was $CURRENT)"

# Update project.pbxproj (all occurrences)
sed -i '' "s/CURRENT_PROJECT_VERSION = ${CURRENT};/CURRENT_PROJECT_VERSION = ${BUILD_NUMBER};/g" "$PBXPROJ"

# ── Step 2: Clean exports ──────────────────────────────────────────────────────
rm -rf "$ARCHIVE" "$EXPORT_DIR"

# ── Step 3: Archive ───────────────────────────────────────────────────────────
echo "▶ Archiving..."
xcodebuild archive \
  -project "$XCPROJECT" \
  -scheme "$SCHEME" \
  -configuration Release \
  -archivePath "$ARCHIVE" \
  -allowProvisioningUpdates \
  CODE_SIGN_STYLE=Manual \
  CODE_SIGNING_ALLOWED=YES \
  CODE_SIGNING_REQUIRED=YES \
  CODE_SIGN_IDENTITY="Apple Distribution" \
  DEVELOPMENT_TEAM="$TEAM_ID" \
  PROVISIONING_PROFILE_SPECIFIER="AgentUnited iOS Distribution" \
  | xcpretty --simple 2>/dev/null || true

# Verify archive exists
[[ -d "$ARCHIVE" ]] || { echo "❌ Archive not created"; exit 1; }
echo "✅ Archive: $ARCHIVE"

# ── Step 4: Export IPA ────────────────────────────────────────────────────────
echo "▶ Exporting IPA..."
cat > "$EXPORT_PLIST" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>method</key>
    <string>app-store-connect</string>
    <key>destination</key>
    <string>export</string>
    <key>signingStyle</key>
    <string>manual</string>
    <key>teamID</key>
    <string>$TEAM_ID</string>
    <key>provisioningProfiles</key>
    <dict>
        <key>$BUNDLE_ID</key>
        <string>AgentUnited iOS Distribution</string>
    </dict>
</dict>
</plist>
EOF

xcodebuild -exportArchive \
  -archivePath "$ARCHIVE" \
  -exportPath "$EXPORT_DIR" \
  -exportOptionsPlist "$EXPORT_PLIST" \
  | xcpretty --simple 2>/dev/null || true

IPA=$(ls "$EXPORT_DIR"/*.ipa 2>/dev/null | head -1)
[[ -n "$IPA" ]] || { echo "❌ IPA not found in $EXPORT_DIR"; exit 1; }
echo "✅ IPA: $IPA"

# ── Step 5: Upload ────────────────────────────────────────────────────────────
echo "▶ Uploading to TestFlight..."
xcrun altool --upload-app \
  --type ios \
  --file "$IPA" \
  --apiKey "$KEY_ID" \
  --apiIssuer "$ISSUER_ID" \
  --show-progress
echo "✅ Upload complete"

# ── Step 6: Export compliance + beta group ────────────────────────────────────
echo "▶ Setting export compliance + adding to beta group..."
pip3 install PyJWT requests --quiet --break-system-packages 2>/dev/null || true

python3 << PYEOF
import jwt, time, requests, os, sys

KEY_ID      = "$KEY_ID"
ISSUER_ID   = "$ISSUER_ID"
APP_ID      = "$APP_ID"
BETA_GROUP  = "$BETA_GROUP_ID"
BUILD_NUM   = "$BUILD_NUMBER"

key = open(os.path.expanduser(f"~/.appstoreconnect/private_keys/AuthKey_{KEY_ID}.p8")).read()
token = jwt.encode(
    {'iss': ISSUER_ID, 'iat': int(time.time()), 'exp': int(time.time()) + 1200, 'aud': 'appstoreconnect-v1'},
    key, algorithm='ES256', headers={'kid': KEY_ID}
)
hdrs = {'Authorization': f'Bearer {token}'}

# Wait for build to appear (Apple processing delay after altool upload)
build_id = None
for attempt in range(20):
    r = requests.get(
        f"https://api.appstoreconnect.apple.com/v1/builds"
        f"?filter[app.id]={APP_ID}&filter[version]={BUILD_NUM}&sort=-uploadedDate&limit=1",
        headers=hdrs
    )
    data = r.json().get('data', [])
    if data:
        build_id = data[0]['id']
        print(f"  Build found: {BUILD_NUM} (id={build_id})")
        break
    print(f"  Waiting for build to appear in ASC ({attempt+1}/20)...")
    time.sleep(15)

if not build_id:
    print("  Build never appeared — skipping compliance. Set it manually in App Store Connect.")
    sys.exit(0)

# Export compliance
r = requests.patch(
    f"https://api.appstoreconnect.apple.com/v1/builds/{build_id}",
    headers={**hdrs, 'Content-Type': 'application/json'},
    json={'data': {'type': 'builds', 'id': build_id, 'attributes': {'usesNonExemptEncryption': False}}}
)
print(f"  Export compliance: {r.status_code}" + ("" if r.ok else f" — {r.text}"))

# Add to beta group (correct endpoint)
r = requests.post(
    f"https://api.appstoreconnect.apple.com/v1/builds/{build_id}/relationships/betaGroups",
    headers={**hdrs, 'Content-Type': 'application/json'},
    json={'data': [{'type': 'betaGroups', 'id': BETA_GROUP}]}
)
print(f"  Beta group:        {r.status_code}" + ("" if r.ok else f" — {r.text}"))
PYEOF

# ── Done ──────────────────────────────────────────────────────────────────────
echo ""
echo "✅ Build $BUILD_NUMBER uploaded to TestFlight."
echo "   Apple processes in 5–15 min, then it appears in TestFlight."
echo "   Build number in project.pbxproj updated to $BUILD_NUMBER — commit it."
