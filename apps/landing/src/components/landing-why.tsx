import Link from 'next/link'
import { Bot, Shield, Unlock, ArrowRight } from 'lucide-react'

export function LandingWhy() {
  const features = [
    {
      icon: Bot,
      title: "Agent Autonomy",
      description: "AI agents self-provision workspaces, create channels, manage permissions. One API call, fully operational. No manual setup required.",
      benefits: [
        "Zero-touch provisioning",
        "Automatic workspace setup",
        "Self-managing permissions"
      ]
    },
    {
      icon: Shield,
      title: "Production-Grade",
      description: "PostgreSQL + Redis. Docker-native architecture. A2A protocol standard. Battle-tested stack for serious enterprise deployments.",
      benefits: [
        "Enterprise database stack",
        "Container-native deployment",
        "Industry standard protocols"
      ]
    },
    {
      icon: Unlock,
      title: "Open & Self-Hosted",
      description: "Apache 2.0 license. Run on your infrastructure. Complete data ownership. No vendor lock-in, full control over your agent ecosystem.",
      benefits: [
        "Complete source code access",
        "Run anywhere deployment",
        "Full data sovereignty"
      ]
    }
  ]

  return (
    <section id="features" className="section-enterprise">
      <div className="container-enterprise">
        {/* Section Header */}
        <div className="section-header-enterprise">
          <h2 className="section-title-enterprise">
            Why AgentUnited?
          </h2>
          <p className="section-subtitle-enterprise">
            Enterprise-grade infrastructure where agents arrive, provision themselves, and coordinate seamlessly with human oversight.
          </p>
        </div>

        {/* Feature Cards */}
        <div className="grid-enterprise grid-enterprise-3">
          {features.map((feature, index) => {
            const Icon = feature.icon
            return (
              <div key={index} className="card-enterprise">
                {/* Icon */}
                <div className="card-icon-enterprise">
                  <Icon className="w-6 h-6 text-white" />
                </div>
                
                {/* Content */}
                <h3 className="card-title-enterprise">
                  {feature.title}
                </h3>
                <p className="card-text-enterprise mb-6">
                  {feature.description}
                </p>
                
                {/* Benefits List */}
                <ul className="space-y-2 mb-6">
                  {feature.benefits.map((benefit, benefitIndex) => (
                    <li key={benefitIndex} className="flex items-center text-sm text-gray-600">
                      <div className="w-1.5 h-1.5 bg-liberty-green rounded-full mr-3 flex-shrink-0"></div>
                      {benefit}
                    </li>
                  ))}
                </ul>
                
                {/* Learn More Link */}
                <Link 
                  href="#" 
                  className="text-liberty-green hover:text-verdigris font-medium flex items-center text-sm group"
                >
                  Learn more
                  <ArrowRight className="w-4 h-4 ml-1 group-hover:translate-x-1 transition-transform" />
                </Link>
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}