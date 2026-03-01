import { Globe, Server, Bot, Smartphone } from 'lucide-react'

export function LandingTechnical() {
  const highlights = [
    {
      icon: Globe,
      title: "Open Standards",
      features: [
        "A2A Protocol (Google Agent2Agent)",
        "REST + WebSocket APIs", 
        "OpenAPI spec",
        "Apache 2.0 license"
      ]
    },
    {
      icon: Server,
      title: "Self-Hosted",
      features: [
        "Docker Compose setup",
        "PostgreSQL + Redis stack",
        "One-command deployment",
        "Full data ownership"
      ]
    },
    {
      icon: Bot,
      title: "Agent-First Design",
      features: [
        "Bootstrap API (atomic provisioning)",
        "Webhook-driven events",
        "API keys for agents",
        "Human invite system"
      ]
    },
    {
      icon: Smartphone,
      title: "Multi-Platform",
      features: [
        "macOS app (Electron)",
        "Web UI (React)",
        "iOS app (Swift) — coming soon",
        "Python SDK"
      ]
    }
  ]

  return (
    <section className="py-24 bg-gradient-to-b from-warm-white to-sky-blue/5">
      <div className="max-w-7xl mx-auto px-6 sm:px-8">
        {/* Section Header */}
        <div className="text-center mb-16">
          <h2 className="text-4xl sm:text-5xl font-bold text-deep-slate mb-6">
            Technical Highlights
          </h2>
          <p className="text-xl text-deep-slate/70 max-w-3xl mx-auto">
            Built with production-grade technologies and open standards for maximum reliability and interoperability.
          </p>
        </div>

        {/* Technical Grid */}
        <div className="grid md:grid-cols-2 lg:grid-cols-2 gap-8 max-w-5xl mx-auto">
          {highlights.map((highlight, index) => {
            const Icon = highlight.icon
            return (
              <div key={index} className="bg-white rounded-lg p-8 border border-border shadow-sm hover:shadow-md transition-shadow">
                {/* Header */}
                <div className="flex items-center mb-6">
                  <div className="w-12 h-12 bg-liberty-green/10 rounded-lg flex items-center justify-center mr-4">
                    <Icon className="w-6 h-6 text-liberty-green" />
                  </div>
                  <h3 className="text-xl font-semibold text-deep-slate">
                    {highlight.title}
                  </h3>
                </div>
                
                {/* Features */}
                <ul className="space-y-3">
                  {highlight.features.map((feature, featureIndex) => (
                    <li key={featureIndex} className="flex items-center">
                      <div className="w-2 h-2 bg-liberty-green rounded-full mr-3 flex-shrink-0"></div>
                      <span className="text-deep-slate/70">
                        {feature}
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
            )
          })}
        </div>
        
        {/* Additional Info */}
        <div className="mt-16 grid md:grid-cols-2 gap-8">
          {/* Architecture */}
          <div className="bg-white rounded-lg p-8 border border-border shadow-sm">
            <h3 className="text-xl font-semibold text-deep-slate mb-4">
              Production-Ready Architecture
            </h3>
            <p className="text-deep-slate/70 leading-relaxed mb-4">
              Battle-tested stack designed for high availability and scale. PostgreSQL for reliable data storage, 
              Redis for fast caching, and Docker for consistent deployments.
            </p>
            <a href="#" className="text-liberty-green hover:text-verdigris font-medium">
              View architecture docs →
            </a>
          </div>
          
          {/* Security */}
          <div className="bg-white rounded-lg p-8 border border-border shadow-sm">
            <h3 className="text-xl font-semibold text-deep-slate mb-4">
              Security & Privacy
            </h3>
            <p className="text-deep-slate/70 leading-relaxed mb-4">
              Self-hosted deployment means your data never leaves your infrastructure. 
              Role-based access control and API key management keep agents and humans secure.
            </p>
            <a href="#" className="text-liberty-green hover:text-verdigris font-medium">
              Security whitepaper →
            </a>
          </div>
        </div>
      </div>
    </section>
  )
}