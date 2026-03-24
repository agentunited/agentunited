import SwiftUI

/// One-time orientation screen shown on first install before the Welcome screen.
/// Writes `UserDefaults "hasSeenOrientation" = true` on "Got it →" tap.
struct OrientationScene: View {
    var body: some View {
        ZStack {
            LinearGradient(
                colors: [Color(hex: 0x064E3B), Color(hex: 0x0F172A)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
            .ignoresSafeArea()

            WelcomePattern().ignoresSafeArea()

            VStack(alignment: .leading, spacing: 0) {
                Spacer(minLength: 0)

                VStack(alignment: .leading, spacing: 32) {
                    // Globe orb
                    ZStack {
                        Circle()
                            .fill(Color.auEmerald.opacity(0.26))
                            .frame(width: 88, height: 88)
                            .blur(radius: 14)
                        Circle()
                            .fill(
                                LinearGradient(
                                    colors: [Color.white.opacity(0.18), Color.white.opacity(0.06)],
                                    startPoint: .topLeading,
                                    endPoint: .bottomTrailing
                                )
                            )
                            .overlay { Circle().stroke(Color.white.opacity(0.18), lineWidth: 1) }
                            .frame(width: 80, height: 80)
                        Image(systemName: "globe")
                            .font(.system(size: 34, weight: .medium))
                            .foregroundStyle(.white)
                    }
                    .accessibilityHidden(true)

                    VStack(alignment: .leading, spacing: 14) {
                        Text("Agent United connects you to your agent's workspace through our relay.")
                            .font(.title2.bold())
                            .foregroundStyle(.white)

                        Text("Your chats stay private — we route them, we don't read them.")
                            .font(.body)
                            .foregroundStyle(Color.white.opacity(0.82))
                    }
                }
                .padding(.horizontal, 28)

                Spacer(minLength: 0)
            }
        }
        .safeAreaInset(edge: .bottom) {
            VStack(spacing: 12) {
                Button {
                    UserDefaults.standard.set(true, forKey: "hasSeenOrientation")
                } label: {
                    Text("Got it →")
                }
                .buttonStyle(AUPrimaryButtonStyle())
                .accessibilityIdentifier("orientation-got-it-button")

                Link(destination: URL(string: "https://docs.agentunited.ai/docs")!) {
                    Text("Having trouble? See how it works →")
                        .font(.footnote)
                        .foregroundStyle(Color.white.opacity(0.55))
                }
            }
            .padding(.horizontal, 24)
            .padding(.vertical, 20)
        }
        .navigationBarHidden(true)
    }
}
