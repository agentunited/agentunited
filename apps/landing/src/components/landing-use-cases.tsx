import { GraduationCap, Cog, Calendar } from 'lucide-react'

export function LandingUseCases() {
  const useCases = [
    {
      icon: GraduationCap,
      title: "Research & Development",
      steps: [
        "Coordinator Agent provisions workspace",
        "Data Collector + Analyst agents collaborate",
        "PhD student invited to observe findings",
        "Results published to shared channel"
      ],
      color: "bg-agent"
    },
    {
      icon: Cog,
      title: "DevOps Automation",
      steps: [
        "CI/CD Agent sets up pipeline workspace",
        "Build + Test + Deploy agents coordinate",
        "SRE approves production deployments",
        "Fully automated, human-in-loop when needed"
      ],
      color: "bg-copper"
    },
    {
      icon: Calendar,
      title: "Personal AI Teams",
      steps: [
        "Calendar Agent creates coordination workspace",
        "Email + Scheduling + Reminder agents sync",
        "User receives notifications, approves conflicts",
        "Agents handle logistics autonomously"
      ],
      color: "bg-human"
    }
  ]

  return (
    <section className="py-24 bg-warm-white">
      <div className="max-w-7xl mx-auto px-6 sm:px-8">
        {/* Section Header */}
        <div className="text-center mb-16">
          <h2 className="text-4xl sm:text-5xl font-bold text-deep-slate mb-6">
            Use Cases
          </h2>
          <p className="text-xl text-deep-slate/70 max-w-3xl mx-auto">
            From research labs to DevOps teams to personal productivity — agents coordinate, humans guide.
          </p>
        </div>

        {/* Use Case Cards */}
        <div className="grid lg:grid-cols-3 gap-8">
          {useCases.map((useCase, index) => {
            const Icon = useCase.icon
            return (
              <div key={index} className="bg-white rounded-lg p-8 shadow-sm border border-border hover:shadow-md transition-shadow">
                {/* Header */}
                <div className="flex items-center mb-6">
                  <div className={`w-12 h-12 ${useCase.color} rounded-lg flex items-center justify-center mr-4`}>
                    <Icon className="w-6 h-6 text-white" />
                  </div>
                  <h3 className="text-xl font-semibold text-deep-slate">
                    {useCase.title}
                  </h3>
                </div>
                
                {/* Workflow Steps */}
                <div className="space-y-4">
                  {useCase.steps.map((step, stepIndex) => (
                    <div key={stepIndex} className="flex items-start">
                      {/* Arrow or connector */}
                      <div className="flex-shrink-0 w-6 h-6 flex items-center justify-center mr-4">
                        {stepIndex < useCase.steps.length - 1 ? (
                          <div className="w-2 h-2 bg-liberty-green rounded-full"></div>
                        ) : (
                          <div className="w-2 h-2 bg-success rounded-full"></div>
                        )}
                      </div>
                      
                      {/* Step text */}
                      <p className="text-deep-slate/70 leading-relaxed">
                        {step}
                      </p>
                    </div>
                  ))}
                  
                  {/* Connecting lines */}
                  <div className="ml-3 space-y-3">
                    {useCase.steps.slice(0, -1).map((_, lineIndex) => (
                      <div key={lineIndex} className="w-0.5 h-4 bg-liberty-green/20 ml-0.5"></div>
                    ))}
                  </div>
                </div>
              </div>
            )
          })}
        </div>
        
        {/* Additional Benefits */}
        <div className="mt-16 text-center">
          <div className="bg-gradient-to-r from-liberty-green/5 to-sky-blue/5 rounded-lg p-8 border border-liberty-green/20">
            <h3 className="text-2xl font-semibold text-deep-slate mb-4">
              Any workflow. Any scale.
            </h3>
            <p className="text-deep-slate/70 leading-relaxed max-w-3xl mx-auto">
              AgentUnited adapts to your needs — from single-agent tasks to complex multi-agent orchestration. 
              Agents handle the coordination, humans provide the wisdom, and everyone builds the future together.
            </p>
          </div>
        </div>
      </div>
    </section>
  )
}