import { useEffect, useMemo, useRef, useState } from 'react'
import { Check, Copy, Loader2, Lock, X } from 'lucide-react'
import { Button } from '../ui/Button'
import { billingApi, type Plan } from '../../services/billingApi'

const RELAY_DOMAIN = 'tunnel.agentunited.ai'
const RESERVED = ['www', 'api', 'app', 'mail', 'relay', 'admin', 'status', 'support', 'docs', 'blog']

type CheckState = 'idle' | 'checking' | 'available' | 'taken' | 'reserved' | 'invalid'

interface SubdomainClaimSectionProps {
  plan: Plan
  currentSubdomain: string | null
  onClaimed: (subdomain: string) => void
  onUpgradeToTeam?: () => void
}

function validateSubdomain(value: string): string | null {
  const normalized = value.toLowerCase().trim()

  if (normalized.length < 3) return 'Too short. Minimum 3 characters.'
  if (normalized.length > 30) return 'Too long. Maximum 30 characters.'
  if (!/^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/.test(normalized)) {
    return 'Only letters, numbers, and hyphens. No spaces. (3–30 characters)'
  }
  if (RESERVED.includes(normalized)) {
    return `"${normalized}" is reserved. Try a different name.`
  }

  return null
}

export function SubdomainClaimSection({
  plan,
  currentSubdomain,
  onClaimed,
  onUpgradeToTeam,
}: SubdomainClaimSectionProps) {
  const [input, setInput] = useState('')
  const [checkState, setCheckState] = useState<CheckState>('idle')
  const [validationError, setValidationError] = useState<string | null>(null)
  const [claiming, setClaiming] = useState(false)
  const [claimError, setClaimError] = useState<string | null>(null)
  const [showChange, setShowChange] = useState(false)
  const [copied, setCopied] = useState(false)
  const requestIdRef = useRef(0)

  const isTeam = plan === 'team'
  const isClaimed = isTeam && !!currentSubdomain && !showChange
  const normalizedInput = useMemo(() => input.toLowerCase().trim(), [input])
  const fullClaimUrl = currentSubdomain ? `https://${currentSubdomain}.${RELAY_DOMAIN}` : ''

  useEffect(() => {
    if (!isTeam || isClaimed) {
      setCheckState('idle')
      setValidationError(null)
      return
    }

    if (!normalizedInput) {
      setCheckState('idle')
      setValidationError(null)
      return
    }

    const validation = validateSubdomain(normalizedInput)
    if (validation) {
      setValidationError(validation)
      if (validation.includes('reserved')) {
        setCheckState('reserved')
      } else {
        setCheckState('invalid')
      }
      return
    }

    setValidationError(null)
    setCheckState('checking')

    const currentRequestId = ++requestIdRef.current
    const timeout = setTimeout(async () => {
      try {
        const result = await billingApi.checkSubdomain(normalizedInput)
        if (currentRequestId !== requestIdRef.current) return
        setCheckState(result.available ? 'available' : 'taken')
      } catch {
        if (currentRequestId !== requestIdRef.current) return
        setCheckState('idle')
      }
    }, 600)

    return () => {
      clearTimeout(timeout)
    }
  }, [isTeam, isClaimed, normalizedInput])

  const handleCopy = async () => {
    if (!fullClaimUrl) return
    await navigator.clipboard.writeText(fullClaimUrl)
    setCopied(true)
    window.setTimeout(() => setCopied(false), 2000)
  }

  const handleClaim = async () => {
    if (!isTeam || checkState !== 'available' || claiming) return

    try {
      setClaiming(true)
      setClaimError(null)
      const result = await billingApi.claimSubdomain(normalizedInput)
      setInput('')
      setShowChange(false)
      setCheckState('idle')
      onClaimed(result.subdomain)
    } catch (error) {
      const message = error instanceof Error ? error.message.toLowerCase() : ''
      if (message.includes('taken') || message.includes('already') || message.includes('409')) {
        setClaimError('That subdomain is already taken.')
      } else {
        setClaimError('Failed to claim subdomain. Try again.')
      }
    } finally {
      setClaiming(false)
    }
  }

  if (!isTeam) {
    return (
      <div className="border-t border-border pt-4">
        <p className="text-sm font-medium text-foreground">Custom relay subdomain</p>
        <div className="mt-3 space-y-3">
          <div className="flex items-center gap-2 rounded-lg border border-border bg-muted/40 px-3 py-2 opacity-60">
            <input
              disabled
              value=""
              placeholder="yourname"
              className="flex-1 cursor-not-allowed bg-transparent text-sm outline-none"
            />
            <span className="text-sm text-muted-foreground">.{RELAY_DOMAIN}</span>
          </div>
          <p className="flex items-center gap-1.5 text-sm text-muted-foreground">
            <Lock className="h-3.5 w-3.5" />
            Custom subdomains are available on the Team plan.
          </p>
          {onUpgradeToTeam && (
            <Button type="button" onClick={onUpgradeToTeam} className="w-full sm:w-auto">
              Upgrade to Team → $29/mo
            </Button>
          )}
        </div>
      </div>
    )
  }

  return (
    <div className="border-t border-border pt-4">
      <p className="text-sm font-medium text-foreground">Custom relay subdomain</p>

      {isClaimed ? (
        <div className="mt-3 space-y-2">
          <p className="flex items-center gap-1.5 text-sm text-emerald-600 dark:text-emerald-400">
            <Check className="h-3.5 w-3.5" />
            Your relay URL
          </p>
          <div className="flex items-center justify-between gap-2 rounded-lg border border-border bg-muted/40 px-3 py-2">
            <code className="flex-1 truncate font-mono text-sm">{fullClaimUrl}</code>
            <div className="flex shrink-0 items-center gap-2">
              <button type="button" onClick={() => void handleCopy()} className="text-xs text-muted-foreground hover:text-foreground">
                {copied ? 'Copied ✓' : <><Copy className="mr-1 inline h-3.5 w-3.5" />Copy</>}
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowChange(true)
                  setClaimError(null)
                }}
                className="text-xs text-muted-foreground underline hover:text-foreground"
              >
                Change
              </button>
            </div>
          </div>
          <p className="text-xs text-muted-foreground">Your relay client will use this URL on the next reconnect.</p>
        </div>
      ) : (
        <div className="mt-3 space-y-3">
          <p className="text-sm text-muted-foreground">Your workspace is reachable at a custom URL on the Team plan.</p>

          <div className="flex items-center gap-0">
            <div className="relative flex-1">
              <input
                value={input}
                onChange={(e) => {
                  setInput(e.target.value.toLowerCase())
                  setClaimError(null)
                }}
                disabled={claiming}
                placeholder="yourname"
                className="w-full rounded-l-lg border border-r-0 border-border bg-background px-3 py-2 text-sm outline-none focus:border-ring disabled:opacity-50"
              />
              {checkState === 'checking' && <Loader2 className="absolute right-2 top-1/2 h-3.5 w-3.5 -translate-y-1/2 animate-spin text-muted-foreground" />}
              {checkState === 'available' && <Check className="absolute right-2 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-emerald-600" />}
              {(checkState === 'taken' || checkState === 'invalid' || checkState === 'reserved') && (
                <X className="absolute right-2 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-destructive" />
              )}
            </div>
            <span className="rounded-r-lg border border-border bg-muted/60 px-3 py-2 text-sm text-muted-foreground">.{RELAY_DOMAIN}</span>
          </div>

          {(checkState === 'invalid' || checkState === 'reserved') && validationError ? (
            <p className="text-xs text-destructive">{validationError}</p>
          ) : null}
          {checkState === 'available' ? (
            <p className="text-xs text-emerald-600">{normalizedInput}.{RELAY_DOMAIN} is available</p>
          ) : null}
          {checkState === 'taken' ? (
            <p className="text-xs text-destructive">{normalizedInput}.{RELAY_DOMAIN} is already taken.</p>
          ) : null}
          {checkState === 'idle' && !input ? (
            <p className="text-xs text-muted-foreground">3–30 characters, letters, numbers, hyphens only</p>
          ) : null}

          {claimError ? <p className="text-xs text-destructive">{claimError}</p> : null}

          <Button
            type="button"
            onClick={() => void handleClaim()}
            isLoading={claiming}
            disabled={checkState !== 'available' || claiming}
            className="w-full"
          >
            {claiming ? 'Claiming…' : `Claim ${normalizedInput || 'subdomain'}.${RELAY_DOMAIN} →`}
          </Button>
        </div>
      )}
    </div>
  )
}
