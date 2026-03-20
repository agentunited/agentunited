import { getApiBaseUrl } from './apiConfig'
import { getAuthToken } from './authService'

export type Plan = 'free' | 'pro' | 'team'

export interface BillingStatus {
  plan: Plan
  entity_count: number
  entity_limit: number
  relay_enabled: boolean
  subscription_status: 'active' | 'past_due' | 'canceled' | 'trialing' | 'free'
  subscription_period_end: string | null
  stripe_customer_id: string | null
  relay_hostname?: string | null
  relay_subdomain?: string | null
}

interface CheckoutBody {
  plan?: 'pro' | 'team'
  price_id?: string
  success_url: string
  cancel_url: string
  billing_cycle?: 'monthly' | 'annual'
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getAuthToken()
  const response = await fetch(`${getApiBaseUrl()}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options.headers ?? {}),
    },
  })

  if (!response.ok) {
    const errorBody = await response.json().catch(() => null)
    const message = errorBody?.message || `Request failed (${response.status})`
    throw new Error(message)
  }

  return response.json()
}

export const billingApi = {
  getStatus() {
    return request<BillingStatus>('/api/v1/billing/status')
  },
  async createCheckoutSession(body: CheckoutBody) {
    const checkout = await request<{ checkout_url?: string; session_url?: string }>('/api/v1/billing/checkout', {
      method: 'POST',
      body: JSON.stringify(body),
    })

    return { checkout_url: checkout.session_url ?? checkout.checkout_url ?? '' }
  },
  async createPortalSession() {
    try {
      const portal = await request<{ portal_url?: string; url?: string }>('/api/v1/billing/portal', {
        method: 'GET',
      })
      return { portal_url: portal.portal_url ?? portal.url ?? '' }
    } catch {
      const portal = await request<{ portal_url?: string; url?: string }>('/api/v1/billing/portal', {
        method: 'POST',
        body: JSON.stringify({}),
      })
      return { portal_url: portal.portal_url ?? portal.url ?? '' }
    }
  },
}
