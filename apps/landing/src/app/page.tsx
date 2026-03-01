import { LandingHero } from "@/components/landing-hero"
import { LandingProblem } from "@/components/landing-problem"
import { LandingHowItWorks } from "@/components/landing-how-it-works"
import { LandingBenefits } from "@/components/landing-benefits"
import { LandingQuickstart } from "@/components/landing-quickstart"
import { LandingFAQ } from "@/components/landing-faq"
import { LandingFooter } from "@/components/landing-footer"

export default function Page() {
  return (
    <div className="min-h-screen bg-white">
      {/* Hero takes full viewport */}
      <LandingHero />
      
      {/* Main content with production-grade spacing */}
      <main className="relative z-10 bg-white">
        <div className="space-y-24 md:space-y-32 lg:space-y-40">
          <LandingProblem />
          <LandingHowItWorks />
          <LandingBenefits />
          <LandingQuickstart />
          <LandingFAQ />
        </div>
      </main>
      
      {/* Footer */}
      <LandingFooter />
    </div>
  )
}