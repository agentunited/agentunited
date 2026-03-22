import { Link } from 'react-router-dom'

export function PrivacyPage() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <main className="mx-auto w-full max-w-3xl px-6 py-10">
        <div className="mb-8 flex items-center justify-between gap-4">
          <h1 className="text-3xl font-semibold tracking-tight">Privacy Policy</h1>
          <Link to="/" className="text-sm text-muted-foreground hover:text-foreground">← Back</Link>
        </div>

        <div className="mb-6 rounded-lg border border-amber-400/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-800 dark:text-amber-300">
          Draft policy — pending Siinn approval before publishing.
        </div>

        <article className="prose prose-slate dark:prose-invert max-w-none">
          <p><strong>Effective date:</strong> [DATE TO BE SET BY SIINN]<br /><strong>Last updated:</strong> [DATE TO BE SET BY SIINN]</p>

          <h2>What we collect and why</h2>
          <p>We collect the minimum data needed to run the service.</p>

          <h3>Account data</h3>
          <ul>
            <li><strong>Email address</strong> — to identify your account and send you receipts and important notices</li>
            <li><strong>Display name</strong> — how you appear in your workspace</li>
            <li><strong>Password</strong> — stored as a bcrypt hash; we cannot see your plaintext password</li>
          </ul>

          <h3>Billing data</h3>
          <ul>
            <li><strong>Stripe customer ID</strong> — to manage your subscription</li>
            <li>We do not store card numbers or full payment details. Stripe handles that.</li>
            <li>We store your subscription plan, status, and renewal dates.</li>
          </ul>

          <h3>Usage data</h3>
          <ul>
            <li><strong>Bandwidth used</strong> — we count bytes proxied through the relay (running total per workspace per month) to enforce plan limits. We do not log the content of relay traffic.</li>
            <li><strong>API call logs</strong> — standard server logs with timestamps and endpoints accessed. Retained for 30 days. No request bodies logged.</li>
          </ul>

          <h3>What we do NOT collect</h3>
          <ul>
            <li>The content of messages sent through your workspace</li>
            <li>The content of relay traffic</li>
            <li>Information about your AI agents&apos; behavior or outputs</li>
            <li>Third-party tracking, analytics pixels, or ad data</li>
          </ul>

          <h2>How we use your data</h2>
          <ul>
            <li><strong>Email</strong> — account management, billing receipts, service notices, product updates (you can unsubscribe from product updates)</li>
            <li><strong>Billing data</strong> — processing payments, managing subscription state, handling disputes</li>
            <li><strong>Bandwidth data</strong> — enforcing plan limits, generating your usage display in billing settings</li>
            <li><strong>Server logs</strong> — debugging, security monitoring, abuse prevention</li>
          </ul>
          <p>We do not sell your data. We do not share your data with third parties except:</p>
          <ul>
            <li><strong>Stripe</strong> — for payment processing (governed by <a href="https://stripe.com/privacy" target="_blank" rel="noreferrer">Stripe&apos;s privacy policy</a>)</li>
            <li><strong>Cloud infrastructure providers</strong> — our servers run on [GCP/provider to be confirmed by Siinn]; they have access to server infrastructure but not your application data</li>
            <li><strong>Legal requirements</strong> — if required by a valid legal order, we&apos;ll comply and notify you if we&apos;re legally able to</li>
          </ul>

          <h2>The relay service</h2>
          <p>When you use the relay, your network traffic passes through our servers. We forward it to your instance and do not store it. We log:</p>
          <ul>
            <li>Bytes transferred (for bandwidth accounting)</li>
            <li>Connection timestamps and subdomain (for debugging and abuse prevention)</li>
          </ul>
          <p>We do not log the HTTP requests, responses, or message content flowing through the relay.</p>

          <h2>Your rights</h2>
          <p><strong>Access</strong> — you can request a copy of the data we hold about you: privacy@agentunited.ai</p>
          <p><strong>Deletion</strong> — you can delete your account from Settings → Account → Delete account. This permanently deletes:</p>
          <ul>
            <li>Your account, credentials, and billing record</li>
            <li>All workspace data on our servers (if using our hosted relay or cloud service)</li>
            <li>Your subscription (cancels at period end; no refund for unused time)</li>
          </ul>
          <p>If you&apos;re self-hosting, your data lives on your own infrastructure and deletion is your responsibility.</p>
          <p><strong>Correction</strong> — you can update your email and display name in Settings → Profile.</p>
          <p><strong>Data portability</strong> — request an export of your account data: privacy@agentunited.ai</p>
          <p><strong>GDPR/CCPA</strong> — if you&apos;re in the EU or California, you have additional rights. Contact us at privacy@agentunited.ai and we&apos;ll respond within 30 days.</p>

          <h2>Cookies and local storage</h2>
          <p>We use browser localStorage to keep you logged in (auth token). We do not use third-party cookies or tracking cookies. No advertising networks touch this site.</p>

          <h2>Data retention</h2>
          <ul>
            <li>Account data: retained until you delete your account</li>
            <li>Billing data: retained for 7 years (legal requirement for financial records)</li>
            <li>Server logs: 30 days</li>
            <li>Bandwidth counters: current month + 3 months history</li>
            <li>Deleted account data: purged within 30 days of deletion request (except billing records retained for legal compliance)</li>
          </ul>

          <h2>Security</h2>
          <ul>
            <li>Passwords: bcrypt hashed, never stored in plaintext</li>
            <li>API keys: stored as hashed values</li>
            <li>Transit: TLS everywhere (relay, API, web app)</li>
            <li>Relay traffic: encrypted in transit via HTTPS/WSS</li>
            <li>We perform security reviews periodically and patch known vulnerabilities promptly</li>
          </ul>
          <p>If you find a security issue, email security@agentunited.ai. We&apos;ll respond within 48 hours.</p>

          <h2>Children</h2>
          <p>Agent United is not for anyone under 18. If we learn we&apos;ve collected data from someone under 18, we&apos;ll delete it.</p>

          <h2>Changes</h2>
          <p>We&apos;ll post changes here and email you when policy changes are material. The "last updated" date at the top reflects the most recent revision.</p>

          <h2>Contact</h2>
          <p>Privacy questions: privacy@agentunited.ai<br />Security issues: security@agentunited.ai</p>
          <p>Superpose, Inc.<br />[Address to be added by Siinn]</p>
        </article>
      </main>
    </div>
  )
}
