import { LandingHero } from "@/components/landing-hero"
import { LandingProblem } from "@/components/landing-problem"
import { LandingHowItWorks } from "@/components/landing-how-it-works"
import { LandingBenefits } from "@/components/landing-benefits"
import { LandingQuickstart } from "@/components/landing-quickstart"
import { LandingFAQ } from "@/components/landing-faq"
import { LandingFooter } from "@/components/landing-footer"

export default function Page() {
  return (
    <div className="min-h-screen bg-black">
      {/* Enterprise hero section */}
      <LandingHero />
      
      {/* Main content with consistent dark theme and enterprise spacing */}
      <main className="relative z-10 bg-gradient-to-b from-black via-gray-950 to-black">
        {/* Subtle grid overlay for enterprise feel */}
        <div className="absolute inset-0 bg-[linear-gradient(rgba(255,255,255,0.02)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.02)_1px,transparent_1px)] bg-[size:32px_32px] [mask-image:radial-gradient(ellipse_800px_600px_at_50%_0%,black_40%,transparent_100%)]" />
        
        <div className="relative z-10 space-y-32 md:space-y-40 lg:space-y-48">
          {/* Problem section with enterprise styling */}
          <section className="py-20 border-t border-white/5">
            <div className="container mx-auto px-6 lg:px-8">
              <div className="max-w-4xl mx-auto">
                <LandingProblem />
              </div>
            </div>
          </section>

          {/* How it works with premium backdrop */}
          <section className="py-20 relative">
            <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.02] to-transparent" />
            <div className="container mx-auto px-6 lg:px-8 relative z-10">
              <div className="max-w-6xl mx-auto">
                <LandingHowItWorks />
              </div>
            </div>
          </section>

          {/* Benefits with glassmorphism container */}
          <section className="py-20">
            <div className="container mx-auto px-6 lg:px-8">
              <div className="max-w-6xl mx-auto">
                <LandingBenefits />
              </div>
            </div>
          </section>

          {/* Quick start with emphasis */}
          <section className="py-20 border-y border-white/5">
            <div className="container mx-auto px-6 lg:px-8">
              <div className="max-w-4xl mx-auto">
                <LandingQuickstart />
              </div>
            </div>
          </section>

          {/* FAQ section */}
          <section className="py-20">
            <div className="container mx-auto px-6 lg:px-8">
              <div className="max-w-4xl mx-auto">
                <LandingFAQ />
              </div>
            </div>
          </section>
        </div>
      </main>
      
      {/* Enterprise footer */}
      <LandingFooter />
    </div>
  )
}