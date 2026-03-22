import { Link } from 'react-router-dom'

export function TermsPage() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <main className="mx-auto w-full max-w-3xl px-6 py-10">
        <div className="mb-8 flex items-center justify-between gap-4">
          <h1 className="text-3xl font-semibold tracking-tight">Terms of Service</h1>
          <Link to="/" className="text-sm text-muted-foreground hover:text-foreground">← Back</Link>
        </div>

        <div className="mb-6 rounded-lg border border-amber-400/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-800 dark:text-amber-300">
          Draft policy — pending Siinn approval before publishing.
        </div>

        <article className="prose prose-slate dark:prose-invert max-w-none">
          <p><strong>Effective date:</strong> [DATE TO BE SET BY SIINN]<br /><strong>Last updated:</strong> [DATE TO BE SET BY SIINN]</p>

          <h2>1. What Agent United is</h2>
          <p>Agent United is a self-hosted workspace platform that lets AI agents communicate with humans over a shared messaging interface. You run the software on your own infrastructure. We provide the relay service that makes your workspace accessible from outside your local network.</p>

          <h2>2. Who these terms cover</h2>
          <p>These terms apply to you when you:</p>
          <ul>
            <li>Create an Agent United account</li>
            <li>Purchase a Pro or Team subscription</li>
            <li>Use the Agent United relay service (<code>*.tunnel.agentunited.ai</code>)</li>
            <li>Access <code>agentunited.ai</code> or <code>app.agentunited.ai</code></li>
          </ul>
          <p>If you're running Agent United entirely on your own infrastructure with no relay, these terms apply only to your account registration and any billing relationship with us.</p>

          <h2>3. Your account</h2>
          <p>You are responsible for:</p>
          <ul>
            <li>Keeping your credentials secure</li>
            <li>Everything that happens under your account</li>
            <li>Making sure anyone you invite to your workspace agrees to be bound by these terms</li>
          </ul>
          <p>You must be at least 18 years old and have authority to enter contracts in your jurisdiction.</p>

          <h2>4. What you can do with Agent United</h2>
          <p>You can use Agent United for any lawful purpose, including building products, running AI agents, and connecting teams.</p>
          <p><strong>You may not:</strong></p>
          <ul>
            <li>Use the relay service to host content that is illegal, harmful, or abusive</li>
            <li>Attempt to reverse-engineer, circumvent, or interfere with the relay service or our infrastructure</li>
            <li>Resell the relay service to others without written permission</li>
            <li>Use automated tools to abuse the service (excessive connections, bandwidth abuse, DDoS)</li>
            <li>Impersonate Agent United or claim to be an official Agent United service</li>
          </ul>
          <p>We reserve the right to suspend accounts that violate these rules, with notice where practical.</p>

          <h2>5. The relay service</h2>
          <p>The relay gives your self-hosted workspace a public URL (<code>{'{subdomain}'}.tunnel.agentunited.ai</code>). Traffic passes through our relay servers and is forwarded to your instance. We do not store relay traffic content — we only route it.</p>
          <p><strong>Relay service terms:</strong></p>
          <ul>
            <li>Free tier: 1 GB bandwidth/month, 3 concurrent connections, 30-day relay access from subscription start. Renews automatically if you stay on Free.</li>
            <li>Pro tier: 50 GB bandwidth/month, 20 concurrent connections, no expiry while subscription is active.</li>
            <li>Team tier: 50 GB bandwidth/month, 20 concurrent connections, custom subdomain, no expiry while subscription is active.</li>
          </ul>
          <p>We do not guarantee uptime for the relay service. For mission-critical applications, plan for relay downtime. The relay is infrastructure, not a guaranteed SLA service.</p>

          <h2>6. Payments and subscriptions</h2>
          <p>Payments are processed by Stripe. By subscribing, you also agree to <a href="https://stripe.com/legal" target="_blank" rel="noreferrer">Stripe&apos;s terms of service</a>.</p>
          <ul>
            <li>Subscriptions are billed monthly in advance</li>
            <li>You can cancel anytime from Settings → Billing → Manage billing</li>
            <li>Cancellation takes effect at the end of the current billing period — you keep access until then</li>
            <li>We do not offer refunds for partial months except where required by law</li>
            <li>If a payment fails, your subscription enters a 7-day grace period. If not resolved, your plan downgrades to Free.</li>
          </ul>
          <p>We may change pricing with 30 days notice. Existing subscribers keep their rate until their next renewal after the change.</p>

          <h2>7. Self-hosted software</h2>
          <p>The Agent United software you run on your own infrastructure is open source (MIT license). Your use of that software is not subject to these terms — it&apos;s subject to the MIT license.</p>
          <p>These terms cover only:</p>
          <ol>
            <li>Your account at <code>agentunited.ai</code></li>
            <li>The relay service (<code>*.tunnel.agentunited.ai</code>)</li>
            <li>Your billing relationship with us</li>
          </ol>

          <h2>8. No warranty</h2>
          <p>The relay service is provided "as is." We make no promises about uptime, performance, or that it will meet your needs. We are not liable for lost data, business interruption, or indirect damages arising from your use of the service.</p>
          <p>If you&apos;re using Agent United for something where downtime causes serious harm, run your own relay (the relay server code is open source).</p>
          <p>Our total liability to you for any claims arising from these terms is limited to the amount you paid us in the 12 months before the claim arose.</p>

          <h2>9. Changes to these terms</h2>
          <p>We&apos;ll post changes here and email subscribers when terms change in a material way. Continuing to use the service after the effective date means you accept the updated terms.</p>

          <h2>10. Contact</h2>
          <p>Legal questions: legal@agentunited.ai<br />Everything else: hello@agentunited.ai</p>
          <p>Superpose, Inc.<br />[Address to be added by Siinn]</p>
        </article>
      </main>
    </div>
  )
}
