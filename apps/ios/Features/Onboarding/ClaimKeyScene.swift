import SwiftUI
import UIKit

extension String {
    func decodedJWTPayload() -> [String: Any]? {
        let parts = self.split(separator: ".")
        guard parts.count == 3 else { return nil }
        var base64 = String(parts[1])
            .replacingOccurrences(of: "-", with: "+")
            .replacingOccurrences(of: "_", with: "/")
        while base64.count % 4 != 0 { base64 += "=" }
        guard let data = Data(base64Encoded: base64) else { return nil }
        return try? JSONSerialization.jsonObject(with: data) as? [String: Any]
    }
}

struct ClaimKeyScene: View {
    @Environment(\.dismiss) private var dismiss
    let claimKey: String
    let onConnected: (URL, String) -> Void

    @State private var copied = false
    @State private var isSharePresented = false
    @StateObject private var viewModel = ClaimKeyViewModel()

    init(claimKey: String, centralJWT: String, onConnected: @escaping (URL, String) -> Void) {
        self.claimKey = claimKey
        self.onConnected = onConnected
        viewModel.setCentralJWT(centralJWT)
    }

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
                                onConnected(relayURL, viewModel.centralJWT)
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
            await viewModel.startPolling()
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
    @Published var centralJWT: String = ""
    private var isPolling = false

    func setCentralJWT(_ jwt: String) {
        self.centralJWT = jwt
    }

    func startPolling() async {
        if workspaceReady || isPolling { return }
        isPolling = true

        let client = CentralAPIClient(authToken: centralJWT)

        while !Task.isCancelled && !workspaceReady {
            do {
                let workspaces = try await client.listWorkspaces()
                if let workspace = workspaces.first,
                   let url = URL(string: workspace.relayURL) {
                    relayURL = url
                    workspaceReady = true
                    return
                }
            } catch {
                // Continue polling on error
            }

            try? await Task.sleep(nanoseconds: 10_000_000_000)
        }

        isPolling = false
    }
}

private struct ShareSheet: UIViewControllerRepresentable {
    let items: [Any]

    func makeUIViewController(context: Context) -> UIActivityViewController {
        UIActivityViewController(activityItems: items, applicationActivities: nil)
    }

    func updateUIViewController(_ uiViewController: UIActivityViewController, context: Context) {}
}
