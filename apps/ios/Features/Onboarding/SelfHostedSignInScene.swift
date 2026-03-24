import SwiftUI

// MARK: - Recent workspace model

struct RecentWorkspace: Codable, Identifiable {
    let url: String
    let name: String?
    let lastUsed: Date

    var id: String { url }

    var displayName: String {
        name ?? url
            .replacingOccurrences(of: "https://", with: "")
            .replacingOccurrences(of: "http://", with: "")
    }

    static func load() -> [RecentWorkspace] {
        guard let data = UserDefaults.standard.data(forKey: "recentWorkspaces"),
              let items = try? JSONDecoder().decode([RecentWorkspace].self, from: data)
        else { return [] }
        return items.sorted { $0.lastUsed > $1.lastUsed }
    }

    static func save(url: String, name: String?) {
        var items = load().filter { $0.url != url }
        items.insert(RecentWorkspace(url: url, name: name, lastUsed: Date()), at: 0)
        let trimmed = Array(items.prefix(3))
        if let data = try? JSONEncoder().encode(trimmed) {
            UserDefaults.standard.set(data, forKey: "recentWorkspaces")
        }
    }
}

// MARK: - Scene

/// Sign-in form for self-hosted / local workspaces — URL + email + password.
/// Accessed via GetStartedScene or agentunited://workspace?url=xxx deep link.
struct SelfHostedSignInScene: View {
    @EnvironmentObject private var coordinator: AppCoordinator
    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel: SelfHostedSignInViewModel

    init(sessionStore: AppSessionStore, prefilledURL: String = "") {
        _viewModel = StateObject(
            wrappedValue: SelfHostedSignInViewModel(sessionStore: sessionStore, prefilledURL: prefilledURL)
        )
    }

    @State private var showForgotURLHelp = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Connect workspace")
                        .font(.title.bold())
                    Text("Sign in using your workspace's server address.")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                VStack(alignment: .leading, spacing: 20) {
                    // MARK: Workspace URL
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Workspace URL (your server)")
                            .font(.subheadline.weight(.medium))
                            .foregroundStyle(Color.auSecondaryLabel)

                        TextField("https://acme.tunnel.agentunited.ai", text: $viewModel.workspaceURL)
                            .keyboardType(.URL)
                            .textContentType(.URL)
                            .textInputAutocapitalization(.never)
                            .autocorrectionDisabled(true)
                            .fieldStyled()
                            .accessibilityIdentifier("workspace-url-field")

                        // Recent workspaces chips
                        if !viewModel.recentWorkspaces.isEmpty {
                            ScrollView(.horizontal, showsIndicators: false) {
                                HStack(spacing: 8) {
                                    ForEach(viewModel.recentWorkspaces) { ws in
                                        Button {
                                            viewModel.workspaceURL = ws.url
                                        } label: {
                                            Text(ws.displayName)
                                                .font(.caption.weight(.medium))
                                                .foregroundStyle(Color.auEmerald)
                                                .lineLimit(1)
                                                .padding(.horizontal, 10)
                                                .padding(.vertical, 6)
                                                .background(Color.auEmerald.opacity(0.10))
                                                .clipShape(Capsule())
                                                .overlay(Capsule().stroke(Color.auEmerald.opacity(0.30), lineWidth: 1))
                                        }
                                        .accessibilityLabel("Use workspace: \(ws.displayName)")
                                    }
                                }
                                .padding(.horizontal, 2)
                            }
                        }

                        // Forgot workspace URL
                        VStack(alignment: .leading, spacing: 4) {
                            Button("Forgot workspace URL?") {
                                withAnimation(.easeInOut(duration: 0.2)) {
                                    showForgotURLHelp.toggle()
                                }
                            }
                            .font(.footnote)
                            .foregroundStyle(Color.auEmerald)
                            .accessibilityIdentifier("forgot-url-button")

                            if showForgotURLHelp {
                                Text("Ask your agent: \"What's your workspace URL?\"")
                                    .font(.footnote)
                                    .foregroundStyle(.secondary)
                                    .transition(.opacity.combined(with: .move(edge: .top)))
                                    .accessibilityIdentifier("forgot-url-help-text")
                            }
                        }
                        .animation(.easeInOut(duration: 0.2), value: showForgotURLHelp)
                    }

                    // MARK: Email
                    SignInFieldGroup(title: "Email") {
                        TextField("you@example.com", text: $viewModel.email)
                            .keyboardType(.emailAddress)
                            .textContentType(.emailAddress)
                            .textInputAutocapitalization(.never)
                            .autocorrectionDisabled(true)
                            .accessibilityIdentifier("email-field")
                    }

                    // MARK: Password
                    SignInFieldGroup(title: "Password") {
                        SecureField("Enter your password", text: $viewModel.password)
                            .textContentType(.password)
                            .accessibilityIdentifier("password-field")
                    }
                }

                if let error = viewModel.errorMessage {
                    Text(error)
                        .font(.subheadline)
                        .foregroundStyle(.red)
                }

                Button {
                    Task {
                        let ok = await viewModel.signIn()
                        if ok { coordinator.setAuthenticated(true) }
                    }
                } label: {
                    if viewModel.isSubmitting {
                        ProgressView().frame(maxWidth: .infinity)
                    } else {
                        Text("Sign in")
                    }
                }
                .buttonStyle(AUPrimaryButtonStyle())
                .disabled(!viewModel.canSubmit)
                .accessibilityIdentifier("submit-sign-in-button")
            }
            .padding(.horizontal, 20)
            .padding(.top, 20)
            .padding(.bottom, 32)
        }
        .background(Color.auGrouped.ignoresSafeArea())
        .navigationTitle("Connect Workspace")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarLeading) {
                Button("Close") { dismiss() }
            }
        }
    }
}

// MARK: - Shared field group

struct SignInFieldGroup<Content: View>: View {
    let title: String
    @ViewBuilder let content: () -> Content

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.subheadline.weight(.medium))
                .foregroundStyle(Color.auSecondaryLabel)
            content().fieldStyled()
        }
    }
}

// MARK: - Field style

extension View {
    func fieldStyled() -> some View {
        self
            .font(.body)
            .padding(.horizontal, 14)
            .frame(height: 50)
            .background(Color.auSecondary)
            .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
    }
}

// MARK: - ViewModel

@MainActor
final class SelfHostedSignInViewModel: ObservableObject {
    @Published var workspaceURL: String
    @Published var email = ""
    @Published var password = ""
    @Published var isSubmitting = false
    @Published var errorMessage: String?
    @Published var recentWorkspaces: [RecentWorkspace] = []

    private let sessionStore: AppSessionStore

    init(sessionStore: AppSessionStore, prefilledURL: String = "") {
        self.sessionStore = sessionStore
        self.workspaceURL = prefilledURL
        self.recentWorkspaces = RecentWorkspace.load()
    }

    var canSubmit: Bool {
        !workspaceURL.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty &&
        !email.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty &&
        !password.isEmpty &&
        !isSubmitting
    }

    /// Returns true on success, persists session and saves to recent workspaces.
    func signIn() async -> Bool {
        errorMessage = nil
        isSubmitting = true
        defer { isSubmitting = false }

        let urlString = workspaceURL.trimmingCharacters(in: .whitespacesAndNewlines)
        guard let url = URL(string: urlString) else {
            errorMessage = "Enter a valid workspace URL."
            return false
        }

        do {
            let client = LiveAUAPIClient(instanceURL: url)
            let response = try await client.login(
                email: email.trimmingCharacters(in: .whitespacesAndNewlines),
                password: password
            )
            let fallbackDisplayName = email.split(separator: "@").first.map(String.init) ?? "User"
            try sessionStore.persistInviteSession(
                instanceURL: url,
                userID: response.userID,
                email: email.trimmingCharacters(in: .whitespacesAndNewlines),
                displayName: fallbackDisplayName,
                token: response.token,
                expiresAt: response.expiresAt,
                userType: "human"
            )
            RecentWorkspace.save(url: urlString, name: nil)
            return true
        } catch {
            errorMessage = "Sign in failed. Check workspace URL, email, and password."
            return false
        }
    }
}
