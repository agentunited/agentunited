import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { CheckCircle2, Loader2, ArrowRight } from 'lucide-react'
import { billingApi, type BillingStatus } from '../services/billingApi'

const MAX_WAIT_MS = 30_000
const POLL_MS = 3_000

type PaidPlan = 'pro' | 'team'

function planFromQuery(raw: string | null): PaidPlan | null {
  if (raw === 'pro' || raw === 'team') return raw
  return null
}

export function WelcomePage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const requestedPlan = planFromQuery(searchParams.get('plan'))

  const [status, setStatus] = useState<BillingStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [timedOut, setTimedOut] = useState(false)

  useEffect(() => {
    if (!requestedPlan) {
      navigate('/settings/billing', { replace: true })
      return
    }

    let cancelled = false
    const startedAt = Date.now()

    const poll = async () => {
      try {
        const nextStatus = await billingApi.getStatus()
        if (cancelled) return

        setStatus(nextStatus)

        if (nextStatus.plan === 'pro' || nextStatus.plan === 'team') {
          sessionStorage.setItem('au:plan-upgraded', '1')
          sessionStorage.setItem('au:plan-upgraded-plan', nextStatus.plan)
          sessionStorage.setItem('au:plan-upgraded-at', String(Date.now()))
          setTimedOut(false)
          setLoading(false)
          return true
        }
      } catch {
        // keep polling
      }

      return false
    }

    void poll().then((done) => {
      if (done || cancelled) return

      const interval = window.setInterval(async () => {
        const elapsed = Date.now() - startedAt
        if (elapsed >= MAX_WAIT_MS) {
          window.clearInterval(interval)
          if (!cancelled) {
            setTimedOut(true)
            setLoading(false)
          }
          return
        }

        const confirmed = await poll()
        if (confirmed) {
          window.clearInterval(interval)
        }
      }, POLL_MS)

      const timeout = window.setTimeout(() => {
        window.clearInterval(interval)
      }, MAX_WAIT_MS + 500)

      return () => {
        window.clearInterval(interval)
        window.clearTimeout(timeout)
      }
    })

    return () => {
      cancelled = true
    }
  }, [requestedPlan, navigate])

  const activePlan: PaidPlan = useMemo(() => {
    if (status?.plan === 'pro' || status?.plan === 'team') return status.plan
    return requestedPlan ?? 'pro'
  }, [status?.plan, requestedPlan])

  if (!requestedPlan) return null

  return (
    <div className="mx-auto w-full max-w-3xl px-4 py-10">
      <Link to="/settings/billing" className="text-sm text-muted-foreground hover:text-foreground">← Back to settings</Link>

      <div className="mt-6 rounded-2xl border border-border bg-card p-6 sm:p-8">
        {loading ? (
          <div className="space-y-4">
            <div className="flex items-center gap-2 text-foreground">
              <Loader2 className="h-5 w-5 animate-spin text-emerald-600" />
              <h1 className="text-xl font-semibold">Confirming your plan…</h1>
            </div>
            <p className="text-sm text-muted-foreground">Your payment was received. Plan activation usually takes a few seconds.</p>
          </div>
        ) : timedOut ? (
          <div className="space-y-4">
            <h1 className="text-xl font-semibold text-foreground">Payment received. Your plan may take a moment to activate.</h1>
            <p className="text-sm text-muted-foreground">Check back in a minute, or refresh your billing status now.</p>
            <div className="flex gap-3">
              <button
                type="button"
                onClick={() => window.location.reload()}
                className="rounded-lg bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600"
              >
                Refresh status
              </button>
              <button
                type="button"
                onClick={() => navigate('/settings/billing')}
                className="rounded-lg border border-border px-4 py-2 text-sm font-medium hover:bg-muted"
              >
                Open billing
              </button>
            </div>
          </div>
        ) : (
          <>
            <div className="flex items-center gap-2 text-emerald-600">
              <CheckCircle2 className="h-6 w-6" />
              <h1 className="text-2xl font-semibold text-foreground">You&apos;re on {activePlan === 'team' ? 'Team' : 'Pro'} 🎉</h1>
            </div>

            <p className="mt-2 text-sm text-muted-foreground">
              {activePlan === 'team'
                ? '50 GB relay bandwidth, 20 concurrent connections, unlimited entities, and a custom relay subdomain.'
                : '50 GB relay bandwidth, 20 concurrent connections, 10 entities. Your workspace is ready.'}
            </p>

            <div className="mt-6 space-y-4 border-t border-border pt-6">
              <h2 className="text-sm font-semibold text-foreground">What&apos;s next:</h2>

              <a href="https://docs.agentunited.ai/docs/agent-connect" target="_blank" rel="noreferrer" className="flex items-start justify-between rounded-lg border border-border p-3 hover:bg-muted/40">
                <div>
                  <p className="text-sm font-medium text-foreground">① Connect your agent</p>
                  <p className="text-xs text-muted-foreground">Link your agent to this workspace</p>
                </div>
                <ArrowRight className="mt-0.5 h-4 w-4 text-muted-foreground" />
              </a>

              {activePlan === 'team' && (
                <button type="button" onClick={() => navigate('/settings/billing')} className="flex w-full items-start justify-between rounded-lg border border-border p-3 text-left hover:bg-muted/40">
                  <div>
                    <p className="text-sm font-medium text-foreground">② Set relay URL</p>
                    <p className="text-xs text-muted-foreground">Claim your custom subdomain</p>
                  </div>
                  <ArrowRight className="mt-0.5 h-4 w-4 text-muted-foreground" />
                </button>
              )}

              <button type="button" onClick={() => navigate('/chat')} className="flex w-full items-start justify-between rounded-lg border border-border p-3 text-left hover:bg-muted/40">
                <div>
                  <p className="text-sm font-medium text-foreground">{activePlan === 'team' ? '③' : '②'} Invite your first human</p>
                  <p className="text-xs text-muted-foreground">Send an invite link to a teammate from your workspace</p>
                </div>
                <ArrowRight className="mt-0.5 h-4 w-4 text-muted-foreground" />
              </button>
            </div>

            <button
              type="button"
              onClick={() => navigate('/chat')}
              className="mt-6 inline-flex w-full items-center justify-center gap-2 rounded-lg bg-emerald-500 px-4 py-2.5 text-sm font-medium text-white hover:bg-emerald-600"
            >
              Go to workspace →
            </button>
          </>
        )}
      </div>
    </div>
  )
}
