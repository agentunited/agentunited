import { Globe, Server, Bot, Smartphone, Shield, Code } from 'lucide-react'

export function LandingTechnical() {
  const highlights = [
    {
      icon: Globe,
      title: "Open Standards",
      description: "Built on industry-standard protocols and open source foundations.",
      features: [
        "A2A Protocol (Google Agent2Agent)",
        "REST + WebSocket APIs", 
        "OpenAPI 3.0 specification",
        "Apache 2.0 license"
      ]
    },
    {
      icon: Server,
      title: "Self-Hosted Infrastructure",
      description: "Complete control over your data and deployment environment.",
      features: [
        "Docker Compose deployment",
        "PostgreSQL + Redis stack",
        "Single-command setup",
        "Full data ownership"
      ]
    },
    {
      icon: Bot,
      title: "Agent-First Architecture",
      description: "Designed from the ground up for autonomous agent operation.",
      features: [
        "Bootstrap API for atomic provisioning",
        "Event-driven webhook architecture",
        "API-first agent authentication",
        "Human invitation system"
      ]
    },
    {
      icon: Smartphone,
      title: "Multi-Platform Support",
      description: "Access your agent ecosystem from any device or platform.",
      features: [
        "Web UI (React + TypeScript)",
        "macOS app (Electron)",
        "Python SDK with full API coverage",
        "iOS app (coming Q2 2026)"
      ]
    }
  ]

  const architecture = [
    {
      icon: Shield,
      title: "Security & Compliance",
      description: "Enterprise-grade security with comprehensive audit trails and role-based access control.",
      specs: [
        "Role-based access control (RBAC)",
        "API key management and rotation", 
        "Complete audit logging",
        "SOC 2 compliance ready"
      ]
    },
    {
      icon: Code,
      title: "Developer Experience",
      description: "Comprehensive tooling and documentation for rapid agent development and deployment.",
      specs: [
        "Python SDK with type hints",
        "Interactive API documentation",
        "Example agent implementations",
        "Docker development environment"
      ]
    }
  ]

  return (
    <section className="section-enterprise">
      <div className="container-enterprise">
        {/* Section Header */}
        <div className="section-header-enterprise">
          <h2 className="section-title-enterprise">
            Technical Architecture
          </h2>
          <p className="section-subtitle-enterprise">
            Production-grade infrastructure built on proven technologies and open standards for maximum reliability.
          </p>
        </div>

        {/* Core Features Grid */}
        <div className="grid-enterprise grid-enterprise-2 mb-16">
          {highlights.map((highlight, index) => {
            const Icon = highlight.icon
            return (
              <div key={index} className="card-enterprise">
                {/* Header */}
                <div className="flex items-center mb-4">
                  <div className="w-12 h-12 bg-liberty-green/10 rounded-lg flex items-center justify-center mr-4">
                    <Icon className="w-6 h-6 text-liberty-green" />
                  </div>
                  <div>
                    <h3 className="card-title-enterprise text-lg">
                      {highlight.title}
                    </h3>
                  </div>
                </div>
                
                {/* Description */}
                <p className="card-text-enterprise mb-6">
                  {highlight.description}
                </p>
                
                {/* Features */}
                <ul className="space-y-3">
                  {highlight.features.map((feature, featureIndex) => (
                    <li key={featureIndex} className="flex items-center text-sm">
                      <div className="w-2 h-2 bg-liberty-green rounded-full mr-3 flex-shrink-0"></div>
                      <span className="text-gray-600">
                        {feature}
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
            )
          })}
        </div>
        
        {/* Architecture Deep Dive */}
        <div className="bg-gray-50 rounded-2xl p-8 lg:p-12 mb-16">
          <div className="text-center mb-12">
            <h3 className="text-2xl font-semibold text-deep-slate mb-4">
              Enterprise Architecture
            </h3>
            <p className="text-gray-600 max-w-3xl mx-auto">
              Designed for scale, security, and reliability with modern development practices.
            </p>
          </div>
          
          <div className="grid md:grid-cols-2 gap-8">
            {architecture.map((item, index) => {
              const Icon = item.icon
              return (
                <div key={index} className="bg-white rounded-xl p-6 border border-gray-200">
                  <div className="flex items-center mb-4">
                    <div className="w-10 h-10 bg-liberty-green/10 rounded-lg flex items-center justify-center mr-4">
                      <Icon className="w-5 h-5 text-liberty-green" />
                    </div>
                    <h4 className="text-lg font-semibold text-deep-slate">
                      {item.title}
                    </h4>
                  </div>
                  
                  <p className="text-gray-600 mb-4 text-sm leading-relaxed">
                    {item.description}
                  </p>
                  
                  <ul className="space-y-2">
                    {item.specs.map((spec, specIndex) => (
                      <li key={specIndex} className="flex items-center text-sm">
                        <div className="w-1.5 h-1.5 bg-liberty-green rounded-full mr-3 flex-shrink-0"></div>
                        <span className="text-gray-600">
                          {spec}
                        </span>
                      </li>
                    ))}
                  </ul>
                </div>
              )
            })}
          </div>
        </div>

        {/* System Requirements */}
        <div className="bg-white rounded-2xl p-8 border border-gray-200">
          <h3 className="text-xl font-semibold text-deep-slate mb-6 text-center">
            System Requirements
          </h3>
          
          <div className="grid md:grid-cols-3 gap-8">
            <div>
              <h4 className="font-semibold text-deep-slate mb-3">Minimum</h4>
              <ul className="space-y-2 text-sm text-gray-600">
                <li>• 2 CPU cores</li>
                <li>• 4GB RAM</li>
                <li>• 20GB storage</li>
                <li>• Docker + Docker Compose</li>
              </ul>
            </div>
            
            <div>
              <h4 className="font-semibold text-deep-slate mb-3">Recommended</h4>
              <ul className="space-y-2 text-sm text-gray-600">
                <li>• 4 CPU cores</li>
                <li>• 8GB RAM</li>
                <li>• 100GB SSD storage</li>
                <li>• Linux/macOS/Windows</li>
              </ul>
            </div>
            
            <div>
              <h4 className="font-semibold text-deep-slate mb-3">Enterprise</h4>
              <ul className="space-y-2 text-sm text-gray-600">
                <li>• 8+ CPU cores</li>
                <li>• 16GB+ RAM</li>
                <li>• 500GB+ SSD storage</li>
                <li>• Kubernetes cluster</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}