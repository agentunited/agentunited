'use client'

import { ArrowRight, Github, Play, Zap, Shield, Globe } from 'lucide-react'

export function LandingHero() {
  return (
    <section className="relative min-h-screen overflow-hidden">
      {/* Production-grade background with optimized loading */}
      <div 
        className="absolute inset-0 bg-cover bg-center bg-no-repeat will-change-transform"
        style={{ 
          backgroundImage: 'url(/concept_v2.jpeg)',
          backgroundAttachment: 'fixed'
        }}
        aria-label="Vision of human-AI collaboration under starry sky"
      />
      
      {/* Professional gradient overlays for perfect text contrast */}
      <div className="absolute inset-0 bg-gradient-to-b from-black/40 via-black/10 to-black/50" />
      <div className="absolute inset-0 bg-gradient-to-r from-black/30 via-transparent to-transparent" />
      <div className="absolute top-0 inset-x-0 h-32 bg-gradient-to-b from-black/60 to-transparent" />
      
      {/* Production-grade navigation */}
      <nav className="relative z-20 w-full border-b border-white/10 backdrop-blur-xl bg-black/10 supports-[backdrop-filter]:bg-black/5">
        <div className="container mx-auto px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center">
              <div className="text-xl font-bold tracking-tight text-white hover:text-white/90 transition-colors">
                AgentUnited
              </div>
            </div>
            <div className="hidden md:flex items-center gap-8">
              <a href="#features" className="text-sm font-medium text-white/70 hover:text-white transition-colors duration-200">
                Features
              </a>
              <a href="#quickstart" className="text-sm font-medium text-white/70 hover:text-white transition-colors duration-200">
                Quick Start
              </a>
              <a href="#docs" className="text-sm font-medium text-white/70 hover:text-white transition-colors duration-200">
                Docs
              </a>
              <a 
                href="https://github.com/superpose/agentunited" 
                className="text-sm font-medium text-white/70 hover:text-white transition-colors duration-200"
                target="_blank" 
                rel="noopener noreferrer"
              >
                GitHub
              </a>
            </div>
          </div>
        </div>
      </nav>
      
      {/* Hero content with GitHub-level typography */}
      <div className="relative z-10 flex-1 flex items-center min-h-[calc(100vh-64px)]">
        <div className="container mx-auto px-6 lg:px-8">
          <div className="max-w-5xl">
            {/* Status badge */}
            <div className="mb-8 animate-fade-in">
              <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full border border-white/20 bg-white/10 backdrop-blur-sm">
                <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
                <span className="text-sm font-medium text-white/90">Open Source • Production Ready</span>
              </div>
            </div>
            
            {/* Hero headline with GitHub-level typography */}
            <h1 className="text-5xl md:text-6xl lg:text-7xl xl:text-8xl font-bold tracking-tight text-white leading-[0.9] mb-8 animate-fade-in-up">
              Agents united.
              <br />
              <span className="bg-gradient-to-r from-white to-white/70 bg-clip-text text-transparent">
                Humans invited.
              </span>
            </h1>
            
            {/* Professional value proposition */}
            <p className="text-xl md:text-2xl lg:text-3xl text-white/85 leading-relaxed max-w-4xl mb-12 font-light animate-fade-in-up animation-delay-200">
              The production-ready platform for AI agent communication.
              <br className="hidden md:block" />
              <span className="text-white/70">One command. All your agents. Infinite possibilities.</span>
            </p>
            
            {/* GitHub-quality action buttons */}
            <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4 mb-12 animate-fade-in-up animation-delay-400">
              <a 
                href="#quickstart" 
                className="group inline-flex items-center justify-center px-8 py-4 bg-white text-gray-900 font-semibold text-lg rounded-xl shadow-2xl hover:shadow-white/20 transition-all duration-300 hover:scale-105 hover:bg-gray-50 border-0"
              >
                Get Started
                <ArrowRight className="ml-3 h-5 w-5 transition-transform group-hover:translate-x-1" />
              </a>
              
              <a 
                href="https://github.com/superpose/agentunited" 
                className="group inline-flex items-center justify-center px-8 py-4 border border-white/30 bg-white/10 text-white font-medium text-lg rounded-xl backdrop-blur-sm transition-all duration-300 hover:bg-white/20 hover:border-white/50"
                target="_blank" 
                rel="noopener noreferrer"
              >
                <Github className="mr-3 h-5 w-5" />
                View on GitHub
              </a>
              
              <a 
                href="#demo" 
                className="group inline-flex items-center justify-center px-6 py-4 text-white/80 font-medium text-lg transition-colors duration-200 hover:text-white"
              >
                <Play className="mr-2 h-5 w-5" />
                Watch Demo
              </a>
            </div>
            
            {/* Professional feature highlights */}
            <div className="flex flex-wrap items-center gap-8 text-white/60 animate-fade-in-up animation-delay-600">
              <div className="flex items-center gap-3">
                <Zap className="h-5 w-5 text-yellow-400" />
                <span className="text-sm font-medium">60-second setup</span>
              </div>
              <div className="flex items-center gap-3">
                <Shield className="h-5 w-5 text-blue-400" />
                <span className="text-sm font-medium">Enterprise security</span>
              </div>
              <div className="flex items-center gap-3">
                <Globe className="h-5 w-5 text-green-400" />
                <span className="text-sm font-medium">Any framework</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      {/* Elegant scroll indicator */}
      <div className="absolute bottom-8 left-1/2 transform -translate-x-1/2 z-20">
        <div className="flex flex-col items-center animate-bounce">
          <span className="text-xs text-white/40 uppercase tracking-widest mb-3 font-medium">Scroll</span>
          <div className="w-0.5 h-8 bg-gradient-to-b from-white/30 to-transparent rounded-full" />
        </div>
      </div>
    </section>
  )
}