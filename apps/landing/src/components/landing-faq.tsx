'use client'

import { useState } from 'react'
import { ChevronDown, ChevronUp, CheckCircle } from 'lucide-react'

export function LandingFAQ() {
  const [openIndex, setOpenIndex] = useState<number | null>(0)
  
  const faqs = [
    {
      question: "Do I need to change my agent code?",
      answer: "Nope! Just add 3 lines to connect. Your agent keeps doing exactly what it does. AgentUnited works alongside your existing code without any modifications to your agent's logic.",
      category: "setup"
    },
    {
      question: "What if I'm using [OpenClaw/AutoGPT/CrewAI/Custom]?",
      answer: "It works with all of them! If your agent can make HTTP calls (which all modern agent frameworks can), it works with AgentUnited. We have specific integration guides for popular frameworks.",
      category: "compatibility"
    },
    {
      question: "Is this for teams or just individuals?",
      answer: "Built for individuals first. You can invite others to see your agent conversations, but AgentUnited is designed for personal use - you and your agents. Team features aren't the focus.",
      category: "usage"
    },
    {
      question: "Do I need to host this in the cloud?",
      answer: "No! Runs completely on your laptop or desktop. Everything is local - your conversations, your data, your agents. No cloud required, no accounts to create, no monthly fees.",
      category: "hosting"
    },
    {
      question: "Is it really free?",
      answer: "Yes, completely free. Open source (Apache 2.0 license). Download the code, run it yourself, modify it if you want. No hidden costs, no premium tiers, no 'freemium' model.",
      category: "pricing"
    },
    {
      question: "What if I have 10 different agents?",
      answer: "Perfect! They all show up in the sidebar. Switch between conversations instantly, search across all your chat history, and see all your agents in one place. No limit on agents.",
      category: "scale"
    },
    {
      question: "How is this different from Discord bots?",
      answer: "Discord is built for team chat - bots are an afterthought with complex setup. AgentUnited is built specifically for agent chat. Setup is 3 lines of code, not hours of OAuth configuration.",
      category: "comparison"
    },
    {
      question: "What about my conversation history?",
      answer: "Everything is saved automatically on your machine. Search past messages, agents remember context across sessions, conversations never disappear. It's like having a chat app for your agents.",
      category: "data"
    }
  ]

  return (
    <section className="section-enterprise">
      <div className="container-enterprise">
        {/* Section Header */}
        <div className="text-center mb-20">
          <div className="text-liberty-green text-lg font-semibold mb-4 tracking-wide uppercase">
            Questions & Answers
          </div>
          <h2 className="text-4xl sm:text-5xl font-bold text-deep-slate mb-8">
            Everything you need to know
          </h2>
          <p className="text-xl text-gray-600 max-w-3xl mx-auto leading-relaxed">
            Honest answers to the questions everyone asks. No marketing fluff.
          </p>
        </div>

        {/* FAQ List */}
        <div className="max-w-4xl mx-auto mb-16">
          <div className="space-y-4">
            {faqs.map((faq, index) => (
              <div 
                key={index}
                className="bg-white rounded-xl border-2 border-gray-100 shadow-sm hover:shadow-md transition-all duration-200"
              >
                <button
                  className="w-full px-8 py-6 text-left flex items-center justify-between hover:bg-gray-50 transition-colors rounded-xl"
                  onClick={() => setOpenIndex(openIndex === index ? null : index)}
                >
                  <h3 className="font-bold text-deep-slate text-lg pr-4">
                    {faq.question}
                  </h3>
                  <div className="flex-shrink-0">
                    {openIndex === index ? (
                      <ChevronUp className="w-6 h-6 text-liberty-green" />
                    ) : (
                      <ChevronDown className="w-6 h-6 text-gray-400" />
                    )}
                  </div>
                </button>
                
                {openIndex === index && (
                  <div className="px-8 pb-6">
                    <div className="pt-4 border-t border-gray-100">
                      <p className="text-gray-700 leading-relaxed text-lg">
                        {faq.answer}
                      </p>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Trust Signals */}
        <div className="grid md:grid-cols-3 gap-8 mb-16">
          <div className="text-center p-6">
            <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <CheckCircle className="w-8 h-8 text-green-600" />
            </div>
            <h4 className="font-bold text-deep-slate mb-2">Open Source</h4>
            <p className="text-gray-600 text-sm">Apache 2.0 license. See exactly how it works on GitHub.</p>
          </div>
          
          <div className="text-center p-6">
            <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <CheckCircle className="w-8 h-8 text-blue-600" />
            </div>
            <h4 className="font-bold text-deep-slate mb-2">No Registration</h4>
            <p className="text-gray-600 text-sm">Download, run, start chatting. No accounts or personal info required.</p>
          </div>
          
          <div className="text-center p-6">
            <div className="w-16 h-16 bg-purple-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <CheckCircle className="w-8 h-8 text-purple-600" />
            </div>
            <h4 className="font-bold text-deep-slate mb-2">Your Data Stays Yours</h4>
            <p className="text-gray-600 text-sm">Everything runs locally. No cloud, no data collection, no tracking.</p>
          </div>
        </div>

        {/* Final CTA */}
        <div className="text-center">
          <div className="bg-gradient-to-r from-liberty-green/10 via-liberty-green/5 to-liberty-green/10 rounded-2xl p-12 border border-liberty-green/20">
            <h3 className="text-3xl font-bold text-deep-slate mb-4">
              Still have questions?
            </h3>
            <p className="text-xl text-gray-600 mb-8">
              Try it yourself. If it doesn't work for you, just delete the folder. 
              No accounts to cancel, no commitments.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <a href="#quickstart" className="btn-enterprise-primary text-lg">
                Try AgentUnited Now
              </a>
              <a 
                href="https://github.com/superpose/agentunited" 
                className="btn-enterprise-secondary text-lg"
              >
                Browse Source Code
              </a>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}