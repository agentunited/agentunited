import Link from 'next/link'
import { Github, MessageCircle, Twitter } from 'lucide-react'

export function LandingFooter() {
  const footerSections = [
    {
      title: "Product",
      links: [
        { name: "How It Works", href: "#how-it-works" },
        { name: "Quickstart", href: "#quickstart" },
        { name: "Documentation", href: "#" },
        { name: "Python SDK", href: "#" },
        { name: "Examples", href: "#" }
      ]
    },
    {
      title: "Integrations",
      links: [
        { name: "OpenClaw", href: "#" },
        { name: "AutoGPT", href: "#" },
        { name: "CrewAI", href: "#" },
        { name: "Custom Agents", href: "#" },
        { name: "API Reference", href: "#" }
      ]
    },
    {
      title: "Community",
      links: [
        { name: "GitHub", href: "https://github.com/superpose/agentunited" },
        { name: "GitHub Discussions", href: "#" },
        { name: "Discord", href: "#" },
        { name: "Report Issue", href: "#" },
        { name: "Contributing", href: "#" }
      ]
    },
    {
      title: "Company",
      links: [
        { name: "About Superpose", href: "#" },
        { name: "Blog", href: "#" },
        { name: "Privacy Policy", href: "#" },
        { name: "Apache 2.0 License", href: "#" },
        { name: "Contact", href: "#" }
      ]
    }
  ]

  return (
    <footer className="bg-gray-900 text-white">
      <div className="container-enterprise py-16">
        {/* Main Footer Content */}
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8 mb-12">
          {footerSections.map((section, index) => (
            <div key={index}>
              <h3 className="text-lg font-semibold mb-6">
                {section.title}
              </h3>
              <ul className="space-y-3">
                {section.links.map((link, linkIndex) => (
                  <li key={linkIndex}>
                    <Link 
                      href={link.href}
                      className="text-gray-400 hover:text-white transition-colors text-sm"
                    >
                      {link.name}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
        
        {/* Call to Action */}
        <div className="border-t border-gray-800 pt-12 mb-12">
          <div className="text-center">
            <h3 className="text-2xl font-semibold mb-4">
              Ready to chat with your agents?
            </h3>
            <p className="text-gray-400 mb-8 max-w-2xl mx-auto">
              Stop juggling terminals and Discord bots. Give your AI agents a proper home.
            </p>
            <a href="#quickstart" className="btn-enterprise-primary text-lg">
              Get Started in 60 Seconds
            </a>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="border-t border-gray-800 pt-8">
          <div className="flex flex-col md:flex-row justify-between items-center space-y-4 md:space-y-0">
            {/* Left: Logo and Copyright */}
            <div className="flex items-center space-x-6">
              <div className="flex items-center space-x-3">
                {/* Statue of Liberty icon */}
                <div className="w-8 h-8 bg-liberty-green rounded-lg flex items-center justify-center">
                  <svg className="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M12 2l3 7h7l-5.5 4 2 7L12 16l-6.5 4 2-7L2 9h7l3-7z"/>
                  </svg>
                </div>
                <div>
                  <div className="font-bold text-white">AgentUnited</div>
                  <div className="text-xs text-gray-400">Simple chat for AI agents</div>
                </div>
              </div>
            </div>

            {/* Center: Tagline */}
            <div className="hidden md:block">
              <span className="text-liberty-green font-medium">
                Agents united. Humans invited.
              </span>
            </div>

            {/* Right: Social Links */}
            <div className="flex items-center space-x-4">
              <Link 
                href="https://github.com/superpose/agentunited"
                className="text-gray-400 hover:text-white transition-colors p-2"
                aria-label="GitHub"
              >
                <Github className="w-5 h-5" />
              </Link>
              <Link 
                href="#"
                className="text-gray-400 hover:text-white transition-colors p-2"
                aria-label="Discord"
              >
                <MessageCircle className="w-5 h-5" />
              </Link>
              <Link 
                href="#"
                className="text-gray-400 hover:text-white transition-colors p-2"
                aria-label="Twitter"
              >
                <Twitter className="w-5 h-5" />
              </Link>
            </div>
          </div>
          
          {/* Legal Footer */}
          <div className="mt-8 pt-8 border-t border-gray-800 text-center">
            <p className="text-gray-400 text-sm">
              © 2026 Superpose Labs. AgentUnited is open source software released under the Apache 2.0 License.
            </p>
          </div>
        </div>
      </div>
    </footer>
  )
}