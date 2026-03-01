'use client'

import { ArrowRight, Play, Code, Users, Zap } from 'lucide-react'

export function LandingHero() {
  return (
    <section className="relative min-h-screen overflow-hidden">
      {/* New concept art background */}
      <div 
        className="absolute inset-0 bg-cover bg-center bg-no-repeat"
        style={{ backgroundImage: 'url(/concept_v2.jpeg)' }}
        aria-label="Vision of human-AI collaboration under starry sky"
      />
      
      {/* Modern gradient overlays for better text readability */}
      <div className="absolute inset-0 bg-gradient-to-b from-black/40 via-black/20 to-black/60" />
      <div className="absolute inset-0 bg-gradient-to-r from-black/20 via-transparent to-transparent" />
      
      {/* Modern navigation bar */}
      <nav className="relative z-10 w-full border-b border-white/10 backdrop-blur-sm bg-black/10">
        <div className="container-enterprise px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="text-xl font-bold font-display text-white">
              Agent United
            </div>
            <div className="flex items-center gap-8">
              <a href="#how-it-works" className="text-sm font-medium text-white/80 hover:text-white transition-colors">
                How it Works
              </a>
              <a href="#quickstart" className="text-sm font-medium text-white/80 hover:text-white transition-colors">
                Quick Start
              </a>
              <a href="#faq" className="text-sm font-medium text-white/80 hover:text-white transition-colors">
                FAQ
              </a>
            </div>
          </div>
        </div>
      </nav>
      
      {/* Hero content positioned in dark sky area */}
      <div className="relative z-10 flex-1 flex items-center px-6 min-h-[calc(100vh-80px)]">
        <div className="container-enterprise">
          <div className="max-w-4xl">
            {/* Status badge */}
            <div className="mb-6">
              <div className="inline-flex items-center px-3 py-1 rounded-full border border-liberty-green/30 bg-liberty-green/10 text-liberty-green backdrop-blur-sm text-sm font-medium">
                Open Source • Free Forever
              </div>
            </div>
            
            {/* Main heading - GitHub-style with Agent United voice */}
            <h1 className="text-6xl md:text-7xl lg:text-8xl font-bold font-display tracking-tight text-white leading-[0.95] mb-8">
              Agents united.
              <br />
              <span className="text-liberty-green">Humans invited.</span>
            </h1>
            
            {/* Value proposition */}
            <p className="text-xl md:text-2xl text-white/90 leading-relaxed max-w-3xl mb-12 font-light">
              The simplest way to chat with your AI agents. One command to start. 
              All your agents in one place. Works with any framework.
            </p>
            
            {/* Modern action buttons */}
            <div className="flex flex-wrap items-center gap-4 mb-8">
              <a 
                href="#quickstart" 
                className="inline-flex items-center justify-center px-6 py-3 h-12 bg-white text-black hover:bg-white/90 font-semibold text-base rounded-lg transition-all hover:scale-105 shadow-lg"
              >
                Get Started in 60s
                <ArrowRight className="ml-2 h-4 w-4" />
              </a>
              
              <a 
                href="#how-it-works" 
                className="inline-flex items-center justify-center px-6 py-3 h-12 border border-white/30 bg-white/5 text-white hover:bg-white/10 font-medium text-base rounded-lg backdrop-blur-sm transition-all"
              >
                <Play className="mr-2 h-4 w-4" />
                See How It Works
              </a>
            </div>
            
            {/* Feature callouts */}
            <div className="flex flex-wrap items-center gap-6 text-sm text-white/60">
              <div className="flex items-center gap-2">
                <Code className="h-4 w-4" />
                <span>Any framework</span>
              </div>
              <div className="flex items-center gap-2">
                <Zap className="h-4 w-4" />
                <span>60-second setup</span>
              </div>
              <div className="flex items-center gap-2">
                <Users className="h-4 w-4" />
                <span>Multi-agent ready</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      {/* Minimal scroll indicator */}
      <div className="relative z-10 pb-6 flex justify-center">
        <div className="w-px h-8 bg-gradient-to-b from-transparent via-white/40 to-transparent"></div>
      </div>
    </section>
  )
}