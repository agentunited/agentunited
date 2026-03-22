import { useEffect, useMemo, useRef, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { Check, ChevronLeft, Circle, Loader2, Pencil } from 'lucide-react'
import { Card } from '../components/ui/Card'
import { Input } from '../components/ui/Input'
import { Button } from '../components/ui/Button'
import { userApi, type MeProfile } from '../services/userApi'
import { billingApi, type BillingStatus, type Plan } from '../services/billingApi'
import { UpgradePrompt } from '../components/billing/UpgradePrompt'

interface UserSettingsPageProps {
  initialTab?: 'profile' | 'password' | 'billing'
}

type SettingsTab = 'profile' | 'password' | 'billing'

const STRIPE_PRICE_IDS = {
  pro: 'price_1T8leTJep8wNHLzTS1wP6gOF',
  team: 'price_1T8leoJep8wNHLzTjB8gVZSj',
} as const

function initialsFor(email?: string, displayName?: string) {
  const source = (displayName || email || 'U').trim()
  return source
    .split(' ')
    .map((x) => x[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
}

function getPasswordStrength(password: string): { label: 'Weak' | 'Good' | 'Strong'; width: string; color: string } {
  if (!password || password.length < 12) {
    return { label: 'Weak', width: '33%', color: 'var(--color-destructive)' }
  }

  const hasNumber = /\d/.test(password)
  const hasSymbol = /[^A-Za-z0-9]/.test(password)

  if (hasNumber && hasSymbol) {
    return { label: 'Strong', width: '100%', color: 'oklch(0.696 0.17 162)' }
  }

  return { label: 'Good', width: '66%', color: 'oklch(0.769 0.137 72)' }
}

function formatRenewalDate(raw?: string | null) {
  if (!raw) return null
  const d = new Date(raw)
  if (Number.isNaN(d.getTime())) return null
  return d.toLocaleDateString([], { month: 'long', day: 'numeric', year: 'numeric' })
}

function planBadgeClass(plan: Plan) {
  if (plan === 'pro') return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300'
  if (plan === 'team') return 'bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300'
  return 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300'
}

function usageBarClass(ratio: number) {
  if (ratio >= 1) return 'bg-red-500'
  if (ratio >= 0.8) return 'bg-amber-400'
  return 'bg-emerald-500'
}

export function UserSettingsPage({ initialTab = 'profile' }: UserSettingsPageProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [activeTab, setActiveTab] = useState<SettingsTab>(initialTab)

  const [profile, setProfile] = useState<MeProfile | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [displayName, setDisplayName] = useState('')
  const [avatarUrl, setAvatarUrl] = useState('')
  const [savingProfile, setSavingProfile] = useState(false)
  const [profileError, setProfileError] = useState<string | null>(null)
  const [profileSaved, setProfileSaved] = useState(false)

  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [savingPassword, setSavingPassword] = useState(false)
  const [passwordError, setPasswordError] = useState<string | null>(null)
  const [passwordSaved, setPasswordSaved] = useState(false)

  const [billingStatus, setBillingStatus] = useState<BillingStatus | null>(null)
  const [billingLoading, setBillingLoading] = useState(true)
  const [billingError, setBillingError] = useState<string | null>(null)
  const [checkoutLoadingPlan, setCheckoutLoadingPlan] = useState<'pro' | 'team' | null>(null)
  const [portalLoading, setPortalLoading] = useState(false)
  const [billingActionError, setBillingActionError] = useState<string | null>(null)
  const [billingRedirectStatus, setBillingRedirectStatus] = useState<'upgraded' | 'canceled' | null>(() => {
    const params = new URLSearchParams(window.location.search)
    if (params.get('upgraded') === 'true') return 'upgraded'
    if (params.get('canceled') === 'true') return 'canceled'
    return null
  })

  useEffect(() => {
    setActiveTab(initialTab)
  }, [initialTab])

  useEffect(() => {
    async function loadMe() {
      try {
        setLoading(true)
        setLoadError(null)
        const me = await userApi.getMe()
        setProfile(me)
        setDisplayName(me.display_name || '')
        setAvatarUrl(me.avatar_url || '')
      } catch (err) {
        setLoadError(err instanceof Error ? err.message : 'Could not load profile. Try again.')
      } finally {
        setLoading(false)
      }
    }

    void loadMe()
  }, [])

  const loadBilling = async () => {
    try {
      setBillingLoading(true)
      setBillingError(null)
      const status = await billingApi.getStatus()
      setBillingStatus(status)
    } catch {
      setBillingError('Unable to load billing info. Try refreshing.')
    } finally {
      setBillingLoading(false)
    }
  }

  const clearBillingRedirectParams = () => {
    const params = new URLSearchParams(window.location.search)
    params.delete('upgraded')
    params.delete('canceled')
    const next = params.toString()
    const newUrl = `${window.location.pathname}${next ? `?${next}` : ''}`
    window.history.replaceState({}, '', newUrl)
  }

  useEffect(() => {
    if (activeTab !== 'billing' || billingStatus) return
    void loadBilling()
  }, [activeTab, billingStatus])

  useEffect(() => {
    if (!profileSaved) return
    const t = window.setTimeout(() => setProfileSaved(false), 3000)
    return () => window.clearTimeout(t)
  }, [profileSaved])

  useEffect(() => {
    if (!passwordSaved) return
    const t = window.setTimeout(() => setPasswordSaved(false), 3000)
    return () => window.clearTimeout(t)
  }, [passwordSaved])

  useEffect(() => {
    if (!billingRedirectStatus) return
    const t = window.setTimeout(() => {
      setBillingRedirectStatus(null)
      clearBillingRedirectParams()
    }, 8000)
    return () => window.clearTimeout(t)
  }, [billingRedirectStatus])

  const joinedDate = useMemo(() => {
    if (!profile?.created_at) return '—'
    return new Date(profile.created_at).toLocaleDateString([], {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
  }, [profile?.created_at])

  const avatarLabel = initialsFor(profile?.email, displayName)
  const profileDirty =
    displayName.trim() !== (profile?.display_name || '') ||
    avatarUrl.trim() !== (profile?.avatar_url || '')

  const passwordStrength = getPasswordStrength(newPassword)

  const handleAvatarFile = async (file?: File) => {
    if (!file) return

    const maxSize = 5 * 1024 * 1024
    if (file.size > maxSize) {
      setProfileError('Avatar must be 5MB or smaller')
      return
    }

    if (!['image/jpeg', 'image/png', 'image/webp'].includes(file.type)) {
      setProfileError('Avatar must be JPG, PNG, or WebP')
      return
    }

    const reader = new FileReader()
    reader.onload = () => {
      if (typeof reader.result === 'string') {
        setAvatarUrl(reader.result)
        setProfileError(null)
      }
    }
    reader.readAsDataURL(file)
  }

  const saveProfile = async (e: FormEvent) => {
    e.preventDefault()
    try {
      setSavingProfile(true)
      setProfileError(null)
      const updated = await userApi.updateMe({
        display_name: displayName.trim() || undefined,
        avatar_url: avatarUrl.trim() || undefined,
      })
      setProfile(updated)
      setProfileSaved(true)
    } catch (err) {
      setProfileError(err instanceof Error ? err.message : 'Failed to save profile')
    } finally {
      setSavingProfile(false)
    }
  }

  const savePassword = async (e: FormEvent) => {
    e.preventDefault()

    if (newPassword.length < 12) {
      setPasswordError('Password must be at least 12 characters')
      return
    }

    if (newPassword !== confirmPassword) {
      setPasswordError('Passwords do not match')
      return
    }

    try {
      setSavingPassword(true)
      setPasswordError(null)
      await userApi.updatePassword({
        current_password: currentPassword,
        new_password: newPassword,
      })
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
      setPasswordSaved(true)
    } catch (err) {
      setPasswordError(err instanceof Error ? err.message : 'Failed to update password')
    } finally {
      setSavingPassword(false)
    }
  }

  const checkout = async (plan: 'pro' | 'team') => {
    try {
      setBillingActionError(null)
      setCheckoutLoadingPlan(plan)
      const success_url = `${window.location.origin}/welcome?plan=${plan}`
      const cancel_url = `${window.location.origin}/settings/billing?canceled=true`
      const { checkout_url } = await billingApi.createCheckoutSession({
        plan,
        price_id: STRIPE_PRICE_IDS[plan],
        success_url,
        cancel_url,
      })
      window.location.href = checkout_url
    } catch (error) {
      setBillingActionError(error instanceof Error ? error.message : 'Something went wrong. Try again.')
    } finally {
      setCheckoutLoadingPlan(null)
    }
  }

  const openPortal = async () => {
    try {
      setBillingActionError(null)
      setPortalLoading(true)
      const { portal_url } = await billingApi.createPortalSession()
      if (!portal_url) throw new Error('Billing portal unavailable. Try again.')
      window.location.href = portal_url
    } catch (error) {
      setBillingActionError(error instanceof Error ? error.message : 'Failed to open billing portal.')
    } finally {
      setPortalLoading(false)
    }
  }

  if (loading) {
    return <div className="flex min-h-screen items-center justify-center text-sm text-muted-foreground">Loading settings…</div>
  }

  if (loadError) {
    return (
      <div className="flex min-h-screen items-center justify-center px-4">
        <Card className="w-full max-w-md">
          <h2 className="text-lg font-semibold text-foreground">Could not load profile</h2>
          <p className="mt-1 text-sm text-muted-foreground">Try again.</p>
          <Button className="mt-4" onClick={() => window.location.reload()}>
            Retry
          </Button>
        </Card>
      </div>
    )
  }

  const ratio = billingStatus ? Math.min(1, billingStatus.entity_count / Math.max(1, billingStatus.entity_limit)) : 0
  const renewal = formatRenewalDate(billingStatus?.subscription_period_end)
  const showUpgradePrompt =
    billingStatus &&
    ((billingStatus.plan === 'free' && billingStatus.entity_count >= billingStatus.entity_limit) ||
      (billingStatus.plan === 'pro' && billingStatus.entity_count >= billingStatus.entity_limit))

  return (
    <div className="mx-auto w-full max-w-2xl px-4 py-8">
      <Link
        to="/chat"
        className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ChevronLeft className="h-4 w-4" />
        Back to chat
      </Link>

      <h1 className="mb-4 text-2xl font-semibold text-foreground">Settings</h1>

      <div className="mb-4 flex flex-wrap gap-2 border-b border-border pb-3">
        {([
          ['profile', 'Profile'],
          ['password', 'Password'],
          ['billing', 'Billing'],
        ] as Array<[SettingsTab, string]>).map(([key, label]) => (
          <button
            key={key}
            type="button"
            onClick={() => setActiveTab(key)}
            className={`rounded-lg px-3 py-1.5 text-sm font-medium transition ${
              activeTab === key ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {activeTab === 'profile' ? (
        <Card className="p-6">
          <h2 className="mb-1 text-base font-semibold text-foreground">Profile</h2>
          <p className="mb-6 text-sm text-muted-foreground">Update your display details.</p>
          <hr className="mb-6 border-border" />

          <form className="space-y-5" onSubmit={saveProfile}>
            <div>
              <p className="mb-2 text-sm font-medium text-foreground">Avatar</p>
              <div
                className="group relative h-16 w-16 cursor-pointer"
                onClick={() => fileInputRef.current?.click()}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault()
                    fileInputRef.current?.click()
                  }
                }}
                role="button"
                tabIndex={0}
                aria-label="Change avatar"
              >
                {avatarUrl ? (
                  <img src={avatarUrl} alt="Avatar" className="h-16 w-16 rounded-full object-cover" />
                ) : (
                  <div className="flex h-16 w-16 items-center justify-center rounded-full bg-emerald-500/15 text-lg font-semibold text-emerald-700 dark:text-emerald-300">
                    {avatarLabel}
                  </div>
                )}

                <div className="absolute inset-0 flex items-center justify-center rounded-full bg-black/40 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100">
                  <Pencil className="h-5 w-5 text-white" />
                </div>
              </div>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/jpeg,image/png,image/webp"
                className="sr-only"
                onChange={(e) => void handleAvatarFile(e.target.files?.[0])}
              />
              <p className="mt-2 text-xs text-muted-foreground">Click to change · JPG, PNG, WebP · Max 5MB</p>
            </div>

            <Input
              label="Display name"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="How your name appears in conversations"
            />

            {profileError && <p className="text-sm text-destructive">{profileError}</p>}

            <div className="flex items-center justify-end gap-3">
              {profileSaved && (
                <span className="flex items-center gap-1 text-sm text-emerald-600">
                  <Check className="h-3.5 w-3.5" />
                  Saved
                </span>
              )}
              <Button type="submit" isLoading={savingProfile} disabled={!profileDirty || savingProfile}>
                {savingProfile ? 'Saving…' : 'Save profile'}
              </Button>
            </div>
          </form>
        </Card>
      ) : null}

      {activeTab === 'password' ? (
        <Card className="p-6">
          <h2 className="mb-1 text-base font-semibold text-foreground">Password</h2>
          <p className="mb-6 text-sm text-muted-foreground">Change your password securely.</p>
          <hr className="mb-6 border-border" />

          <form className="space-y-4" onSubmit={savePassword}>
            <Input
              label="Current password"
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              required
            />
            <div>
              <Input
                label="New password"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                required
              />
              <div className="mt-2 h-1 w-full rounded-full bg-border">
                <div
                  className="h-1 rounded-full transition-all"
                  style={{ width: passwordStrength.width, backgroundColor: passwordStrength.color }}
                />
              </div>
              <p className="mt-1 text-xs text-muted-foreground">{passwordStrength.label}</p>
            </div>
            <Input
              label="Confirm password"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              error={confirmPassword && confirmPassword !== newPassword ? 'Passwords do not match' : undefined}
              required
            />

            {passwordError && <p className="text-sm text-destructive">{passwordError}</p>}

            <div className="flex items-center justify-end gap-3">
              {passwordSaved && (
                <span className="flex items-center gap-1 text-sm text-emerald-600">
                  <Check className="h-3.5 w-3.5" />
                  Password updated
                </span>
              )}
              <Button type="submit" isLoading={savingPassword}>
                {savingPassword ? 'Updating…' : 'Update password'}
              </Button>
            </div>
          </form>
        </Card>
      ) : null}

      {activeTab === 'billing' ? (
        <Card className="space-y-4 p-6">
          <div>
            <h2 className="text-base font-semibold text-foreground">Current Plan</h2>
            <p className="text-sm text-muted-foreground">Live plan, usage, and relay status.</p>
          </div>

          {billingRedirectStatus === 'upgraded' ? (
            <div className="rounded-md border border-emerald-500/30 bg-emerald-500/10 p-3 text-sm text-emerald-700 dark:text-emerald-300">
              ✓ Billing update successful. Your new plan is now active.
            </div>
          ) : null}

          {billingRedirectStatus === 'canceled' ? (
            <div className="rounded-md border border-amber-500/30 bg-amber-500/10 p-3 text-sm text-amber-700 dark:text-amber-300">
              Checkout canceled. No changes were made.
            </div>
          ) : null}

          {billingLoading ? (
            <div className="space-y-3">
              <div className="h-5 w-24 animate-pulse rounded bg-muted" />
              <div className="h-2 w-full animate-pulse rounded bg-muted" />
              <div className="h-4 w-40 animate-pulse rounded bg-muted" />
            </div>
          ) : billingError ? (
            <div className="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">
              <p>{billingError}</p>
              <button className="mt-2 underline" onClick={() => void loadBilling()}>
                Retry
              </button>
            </div>
          ) : billingStatus ? (
            <>
              <div className={`inline-flex rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-[0.08em] ${planBadgeClass(billingStatus.plan)}`}>
                {billingStatus.plan}
              </div>

              {billingStatus.subscription_status === 'past_due' ? (
                <div className="rounded-md border border-amber-300 bg-amber-50 p-3 text-sm text-amber-800 dark:border-amber-700 dark:bg-amber-950/30 dark:text-amber-300">
                  ⚠ Payment failed. Update your payment method to avoid service interruption.
                </div>
              ) : null}

              <div>
                <p className="text-sm font-medium text-foreground">Entities used</p>
                <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-muted">
                  <div className={`h-full ${usageBarClass(ratio)}`} style={{ width: `${ratio * 100}%` }} />
                </div>
                <p className="mt-2 text-sm text-foreground">
                  {billingStatus.entity_count} of {billingStatus.entity_limit} entities used
                </p>
                <p className="text-xs text-muted-foreground">AI agents and humans count together.</p>
              </div>

              <div>
                <p className="text-sm font-medium text-foreground">Relay status</p>
                <p
                  className={`mt-1 inline-flex items-center gap-2 text-sm ${
                    billingStatus.relay_enabled ? 'text-emerald-600' : 'text-slate-500'
                  }`}
                >
                  <Circle className="h-3 w-3 fill-current" />
                  {billingStatus.subscription_status === 'past_due'
                    ? 'Service may be interrupted'
                    : billingStatus.relay_enabled
                      ? billingStatus.relay_hostname || 'Active'
                      : 'Not active (localhost only)'}
                </p>
                {renewal ? (
                  <p className={`mt-1 text-sm ${billingStatus.subscription_status === 'past_due' ? 'text-amber-600' : 'text-muted-foreground'}`}>
                    {billingStatus.subscription_status === 'past_due' ? 'Payment due' : 'Renewal'}: {renewal}
                  </p>
                ) : null}
              </div>

              {billingActionError ? <p className="text-sm text-destructive">{billingActionError}</p> : null}

              {billingStatus.plan === 'free' ? (
                <button
                  type="button"
                  onClick={() => void checkout('pro')}
                  disabled={checkoutLoadingPlan !== null}
                  className="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-emerald-500 px-4 py-2.5 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-60"
                >
                  {checkoutLoadingPlan === 'pro' ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
                  {checkoutLoadingPlan === 'pro' ? 'Redirecting to Stripe…' : 'Upgrade to Pro — $9/mo →'}
                </button>
              ) : null}

              {billingStatus.plan !== 'free' ? (
                <button
                  type="button"
                  onClick={() => void openPortal()}
                  disabled={portalLoading}
                  className="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-border bg-background px-4 py-2.5 text-sm font-medium text-foreground hover:bg-muted disabled:opacity-60"
                >
                  {portalLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
                  {portalLoading ? 'Opening billing portal…' : 'Manage billing →'}
                </button>
              ) : null}

              {billingStatus.plan === 'pro' ? (
                <button
                  type="button"
                  onClick={() => void checkout('team')}
                  disabled={checkoutLoadingPlan !== null}
                  className="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-emerald-500 px-4 py-2.5 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-60"
                >
                  {checkoutLoadingPlan === 'team' ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
                  {checkoutLoadingPlan === 'team' ? 'Redirecting to Stripe…' : 'Upgrade to Team — $29/mo →'}
                </button>
              ) : null}

              {showUpgradePrompt ? (
                <UpgradePrompt
                  plan={billingStatus.plan}
                  onUpgrade={() => checkout(billingStatus.plan === 'pro' ? 'team' : 'pro')}
                  loading={checkoutLoadingPlan !== null}
                  error={billingActionError}
                />
              ) : null}
            </>
          ) : null}
        </Card>
      ) : null}

      {activeTab !== 'billing' ? (
        <Card className="mt-4 p-6">
          <h2 className="mb-1 text-base font-semibold text-foreground">Account info</h2>
          <p className="mb-6 text-sm text-muted-foreground">Read-only account details.</p>
          <hr className="mb-6 border-border" />

          <dl className="space-y-2 text-sm">
            <div className="flex gap-2">
              <dt className="w-28 shrink-0 text-muted-foreground">Email</dt>
              <dd className="text-foreground">{profile?.email || '—'}</dd>
            </div>
            <div className="flex gap-2">
              <dt className="w-28 shrink-0 text-muted-foreground">Member since</dt>
              <dd className="text-foreground">{joinedDate}</dd>
            </div>
            <div className="flex gap-2">
              <dt className="w-28 shrink-0 text-muted-foreground">Account type</dt>
              <dd className="text-foreground">Human</dd>
            </div>
          </dl>
        </Card>
      ) : null}
    </div>
  )
}
