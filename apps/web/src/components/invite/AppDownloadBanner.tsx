import { useState } from 'react'

interface AppDownloadBannerProps {
  inviteToken: string
  instanceUrl: string
}

export function AppDownloadBanner({ inviteToken, instanceUrl }: AppDownloadBannerProps) {
  const [dismissed, setDismissed] = useState(
    () => localStorage.getItem('au:app-banner-dismissed') === '1'
  )

  if (dismissed) return null

  const deepLink = `agentunited://invite?token=${encodeURIComponent(inviteToken)}&instance=${encodeURIComponent(instanceUrl)}`
  const macOSUrl = 'https://github.com/agentunited/agentunited/releases/latest'

  const dismiss = () => {
    localStorage.setItem('au:app-banner-dismissed', '1')
    setDismissed(true)
  }

  return (
    <div className="mb-4 rounded-2xl border border-slate-200/80 bg-white/80 p-4 shadow-sm backdrop-blur">
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-center gap-2.5">
          <span className="text-xl">📱</span>
          <div>
            <p className="text-sm font-semibold text-slate-900">Get the Agent United app</p>
            <p className="text-xs text-slate-500 mt-0.5">Open your invite link directly in the app.</p>
          </div>
        </div>
        <button
          onClick={dismiss}
          aria-label="Dismiss"
          className="text-slate-400 hover:text-slate-600 transition-colors mt-0.5 shrink-0"
        >
          <svg className="h-4 w-4" viewBox="0 0 16 16" fill="none">
            <path d="M12 4L4 12M4 4l8 8" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
          </svg>
        </button>
      </div>

      <div className="mt-3 flex flex-wrap gap-2">
        <a
          href={macOSUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 hover:bg-slate-50 transition-colors"
        >
          💻 <span>macOS — Download</span>
        </a>
      </div>

      <a
        href={deepLink}
        className="mt-2.5 inline-block text-xs font-medium text-emerald-600 hover:text-emerald-700 transition-colors"
      >
        Already installed? Open invite in app →
      </a>
    </div>
  )
}
