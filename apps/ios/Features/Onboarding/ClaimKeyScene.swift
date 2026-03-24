import SwiftUI
import UIKit

struct ClaimKeyScene: View {
    @Environment(\.dismiss) private var dismiss
    let claimKey: String
    let centralJWT: String
    let onConnected: (URL, String) -> Void

    @State private var copied = false
    @State private var isSharePresented = false
    @StateObject private var viewModel = ClaimKeyViewModel()

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Claim Key")
                        .font(.title.bold())
                    Text("Share this with your agent to set up a workspace")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                Text(claimKey)
                    .font(.system(.title3, design: .monospaced).weight(.semibold))
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 18)
                    .padding(.horizontal, 12)
                    .background(Color.auSecondary)
                    .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    .textSelection(.enabled)
                    .accessibilityIdentifier("claim-key-label")

                HStack(spacing: 12) {
                    Button(copied ? "Copied!" : "Copy") {
                        UIPasteboard.general.string = claimKey
                        copied = true
                    }
                    .buttonStyle(AUPrimaryButtonStyle())
                    .accessibilityIdentifier("claim-copy-button")

                    Button("Share") {
                        isSharePresented = true
                    }
                    .buttonStyle(AUSecondaryButtonStyle())
                    .accessibilityIdentifier("claim-share-button")
                }

                VStack(alignment: .leading, spacing: 8) {
                    if viewModel.workspaceReady {
                        Text("✅ Workspace ready! Tap to connect.")
                            .font(.body)
                        Button("Connect workspace") {
                            if let relayURL = viewModel.relayURL {
                                onConnected(relayURL, centralJWT)
                            }
                        }
                        .buttonStyle(AUPrimaryButtonStyle())
                    } else {
                        HStack(spacing: 8) {
                            ProgressView()
                            Text("Waiting for your agent…")
                                .font(.body)
                        }
                    }
                }
                .accessibilityIdentifier("claim-poll-state")
            }
            .padding(.horizontal, 20)
            .padding(.top, 20)
            .padding(.bottom, 28)
        }
        .navigationTitle("Claim Key")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarLeading) {
                Button("Cancel") { dismiss() }
                    .foregroundStyle(Color.auEmerald)
            }
        }
        .task {
            await viewModel.startMockPolling()
        }
        .sheet(isPresented: $isSharePresented) {
            ShareSheet(items: [claimKey])
        }
    }
}

@MainActor
final class ClaimKeyViewModel: ObservableObject {
    @Published var workspaceReady = false
    @Published var relayURL: URL?

    func startMockPolling() async {
        if workspaceReady { return }
        for _ in 0..<3 {
            try? await Task.sleep(nanoseconds: 10_000_000_000) // 10s x3
        }
        relayURL = URL(string: "https://example.tunnel.agentunited.ai")
        workspaceReady = true
    }
}

private struct ShareSheet: UIViewControllerRepresentable {
    let items: [Any]

    func makeUIViewController(context: Context) -> UIActivityViewController {
        UIActivityViewController(activityItems: items, applicationActivities: nil)
    }

    func updateUIViewController(_ uiViewController: UIActivityViewController, context: Context) {}
}
