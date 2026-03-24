import SwiftUI
import SafariServices

struct GetStartedScene: View {
    @Environment(\.dismiss) private var dismiss
    let sessionStore: AppSessionStore
    @State private var isSafariPresented = false
    @State private var isPresentingSelfHosted = false

    var body: some View {
        ZStack {
            LinearGradient(
                colors: [Color(hex: 0x064E3B), Color(hex: 0x0F172A)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
            .ignoresSafeArea()

            WelcomePattern().ignoresSafeArea()

            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    Spacer(minLength: 24)

                    Text("How would you like to get started?")
                        .font(.title2.bold())
                        .foregroundStyle(.white)

                    // Option A: My agent sets it up
                    Button {
                        isSafariPresented = true
                    } label: {
                        GetStartedOptionCard(
                            emoji: "🤖",
                            title: "My agent sets it up",
                            description: "Agent bootstraps your workspace, then sends you an invite link. Perfect for agents that can run curl commands or call APIs."
                        )
                    }
                    .accessibilityIdentifier("get-started-agent-option")

                    // Option B: Self-hosted workspace
                    Button {
                        isPresentingSelfHosted = true
                    } label: {
                        GetStartedOptionCard(
                            emoji: "🖥️",
                            title: "Self-hosted workspace",
                            description: "Connect to a workspace you run yourself on your own server."
                        )
                    }
                    .accessibilityIdentifier("get-started-selfhosted-option")

                    HStack {
                        Rectangle()
                            .fill(Color.white.opacity(0.15))
                            .frame(height: 1)
                        Text("or")
                            .font(.caption)
                            .foregroundStyle(Color.white.opacity(0.45))
                            .padding(.horizontal, 8)
                        Rectangle()
                            .fill(Color.white.opacity(0.15))
                            .frame(height: 1)
                    }

                    Link(destination: URL(string: "https://docs.agentunited.ai/docs/quickstart")!) {
                        Text("See the full setup guide →")
                            .font(.footnote)
                            .foregroundStyle(Color.white.opacity(0.60))
                            .frame(maxWidth: .infinity, alignment: .leading)
                    }

                    Spacer(minLength: 32)
                }
                .padding(.horizontal, 24)
            }
        }
        .navigationTitle("Get Started")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarLeading) {
                Button("Close") { dismiss() }
                    .foregroundStyle(.white)
            }
        }
        .sheet(isPresented: $isSafariPresented) {
            AgentQuickstartSafariView()
        }
        .sheet(isPresented: $isPresentingSelfHosted) {
            NavigationStack {
                SelfHostedSignInScene(sessionStore: sessionStore)
            }
            .tint(.auEmerald)
        }
    }
}

// MARK: - Option card

private struct GetStartedOptionCard: View {
    let emoji: String
    let title: String
    let description: String

    var body: some View {
        HStack(alignment: .top, spacing: 14) {
            Text(emoji)
                .font(.title2)
                .accessibilityHidden(true)
            VStack(alignment: .leading, spacing: 4) {
                Text(title)
                    .font(.body.weight(.semibold))
                    .foregroundStyle(.white)
                Text(description)
                    .font(.footnote)
                    .foregroundStyle(Color.white.opacity(0.70))
                    .fixedSize(horizontal: false, vertical: true)
                    .multilineTextAlignment(.leading)
            }
            Spacer(minLength: 0)
            Image(systemName: "chevron.right")
                .font(.footnote.weight(.semibold))
                .foregroundStyle(Color.white.opacity(0.35))
                .accessibilityHidden(true)
        }
        .padding(16)
        .background(Color.white.opacity(0.08))
        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .stroke(Color.white.opacity(0.15), lineWidth: 1)
        )
    }
}

// MARK: - Safari view

struct AgentQuickstartSafariView: UIViewControllerRepresentable {
    private let url = URL(string: "https://docs.agentunited.ai/docs/agent-quickstart")!

    func makeUIViewController(context: Context) -> SFSafariViewController {
        SFSafariViewController(url: url)
    }

    func updateUIViewController(_ uiViewController: SFSafariViewController, context: Context) {}
}
