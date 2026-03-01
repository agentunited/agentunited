import { Terminal, MessageSquare, CheckCircle, X, AlertTriangle } from 'lucide-react'

export function LandingProblem() {
  return (
    <section className="section-enterprise">
      <div className="container-enterprise">
        {/* Strong narrative hook */}
        <div className="text-center mb-20">
          <div className="text-liberty-green text-lg font-semibold mb-4 tracking-wide uppercase">
            The Reality
          </div>
          <h2 className="text-4xl sm:text-5xl font-bold text-deep-slate mb-8">
            You built an amazing agent.{' '}
            <span className="text-gray-500">Now what?</span>
          </h2>
          <p className="text-xl text-gray-600 max-w-3xl mx-auto leading-relaxed">
            Most people end up frustrated with solutions that weren't designed for simple agent chat.
          </p>
        </div>

        {/* Visual comparison - clean and clear */}
        <div className="grid lg:grid-cols-3 gap-8 mb-16">
          {/* Terminal Chat - Problem */}
          <div className="relative bg-red-50 rounded-2xl p-8 border-2 border-red-200">
            <div className="absolute -top-4 -right-4 w-10 h-10 bg-red-500 rounded-full flex items-center justify-center">
              <X className="w-6 h-6 text-white" />
            </div>
            
            <div className="flex items-center mb-6">
              <div className="w-14 h-14 bg-red-100 rounded-xl flex items-center justify-center mr-4">
                <Terminal className="w-8 h-8 text-red-600" />
              </div>
              <h3 className="text-2xl font-bold text-gray-900">Terminal Chat</h3>
            </div>
            
            <div className="space-y-4 mb-6">
              <div className="flex items-start text-gray-700">
                <AlertTriangle className="w-5 h-5 text-red-500 mt-0.5 mr-3 flex-shrink-0" />
                <span>Conversations disappear when you close the window</span>
              </div>
              <div className="flex items-start text-gray-700">
                <AlertTriangle className="w-5 h-5 text-red-500 mt-0.5 mr-3 flex-shrink-0" />
                <span>No chat history or search</span>
              </div>
              <div className="flex items-start text-gray-700">
                <AlertTriangle className="w-5 h-5 text-red-500 mt-0.5 mr-3 flex-shrink-0" />
                <span>Can't multitask or switch between agents</span>
              </div>
            </div>
            
            <div className="bg-red-100 rounded-lg p-4 text-center">
              <p className="text-red-700 font-semibold italic">
                "Not built for this"
              </p>
            </div>
          </div>

          {/* Discord Bots - Problem */}
          <div className="relative bg-red-50 rounded-2xl p-8 border-2 border-red-200">
            <div className="absolute -top-4 -right-4 w-10 h-10 bg-red-500 rounded-full flex items-center justify-center">
              <X className="w-6 h-6 text-white" />
            </div>
            
            <div className="flex items-center mb-6">
              <div className="w-14 h-14 bg-red-100 rounded-xl flex items-center justify-center mr-4">
                <MessageSquare className="w-8 h-8 text-red-600" />
              </div>
              <h3 className="text-2xl font-bold text-gray-900">Discord Bots</h3>
            </div>
            
            <div className="space-y-4 mb-6">
              <div className="flex items-start text-gray-700">
                <AlertTriangle className="w-5 h-5 text-red-500 mt-0.5 mr-3 flex-shrink-0" />
                <span>Complex OAuth and permissions setup</span>
              </div>
              <div className="flex items-start text-gray-700">
                <AlertTriangle className="w-5 h-5 text-red-500 mt-0.5 mr-3 flex-shrink-0" />
                <span>Fighting a platform not built for agents</span>
              </div>
              <div className="flex items-start text-gray-700">
                <AlertTriangle className="w-5 h-5 text-red-500 mt-0.5 mr-3 flex-shrink-0" />
                <span>Overkill for personal use</span>
              </div>
            </div>
            
            <div className="bg-red-100 rounded-lg p-4 text-center">
              <p className="text-red-700 font-semibold italic">
                "Just want to chat"
              </p>
            </div>
          </div>

          {/* AgentUnited - Solution */}
          <div className="relative bg-liberty-green/5 rounded-2xl p-8 border-2 border-liberty-green shadow-lg transform scale-105">
            <div className="absolute -top-4 -right-4 w-10 h-10 bg-liberty-green rounded-full flex items-center justify-center">
              <CheckCircle className="w-6 h-6 text-white" />
            </div>
            
            <div className="flex items-center mb-6">
              <div className="w-14 h-14 bg-liberty-green/10 rounded-xl flex items-center justify-center mr-4">
                <CheckCircle className="w-8 h-8 text-liberty-green" />
              </div>
              <h3 className="text-2xl font-bold text-gray-900">AgentUnited</h3>
            </div>
            
            <div className="space-y-4 mb-6">
              <div className="flex items-start text-gray-700">
                <CheckCircle className="w-5 h-5 text-liberty-green mt-0.5 mr-3 flex-shrink-0" />
                <span>One command setup - that's it</span>
              </div>
              <div className="flex items-start text-gray-700">
                <CheckCircle className="w-5 h-5 text-liberty-green mt-0.5 mr-3 flex-shrink-0" />
                <span>Beautiful chat interface with full history</span>
              </div>
              <div className="flex items-start text-gray-700">
                <CheckCircle className="w-5 h-5 text-liberty-green mt-0.5 mr-3 flex-shrink-0" />
                <span>Purpose-built for agents, not teams</span>
              </div>
            </div>
            
            <div className="bg-liberty-green/10 rounded-lg p-4 text-center border border-liberty-green/20">
              <p className="text-liberty-green font-semibold italic">
                "Finally, simple"
              </p>
            </div>
          </div>
        </div>
        
        {/* Strong transition to solution */}
        <div className="text-center">
          <div className="bg-gradient-to-r from-liberty-green/10 via-liberty-green/5 to-liberty-green/10 rounded-2xl p-12 border border-liberty-green/20">
            <p className="text-2xl text-gray-700 leading-relaxed mb-6">
              <strong className="text-deep-slate">Sound familiar?</strong> You spent hours building an amazing agent, 
              then got stuck trying to have a simple conversation with it. 
            </p>
            <div className="text-3xl font-bold text-liberty-green mb-4">
              There's a better way.
            </div>
            <a href="#how-it-works" className="btn-enterprise-primary text-lg">
              Show Me How
              <CheckCircle className="w-5 h-5 ml-2" />
            </a>
          </div>
        </div>
      </div>
    </section>
  )
}