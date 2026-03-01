import { LandingHero } from "@/components/landing-hero"
import { LandingProblem } from "@/components/landing-problem"
import { LandingHowItWorks } from "@/components/landing-how-it-works"
import { LandingBenefits } from "@/components/landing-benefits"
import { LandingQuickstart } from "@/components/landing-quickstart"
import { LandingFAQ } from "@/components/landing-faq"
import { LandingFooter } from "@/components/landing-footer"

export default function Page() {
  return (
    <div className="min-h-screen">
      <LandingHero />
      <main>
        <LandingProblem />
        <LandingHowItWorks />
        <LandingBenefits />
        <LandingQuickstart />
        <LandingFAQ />
      </main>
      <LandingFooter />
    </div>
  )
}