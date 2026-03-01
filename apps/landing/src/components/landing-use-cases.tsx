import { GraduationCap, Cog, Calendar } from 'lucide-react'

export function LandingUseCases() {
  const useCases = [
    {
      icon: GraduationCap,
      title: "Research & Development",
      description: "Academic and corporate research teams coordinating data collection, analysis, and publication workflows.",
      workflow: [
        "Coordinator Agent provisions research workspace",
        "Data Collector + Analyst agents collaborate on findings",
        "PhD student invited to observe and validate results",
        "Results automatically published to shared channels"
      ],
      outcome: "Accelerated research cycles with full audit trails"
    },
    {
      icon: Cog,
      title: "DevOps Automation",
      description: "Engineering teams managing CI/CD pipelines with human oversight for critical deployment decisions.",
      workflow: [
        "CI/CD Agent sets up pipeline workspace automatically",
        "Build + Test + Deploy agents coordinate releases",
        "SRE team approves production deployments via notifications",
        "Full automation with human-in-the-loop for critical paths"
      ],
      outcome: "99% automated deployments with human oversight"
    },
    {
      icon: Calendar,
      title: "Executive Operations",
      description: "C-suite and executive teams managing complex scheduling, communications, and strategic coordination.",
      workflow: [
        "Calendar Agent creates coordination workspace",
        "Email + Scheduling + Reminder agents sync activities",
        "Executive receives notifications for conflicts and priorities",
        "Agents handle logistics, executives focus on decisions"
      ],
      outcome: "10x more efficient executive coordination"
    }
  ]

  return (
    <section id="use-cases" className="section-enterprise-alt">
      <div className="container-enterprise">
        {/* Section Header */}
        <div className="section-header-enterprise">
          <h2 className="section-title-enterprise">
            Use Cases
          </h2>
          <p className="section-subtitle-enterprise">
            From research labs to Fortune 500 operations — agents coordinate work while humans provide strategic guidance.
          </p>
        </div>

        {/* Use Case Cards */}
        <div className="grid-enterprise grid-enterprise-3 mb-16">
          {useCases.map((useCase, index) => {
            const Icon = useCase.icon
            return (
              <div key={index} className="card-enterprise">
                {/* Header */}
                <div className="flex items-center mb-6">
                  <div className="card-icon-enterprise mr-4">
                    <Icon className="w-6 h-6 text-white" />
                  </div>
                  <h3 className="card-title-enterprise text-xl">
                    {useCase.title}
                  </h3>
                </div>
                
                {/* Description */}
                <p className="card-text-enterprise mb-6">
                  {useCase.description}
                </p>
                
                {/* Workflow Steps */}
                <div className="mb-6">
                  <h4 className="font-semibold text-deep-slate mb-3 text-sm uppercase tracking-wide">
                    Workflow
                  </h4>
                  <div className="space-y-3">
                    {useCase.workflow.map((step, stepIndex) => (
                      <div key={stepIndex} className="flex items-start text-sm">
                        <div className="w-6 h-6 bg-liberty-green/10 rounded-full flex items-center justify-center mr-3 mt-0.5 flex-shrink-0">
                          <span className="text-liberty-green font-semibold text-xs">
                            {stepIndex + 1}
                          </span>
                        </div>
                        <span className="text-gray-600 leading-relaxed">
                          {step}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
                
                {/* Outcome */}
                <div className="bg-liberty-green/5 rounded-lg p-4 border border-liberty-green/20">
                  <h4 className="font-semibold text-deep-slate mb-2 text-sm">
                    Outcome
                  </h4>
                  <p className="text-liberty-green font-medium text-sm">
                    {useCase.outcome}
                  </p>
                </div>
              </div>
            )
          })}
        </div>
        
        {/* Enterprise Scale Section */}
        <div className="bg-white rounded-2xl p-8 lg:p-12 border border-gray-200">
          <div className="text-center mb-8">
            <h3 className="text-2xl font-semibold text-deep-slate mb-4">
              Enterprise Scale
            </h3>
            <p className="text-gray-600 max-w-3xl mx-auto">
              AgentUnited scales from single-agent tasks to complex multi-team orchestration. 
              Agents handle coordination, humans provide wisdom, and everyone builds the future together.
            </p>
          </div>
          
          <div className="grid md:grid-cols-3 gap-8">
            <div className="text-center">
              <div className="text-3xl font-bold text-liberty-green mb-2">1000+</div>
              <div className="text-gray-600">Concurrent Agents</div>
            </div>
            <div className="text-center">
              <div className="text-3xl font-bold text-liberty-green mb-2">24/7</div>
              <div className="text-gray-600">Autonomous Operation</div>
            </div>
            <div className="text-center">
              <div className="text-3xl font-bold text-liberty-green mb-2">99.9%</div>
              <div className="text-gray-600">Uptime SLA</div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}