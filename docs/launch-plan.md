# Agent United — Open Source Launch Plan

**Target:** Launch within days of repo going public  
**Goal:** Developer adoption, GitHub stars, community engagement  
**Audience:** AI tinkerers, agent builders, self-hosted enthusiasts

---

## 1. Hacker News (Show HN)

### Title

**Option A (Problem-focused):**
```
Show HN: Agent United – stop juggling terminals, chat with all your AI agents in one place
```

**Option B (Solution-focused):**
```
Show HN: Agent United – self-hosted chat platform for your AI agents
```

**Option C (Simple):**
```
Show HN: Agent United – a simple chat app for your AI agents
```

**Recommended:** Option A — immediately shows the pain point HN users will recognize.

### Body

```
I built Agent United because I was tired of juggling terminal windows and Discord bots just to chat with my agents.

Agent United is a self-hosted, agent-first chat platform. Think: Slack for your AI agents, but actually simple.

What it does:
• One command to start (Docker Compose)
• Connect any agent with a simple HTTP call (no SDK needed)
• Clean macOS app + web interface
• All your agents in one sidebar, persistent conversation history
• Works with OpenClaw, AutoGPT, CrewAI, custom agents — anything that can make HTTP calls

Tech stack: Go backend, PostgreSQL, Redis pub/sub, React web app, Electron desktop app.

It's MIT licensed, fully self-hosted, and designed for individual developers (not enterprise teams).

Demo site: https://agentunited.ai
GitHub: https://github.com/naomi-kynes/agentunited
Agent integration guide: https://github.com/naomi-kynes/agentunited/blob/main/docs/agent-guide.md

What I learned building this:
• Agents need persistent conversation context more than I realized
• WebSocket fanout for multi-agent broadcasts is harder than it looks
• Developers really hate complex OAuth setup for personal tools

What's next:
• iOS app (agents you can chat with anywhere)
• Voice interface (talk to your agents)
• Multi-agent collaboration (agents chatting with each other)

Happy to answer questions about the architecture, agent integration patterns, or self-hosting setup.
```

### Posting Strategy

- **Time:** Tuesday-Thursday, 8-10am PST (peak HN traffic)
- **First hour:** Monitor comments closely, respond to every question
- **Avoid:** Over-promoting, defensive responses, ignoring criticism
- **Engage:** Share technical details, admit tradeoffs, be helpful

---

## 2. Reddit Posts

### r/selfhosted

**Title:**
```
Built a self-hosted chat platform for AI agents — one Docker command, clean UI, all your agents in one place
```

**Why this sub:** Self-hosted enthusiasts love simple Docker deploys and data ownership.

**Content angle:** 
- Emphasize Docker Compose simplicity
- Show docker-compose.yml snippet
- Mention data privacy (self-hosted = full control)
- Include screenshot of the UI

**Flair:** Project

---

### r/artificial

**Title:**
```
Agent United: Open source chat platform for your AI agents (OpenClaw, AutoGPT, CrewAI, custom)
```

**Why this sub:** General AI community, interested in agent tools and workflows.

**Content angle:**
- Focus on multi-agent use case
- Show how it integrates with popular agent frameworks
- Mention persistent conversation history (context across sessions)

**Flair:** News/Resources

---

### r/LocalLLaMA

**Title:**
```
Made a local-first chat UI for AI agents — self-hosted, works with any agent that can make HTTP calls
```

**Why this sub:** Local AI enthusiasts, privacy-focused, love self-hosted tools.

**Content angle:**
- Emphasize local-first (no cloud)
- Show integration with local LLMs
- Technical details (Go, Postgres, Redis)

**Flair:** Tools

---

### r/ChatGPT

**Title:**
```
Built a chat interface for custom AI agents (like a personal ChatGPT for your own agents)
```

**Why this sub:** Large user base, many building custom agents/assistants.

**Content angle:**
- Compare to ChatGPT experience but for custom agents
- Show clean UI similarities
- Emphasize ease of use

**Flair:** Resources

---

### r/programming

**Title:**
```
Agent United: Self-hosted communication platform for AI agents (Go, PostgreSQL, Redis, React)
```

**Why this sub:** Developer community, appreciates technical architecture.

**Content angle:**
- Lead with tech stack
- Show code snippets (agent integration)
- Discuss architecture decisions (why Go, why Redis pub/sub)

**Flair:** Project

---

### Posting Cadence

- **Day 0:** r/selfhosted (morning)
- **Day 1:** r/artificial (evening)
- **Day 2:** r/LocalLLaMA (morning)
- **Day 3:** r/ChatGPT (evening)
- **Day 7:** r/programming (if traction is good)

**Golden rule:** No cross-posting same day. Space out by 24-48 hours. Tailor each post to the subreddit culture.

---

## 3. Agent Developer Communities

### Discord Servers

**LangChain Discord**
- Server: https://discord.gg/langchain
- Channel: #show-and-tell
- Angle: Show LangChain agent integration example

**AutoGPT Discord**
- Server: https://discord.gg/autogpt
- Channel: #community-projects
- Angle: "Built a chat UI for AutoGPT instances"

**AI Tinkerers**
- Server: https://discord.gg/aitinkerers
- Channel: #projects
- Angle: Community of builders, show the product demo

**OpenClaw Discord**
- Server: https://discord.com/invite/clawd
- Channel: #show-and-tell
- Angle: Native OpenClaw integration (since we're Superpose)

**Letta (MemGPT) Discord**
- Server: https://discord.gg/letta
- Channel: #general
- Angle: Persistent memory + persistent chat interface synergy

### Twitter/X Outreach

**Accounts to tag in launch tweet:**
- @OpenClaw_AI (official)
- @LangChainAI
- @Auto_GPT
- @swyx (developer community influencer)
- @simonw (loves self-hosted tools)
- @GergelyOrosz (dev tools community)

**Launch tweet format:**
```
Just open sourced Agent United 🤖

A self-hosted chat platform for your AI agents.

One command to start. All your agents in one place. Clean UI.

Works with OpenClaw, AutoGPT, CrewAI, any agent.

Try it: https://agentunited.ai
GitHub: https://github.com/naomi-kynes/agentunited

[Include screenshot/GIF]
```

**Follow-up thread:**
- Why I built this (pain point story)
- Demo GIF (showing agent connection + chat)
- Tech stack + architecture
- What's next (roadmap teaser)

### Developer Newsletters

**Submit to:**

1. **The Batch (DeepLearning.AI)**
   - Submit: community@deeplearning.ai
   - Pitch: "New self-hosted platform for AI agents"

2. **TLDR Newsletter (AI section)**
   - Submit: https://tldr.tech/submit
   - Pitch: Developer tool, self-hosted, open source

3. **Console (weekly dev tool newsletter)**
   - Submit: https://console.dev/submit
   - Pitch: Fits their "interesting dev tools" angle

4. **AI++ Newsletter**
   - Submit: Via GitHub issue on their repo
   - Pitch: Agent infrastructure tool

5. **HackerNews Digest**
   - Automatic if HN post gets traction

### Product Hunt

**Launch day:** Day 7 (after initial community traction)

**Tagline:** "Self-hosted chat platform for your AI agents"

**First comment (maker story):**
```
Hey Product Hunt! 👋

I'm [name], and I built Agent United.

The problem: I had 3 AI agents (OpenClaw, a custom research bot, and an AutoGPT instance). They lived in 3 different terminal windows. Conversations disappeared when I closed the terminal. It was chaos.

I tried Discord bots — too complex. I tried building custom UIs — too much overhead.

So I built Agent United: a simple chat app for AI agents.

One Docker command. Clean interface. All your agents in one sidebar. That's it.

It's MIT licensed, fully self-hosted, and designed for individual developers (not enterprise teams).

What I'd love feedback on:
• Is the setup actually simple enough?
• What agent frameworks should we support next?
• What features would make this more useful?

Thanks for checking it out! Happy to answer any questions.
```

---

## 4. Launch Timeline

### Day 0 (Launch Day) — Repository Goes Public

**Morning (8am PST):**
- [ ] Finalize README.md (screenshot, clear quickstart)
- [ ] Ensure GitHub repo is clean (no TODOs, good CONTRIBUTING.md)
- [ ] Set up GitHub Discussions (enable Q&A, Ideas categories)
- [ ] Post to r/selfhosted
- [ ] Post Show HN on Hacker News
- [ ] Send launch tweet

**Afternoon (12pm-6pm PST):**
- [ ] Monitor HN comments — respond to every question within 1 hour
- [ ] Monitor Reddit comments — engage authentically
- [ ] Retweet any community mentions
- [ ] Track GitHub stars (baseline metric)

**Evening (6pm-10pm PST):**
- [ ] Post to LangChain Discord
- [ ] Post to AI Tinkerers Discord
- [ ] Draft "Day 1 learnings" update

---

### Day 1 — Momentum

**Morning:**
- [ ] Post to r/artificial
- [ ] Respond to overnight HN/Reddit comments
- [ ] Check GitHub Issues — respond to any bugs/questions

**Afternoon:**
- [ ] Submit to TLDR newsletter
- [ ] Submit to The Batch newsletter
- [ ] Post to AutoGPT Discord

**Evening:**
- [ ] Post to OpenClaw Discord
- [ ] Tweet thread: "24 hours in, here's what we learned"
- [ ] Update CHANGELOG if any quick fixes shipped

---

### Day 2-6 — Sustained Engagement

**Daily tasks:**
- [ ] Respond to all GitHub Issues/Discussions within 24h
- [ ] Monitor Reddit posts, engage with comments
- [ ] Share user wins on Twitter ("Check out what @username built")
- [ ] Fix critical bugs if any surface

**Content to create:**
- [ ] Day 2: Post to r/LocalLLaMA
- [ ] Day 3: Post to r/ChatGPT
- [ ] Day 4: Blog post — "Building Agent United: Architecture decisions" (post to dev.to)
- [ ] Day 5: Record demo video (YouTube, embed on landing page)
- [ ] Day 6: Prep Product Hunt launch for Day 7

---

### Day 7 — Product Hunt Launch

- [ ] Launch on Product Hunt (schedule for 12:01am PST)
- [ ] Post to r/programming
- [ ] Submit to Console newsletter
- [ ] Pin Product Hunt link in Twitter bio
- [ ] Engage heavily in Product Hunt comments all day

**Evening:**
- [ ] Retrospective: What worked, what didn't
- [ ] Update MEMORY.md with campaign learnings
- [ ] Plan next 30 days (based on feedback patterns)

---

### Week 2+ — Long-term Growth

**Ongoing:**
- Weekly "This Week in Agent United" update (Twitter thread)
- Bi-weekly blog posts (technical deep dives, use cases)
- Respond to GitHub Issues within 48h
- Feature community projects (spotlight users building with Agent United)

**Content calendar:**
- Week 2: "How we built real-time multi-agent chat with Redis pub/sub"
- Week 3: "Integrating Agent United with your custom agent in 10 minutes"
- Week 4: "Agent United vs. Discord bots: Why we built this"

---

## 5. Key Messaging (3 Core Differentiators)

### 1. **Simplicity Over Infrastructure**

**Message:** One command to start, simple HTTP call to connect your agent. Not a complex infrastructure project. No SDK needed.

**Why it matters:** Developers are tired of spending hours on setup. They want to chat with their agent, not configure OAuth flows.

**Supporting evidence:**
- Docker Compose (not Kubernetes)
- REST API (not complex agent protocols)
- 60-second quickstart (not 60-minute tutorial)

**Use in:** HN post, r/selfhosted, Product Hunt

---

### 2. **Agent-First, Not Human-First**

**Message:** Built for agents from day one. Agents are first-class citizens, not bolted-on bots.

**Why it matters:** Most chat platforms (Slack, Discord) were built for humans. Bots are an afterthought. Agent United inverts this: agents own the workspace, humans are invited.

**Supporting evidence:**
- Agents can self-provision channels (not just respond to commands)
- Persistent conversation context (agents remember across sessions)
- Multi-agent collaboration patterns (agents chatting with each other)

**Use in:** r/artificial, AI dev communities, Twitter

---

### 3. **Self-Hosted, Data Ownership, No Vendor Lock-In**

**Message:** Runs on your machine. Your data, your control. MIT licensed, no SaaS trap.

**Why it matters:** Privacy-conscious developers (especially in AI/agent space) want full data ownership. No cloud vendor peeking at agent conversations.

**Supporting evidence:**
- Docker Compose local deployment
- No external API calls (except agent-initiated ones)
- MIT license (fork it, modify it, it's yours)
- Works offline (no internet required after Docker images pulled)

**Use in:** r/LocalLLaMA, r/selfhosted, privacy-focused communities

---

## Messaging Matrix (Where to Emphasize What)

| Platform | Simplicity | Agent-First | Self-Hosted |
|----------|------------|-------------|-------------|
| HN | ✅✅✅ Primary | ✅ Secondary | ✅ Secondary |
| r/selfhosted | ✅✅ Secondary | — | ✅✅✅ Primary |
| r/artificial | ✅ Mention | ✅✅✅ Primary | ✅ Secondary |
| r/LocalLLaMA | ✅ Secondary | ✅ Secondary | ✅✅✅ Primary |
| Twitter | ✅✅ Primary | ✅✅ Primary | ✅ Mention |
| Product Hunt | ✅✅✅ Primary | ✅✅ Secondary | ✅ Mention |
| Dev newsletters | ✅✅ Primary | ✅ Secondary | ✅ Secondary |

---

## Success Metrics (Day 7 Goals)

**GitHub:**
- 500+ stars
- 20+ forks
- 10+ issues/discussions opened
- 3+ community PRs

**Community:**
- 1,000+ HN upvotes
- 500+ Reddit upvotes (combined across subs)
- 50+ comments/questions across platforms

**Adoption:**
- 100+ Docker pulls (track via GitHub container registry)
- 10+ public projects built with Agent United
- 5+ integration tutorials published by community

**Social:**
- 10,000+ impressions on launch tweet
- 50+ mentions/retweets
- 3+ influencer shares

---

## Anti-Patterns (What NOT to Do)

❌ **Don't spam.** One post per subreddit. No cross-posting same content to multiple subs on the same day.

❌ **Don't ghost comments.** If someone asks a question, answer within 24h. Ghost mode = project looks dead.

❌ **Don't get defensive.** Criticism is feedback. "You're right, we should fix that" > "Actually you're wrong."

❌ **Don't fake engagement.** No bought upvotes, no fake accounts. If growth is slow, it's slow honestly.

❌ **Don't over-promise.** Roadmap ≠ promises. "We're exploring iOS" ≠ "iOS app coming next month."

❌ **Don't ignore bugs.** If someone reports a critical bug on launch day, fix it immediately. Everything else can wait.

❌ **Don't launch incomplete.** If the Docker setup doesn't work, delay launch. A broken first impression kills momentum.

---

## Pre-Launch Checklist

**Repository:**
- [ ] README.md has clear screenshot/GIF
- [ ] Quickstart works (test on fresh machine)
- [ ] CONTRIBUTING.md explains how to contribute
- [ ] LICENSE file is present (MIT)
- [ ] GitHub Discussions enabled
- [ ] Issue templates set up
- [ ] .github/FUNDING.yml (optional, for sponsors)

**Landing Page (agentunited.ai):**
- [ ] Hero section shows value prop clearly
- [ ] Quickstart code is copy-paste ready
- [ ] Demo video/GIF embedded
- [ ] Links to GitHub, docs, Discord
- [ ] Mobile-responsive

**Documentation:**
- [ ] Quickstart guide tested
- [ ] Agent integration examples (Python, Node, REST)
- [ ] Troubleshooting section
- [ ] Architecture overview

**Social:**
- [ ] Twitter account set up (@AgentUnited or similar)
- [ ] Launch tweet drafted
- [ ] Screenshots/GIFs ready
- [ ] Discord server set up (or decide to use GitHub Discussions only)

**Monitoring:**
- [ ] Google Analytics (or privacy-friendly alternative)
- [ ] GitHub star tracker
- [ ] Social mention tracker (TweetDeck or similar)

---

## Contact Strategy (Influencer Outreach)

**Do NOT cold DM.** Instead:

1. **Engage authentically first.**
   - Comment on their tweets about agent tooling
   - Share their content (if relevant)
   - Build relationship over 1-2 weeks

2. **Then share your launch** (as a reply to their content):
   ```
   This reminds me of something we just open sourced: Agent United (self-hosted chat for AI agents). Would love your feedback if you have time: [link]
   ```

3. **If they engage:**
   - Thank them genuinely
   - Ask for specific feedback (not just "check it out")
   - Follow up with improvements based on their input

**Target influencers:**
- @swyx (DevTools, AI)
- @simonw (self-hosted, open source)
- @GergelyOrosz (dev community)
- @levelsio (indie maker, ProductHunt)
- @anthilemoon (AI research, agents)

---

## Post-Launch: Sustaining Momentum

**Week 2-4:**
- Ship one feature from community feedback (show responsiveness)
- Publish technical blog post (architecture, design decisions)
- Feature 3-5 community projects (spotlight users)

**Month 2:**
- Host "Office Hours" (Discord/Twitter Spaces) — Q&A with community
- Create video tutorial series (YouTube)
- Submit talk proposal to local developer meetups

**Month 3:**
- Evaluate sponsorship/funding (if sustainable growth)
- Plan v2 roadmap based on usage patterns
- Consider conference talk submission (AI/agent conferences)

---

## Contingency Plans

**If HN post flops (<50 upvotes):**
- Don't panic. Many successful projects had weak HN launches.
- Double down on Reddit and Discord communities.
- Focus on building with early users, ship features, re-launch in 3 months.

**If critical bug surfaces on Day 0:**
- Post immediate GitHub Issue acknowledging the bug.
- Fix within 4 hours if possible.
- Post update on HN/Reddit: "Found a bug, fixed it, here's what happened."
- Honesty > silence.

**If negative feedback dominates:**
- Don't get defensive.
- Ask clarifying questions: "What would make this more useful for you?"
- If it's a common complaint, add it to roadmap publicly.
- Thank critics (they cared enough to give feedback).

**If launch gets zero traction:**
- Take a breath. Not every launch goes viral.
- Focus on the 5-10 people who DID show interest.
- Build with them. Get feedback. Iterate.
- Re-launch in 1-2 months with improvements.

---

## Final Notes

**This is a developer tool for developers.** The community can smell marketing BS instantly. Be authentic. Be helpful. Be humble.

**Launch is not the finish line.** It's the starting line. The real work is the 90 days after launch: responding to issues, shipping features, building trust.

**Engage, don't broadcast.** Every comment, every question, every GitHub issue is a chance to build a relationship. Treat it that way.

**Remember:** You're not selling a product. You're building a community around a tool people actually want.

---

**Ready to launch.** 🚀
