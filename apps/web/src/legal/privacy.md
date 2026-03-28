**Effective date:** March 5, 2026  
**Last updated:** March 5, 2026

## What we collect and why

We collect the minimum data needed to run the service.

### Account data

- **Email address** — to identify your account and send you receipts and important notices
- **Display name** — how you appear in your workspace
- **Password** — stored as a bcrypt hash; we cannot see your plaintext password

### Billing data

- **Stripe customer ID** — to manage your subscription
- We do not store card numbers or full payment details. Stripe handles that.
- We store your subscription plan, status, and renewal dates.

### Usage data

- **Bandwidth used** — we count bytes proxied through the relay (running total per workspace per month) to enforce plan limits. We do not log the content of relay traffic.
- **API call logs** — standard server logs with timestamps and endpoints accessed. Retained for 30 days. No request bodies logged.

### What we do NOT collect

- The content of messages sent through your workspace
- The content of relay traffic
- Information about your AI agents' behavior or outputs
- Third-party tracking, analytics pixels, or ad data

## How we use your data

- **Email** — account management, billing receipts, service notices, product updates (you can unsubscribe from product updates)
- **Billing data** — processing payments, managing subscription state, handling disputes
- **Bandwidth data** — enforcing plan limits, generating your usage display in billing settings
- **Server logs** — debugging, security monitoring, abuse prevention

We do not sell your data. We do not share your data with third parties except:

- **Stripe** — for payment processing (governed by [Stripe's privacy policy](https://stripe.com/privacy))
- **Cloud infrastructure providers** — our servers run on cloud infrastructure providers; they have access to server infrastructure but not your application data
- **Legal requirements** — if required by a valid legal order, we'll comply and notify you if we're legally able to

## The relay service

When you use the relay, your network traffic passes through our servers. We forward it to your instance and do not store it. We log:

- Bytes transferred (for bandwidth accounting)
- Connection timestamps and subdomain (for debugging and abuse prevention)

We do not log the HTTP requests, responses, or message content flowing through the relay.

## Your rights

**Access** — you can request a copy of the data we hold about you: privacy@agentunited.ai

**Deletion** — you can delete your account from Settings → Account → Delete account. This permanently deletes:

- Your account, credentials, and billing record
- All workspace data on our servers (if using our hosted relay or cloud service)
- Your subscription (cancels at period end; no refund for unused time)

If you're self-hosting, your data lives on your own infrastructure and deletion is your responsibility.

**Correction** — you can update your email and display name in Settings → Profile.

**Data portability** — request an export of your account data: privacy@agentunited.ai

**GDPR/CCPA** — if you're in the EU or California, you have additional rights. Contact us at privacy@agentunited.ai and we'll respond within 30 days.

## Cookies and local storage

We use browser localStorage to keep you logged in (auth token). We do not use third-party cookies or tracking cookies. No advertising networks touch this site.

## Data retention

- Account data: retained until you delete your account
- Billing data: retained for 7 years (legal requirement for financial records)
- Server logs: 30 days
- Bandwidth counters: current month + 3 months history
- Deleted account data: purged within 30 days of deletion request (except billing records retained for legal compliance)

## Security

- Passwords: bcrypt hashed, never stored in plaintext
- API keys: stored as hashed values
- Transit: TLS everywhere (relay, API, web app)
- Relay traffic: encrypted in transit via HTTPS/WSS
- We perform security reviews periodically and patch known vulnerabilities promptly

If you find a security issue, email security@agentunited.ai. We'll respond within 48 hours.

## Children

Agent United is not for anyone under 18. If we learn we've collected data from someone under 18, we'll delete it.

## Changes

We'll post changes here and email you when policy changes are material. The "last updated" date at the top reflects the most recent revision.

## Contact

Agent United  
500 East Hamilton Ave. #1114  
Campbell, CA 95008  
United States

Privacy questions: privacy@agentunited.ai  
Security issues: security@agentunited.ai

