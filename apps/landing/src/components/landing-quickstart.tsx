'use client'

import { useState } from 'react'
import { Copy, Check, ExternalLink, Book, Code, Puzzle, FileText, PlayCircle } from 'lucide-react'

export function LandingQuickstart() {
  const [copiedStep, setCopiedStep] = useState<number | null>(null)
  
  const handleCopy = async (step: number, text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopiedStep(step)
      setTimeout(() => setCopiedStep(null), 2000)
    } catch (err) {
      console.error('Failed to copy text: ', err)
    }
  }
  
  const step1Code = `git clone https://github.com/superpose/agentunited
cd agentunited && docker-compose up`

  const step2Code = `pip install agentunited

# Your existing agent
class MyAgent:
    def respond(self, message):
        return f"Hello! You said: {message}"

# Add these 3 lines:
from agentunited import connect
au = connect("http://localhost:8080")
au.listen(MyAgent().respond)`

  return (
    <section id="quickstart" className="relative py-32 bg-transparent">
      <div className="container mx-auto px-6 lg:px-8">
        {/* Enterprise Hero Section Header */}
        <div className="text-center mb-20 animate-fade-in-up">
          <div className="inline-flex items-center px-4 py-2 rounded-full bg-white/5 border border-white/10 backdrop-blur-sm mb-6">
            <span className="text-sm font-semibold text-white/80 uppercase tracking-wider">Ready? Let's Go</span>
          </div>
          <h2 className="text-5xl md:text-6xl lg:text-7xl font-bold text-white mb-10 leading-tight tracking-tight">
            Get started in{' '}
            <span className="text-emerald-400">60 seconds</span>
          </h2>
          <p className="text-xl md:text-2xl text-white/70 max-w-4xl mx-auto leading-relaxed font-light mb-10">
            Seriously. Copy these commands, paste them, and you'll be chatting with your agent before you finish reading this section.
          </p>
          
          {/* Quick stats */}
          <div className="flex justify-center gap-8 text-sm text-gray-500">
            <div className="flex items-center">
              <div className="w-2 h-2 bg-green-500 rounded-full mr-2"></div>
              Average setup: 90 seconds
            </div>
            <div className="flex items-center">
              <div className="w-2 h-2 bg-blue-500 rounded-full mr-2"></div>
              No registration required
            </div>
            <div className="flex items-center">
              <div className="w-2 h-2 bg-purple-500 rounded-full mr-2"></div>
              Works on any computer
            </div>
          </div>
        </div>

        {/* Step-by-step with prominent copy buttons */}
        <div className="max-w-5xl mx-auto space-y-12 mb-16">
          
          {/* Step 1 */}
          <div className="bg-white/5 backdrop-blur-xl rounded-3xl p-10 border border-white/20 shadow-2xl hover:shadow-emerald-500/10 transition-all duration-500 hover:scale-[1.02]">
            <div className="flex items-center justify-between mb-8">
              <div className="flex items-center">
                <div className="w-14 h-14 bg-emerald-500 rounded-2xl flex items-center justify-center text-white font-bold text-xl mr-6 shadow-xl shadow-emerald-500/30">1</div>
                <div>
                  <h3 className="text-3xl font-bold text-white mb-2">Start AgentUnited</h3>
                  <p className="text-white/70 text-lg">One command. Works on Mac, Linux, Windows with Docker.</p>
                </div>
              </div>
              <div className="text-right">
                <div className="text-sm text-white/50 font-semibold uppercase tracking-wider">Time needed</div>
                <div className="font-bold text-emerald-400 text-lg">30 seconds</div>
              </div>
            </div>
            
            <div className="bg-gray-900 rounded-xl overflow-hidden">
              <div className="flex items-center justify-between px-6 py-3 bg-gray-800">
                <div className="flex items-center space-x-3">
                  <div className="flex space-x-1.5">
                    <div className="w-3 h-3 bg-red-500 rounded-full"></div>
                    <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
                    <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                  </div>
                  <span className="text-gray-400 text-sm font-mono">terminal</span>
                </div>
                <button
                  onClick={() => handleCopy(1, step1Code)}
                  className="flex items-center space-x-2 px-4 py-2 bg-liberty-green hover:bg-[#6E9589] text-white text-sm rounded-lg transition-colors"
                >
                  {copiedStep === 1 ? (
                    <>
                      <Check className="w-4 h-4" />
                      <span>Copied!</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-4 h-4" />
                      <span>Copy & Paste</span>
                    </>
                  )}
                </button>
              </div>
              <div className="p-6 font-mono text-sm">
                <div className="space-y-2">
                  <div className="text-blue-300">git clone https://github.com/superpose/agentunited</div>
                  <div className="text-blue-300">cd agentunited && docker-compose up</div>
                  <div className="text-gray-500 mt-4"># Wait for this message:</div>
                  <div className="text-liberty-green">✓ AgentUnited ready at http://localhost:3000</div>
                </div>
              </div>
            </div>
          </div>

          {/* Step 2 */}
          <div className="bg-white/5 backdrop-blur-xl rounded-3xl p-10 border border-white/20 shadow-2xl hover:shadow-emerald-500/10 transition-all duration-500 hover:scale-[1.02]">
            <div className="flex items-center justify-between mb-8">
              <div className="flex items-center">
                <div className="w-14 h-14 bg-emerald-500 rounded-2xl flex items-center justify-center text-white font-bold text-xl mr-6 shadow-xl shadow-emerald-500/30">2</div>
                <div>
                  <h3 className="text-3xl font-bold text-white mb-2">Connect Your Agent</h3>
                  <p className="text-white/70 text-lg">Works with OpenClaw, AutoGPT, CrewAI, any Python agent.</p>
                </div>
              </div>
              <div className="text-right">
                <div className="text-sm text-white/50 font-semibold uppercase tracking-wider">Time needed</div>
                <div className="font-bold text-emerald-400 text-lg">30 seconds</div>
              </div>
            </div>
            
            <div className="bg-gray-900 rounded-xl overflow-hidden">
              <div className="flex items-center justify-between px-6 py-3 bg-gray-800">
                <div className="flex items-center space-x-3">
                  <div className="flex space-x-1.5">
                    <div className="w-3 h-3 bg-red-500 rounded-full"></div>
                    <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
                    <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                  </div>
                  <span className="text-gray-400 text-sm font-mono">agent.py</span>
                </div>
                <button
                  onClick={() => handleCopy(2, step2Code)}
                  className="flex items-center space-x-2 px-4 py-2 bg-liberty-green hover:bg-[#6E9589] text-white text-sm rounded-lg transition-colors"
                >
                  {copiedStep === 2 ? (
                    <>
                      <Check className="w-4 h-4" />
                      <span>Copied!</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-4 h-4" />
                      <span>Copy & Paste</span>
                    </>
                  )}
                </button>
              </div>
              <div className="p-6 font-mono text-sm">
                <div className="space-y-1">
                  <div className="text-blue-300">pip install agentunited</div>
                  <div className="text-gray-500 mt-3"># Your existing agent</div>
                  <div className="text-white">class MyAgent:</div>
                  <div className="text-white pl-4">def respond(self, message):</div>
                  <div className="text-white pl-8">return f"Hello! You said: {'{message}'}"</div>
                  <div className="text-gray-500 mt-3"># Add these 3 lines:</div>
                  <div className="text-yellow-300">from agentunited import connect</div>
                  <div className="text-yellow-300">au = connect("http://localhost:8080")</div>
                  <div className="text-yellow-300">au.listen(MyAgent().respond)</div>
                  <div className="text-liberty-green mt-3">✓ Agent connected!</div>
                </div>
              </div>
            </div>
            
            {/* Framework support */}
            <div className="mt-6 flex flex-wrap gap-3">
              <span className="px-3 py-1 bg-liberty-green/10 text-liberty-green rounded-full text-xs font-medium">✓ OpenClaw</span>
              <span className="px-3 py-1 bg-liberty-green/10 text-liberty-green rounded-full text-xs font-medium">✓ AutoGPT</span>
              <span className="px-3 py-1 bg-liberty-green/10 text-liberty-green rounded-full text-xs font-medium">✓ CrewAI</span>
              <span className="px-3 py-1 bg-liberty-green/10 text-liberty-green rounded-full text-xs font-medium">✓ Custom</span>
            </div>
          </div>

          {/* Step 3 */}
          <div className="bg-gradient-to-r from-liberty-green/10 to-liberty-green/5 rounded-2xl p-8 border-2 border-liberty-green/20">
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center">
                <div className="w-12 h-12 bg-liberty-green rounded-full flex items-center justify-center text-white font-bold text-xl mr-4">3</div>
                <div>
                  <h3 className="text-2xl font-bold text-deep-slate">Start Chatting!</h3>
                  <p className="text-gray-600">Open the app and see your agent waiting for you.</p>
                </div>
              </div>
              <div className="text-right">
                <div className="text-sm text-gray-500">Time needed</div>
                <div className="font-bold text-liberty-green">0 seconds</div>
              </div>
            </div>
            
            <div className="grid md:grid-cols-3 gap-6">
              <div className="bg-white rounded-xl p-6 text-center shadow-sm">
                <div className="text-3xl mb-3">🖥️</div>
                <div className="font-semibold text-gray-900 mb-2">macOS App</div>
                <div className="text-gray-600 text-sm mb-3">Download and open</div>
                <a href="#" className="text-liberty-green text-sm font-medium">Get macOS App →</a>
              </div>
              <div className="bg-white rounded-xl p-6 text-center shadow-sm">
                <div className="text-3xl mb-3">🌐</div>
                <div className="font-semibold text-gray-900 mb-2">Web Browser</div>
                <div className="text-gray-600 text-sm mb-3">http://localhost:3000</div>
                <a href="http://localhost:3000" className="text-liberty-green text-sm font-medium">Open in Browser →</a>
              </div>
              <div className="bg-white rounded-xl p-6 text-center shadow-sm">
                <div className="text-3xl mb-3">💬</div>
                <div className="font-semibold text-gray-900 mb-2">Start Chatting</div>
                <div className="text-gray-600 text-sm mb-3">Click your agent, type hi!</div>
                <div className="text-liberty-green text-sm font-medium">That's it! ✨</div>
              </div>
            </div>
          </div>
        </div>

        {/* Integration Resources */}
        <div className="bg-white rounded-2xl p-12 border border-gray-200 mb-16">
          <h3 className="text-2xl font-bold text-center text-deep-slate mb-8">
            Need help with your specific framework?
          </h3>
          
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
            {[
              { title: "OpenClaw Guide", icon: Code, desc: "Step-by-step integration" },
              { title: "AutoGPT Guide", icon: FileText, desc: "Connect AutoGPT agents" },
              { title: "CrewAI Guide", icon: Puzzle, desc: "Multi-agent setup" },
              { title: "Documentation", icon: Book, desc: "Full API reference" }
            ].map((resource, index) => {
              const Icon = resource.icon
              return (
                <a key={index} href="#" className="group text-center p-6 rounded-xl border border-gray-200 hover:border-liberty-green hover:shadow-md transition-all">
                  <Icon className="w-8 h-8 text-liberty-green mx-auto mb-3 group-hover:scale-110 transition-transform" />
                  <div className="font-semibold text-gray-900 mb-2">{resource.title}</div>
                  <div className="text-sm text-gray-600 mb-3">{resource.desc}</div>
                  <div className="text-liberty-green text-sm group-hover:text-verdigris">Read guide →</div>
                </a>
              )
            })}
          </div>
        </div>
        
        {/* Success state */}
        <div className="text-center">
          <div className="max-w-3xl mx-auto bg-gradient-to-r from-green-50 to-liberty-green/5 rounded-2xl p-12 border border-green-200">
            <div className="w-20 h-20 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-6">
              <Check className="w-10 h-10 text-green-600" />
            </div>
            <h3 className="text-2xl font-bold text-deep-slate mb-4">
              Congratulations! 🎉
            </h3>
            <p className="text-lg text-gray-600 mb-6 leading-relaxed">
              You now have a beautiful chat interface for all your AI agents. 
              Your agents have a proper home, and you have a simple way to talk to them.
            </p>
            <div className="flex items-center justify-center text-sm text-gray-500">
              <PlayCircle className="w-4 h-4 mr-2" />
              Join 1,000+ developers who've made the switch from terminals and Discord
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}