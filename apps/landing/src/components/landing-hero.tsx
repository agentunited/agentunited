'use client'

import { ArrowRight, Play } from 'lucide-react'

export function LandingHero() {
  return (
    <section className="relative min-h-screen overflow-hidden bg-black">
      {/* Full-screen concept art - no overlays or containers blocking it */}
      <div 
        className="absolute inset-0 bg-cover bg-center bg-no-repeat"
        style={{ backgroundImage: 'url(/concept-art.jpg)' }}
        aria-label="AI agents and humans united under the Statue of Liberty at night"
      />
      
      {/* Subtle gradient only at the bottom for text readability */}
      <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent" />
      
      {/* Navigation spacer */}
      <div className="h-20" />
      
      {/* Content positioned strategically to complement the art */}
      <div className="relative z-10 min-h-[calc(100vh-80px)] flex items-end pb-20">
        <div className="container-enterprise">
          <div className="max-w-4xl">
            {/* 1. Brand tagline - 48px, Rajdhani Bold, Liberty Green */}
            <h1 className="text-5xl font-bold font-display text-liberty-green mb-6 leading-tight">
              Agents united. Humans invited.
            </h1>
            
            {/* 2. Value proposition - 28px, Inter Medium, dark gray */}
            <h2 className="text-3xl font-medium text-gray-200 mb-4 leading-relaxed">
              The simplest way to chat with your AI agents.
            </h2>
            
            {/* 3. Supporting detail - 18px, Inter Regular, muted */}
            <p className="text-lg text-white/70 mb-8 leading-relaxed">
              One command to start. All your agents in one place.
            </p>
            
            {/* CTA buttons */}
            <div className="flex flex-col sm:flex-row gap-4 mb-8">
              <a 
                href="#quickstart" 
                className="inline-flex items-center justify-center px-8 py-4 bg-liberty-green hover:bg-[#6E9589] text-white font-semibold rounded-lg transition-all duration-200 text-lg shadow-lg hover:shadow-xl hover:-translate-y-0.5"
              >
                Get Started in 60 Seconds
                <ArrowRight className="w-5 h-5 ml-2" />
              </a>
              <a 
                href="#how-it-works" 
                className="inline-flex items-center justify-center px-8 py-4 bg-transparent border-2 border-white/80 hover:border-white text-white font-semibold rounded-lg transition-all duration-200 text-lg hover:bg-white/10"
              >
                <Play className="w-5 h-5 mr-2" />
                See How It Works
              </a>
            </div>
            
            {/* Quick proof point */}
            <div className="text-white/60 text-sm">
              ✓ Works with OpenClaw, AutoGPT, CrewAI, any agent
            </div>
          </div>
        </div>
      </div>
      
      {/* Elegant scroll indicator */}
      <div className="absolute bottom-6 left-1/2 transform -translate-x-1/2 z-10 text-center">
        <div className="text-white/50 text-xs mb-2 uppercase tracking-wider">Scroll</div>
        <div className="w-px h-8 bg-white/30 mx-auto animate-pulse" />
      </div>
    </section>
  )
}