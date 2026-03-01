'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Github, Menu, X } from 'lucide-react'

export function LandingNav() {
  const [isScrolled, setIsScrolled] = useState(false)
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)

  useEffect(() => {
    const handleScroll = () => {
      const scrollTop = window.scrollY
      setIsScrolled(scrollTop > 100)
    }

    window.addEventListener('scroll', handleScroll, { passive: true })
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  return (
    <nav className={`nav-transparent ${isScrolled ? 'nav-scrolled' : ''}`}>
      <div className="max-w-7xl mx-auto px-6 sm:px-8">
        <div className="flex items-center justify-between h-20">
          {/* Logo */}
          <Link href="/" className="flex items-center space-x-3 z-10">
            <div className="w-10 h-10 bg-liberty-green/90 backdrop-blur-md rounded-xl flex items-center justify-center border border-white/20 shadow-lg">
              {/* Simplified Statue of Liberty torch icon */}
              <svg className="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 2l3 7h7l-5.5 4 2 7L12 16l-6.5 4 2-7L2 9h7l3-7z"/>
              </svg>
            </div>
            <span className={`text-2xl font-bold tracking-tight transition-colors ${
              isScrolled ? 'text-deep-slate' : 'text-white drop-shadow-lg'
            }`}>
              AgentUnited
            </span>
          </Link>

          {/* Desktop Navigation Links */}
          <div className="hidden md:flex items-center space-x-8">
            <Link 
              href="#features" 
              className={`font-medium transition-colors ${
                isScrolled 
                  ? 'text-deep-slate/80 hover:text-liberty-green' 
                  : 'text-white/90 hover:text-white drop-shadow-md'
              }`}
            >
              Features
            </Link>
            <Link 
              href="#how-it-works" 
              className={`font-medium transition-colors ${
                isScrolled 
                  ? 'text-deep-slate/80 hover:text-liberty-green' 
                  : 'text-white/90 hover:text-white drop-shadow-md'
              }`}
            >
              How It Works
            </Link>
            <Link 
              href="#quickstart" 
              className={`font-medium transition-colors ${
                isScrolled 
                  ? 'text-deep-slate/80 hover:text-liberty-green' 
                  : 'text-white/90 hover:text-white drop-shadow-md'
              }`}
            >
              Quickstart
            </Link>
            <Link 
              href="#docs" 
              className={`font-medium transition-colors ${
                isScrolled 
                  ? 'text-deep-slate/80 hover:text-liberty-green' 
                  : 'text-white/90 hover:text-white drop-shadow-md'
              }`}
            >
              Docs
            </Link>
          </div>

          {/* Desktop CTA Buttons */}
          <div className="hidden md:flex items-center space-x-4">
            {isScrolled ? (
              <>
                <Button variant="outline" size="sm" asChild>
                  <Link href="https://github.com/superpose/agentunited" className="flex items-center space-x-2">
                    <Github className="w-4 h-4" />
                    <span>GitHub</span>
                  </Link>
                </Button>
                <Button size="sm" className="bg-liberty-green hover:bg-[#6E9589] text-white">
                  Get Started
                </Button>
              </>
            ) : (
              <>
                <button className="btn-glass-secondary flex items-center space-x-2">
                  <Github className="w-4 h-4" />
                  <span>GitHub</span>
                </button>
                <button className="btn-glass-primary">
                  Get Started
                </button>
              </>
            )}
          </div>

          {/* Mobile Menu Button */}
          <button
            className={`md:hidden p-2 rounded-lg transition-colors ${
              isScrolled 
                ? 'text-deep-slate hover:bg-liberty-green/10' 
                : 'text-white hover:bg-white/10'
            }`}
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
          >
            {isMobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
          </button>
        </div>

        {/* Mobile Menu */}
        {isMobileMenuOpen && (
          <div className="md:hidden">
            <div className="px-2 pt-2 pb-3 space-y-1 bg-white/95 backdrop-blur-md rounded-lg mt-2 border border-liberty-green/20 shadow-lg">
              <Link
                href="#features"
                className="block px-3 py-2 text-deep-slate hover:text-liberty-green hover:bg-liberty-green/5 rounded-md transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                Features
              </Link>
              <Link
                href="#how-it-works"
                className="block px-3 py-2 text-deep-slate hover:text-liberty-green hover:bg-liberty-green/5 rounded-md transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                How It Works
              </Link>
              <Link
                href="#quickstart"
                className="block px-3 py-2 text-deep-slate hover:text-liberty-green hover:bg-liberty-green/5 rounded-md transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                Quickstart
              </Link>
              <Link
                href="#docs"
                className="block px-3 py-2 text-deep-slate hover:text-liberty-green hover:bg-liberty-green/5 rounded-md transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                Docs
              </Link>
              <div className="flex flex-col space-y-2 pt-4 border-t border-liberty-green/20">
                <Button variant="outline" size="sm" asChild>
                  <Link href="https://github.com/superpose/agentunited" className="flex items-center justify-center space-x-2">
                    <Github className="w-4 h-4" />
                    <span>GitHub</span>
                  </Link>
                </Button>
                <Button size="sm" className="bg-liberty-green hover:bg-[#6E9589] text-white">
                  Get Started
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </nav>
  )
}