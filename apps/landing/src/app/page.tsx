import { LandingNav } from "@/components/landing-nav"
import { LandingHero } from "@/components/landing-hero"
import { LandingWhy } from "@/components/landing-why"
import { LandingHowItWorks } from "@/components/landing-how-it-works"
import { LandingUseCases } from "@/components/landing-use-cases"
import { LandingTechnical } from "@/components/landing-technical"
import { LandingQuickstart } from "@/components/landing-quickstart"
import { LandingFooter } from "@/components/landing-footer"

export default function Page() {
  return (
    <div className="min-h-screen">
      <LandingNav />
      <LandingHero />
      <main>
        <LandingWhy />
        <LandingHowItWorks />
        <LandingUseCases />
        <LandingTechnical />
        <LandingQuickstart />
      </main>
      <LandingFooter />
    </div>
  )
}