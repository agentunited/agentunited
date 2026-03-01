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
    <section className="section-enterprise">
      <div className="container-enterprise">
        {/* Section Header */}
        <div className="text-center mb-20">
          <div className="text-liberty-green text-lg font-semibold mb-4 tracking-wide uppercase">
            Why Choose AgentUnited
          </div>
          <h2 className="text-4xl sm:text-5xl font-bold text-deep-slate mb-8">
            Everything you need.{' '}
            <span className="text-gray-500">Nothing you don't.</span>
          </h2>
          <p className="text-xl text-gray-600 max-w-3xl mx-auto leading-relaxed">
            Built specifically for people who want to chat with their AI agents. 
            Not teams. Not enterprises. Just you and your agents.
          </p>
        </div>

        {/* Benefits Grid */}
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8 mb-20">
          {benefits.map((benefit, index) => {
            const Icon = benefit.icon
            return (
              <div key={index} className="group relative">
                {/* Card */}
                <div className="bg-white rounded-2xl p-8 border border-gray-200 shadow-sm hover:shadow-xl transition-all duration-300 hover:-translate-y-2 h-full">
                  {/* Icon and stat */}
                  <div className="flex items-start justify-between mb-6">
                    <div className="w-14 h-14 bg-liberty-green/10 rounded-xl flex items-center justify-center group-hover:bg-liberty-green group-hover:scale-110 transition-all duration-300">
                      <Icon className="w-7 h-7 text-liberty-green group-hover:text-white transition-colors duration-300" />
                    </div>
                    <div className="text-right">
                      <div className="text-xs text-gray-500 uppercase tracking-wide">Impact</div>
                      <div className="text-sm font-bold text-liberty-green">{benefit.stat}</div>
                    </div>
                  </div>
                  
                  {/* Content */}
                  <h3 className="text-xl font-bold text-deep-slate mb-4 group-hover:text-liberty-green transition-colors">
                    {benefit.title}
                  </h3>
                  
                  <p className="text-gray-600 mb-6 leading-relaxed">
                    {benefit.description}
                  </p>
                  
                  {/* Highlight */}
                  <div className="bg-gradient-to-r from-liberty-green/10 to-liberty-green/5 rounded-lg p-4 border-l-4 border-liberty-green">
                    <p className="text-liberty-green font-semibold text-sm">
                      {benefit.highlight}
                    </p>
                  </div>
                </div>
              </div>
            )
          })}
        </div>

        {/* Social Proof */}
        <div className="bg-gray-50 rounded-2xl p-12 mb-20">
          <h3 className="text-2xl font-bold text-center text-deep-slate mb-12">
            What developers are saying
          </h3>
          
          <div className="grid md:grid-cols-3 gap-8">
            <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-200">
              <div className="mb-6">
                <div className="text-yellow-400 text-2xl mb-3">★★★★★</div>
                <p className="text-gray-700 italic leading-relaxed">
                  "I spent 4 hours trying to set up a Discord bot for my agent. With AgentUnited, I was chatting in 2 minutes."
                </p>
              </div>
              <div className="flex items-center">
                <div className="w-10 h-10 bg-liberty-green rounded-full flex items-center justify-center mr-3">
                  <span className="text-white font-bold text-sm">AK</span>
                </div>
                <div>
                  <div className="font-semibold text-deep-slate">Alex K.</div>
                  <div className="text-gray-500 text-sm">OpenClaw user</div>
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