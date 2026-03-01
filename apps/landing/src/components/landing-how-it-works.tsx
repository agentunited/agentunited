import { Download, Link2, MessageCircle, ArrowDown, CheckCircle } from 'lucide-react'

export function LandingHowItWorks() {
  return (
    <section id="how-it-works" className="relative py-32 bg-transparent">
      <div className="container mx-auto px-6 lg:px-8">
        {/* Enterprise Section Header */}
        <div className="text-center mb-24 animate-fade-in-up">
          <div className="inline-flex items-center px-4 py-2 rounded-full bg-white/5 border border-white/10 backdrop-blur-sm mb-6">
            <span className="text-sm font-semibold text-white/80 uppercase tracking-wider">How It Works</span>
          </div>
          <h2 className="text-5xl md:text-6xl lg:text-7xl font-bold text-white mb-10 leading-tight tracking-tight">
            Three simple steps.{' '}
            <br className="hidden md:block" />
            <span className="text-white/50">No complex setup.</span>
          </h2>
          <p className="text-xl md:text-2xl text-white/70 max-w-4xl mx-auto leading-relaxed font-light">
            No OAuth. No webhooks. No fighting with platforms. Just simple, beautiful agent chat.
          </p>
        </div>

        {/* Steps Timeline */}
        <div className="max-w-4xl mx-auto">
          {/* Step 1 */}
          <div className="relative mb-16">
            <div className="flex flex-col lg:flex-row items-center gap-12">
              {/* Step Number */}
              <div className="flex-shrink-0 w-32 h-32 bg-liberty-green rounded-full flex items-center justify-center shadow-lg">
                <span className="text-4xl font-bold text-white">1</span>
              </div>
              
              {/* Content */}
              <div className="flex-1">
                <div className="flex items-center mb-6">
                  <Download className="w-10 h-10 text-emerald-400 mr-5" />
                  <h3 className="text-4xl font-bold text-white">Start AgentUnited</h3>
                </div>
                <p className="text-xl text-white/70 mb-8 leading-relaxed">
                  One command. 30 seconds. Done.
                </p>
                
                {/* Code */}
                <div className="bg-gray-900 rounded-xl p-6 font-mono text-sm shadow-lg">
                  <div className="flex items-center mb-4">
                    <div className="flex space-x-2">
                      <div className="w-3 h-3 bg-red-500 rounded-full"></div>
                      <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
                      <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                    </div>
                    <span className="text-gray-400 ml-4">terminal</span>
                  </div>
                  <div className="space-y-2">
                    <div className="text-blue-300">git clone https://github.com/superpose/agentunited</div>
                    <div className="text-blue-300">cd agentunited && docker-compose up</div>
                    <div className="text-liberty-green font-semibold mt-3">✓ AgentUnited running at http://localhost:3000</div>
                  </div>
                </div>
              </div>
            </div>
            
            {/* Arrow */}
            <div className="flex justify-center my-8">
              <ArrowDown className="w-8 h-8 text-liberty-green" />
            </div>
          </div>

          {/* Step 2 */}
          <div className="relative mb-16">
            <div className="flex flex-col lg:flex-row-reverse items-center gap-12">
              {/* Step Number */}
              <div className="flex-shrink-0 w-32 h-32 bg-liberty-green rounded-full flex items-center justify-center shadow-lg">
                <span className="text-4xl font-bold text-white">2</span>
              </div>
              
              {/* Content */}
              <div className="flex-1">
                <div className="flex items-center mb-6">
                  <Link2 className="w-10 h-10 text-emerald-400 mr-5" />
                  <h3 className="text-4xl font-bold text-white">Connect Your Agent</h3>
                </div>
                <p className="text-xl text-white/70 mb-8 leading-relaxed">
                  Add 3 lines to your existing agent. Works with any framework.
                </p>
                
                {/* Code */}
                <div className="bg-gray-900 rounded-xl p-6 font-mono text-sm shadow-lg">
                  <div className="flex items-center mb-4">
                    <div className="flex space-x-2">
                      <div className="w-3 h-3 bg-red-500 rounded-full"></div>
                      <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
                      <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                    </div>
                    <span className="text-gray-400 ml-4">agent.py</span>
                  </div>
                  <div className="space-y-2">
                    <div className="text-green-400"># Your existing agent code stays the same</div>
                    <div className="text-white">class MyAgent:</div>
                    <div className="text-white pl-4">def respond(self, message):</div>
                    <div className="text-white pl-8">return "Hello from my agent!"</div>
                    <div className="text-green-400 mt-4"># Add these 3 lines:</div>
                    <div className="text-blue-300">from agentunited import connect</div>
                    <div className="text-blue-300">au = connect("http://localhost:8080")</div>
                    <div className="text-blue-300">au.listen(MyAgent().respond)</div>
                    <div className="text-liberty-green font-semibold mt-3">✓ Agent connected and ready</div>
                  </div>
                </div>
                
                {/* Framework badges */}
                <div className="flex flex-wrap gap-3 mt-6">
                  <span className="px-4 py-2 bg-liberty-green/10 text-liberty-green rounded-full text-sm font-medium">OpenClaw ✓</span>
                  <span className="px-4 py-2 bg-liberty-green/10 text-liberty-green rounded-full text-sm font-medium">AutoGPT ✓</span>
                  <span className="px-4 py-2 bg-liberty-green/10 text-liberty-green rounded-full text-sm font-medium">CrewAI ✓</span>
                  <span className="px-4 py-2 bg-liberty-green/10 text-liberty-green rounded-full text-sm font-medium">Custom ✓</span>
                </div>
              </div>
            </div>
            
            {/* Arrow */}
            <div className="flex justify-center my-8">
              <ArrowDown className="w-8 h-8 text-liberty-green" />
            </div>
          </div>

          {/* Step 3 */}
          <div className="relative">
            <div className="flex flex-col lg:flex-row items-center gap-12">
              {/* Step Number */}
              <div className="flex-shrink-0 w-32 h-32 bg-liberty-green rounded-full flex items-center justify-center shadow-lg">
                <span className="text-4xl font-bold text-white">3</span>
              </div>
              
              {/* Content */}
              <div className="flex-1">
                <div className="flex items-center mb-6">
                  <MessageCircle className="w-10 h-10 text-emerald-400 mr-5" />
                  <h3 className="text-4xl font-bold text-white">Start Chatting</h3>
                </div>
                <p className="text-xl text-white/70 mb-8 leading-relaxed">
                  Open the app. See your agent in the sidebar. Click and chat.
                </p>
                
                {/* Enterprise Visual representation */}
                <div className="bg-white/5 backdrop-blur-xl rounded-3xl border border-white/20 p-8 shadow-2xl">
                  <div className="grid md:grid-cols-3 gap-8 text-center">
                    <div>
                      <div className="w-16 h-16 bg-emerald-500/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
                        <span className="text-2xl">🖥️</span>
                      </div>
                      <div className="font-semibold text-white mb-2">macOS App</div>
                      <div className="text-white/70 text-sm">Native, fast, beautiful</div>
                    </div>
                    <div>
                      <div className="w-16 h-16 bg-emerald-500/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
                        <span className="text-2xl">🌐</span>
                      </div>
                      <div className="font-semibold text-white mb-2">Web Browser</div>
                      <div className="text-white/70 text-sm">Works anywhere</div>
                    </div>
                    <div>
                      <div className="w-16 h-16 bg-emerald-500/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
                        <span className="text-2xl">💬</span>
                      </div>
                      <div className="font-semibold text-white mb-2">Chat History</div>
                      <div className="text-white/70 text-sm">Never lose context</div>
                    </div>
                  </div>
                  
                  <div className="mt-8 pt-6 border-t border-gray-200 text-center">
                    <CheckCircle className="w-8 h-8 text-liberty-green mx-auto mb-3" />
                    <p className="text-lg font-semibold text-white">
                      That's it! You're chatting with your agent.
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Strong CTA */}
        <div className="text-center mt-20">
          <div className="bg-gradient-to-r from-liberty-green to-verdigris rounded-2xl p-12 text-white">
            <h3 className="text-3xl font-bold mb-4">Ready to try it?</h3>
            <p className="text-xl mb-8 text-white/90">
              Most people are chatting with their agents within 2 minutes.
            </p>
            <a href="#quickstart" className="inline-flex items-center px-8 py-4 bg-white text-liberty-green font-bold rounded-lg text-lg hover:bg-gray-100 transition-colors shadow-lg">
              Start Now - It's Free
              <ArrowDown className="w-5 h-5 ml-2" />
            </a>
          </div>
        </div>
      </div>
    </section>
  )
}