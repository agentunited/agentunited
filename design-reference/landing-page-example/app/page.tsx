import { LandingNav } from "@/components/landing-nav"
import { LandingHero } from "@/components/landing-hero"
import { LandingFeatures } from "@/components/landing-features"
import { LandingChatPreview } from "@/components/landing-chat-preview"
import { LandingHowItWorks } from "@/components/landing-how-it-works"
import { LandingTestimonials } from "@/components/landing-testimonials"
import { LandingCTA } from "@/components/landing-cta"
import { LandingFooter } from "@/components/landing-footer"

export default function Page() {
  return (
    <div className="min-h-dvh bg-background">
      <LandingNav />
      <main>
        <LandingHero />
        <LandingFeatures />
        <LandingChatPreview />
        <LandingHowItWorks />
        <LandingTestimonials />
        <LandingCTA />
      </main>
      <LandingFooter />
    </div>
  )
}
