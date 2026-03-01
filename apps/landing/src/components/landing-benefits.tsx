import { Users, History, Sparkles, Zap, Puzzle, Shield } from 'lucide-react'

export function LandingBenefits() {
  const benefits = [
    {
      icon: Users,
      title: "All Agents in One Place",
      description: "Stop juggling terminal windows and Discord servers. See all your agents in one beautiful sidebar.",
      highlight: "End the tab chaos",
      stat: "10+ agents supported"
    },
    {
      icon: History,
      title: "Never Lose Conversations",
      description: "Every chat is saved automatically. Search past messages. Agents remember context across sessions.",
      highlight: "Full conversation history",
      stat: "Unlimited storage"
    },
    {
      icon: Sparkles,
      title: "Beautiful Chat Interface",
      description: "macOS app, web browser, iOS coming soon. Clean design that feels like texting a friend.",
      highlight: "Actually enjoyable to use",
      stat: "Native performance"
    },
    {
      icon: Zap,
      title: "Setup in 60 Seconds",
      description: "One command to start. 3 lines to connect. No OAuth, webhooks, or configuration hell.",
      highlight: "Seriously that simple",
      stat: "2 minute setup"
    },
    {
      icon: Puzzle,
      title: "Works with Everything",
      description: "OpenClaw, AutoGPT, CrewAI, custom agents. If it can make HTTP calls, it works.",
      highlight: "Bring any agent",
      stat: "Framework agnostic"
    },
    {
      icon: Shield,
      title: "Your Data, Your Control",
      description: "Self-hosted on your machine. No cloud vendor. No data leaves your computer.",
      highlight: "Complete privacy",
      stat: "100% local"
    }
  ]

  return (
    <section className="relative py-32 bg-transparent">
      <div className="container mx-auto px-6 lg:px-8">
        {/* Enterprise Section Header */}
        <div className="text-center mb-24 animate-fade-in-up">
          <div className="inline-flex items-center px-4 py-2 rounded-full bg-white/5 border border-white/10 backdrop-blur-sm mb-6">
            <span className="text-sm font-semibold text-white/80 uppercase tracking-wider">Why Choose AgentUnited</span>
          </div>
          <h2 className="text-5xl md:text-6xl lg:text-7xl font-bold text-white mb-10 leading-tight tracking-tight">
            Everything you need.{' '}
            <br className="hidden md:block" />
            <span className="text-white/50">Nothing you don't.</span>
          </h2>
          <p className="text-xl md:text-2xl text-white/70 max-w-4xl mx-auto leading-relaxed font-light">
            Built specifically for people who want to chat with their AI agents. 
            Not teams. Not enterprises. Just you and your agents.
          </p>
        </div>

        {/* Enterprise Benefits Grid */}
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8 mb-20 animate-fade-in-up animation-delay-200">
          {benefits.map((benefit, index) => {
            const Icon = benefit.icon
            return (
              <div key={index} className="group relative">
                {/* Enterprise Dark Card */}
                <div className="bg-white/5 backdrop-blur-xl rounded-3xl p-8 border border-white/10 hover:border-white/20 shadow-xl hover:shadow-2xl transition-all duration-500 hover:scale-105 h-full">
                  {/* Icon and stat */}
                  <div className="flex items-start justify-between mb-8">
                    <div className="w-16 h-16 bg-white/10 rounded-2xl flex items-center justify-center group-hover:bg-emerald-500/20 group-hover:scale-110 transition-all duration-300">
                      <Icon className="w-8 h-8 text-emerald-400 group-hover:text-emerald-300 transition-colors duration-300" />
                    </div>
                    <div className="text-right">
                      <div className="text-xs text-white/40 uppercase tracking-wider font-semibold">Impact</div>
                      <div className="text-sm font-bold text-emerald-400">{benefit.stat}</div>
                    </div>
                  </div>
                  
                  {/* Content */}
                  <h3 className="text-2xl font-bold text-white mb-4 group-hover:text-emerald-300 transition-colors">
                    {benefit.title}
                  </h3>
                  
                  <p className="text-white/70 mb-8 leading-relaxed text-lg">
                    {benefit.description}
                  </p>
                  
                  {/* Enterprise Highlight */}
                  <div className="bg-gradient-to-r from-emerald-500/20 to-emerald-500/10 backdrop-blur-sm rounded-2xl p-5 border border-emerald-500/30">
                    <p className="text-emerald-300 font-semibold">
                      {benefit.highlight}
                    </p>
                  </div>
                </div>
              </div>
            )
          })}
        </div>

        {/* Enterprise Social Proof */}
        <div className="bg-gradient-to-r from-white/5 via-white/10 to-white/5 backdrop-blur-xl rounded-3xl p-12 lg:p-16 border border-white/20 mb-20 animate-fade-in-up animation-delay-400">
          <h3 className="text-3xl lg:text-4xl font-bold text-center text-white mb-16">
            What developers are saying
          </h3>
          
          <div className="grid md:grid-cols-3 gap-8">
            <div className="bg-white/5 backdrop-blur-sm rounded-2xl p-8 border border-white/10 hover:border-white/20 transition-all duration-300 hover:scale-105">
              <div className="mb-8">
                <div className="text-amber-400 text-3xl mb-4">★★★★★</div>
                <p className="text-white/80 italic leading-relaxed text-lg">
                  "I spent 4 hours trying to set up a Discord bot for my agent. With AgentUnited, I was chatting in 2 minutes."
                </p>
              </div>
              <div className="flex items-center">
                <div className="w-12 h-12 bg-emerald-500 rounded-2xl flex items-center justify-center mr-4">
                  <span className="text-white font-bold">AK</span>
                </div>
                <div>
                  <div className="font-semibold text-white">Alex K.</div>
                  <div className="text-white/60 text-sm">OpenClaw user</div>
                </div>
              </div>
            </div>
            
            <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-200">
              <div className="mb-6">
                <div className="text-yellow-400 text-2xl mb-3">★★★★★</div>
                <p className="text-gray-700 italic leading-relaxed">
                  "Finally, a chat interface that doesn't feel like I'm fighting the tool. It just works."
                </p>
              </div>
              <div className="flex items-center">
                <div className="w-10 h-10 bg-liberty-green rounded-full flex items-center justify-center mr-3">
                  <span className="text-white font-bold text-sm">JM</span>
                </div>
                <div>
                  <div className="font-semibold text-deep-slate">Jordan M.</div>
                  <div className="text-gray-500 text-sm">AI Developer</div>
                </div>
              </div>
            </div>
            
            <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-200">
              <div className="mb-6">
                <div className="text-yellow-400 text-2xl mb-3">★★★★★</div>
                <p className="text-gray-700 italic leading-relaxed">
                  "I have 5 different agents. AgentUnited is the only place I can see them all together."
                </p>
              </div>
              <div className="flex items-center">
                <div className="w-10 h-10 bg-liberty-green rounded-full flex items-center justify-center mr-3">
                  <span className="text-white font-bold text-sm">SL</span>
                </div>
                <div>
                  <div className="font-semibold text-deep-slate">Sam L.</div>
                  <div className="text-gray-500 text-sm">Researcher</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Strong transition CTA */}
        <div className="text-center">
          <div className="max-w-3xl mx-auto">
            <h3 className="text-3xl font-bold text-deep-slate mb-6">
              Stop fighting with terminals and Discord bots
            </h3>
            <p className="text-xl text-gray-600 mb-8 leading-relaxed">
              Your agents deserve a proper home. You deserve a simple way to chat with them. 
              Give both what they need.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <a href="#quickstart" className="btn-enterprise-primary text-lg">
                Start Your Agent Chat Hub
              </a>
              <a href="#how-it-works" className="btn-enterprise-secondary text-lg">
                See It In Action
              </a>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}