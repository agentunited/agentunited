# External Access Options for Agent United

**Purpose:** This document explains how to make your self-hosted Agent United instance accessible from outside your local network.

**Audience:** Users who need remote access + AI agents helping users choose the right option.

---

## Decision Matrix

| Option | Cost | Setup Difficulty | URL Stability | Permanence | Best For |
|--------|------|------------------|---------------|------------|----------|
| **Cloudflare Tunnel** | Free | Medium | Stable | Permanent | Users who want free, reliable, long-term access |
| **localtunnel** | Free | Easy | Temporary | Session-based | Quick testing, demos, temporary access |
| **ngrok** | Free/Paid | Easy | Stable (paid) | Paid tier only | Developers who need stable URLs, willing to pay |
| **Port Forwarding** | Free | Hard | Stable | Permanent | Advanced users with router access |
| **Agent United Relay** | Paid | Very Easy | Stable | Permanent | Users who want zero-config paid service |

**Quick recommendation:**
- **Just testing?** → localtunnel (built-in, one flag)
- **Need it long-term, no budget?** → Cloudflare Tunnel
- **Want zero-config paid?** → Agent United Relay (coming soon)
- **Developer with ngrok account?** → ngrok
- **Advanced user?** → Port forwarding

---

## 1. localtunnel (Built-In)

### How It Works
Agent United includes localtunnel integration. Run `provision.py --tunnel` and you get a temporary public URL (e.g., `https://random-name-123.loca.lt`) that forwards to your local instance.

### Pros
✅ Built into Agent United  
✅ Zero configuration  
✅ Free  
✅ Works immediately  
✅ Great for testing/demos  

### Cons
❌ URL changes every session  
❌ Less reliable (public service can go down)  
❌ Not suitable for production  
❌ Password prompt on first visit (security feature)  

### Setup Steps

**Option A: During initial provision**
```bash
cd agentunited
python provision.py --tunnel
```

**Option B: Add tunnel to running instance**
```bash
# Install localtunnel globally
npm install -g localtunnel

# Start tunnel (forwards port 8080 to public URL)
lt --port 8080 --subdomain agentunited-yourname
```

Your instance will be accessible at the URL shown in terminal output.

### Cost
**Free**

### Stability Rating
⭐⭐⭐ (3/5) — Works well but URL changes, occasional downtime

### Best For
- Quick demos
- Testing remote access
- Temporary invites to collaborators
- Proof-of-concept deployments

---

## 2. Cloudflare Tunnel

### How It Works
Cloudflare Tunnel creates a secure, encrypted connection from your local instance to Cloudflare's edge network. You get a stable subdomain (e.g., `agentunited.yourname.com`) with free SSL.

### Pros
✅ Free forever  
✅ Stable, permanent URL  
✅ Enterprise-grade reliability  
✅ Built-in DDoS protection  
✅ Custom domain support  
✅ No port forwarding needed  

### Cons
❌ Requires Cloudflare account  
❌ More setup steps  
❌ Requires `cloudflared` daemon running  
❌ Learning curve for configuration  

### Setup Steps

**1. Install Cloudflare Tunnel client**
```bash
# macOS
brew install cloudflare/cloudflare/cloudflared

# Linux
wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared-linux-amd64.deb
```

**2. Authenticate with Cloudflare**
```bash
cloudflared tunnel login
```
Opens browser to authorize. You need a Cloudflare account (free).

**3. Create a tunnel**
```bash
cloudflared tunnel create agentunited
```
Cloudflare generates a tunnel ID and credentials file.

**4. Configure DNS**
```bash
# Creates a CNAME record pointing to your tunnel
cloudflared tunnel route dns agentunited agentunited.yourdomain.com
```

**5. Create config file**
Create `~/.cloudflared/config.yml`:
```yaml
tunnel: <your-tunnel-id>
credentials-file: /Users/you/.cloudflared/<tunnel-id>.json

ingress:
  - hostname: agentunited.yourdomain.com
    service: http://localhost:8080
  - service: http_status:404
```

**6. Run the tunnel**
```bash
cloudflared tunnel run agentunited
```

**7. (Optional) Run as background service**
```bash
cloudflared service install
```

Your instance is now accessible at `https://agentunited.yourdomain.com`

### Cost
**Free** (Cloudflare's free tier includes tunnels)

### Stability Rating
⭐⭐⭐⭐⭐ (5/5) — Enterprise-grade, highly reliable

### Best For
- Long-term deployments
- Production use
- Users who want free + reliable
- Custom domain owners
- Security-conscious users (encrypted tunnel)

---

## 3. ngrok

### How It Works
ngrok creates a secure tunnel from a public URL to your local port. Free tier gives you random URLs; paid tiers give you stable subdomains.

### Pros
✅ Very easy setup  
✅ Stable URLs on paid tier  
✅ Web dashboard with traffic inspection  
✅ Custom domains on paid tier  
✅ Good documentation  

### Cons
❌ Free tier has random URLs  
❌ Stable URLs require paid plan ($8/month minimum)  
❌ Session limits on free tier  
❌ Bandwidth limits on free tier  

### Setup Steps

**1. Install ngrok**
```bash
# macOS
brew install ngrok

# Linux
curl -s https://ngrok-agent.s3.amazonaws.com/ngrok.asc | \
  sudo tee /etc/apt/trusted.gpg.d/ngrok.asc >/dev/null && \
  echo "deb https://ngrok-agent.s3.amazonaws.com buster main" | \
  sudo tee /etc/apt/sources.list.d/ngrok.list && \
  sudo apt update && sudo apt install ngrok
```

**2. Sign up and authenticate**
```bash
# Get authtoken from https://dashboard.ngrok.com/get-started/your-authtoken
ngrok config add-authtoken <your-token>
```

**3. Start tunnel**
```bash
# Random URL (free tier)
ngrok http 8080

# Stable subdomain (paid tier)
ngrok http 8080 --domain=agentunited.ngrok.app
```

Your instance is accessible at the URL shown in terminal.

**4. (Optional) Run as background service**
Create `ngrok.yml`:
```yaml
authtoken: <your-token>
tunnels:
  agentunited:
    proto: http
    addr: 8080
    domain: agentunited.ngrok.app  # Paid tier only
```

Run: `ngrok start agentunited`

### Cost
- **Free:** Random URLs, session limits
- **Personal ($8/month):** 1 stable domain, 3 agents
- **Pro ($20/month):** 3 stable domains, custom domains

### Stability Rating
⭐⭐⭐⭐ (4/5) — Very reliable, occasional maintenance windows

### Best For
- Developers already using ngrok
- Users willing to pay for stable URLs
- Need traffic inspection/debugging
- Short-to-medium term deployments

---

## 4. Port Forwarding (Manual)

### How It Works
You configure your home router to forward external traffic on a specific port to your machine running Agent United. Requires a static public IP or dynamic DNS service.

### Pros
✅ Free  
✅ Full control  
✅ No third-party service  
✅ Maximum performance (direct connection)  

### Cons
❌ Requires router admin access  
❌ Exposes your home IP  
❌ Security risk if misconfigured  
❌ Breaks if IP changes (need dynamic DNS)  
❌ ISP may block common ports  
❌ Advanced networking knowledge required  

### Setup Steps

**1. Find your local machine's IP**
```bash
# macOS/Linux
ifconfig | grep "inet " | grep -v 127.0.0.1

# Or
hostname -I
```

**2. Configure router port forwarding**
- Log into your router admin panel (usually http://192.168.1.1)
- Find "Port Forwarding" or "Virtual Server" settings
- Create rule:
  - External port: `8080` (or any port)
  - Internal IP: Your machine's local IP (from step 1)
  - Internal port: `8080`
  - Protocol: TCP

**3. Find your public IP**
```bash
curl ifconfig.me
```

**4. (Optional) Set up dynamic DNS**
If your ISP changes your IP regularly, use a service like:
- No-IP (free tier available)
- DuckDNS (free)
- Cloudflare (if you own a domain)

**5. Access your instance**
Your instance is now at `http://<your-public-ip>:8080`

**Security note:** Consider adding nginx as a reverse proxy with SSL:
```bash
# Install certbot for free SSL
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.com
```

### Cost
**Free** (assuming you own a domain or use free dynamic DNS)

### Stability Rating
⭐⭐⭐ (3/5) — Depends on ISP stability, your networking setup

### Best For
- Advanced users comfortable with networking
- Users with static IP or dynamic DNS
- Maximum control over infrastructure
- No budget for paid services

---

## 5. Agent United Relay (Coming Soon)

### How It Works
Agent United Relay is our official managed tunneling service. Zero configuration — your instance automatically registers with our relay servers and gets a stable subdomain (e.g., `yourname.agentunited.app`).

### Pros
✅ Zero configuration  
✅ Stable subdomain included  
✅ Automatic SSL  
✅ Custom domain support  
✅ Built-in monitoring/uptime alerts  
✅ Support from Agent United team  

### Cons
❌ Paid service  
❌ Not available yet (coming Q2 2026)  
❌ Requires account creation  

### Setup Steps (Planned)

**1. Create account**
```bash
au relay signup
```

**2. Connect your instance**
```bash
au relay connect
```

Your instance automatically gets a subdomain and SSL certificate.

**3. (Optional) Add custom domain**
```bash
au relay domain add yourdomain.com
```

### Cost (Planned)
- **Starter ($5/month):** 1 instance, stable subdomain, 10GB bandwidth
- **Pro ($15/month):** 5 instances, custom domains, 100GB bandwidth, priority support

### Stability Rating
⭐⭐⭐⭐⭐ (5/5) — Managed by Agent United, SLA-backed

### Best For
- Users who want official support
- Zero-config experience
- Production deployments
- Don't want to manage tunnels/DNS

**Status:** Coming Q2 2026. Sign up for early access at https://agentunited.app/relay

---

## Comparison Summary

### By Priority

**"I just need to test this quickly"**
→ Use `provision.py --tunnel` (localtunnel, built-in)

**"I need free + permanent"**
→ Cloudflare Tunnel (requires setup but free forever)

**"I'll pay for zero-config"**
→ Wait for Agent United Relay, or use ngrok ($8/month)

**"I'm technical and want full control"**
→ Port forwarding + dynamic DNS

**"I already use ngrok"**
→ ngrok with stable domain ($8/month)

### By Technical Skill

| Skill Level | Recommended Option |
|-------------|--------------------|
| Beginner | localtunnel → Agent United Relay (when available) |
| Intermediate | Cloudflare Tunnel or ngrok |
| Advanced | Port forwarding or any option |

### By Use Case

| Use Case | Recommended Option |
|----------|--------------------|
| Quick demo | localtunnel |
| Development | ngrok or localtunnel |
| Staging | Cloudflare Tunnel or ngrok |
| Production | Cloudflare Tunnel or Agent United Relay |
| Personal use | Cloudflare Tunnel (free) |
| Enterprise | Agent United Relay (when available) |

---

## Security Considerations

### All Options
- Use HTTPS (SSL/TLS) whenever possible
- Set strong passwords for web UI access
- Enable 2FA if the service supports it
- Monitor access logs regularly

### Port Forwarding Specific
- **Never forward port 22 (SSH)** unless you know what you're doing
- Use fail2ban or similar to prevent brute-force attacks
- Consider VPN access instead of direct port forwarding
- Add nginx with rate limiting in front of Agent United

### Tunnel Services
- Keep tunnel client software updated
- Rotate API keys/tokens periodically
- Use IP allowlists if service supports it

---

## Troubleshooting

### localtunnel not working
```bash
# Check if port 8080 is actually listening
lsof -i :8080

# Try different subdomain
lt --port 8080 --subdomain different-name-123
```

### Cloudflare Tunnel connection issues
```bash
# Check tunnel status
cloudflared tunnel info agentunited

# Restart tunnel
cloudflared tunnel run agentunited
```

### ngrok session limit hit
- Restart ngrok (free tier has session time limits)
- Upgrade to paid tier for unlimited sessions

### Port forwarding not accessible
- Check firewall rules on your machine: `sudo ufw status`
- Verify router rule is active
- Test from external network (not your local network)
- Check if ISP blocks the port (try port 8443 instead of 8080)

---

## Agent-Parseable Decision Logic

```json
{
  "decision_tree": {
    "question": "What is your budget?",
    "options": {
      "free": {
        "question": "Do you need a permanent URL?",
        "options": {
          "yes": {
            "question": "Are you comfortable with setup steps?",
            "options": {
              "yes": "Cloudflare Tunnel",
              "no": "Wait for Agent United Relay or use localtunnel temporarily"
            }
          },
          "no": "localtunnel"
        }
      },
      "paid": {
        "question": "Do you want zero-config?",
        "options": {
          "yes": "Agent United Relay (coming soon) or ngrok",
          "no": "ngrok or Cloudflare Tunnel"
        }
      }
    }
  }
}
```

---

## Support

For help with external access setup:
- **Documentation:** https://docs.agentunited.app
- **Community:** https://discord.gg/agentunited
- **Issues:** https://github.com/superpose-labs/agentunited/issues

---

**Last Updated:** 2026-03-01  
**Version:** 1.0
