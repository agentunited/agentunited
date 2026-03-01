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
      setIsScrolled(scrollTop > 50)
    }

    window.addEventListener('scroll', handleScroll, { passive: true })
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  return (
    <nav className={`nav-enterprise ${isScrolled ? 'nav-enterprise-scrolled' : ''}`}>
      <div className="container-enterprise">
        <div className="flex items-center justify-between h-20">
          {/* Logo */}
          <Link href="/" className="flex items-center space-x-3">
            <div className="w-10 h-10 bg-liberty-green rounded-lg flex items-center justify-center">
              {/* Statue of Liberty torch icon */}
              <svg className="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 2l3 7h7l-5.5 4 2 7L12 16l-6.5 4 2-7L2 9h7l3-7z"/>
              </svg>
            </div>
            <span className="text-2xl font-bold text-white tracking-tight">
              AgentUnited
            </span>
          </Link>

          {/* Desktop Navigation Links */}
          <div className="hidden md:flex items-center space-x-8">
            <a 
              href="#how-it-works" 
              className="text-white/80 hover:text-white font-medium transition-colors"
            >
              How It Works
            </a>
            <a 
              href="#quickstart" 
              className="text-white/80 hover:text-white font-medium transition-colors"
            >
              Quickstart
            </a>
            <Link 
              href="#" 
              className="text-white/80 hover:text-white font-medium transition-colors"
            >
              Documentation
            </Link>
            <Link 
              href="#" 
              className="text-white/80 hover:text-white font-medium transition-colors"
            >
              Examples
            </Link>
          </div>

          {/* Desktop CTA Buttons */}
          <div className="hidden md:flex items-center space-x-4">
            <Link 
              href="https://github.com/superpose/agentunited"
              className="flex items-center space-x-2 px-4 py-2 text-white/80 hover:text-white border border-white/20 hover:border-white/40 rounded-lg transition-all"
            >
              <Github className="w-4 h-4" />
              <span>GitHub</span>
            </Link>
            <a href="#quickstart" className="btn-enterprise-primary">
              Get Started
            </a>
          </div>

          {/* Mobile Menu Button */}
          <button
            className="md:hidden p-2 text-white hover:text-white/80 transition-colors"
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
          >
            {isMobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
          </button>
        </div>

        {/* Mobile Menu */}
        {isMobileMenuOpen && (
          <div className="md:hidden border-t border-white/20">
            <div className="py-4 space-y-4">
              <a
                href="#how-it-works"
                className="block text-white/80 hover:text-white font-medium transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                How It Works
              </a>
              <a
                href="#quickstart"
                className="block text-white/80 hover:text-white font-medium transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                Quickstart
              </a>
              <Link
                href="#"
                className="block text-white/80 hover:text-white font-medium transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                Documentation
              </Link>
              <Link
                href="#"
                className="block text-white/80 hover:text-white font-medium transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                Examples
              </Link>
              
              <div className="pt-4 space-y-3 border-t border-white/20">
                <Link 
                  href="https://github.com/superpose/agentunited"
                  className="flex items-center justify-center space-x-2 w-full px-4 py-2 text-white/80 hover:text-white border border-white/20 hover:border-white/40 rounded-lg transition-all"
                >
                  <Github className="w-4 h-4" />
                  <span>GitHub</span>
                </Link>
                <a href="#quickstart" className="btn-enterprise-primary w-full block text-center">
                  Get Started
                </a>
              </div>
            </div>
          </div>
        )}
      </div>
    </nav>
  )
}