import { Lock, Loader2 } from 'lucide-react'
import type { Plan } from '../../services/billingApi'

interface UpgradePromptProps {
  plan: Plan
  onUpgrade: () => Promise<void> | void
  loading?: boolean
  error?: string | null
  className?: string
}

export function UpgradePrompt({ plan, onUpgrade, loading = false, error = null, className = '' }: UpgradePromptProps) {
  const isPro = plan === 'pro'
  const ctaLabel = isPro ? 'Upgrade to Team — $99/mo →' : 'Upgrade to Pro — $29/mo →'
  const body = isPro
    ? "You've reached your 10-entity limit on the Pro plan. Upgrade to Team for unlimited entities."
    : "You've reached your 3-entity limit on the Free plan. Upgrade to Pro to add up to 10 agents and humans."

  return (
    <div className={`rounded-lg border border-amber-300 bg-amber-50 p-4 dark:border-amber-700 dark:bg-amber-950/30 ${className}`}>
      <div className="flex items-start gap-3">
        <Lock className="mt-0.5 h-4 w-4 text-amber-500" aria-hidden="true" />
        <div className="min-w-0 flex-1">
          <h3 className="font-semibold text-amber-800 dark:text-amber-300">Entity limit reached</h3>
          <p className="mt-1 text-sm text-amber-700 dark:text-amber-400">{body}</p>

          <div className="mt-4 flex flex-wrap items-center gap-3">
            <button
              type="button"
              onClick={() => void onUpgrade()}
              disabled={loading}
              className="inline-flex items-center gap-2 rounded-lg bg-emerald-500 px-3 py-2 text-sm font-medium text-white transition hover:bg-emerald-600 disabled:opacity-60"
            >
              {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
              {loading ? 'Redirecting to Stripe…' : ctaLabel}
            </button>
            <a href="https://agentunited.ai/pricing" target="_blank" rel="noreferrer" className="text-sm font-medium text-amber-800 underline underline-offset-4 hover:text-amber-900 dark:text-amber-300 dark:hover:text-amber-200">
              View pricing
            </a>
          </div>

          {error ? <p className="mt-2 text-sm text-red-600">{error}</p> : null}
        </div>
      </div>
    </div>
  )
}
