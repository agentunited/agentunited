import { Link } from 'react-router-dom'

const MAC_OS_URL = 'https://github.com/agentunited/agentunited/releases/latest'
const IOS_URL = 'https://testflight.apple.com/join/PLACEHOLDER'
const BROWSER_URL = '/login'

export function AppDownloadSection() {
  return (
    <div className="mx-auto max-w-xl rounded-2xl border border-border bg-card px-8 py-8 text-center shadow-sm">
      <h2 className="text-xl font-semibold text-foreground">
        Chat with your agent on any device.
      </h2>
      <p className="mt-2 text-sm text-muted-foreground">
        Native apps for iPhone and Mac, or use any browser — no setup required.
      </p>

      <div className="mt-6 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
        <a
          href={MAC_OS_URL}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-emerald-500 px-5 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-emerald-600 sm:w-auto"
        >
          <span>💻</span> Download for macOS
        </a>
        <a
          href={IOS_URL}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex w-full items-center justify-center gap-2 rounded-xl border border-border bg-background px-5 py-2.5 text-sm font-semibold text-foreground transition-colors hover:bg-muted sm:w-auto"
        >
          <span>🍎</span> Get on TestFlight (iOS)
        </a>
      </div>

      <Link
        to={BROWSER_URL}
        className="mt-3 inline-block text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        Use in browser →
      </Link>

      <p className="mt-4 text-xs text-muted-foreground">
        Already have the app? Open your invite link to launch it directly.
      </p>
    </div>
  )
}
