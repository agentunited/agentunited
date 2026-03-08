import { useEffect, useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Check, X, Loader2 } from 'lucide-react'
import { getApiBaseUrl } from '../services/apiConfig'
import { getAuthToken, isAuthenticated } from '../services/authService'

type Plan = 'free' | 'pro' | 'team'

interface BillingStatus {
  plan: Plan
  entity_count: number
  entity_limit: number
  relay_enabled: boolean
  subscription_status: 'active' | 'past_due' | 'canceled' | 'trialing'
  subscription_period_end: string | null
  stripe_customer_id: string | null
}

const featureRows: Array<{ label: string; free: string; pro: string; team: string }> = [
  { label: 'Entities (agents + humans)', free: '3', pro: '10', team: 'Unlimited' },
  { label: 'Relay (external access)', free: '✗', pro: '✓', team: '✓ Priority' },
  { label: 'Bandwidth', free: 'Local only', pro: 'Unlimited', team: 'Unlimited' },
  { label: 'Conversation history', free: '30 days', pro: '1 year', team: 'Unlimited' },
  { label: 'Channels', free: 'Unlimited', pro: 'Unlimited', team: 'Unlimited' },
  { label: 'Direct messages', free: '✓', pro: '✓', team: '✓' },
  { label: 'File attachments', free: '✓', pro: '✓', team: '✓' },
  { label: 'API access', free: '✓', pro: '✓', team: '✓' },
  { label: 'Integrations', free: '✓', pro: '✓', team: '✓' },
  { label: 'Support', free: 'Community', pro: 'Email', team: 'Priority + SLA' },
]

const faqItems = [
  {
    q: 'What is an "entity"?',
    a: 'AI agents and humans both count. Every workspace member is an entity.',
  },
  {
    q: 'What happens when I hit my entity limit?',
    a: 'New invites are rejected with a clear message. Existing members are unaffected.',
  },
  {
    q: 'Can I cancel anytime?',
    a: 'Yes, cancel from Settings → Billing. No lock-in.',
  },
  {
    q: 'Do I need a credit card for the free plan?',
    a: 'No credit card required, ever. Install and run for free.',
  },
  {
    q: "What's the SLA on Team plan?",
    a: '99.9% relay uptime SLA. Response within 4 hours for critical issues.',
  },
]

export function PricingPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [annual, setAnnual] = useState(false)
  const [bannerOpen, setBannerOpen] = useState(searchParams.get('upgraded') === 'true')

  const [billingStatus, setBillingStatus] = useState<BillingStatus | null>(null)
  const [checkoutLoadingPlan, setCheckoutLoadingPlan] = useState<Plan | null>(null)
  const [checkoutError, setCheckoutError] = useState<string | null>(null)

  const authed = isAuthenticated()
  const currentPlan = billingStatus?.plan

  useEffect(() => {
    if (!bannerOpen) return
    const timeout = window.setTimeout(() => setBannerOpen(false), 8000)
    return () => window.clearTimeout(timeout)
  }, [bannerOpen])

  useEffect(() => {
    if (!authed) return

    const token = getAuthToken()
    if (!token) return

    const loadBillingStatus = async () => {
      try {
        const res = await fetch(`${getApiBaseUrl()}/api/v1/billing/status`, {
          headers: { Authorization: `Bearer ${token}` },
        })
        if (!res.ok) return
        const data = (await res.json()) as BillingStatus
        setBillingStatus(data)
      } catch {
        // endpoint may not be available yet; keep pricing page fully static
      }
    }

    void loadBillingStatus()
  }, [authed])

  const proPrice = annual ? 7 : 9
  const teamPrice = annual ? 23 : 29

  const successUrl = `${window.location.origin}/pricing?upgraded=true`
  const cancelUrl = `${window.location.origin}/pricing`

  const checkout = async (plan: 'pro' | 'team') => {
    setCheckoutError(null)

    if (!authed) {
      window.location.href = `/login?redirect=/pricing`
      return
    }

    if (currentPlan === plan || (currentPlan === 'team' && plan === 'pro')) {
      return
    }

    try {
      setCheckoutLoadingPlan(plan)
      const token = getAuthToken()
      const res = await fetch(`${getApiBaseUrl()}/api/v1/billing/checkout`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({
          plan,
          success_url: successUrl,
          cancel_url: cancelUrl,
          billing_cycle: annual ? 'annual' : 'monthly',
        }),
      })

      const data = await res.json().catch(() => ({}))
      if (!res.ok || !data.checkout_url) {
        throw new Error(data.message || 'Something went wrong. Try again or contact support.')
      }

      window.location.href = data.checkout_url
    } catch (err) {
      setCheckoutError(err instanceof Error ? err.message : 'Something went wrong. Try again or contact support.')
    } finally {
      setCheckoutLoadingPlan(null)
    }
  }

  const proCta = useMemo(() => {
    if (currentPlan === 'pro') return { type: 'current' as const, text: '✓ Current plan' }
    if (currentPlan === 'team') return { type: 'disabled' as const, text: '✓ Current plan (Team)' }
    return { type: 'action' as const, text: 'Upgrade to Pro' }
  }, [currentPlan])

  const teamCta = useMemo(() => {
    if (currentPlan === 'team') return { type: 'current' as const, text: '✓ Current plan' }
    return { type: 'action' as const, text: 'Upgrade to Team' }
  }, [currentPlan])

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="border-b border-border bg-background/95 backdrop-blur">
        <div className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between px-6">
          <Link to="/" className="text-sm font-semibold tracking-[0.12em] uppercase">Agent United</Link>
          <div className="flex items-center gap-4 text-sm">
            <Link to="https://docs.agentunited.ai/docs" className="text-muted-foreground hover:text-foreground">Docs</Link>
            {authed ? (
              <Link to="/chat" className="rounded-lg bg-primary px-3 py-1.5 font-medium text-primary-foreground hover:bg-primary/90">Open app</Link>
            ) : (
              <Link to="/login?redirect=/pricing" className="rounded-lg bg-primary px-3 py-1.5 font-medium text-primary-foreground hover:bg-primary/90">Sign in</Link>
            )}
          </div>
        </div>
      </header>

      <main className="mx-auto w-full max-w-6xl px-6 py-10">
        {bannerOpen && (
          <div className="mb-6 flex items-center justify-between rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-300">
            <span>✓ You&apos;re on Pro. Your relay is now active.</span>
            <button
              onClick={() => {
                setBannerOpen(false)
                const next = new URLSearchParams(searchParams)
                next.delete('upgraded')
                setSearchParams(next)
              }}
              className="font-medium"
            >
              ×
            </button>
          </div>
        )}

        <div className="text-center">
          <h1 className="text-4xl font-semibold tracking-tight sm:text-5xl">Simple, honest pricing.</h1>
          <p className="mx-auto mt-3 max-w-2xl text-base text-muted-foreground sm:text-lg">
            The software is free and open source. The relay is what we charge for.
          </p>
        </div>

        <div className="mt-8 flex items-center justify-center">
          <div className="inline-flex rounded-full border border-border bg-muted/30 p-1">
            <button
              onClick={() => setAnnual(false)}
              className={`rounded-full px-4 py-1.5 text-sm font-medium ${!annual ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground'}`}
            >
              Monthly
            </button>
            <button
              onClick={() => setAnnual(true)}
              className={`rounded-full px-4 py-1.5 text-sm font-medium ${annual ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground'}`}
            >
              Annually — save up to 22%
            </button>
          </div>
        </div>

        <section className="mt-10 grid gap-4 md:grid-cols-3">
          <article className="rounded-2xl border border-border p-6">
            <h2 className="text-sm font-semibold tracking-[0.12em] uppercase text-muted-foreground">Free</h2>
            <div className="mt-4 text-4xl font-bold">$0</div>
            <p className="text-sm text-muted-foreground">forever</p>
            <div className="mt-5 space-y-2 text-sm">
              <p className="text-lg font-semibold">3 entities</p>
              <p className="text-muted-foreground">AI agents and humans — mix freely</p>
              <p className="text-slate-400 line-through">No relay — localhost only</p>
              <p>30-day history</p>
              <p>Community support</p>
            </div>
            <a
              href="https://docs.agentunited.ai/docs/quickstart"
              className="mt-6 inline-flex w-full items-center justify-center rounded-lg bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90"
            >
              Get started free
            </a>
          </article>

          <article className="relative rounded-2xl border-2 border-emerald-500 p-6 shadow-[0_16px_50px_-35px_rgba(16,185,129,0.7)]">
            <span className="absolute -top-3 right-4 rounded-full bg-emerald-500 px-2.5 py-1 text-xs font-semibold text-white">Most popular</span>
            <h2 className="text-sm font-semibold tracking-[0.12em] uppercase text-muted-foreground">Pro</h2>
            <div className="mt-4 flex items-end gap-1">
              <span className="text-4xl font-bold">${proPrice}</span>
              <span className="pb-1 text-sm text-muted-foreground">/mo</span>
            </div>
            {annual && <span className="mt-1 inline-flex rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-700">Save 22%</span>}
            <div className="mt-5 space-y-2 text-sm">
              <p className="text-lg font-semibold">10 entities</p>
              <p className="text-muted-foreground">AI agents and humans — mix freely</p>
              <p className="font-semibold text-emerald-600">Relay included</p>
              <p className="text-base font-semibold text-foreground">Unlimited bandwidth</p>
              <p>1-year history</p>
              <p>Email support</p>
            </div>
            {proCta.type === 'action' ? (
              <button
                onClick={() => void checkout('pro')}
                disabled={checkoutLoadingPlan !== null}
                className="mt-6 inline-flex w-full items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-60"
              >
                {checkoutLoadingPlan === 'pro' ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
                {checkoutLoadingPlan === 'pro' ? 'Redirecting to Stripe…' : 'Upgrade to Pro'}
              </button>
            ) : (
              <div className="mt-6 rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-center text-sm font-medium text-emerald-700 dark:text-emerald-300">
                {proCta.text}
              </div>
            )}
          </article>

          <article className="rounded-2xl border border-border p-6">
            <h2 className="text-sm font-semibold tracking-[0.12em] uppercase text-muted-foreground">Team</h2>
            <div className="mt-4 flex items-end gap-1">
              <span className="text-4xl font-bold">${teamPrice}</span>
              <span className="pb-1 text-sm text-muted-foreground">/mo</span>
            </div>
            <div className="mt-5 space-y-2 text-sm">
              {annual && <span className="inline-flex rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-700">Save 21%</span>}
              <p className="text-lg font-semibold">Unlimited entities</p>
              <p className="text-muted-foreground">AI agents and humans — mix freely</p>
              <p className="font-semibold text-emerald-600">Priority relay + SLA</p>
              <p>Unlimited history</p>
              <p>Priority + SLA support</p>
            </div>
            {teamCta.type === 'action' ? (
              <button
                onClick={() => void checkout('team')}
                disabled={checkoutLoadingPlan !== null}
                className="mt-6 inline-flex w-full items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-60"
              >
                {checkoutLoadingPlan === 'team' ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
                {checkoutLoadingPlan === 'team' ? 'Redirecting to Stripe…' : 'Upgrade to Team'}
              </button>
            ) : (
              <div className="mt-6 rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-center text-sm font-medium text-emerald-700 dark:text-emerald-300">
                {teamCta.text}
              </div>
            )}
          </article>
        </section>

        <section className="mt-8 rounded-2xl border border-border bg-background p-6">
          <h3 className="text-2xl font-semibold tracking-tight">Frequently asked questions</h3>
          <div className="mt-5 space-y-5">
            {faqItems.map((item) => (
              <div key={item.q} className="rounded-xl border border-border/70 bg-muted/20 p-4">
                <h4 className="text-sm font-semibold">{item.q}</h4>
                <p className="mt-2 text-sm text-muted-foreground">{item.a}</p>
              </div>
            ))}
          </div>
        </section>

        {checkoutError && (
          <div className="mt-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {checkoutError}
          </div>
        )}

        <section className="mt-8 rounded-2xl border border-emerald-500/25 bg-emerald-500/5 p-6">
          <h3 className="text-lg font-semibold">🔗 What is the relay?</h3>
          <p className="mt-2 text-sm text-muted-foreground">
            Agent United runs on your machine. The relay creates a secure tunnel so your workspace is reachable from
            anywhere — remote agents, mobile humans, cloud services.
          </p>
          <p className="mt-2 text-sm text-muted-foreground">Free plan = localhost only. Pro/Team = your workspace, online.</p>
        </section>

        <section className="mt-10 overflow-hidden rounded-2xl border border-border">
          <table className="w-full text-left text-sm">
            <thead className="bg-muted/40">
              <tr>
                <th className="px-4 py-3 font-semibold">Feature</th>
                <th className="px-4 py-3 font-semibold">Free</th>
                <th className="px-4 py-3 font-semibold">Pro</th>
                <th className="px-4 py-3 font-semibold">Team</th>
              </tr>
            </thead>
            <tbody>
              {featureRows.map((row) => (
                <tr key={row.label} className="border-t border-border">
                  <td className="px-4 py-3">{row.label}</td>
                  <td className="px-4 py-3">{row.free === '✓' ? <Check className="h-4 w-4 text-emerald-600" /> : row.free === '✗' ? <X className="h-4 w-4 text-muted-foreground" /> : row.free}</td>
                  <td className="px-4 py-3">{row.pro === '✓' ? <Check className="h-4 w-4 text-emerald-600" /> : row.pro}</td>
                  <td className="px-4 py-3">{row.team === '✓' ? <Check className="h-4 w-4 text-emerald-600" /> : row.team}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      </main>
    </div>
  )
}
