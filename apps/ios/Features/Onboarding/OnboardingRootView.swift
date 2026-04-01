import SwiftUI

struct OnboardingRootView: View {
    @EnvironmentObject private var coordinator: AppCoordinator
    @EnvironmentObject private var sessionStore: AppSessionStore

    // Orientation gate — UserDefaults-backed; OrientationScene writes true on "Got it →"
    @AppStorage("hasSeenOrientation") private var hasSeenOrientation = false

    // Relay-first primary flow
    @State private var isPresentingRelaySignIn = false
    @State private var isPresentingSignUp = false
    @State private var isPresentingWorkspaceList = false
    @State private var centralJWT = ""

    // Self-hosted sign-in (workspace deep link entry)
    @State private var isPresentingSelfHostedSignIn = false
    @State private var prefilledWorkspaceURL = ""

    // Claim key deep link (Path B)
    @State private var isPresentingClaimKey = false
    @State private var claimKey = ""

    // Reset password deep link
    @State private var isPresentingResetPassword = false
    @State private var resetPasswordToken = ""
    @State private var resetPasswordEmail = ""

    private var pendingInvite: AppCoordinator.PendingInvite? {
        if case let .invite(invite) = coordinator.pendingRoute { return invite }
        return nil
    }

    var body: some View {
        Group {
            if hasSeenOrientation {
                WelcomeScreen(
                    isSignInPresented: $isPresentingRelaySignIn,
                    isGetStartedPresented: $isPresentingSignUp
                )
            } else {
                OrientationScene()
            }
        }
        // Relay sign-in (R4) → WorkspaceListScene (R5)
        .fullScreenCover(isPresented: $isPresentingRelaySignIn) {
            NavigationStack {
                RelaySignInScene(
                    sessionStore: sessionStore,
                    onAuthenticated: { jwt in
                        centralJWT = jwt
                        isPresentingRelaySignIn = false
                        Task {
                            try? await Task.sleep(nanoseconds: 200_000_000)
                            isPresentingWorkspaceList = true
                        }
                    },
                    onSelfHosted: {
                        isPresentingRelaySignIn = false
                        Task {
                            try? await Task.sleep(nanoseconds: 200_000_000)
                            isPresentingSelfHostedSignIn = true
                        }
                    },
                    onForgotPasswordEmailCaptured: { email in
                        resetPasswordEmail = email
                    }
                )
            }
            .tint(.auEmerald)
        }
        // Sign-up (R3) → WorkspaceListScene (R5)
        .fullScreenCover(isPresented: $isPresentingSignUp) {
            NavigationStack {
                SignUpScene(
                    needsClaimKey: false,
                    onOpenSignIn: {
                        isPresentingSignUp = false
                        Task {
                            try? await Task.sleep(nanoseconds: 200_000_000)
                            isPresentingRelaySignIn = true
                        }
                    },
                    onRegistered: { _, jwt in
                        centralJWT = jwt
                        isPresentingSignUp = false
                        Task {
                            try? await Task.sleep(nanoseconds: 200_000_000)
                            isPresentingWorkspaceList = true
                        }
                    }
                )
            }
            .tint(.auEmerald)
        }
        // Workspace list (R5) — contains nav to GetStartedScene (R7)
        .fullScreenCover(isPresented: $isPresentingWorkspaceList) {
            WorkspaceListScene(centralJWT: centralJWT)
        }
        // Self-hosted sign-in (R6) — workspace deep link or from GetStartedScene
        .fullScreenCover(isPresented: $isPresentingSelfHostedSignIn) {
            NavigationStack {
                SelfHostedSignInScene(
                    sessionStore: sessionStore,
                    prefilledURL: prefilledWorkspaceURL
                )
            }
            .tint(.auEmerald)
        }
        // Claim key deep link (Path B legacy)
        .fullScreenCover(isPresented: $isPresentingClaimKey) {
            NavigationStack {
                ClaimKeyScene(
                    claimKey: claimKey,
                    centralJWT: centralJWT,
                    onConnected: { relayURL, jwt in
                        let payload = jwt.decodedJWTPayload()
                        let userID = payload?["sub"] as? String ?? "central-user"
                        let email = payload?["email"] as? String ?? "user@agentunited.ai"
                        let displayName = payload?["display_name"] as? String ?? "User"
                        try? sessionStore.persistInviteSession(
                            instanceURL: relayURL,
                            userID: userID,
                            email: email,
                            displayName: displayName,
                            token: jwt,
                            expiresAt: nil,
                            userType: "human"
                        )
                        isPresentingClaimKey = false
                        coordinator.setAuthenticated(true)
                    }
                )
            }
            .tint(.auEmerald)
        }
        // Invite deep link (Path A) — unchanged
        .fullScreenCover(isPresented: $coordinator.isPresentingInvite) {
            NavigationStack {
                InviteAcceptScene(
                    pendingInvite: pendingInvite,
                    sessionStore: sessionStore
                )
            }
            .tint(.auEmerald)
        }
        .fullScreenCover(isPresented: $isPresentingResetPassword) {
            ResetPasswordScene(token: resetPasswordToken, email: resetPasswordEmail) {
                isPresentingResetPassword = false
                Task {
                    try? await Task.sleep(nanoseconds: 200_000_000)
                    isPresentingRelaySignIn = true
                }
            }
        }
        .onAppear {
            if pendingInvite != nil {
                coordinator.isPresentingInvite = true
            }
            if case let .claim(key) = coordinator.pendingRoute {
                claimKey = key
                centralJWT = (try? KeychainHelper().readJWT(for: "au.central.jwt")) ?? ""
                isPresentingClaimKey = true
            }
            if case let .workspace(url) = coordinator.pendingRoute {
                prefilledWorkspaceURL = url
                isPresentingSelfHostedSignIn = true
            }
            if case let .resetPassword(token) = coordinator.pendingRoute {
                presentResetPassword(token: token, email: resetPasswordEmail)
            }
        }
        .onChange(of: coordinator.pendingRoute) { _, newValue in
            if case .invite = newValue {
                coordinator.isPresentingInvite = true
            }
            if case let .claim(key) = newValue {
                claimKey = key
                centralJWT = (try? KeychainHelper().readJWT(for: "au.central.jwt")) ?? ""
                isPresentingClaimKey = true
            }
            if case let .workspace(url) = newValue {
                prefilledWorkspaceURL = url
                isPresentingSelfHostedSignIn = true
            }
            if case let .resetPassword(token) = newValue {
                presentResetPassword(token: token, email: resetPasswordEmail)
            }
        }
    }

    private func presentResetPassword(token: String, email: String) {
        resetPasswordToken = token
        resetPasswordEmail = email
        coordinator.pendingRoute = nil

        // Dismiss any currently presented onboarding modals first.
        // SwiftUI won't present a parent fullScreenCover while a child cover is active.
        coordinator.isPresentingInvite = false
        isPresentingClaimKey = false
        isPresentingWorkspaceList = false
        isPresentingSelfHostedSignIn = false
        isPresentingSignUp = false
        isPresentingRelaySignIn = false
        isPresentingResetPassword = false

        Task {
            try? await Task.sleep(nanoseconds: 200_000_000)
            isPresentingResetPassword = true
        }
    }
}

private struct WelcomeScreen: View {
    @Binding var isSignInPresented: Bool
    @Binding var isGetStartedPresented: Bool

    @State private var hasProbableClipboardURL = false
    @State private var isBannerVisible = false

    var body: some View {
        ZStack {
            LinearGradient(
                colors: [Color(hex: 0x064E3B), Color(hex: 0x0F172A)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
            .ignoresSafeArea()

            WelcomePattern().ignoresSafeArea()

            VStack(alignment: .leading, spacing: 24) {
                Spacer(minLength: 0)
                VStack(alignment: .leading, spacing: 24) {
                    logoMark
                    VStack(alignment: .leading, spacing: 10) {
                        Text("Agent United")
                            .font(.title.bold())
                            .foregroundStyle(.white)
                        Text("Your workspace, everywhere.")
                            .font(.subheadline)
                            .foregroundStyle(Color.white.opacity(0.78))
                    }
                }
            }
            .padding(.horizontal, 24)
            .padding(.top, 40)
        }
        .safeAreaInset(edge: .bottom) {
            VStack(spacing: 0) {
                if isBannerVisible, hasProbableClipboardURL {
                    ClipboardInviteBanner(
                        onTap: {
                            withAnimation { isBannerVisible = false }
                            guard let text = UIPasteboard.general.string,
                                  text.range(of: #"agentunited://|agentunited\.ai/invite\?token="#, options: .regularExpression) != nil,
                                  let url = URL(string: text)
                            else { return }
                            UIApplication.shared.open(url)
                        },
                        onDismiss: { withAnimation { isBannerVisible = false } }
                    )
                    .padding(.horizontal, 16)
                    .padding(.bottom, 10)
                    .transition(.move(edge: .bottom).combined(with: .opacity))
                }
                VStack(spacing: 12) {
                    Button("Get started for free") {
                        isGetStartedPresented = true
                    }
                    .buttonStyle(AUPrimaryButtonStyle())
                    .accessibilityIdentifier("get-started-button")

                    Button("Sign in") {
                        isSignInPresented = true
                    }
                    .buttonStyle(AUGhostButtonStyle())
                    .accessibilityIdentifier("sign-in-button")
                }
                .padding(.horizontal, 20)
                .padding(.top, 16)
                .padding(.bottom, 20)
            }
        }
        .navigationBarHidden(true)
        .animation(.easeInOut(duration: 0.25), value: isBannerVisible)
    }

    private var logoMark: some View {
        ZStack {
            Circle()
                .fill(Color.auEmerald.opacity(0.26))
                .frame(width: 104, height: 104)
                .blur(radius: 16)
            Circle()
                .fill(LinearGradient(
                    colors: [Color.white.opacity(0.2), Color.white.opacity(0.06)],
                    startPoint: .topLeading,
                    endPoint: .bottomTrailing
                ))
                .overlay { Circle().stroke(Color.white.opacity(0.18), lineWidth: 1) }
                .frame(width: 96, height: 96)
            Image(systemName: "message.badge.waveform.fill")
                .font(.system(size: 38, weight: .semibold))
                .foregroundStyle(.white)
        }
        .accessibilityHidden(true)
    }

    private func checkClipboard() {
        // iOS 16+: detect URL-like content without reading pasteboard payload (avoids paste permission prompt).
        UIPasteboard.general.detectPatterns(for: [.probableWebURL]) { result in
            guard case let .success(patterns) = result,
                  patterns.contains(.probableWebURL)
            else { return }

            // MUST use MainActor.run to access @State from background queue
            Task { @MainActor in
                hasProbableClipboardURL = true
                withAnimation { isBannerVisible = true }

                try? await Task.sleep(nanoseconds: 10_000_000_000)
                withAnimation { isBannerVisible = false }
            }
        }
    }
}

private struct ClipboardInviteBanner: View {
    let onTap: () -> Void
    let onDismiss: () -> Void

    var body: some View {
        HStack(spacing: 10) {
            Image(systemName: "link")
                .font(.footnote.weight(.semibold))
                .foregroundStyle(Color.auEmerald)
                .accessibilityHidden(true)
            Text("Looks like you have an invite link.")
                .font(.footnote)
                .foregroundStyle(.white)
                .lineLimit(1)
            Spacer(minLength: 4)
            Button("Open it →", action: onTap)
                .font(.footnote.weight(.semibold))
                .foregroundStyle(Color.auEmerald)
            Button(action: onDismiss) {
                Image(systemName: "xmark")
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(Color.white.opacity(0.45))
            }
            .accessibilityLabel("Dismiss invite banner")
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 10)
        .background(Color.white.opacity(0.10))
        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 10, style: .continuous).stroke(Color.white.opacity(0.15), lineWidth: 1))
        .accessibilityIdentifier("clipboard-invite-banner")
    }
}

// MARK: - Relay sign-in (R4) — email + password only, no workspace URL

private struct RelaySignInScene: View {
    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel: RelaySignInViewModel
    let onAuthenticated: (String) -> Void
    let onSelfHosted: () -> Void
    let onForgotPasswordEmailCaptured: (String) -> Void

    @State private var isPresentingForgotPassword = false

    init(sessionStore: AppSessionStore,
         onAuthenticated: @escaping (String) -> Void,
         onSelfHosted: @escaping () -> Void,
         onForgotPasswordEmailCaptured: @escaping (String) -> Void) {
        self.onAuthenticated = onAuthenticated
        self.onSelfHosted = onSelfHosted
        self.onForgotPasswordEmailCaptured = onForgotPasswordEmailCaptured
        _viewModel = StateObject(wrappedValue: RelaySignInViewModel(sessionStore: sessionStore))
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Sign in")
                        .font(.title.bold())
                    Text("Sign in with your Agent United account.")
                        .font(.subheadline)
                        .foregroundStyle(Color.auSecondaryLabel)
                }
                VStack(alignment: .leading, spacing: 20) {
                    SignInFieldSection(title: "Email") {
                        TextField("Email", text: $viewModel.email)
                            .keyboardType(.emailAddress)
                            .textContentType(.emailAddress)
                            .textInputAutocapitalization(.never)
                            .autocorrectionDisabled(true)
                            .accessibilityIdentifier("relay-email-field")
                    }
                    SignInFieldSection(title: "Password") {
                        SecureField("Enter your password", text: $viewModel.password)
                            .textContentType(.password)
                            .accessibilityIdentifier("relay-password-field")
                    }
                }
                if let error = viewModel.errorMessage {
                    Text(error)
                        .font(.subheadline)
                        .foregroundStyle(.red)
                }
                Button {
                    Task {
                        if let jwt = await viewModel.signIn() { onAuthenticated(jwt) }
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
                .accessibilityIdentifier("relay-sign-in-button")

                Button {
                    isPresentingForgotPassword = true
                } label: {
                    Text("Forgot password?")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                }
                .frame(maxWidth: .infinity, alignment: .trailing)
                .accessibilityIdentifier("forgot-password-link")

                Divider().padding(.vertical, 8)

                Button {
                    dismiss()
                    onSelfHosted()
                } label: {
                    (Text("Connecting to a self-hosted workspace?")
                        .font(.footnote)
                        .foregroundStyle(.secondary))
                    + (Text("  →")
                        .font(.footnote.weight(.semibold))
                        .foregroundStyle(Color.auEmerald))
                }
                .accessibilityIdentifier("self-hosted-link")
            }
            .padding(.horizontal, 20)
            .padding(.top, 20)
            .padding(.bottom, 32)
        }
        .background(Color.auGrouped.ignoresSafeArea())
        .navigationTitle("Sign In")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarLeading) {
                Button("Close") { dismiss() }
            }
        }
        .fullScreenCover(isPresented: $isPresentingForgotPassword) {
            ForgotPasswordScene(onSuccessEmailCaptured: onForgotPasswordEmailCaptured)
        }
    }
}

@MainActor
private final class RelaySignInViewModel: ObservableObject {
    @Published var email = ""
    @Published var password = ""
    @Published var isSubmitting = false
    @Published var errorMessage: String?

    private let sessionStore: AppSessionStore

    init(sessionStore: AppSessionStore) {
        self.sessionStore = sessionStore
    }

    var canSubmit: Bool {
        !email.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty && !password.isEmpty && !isSubmitting
    }

    func signIn() async -> String? {
        errorMessage = nil
        isSubmitting = true
        defer { isSubmitting = false }
        do {
            let resp = try await CentralAPIClient().login(
                email: email.trimmingCharacters(in: .whitespacesAndNewlines),
                password: password
            )
            try? KeychainHelper().storeJWT(resp.token, for: "au.central.jwt")
            return resp.token
        } catch let e as CentralAPIError {
            if case .invalidCredentials = e {
                errorMessage = "Incorrect email or password."
            } else {
                errorMessage = "Sign in failed. Please try again."
            }
            return nil
        } catch {
            errorMessage = "Sign in failed. Please try again."
            return nil
        }
    }
}


private struct SignInScene: View {
    enum Mode {
        case workspace
        case central
    }

    @EnvironmentObject private var coordinator: AppCoordinator
    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel: SignInViewModel
    let mode: Mode
    let onOpenSignUp: () -> Void
    let onAuthenticated: ((String) -> Void)?

    init(sessionStore: AppSessionStore, mode: Mode = .workspace, onOpenSignUp: @escaping () -> Void, onAuthenticated: ((String) -> Void)? = nil) {
        self.mode = mode
        self.onOpenSignUp = onOpenSignUp
        self.onAuthenticated = onAuthenticated
        _viewModel = StateObject(wrappedValue: SignInViewModel(sessionStore: sessionStore, mode: mode))
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Sign in")
                        .font(.title.bold())
                    Text("Connect to your workspace with the same credentials you use on the web.")
                        .font(.subheadline)
                        .foregroundStyle(Color.auSecondaryLabel)
                }

                VStack(alignment: .leading, spacing: 20) {
                    if mode == .workspace {
                        SignInFieldSection(title: "Workspace") {
                            TextField("https://workspace.example.com", text: $viewModel.workspaceURL)
                                .keyboardType(.URL)
                                .textInputAutocapitalization(.never)
                                .autocorrectionDisabled(true)
                                .accessibilityIdentifier("workspace-url-field")
                        }
                    }

                    SignInFieldSection(title: "Email") {
                        TextField("Email", text: $viewModel.email)
                            .keyboardType(.emailAddress)
                            .textContentType(.emailAddress)
                            .textInputAutocapitalization(.never)
                            .autocorrectionDisabled(true)
                            .accessibilityIdentifier("email-field")
                    }

                    SignInFieldSection(title: "Password") {
                        SecureField("Enter your password", text: $viewModel.password)
                            .textContentType(.password)
                            .accessibilityIdentifier("password-field")
                    }
                }

                if let errorMessage = viewModel.errorMessage {
                    Text(errorMessage)
                        .font(.subheadline)
                        .foregroundStyle(Color.red)
                }

                Button {
                    Task {
                        let result = await viewModel.signIn()
                        if let result {
                            if mode == .workspace {
                                coordinator.setAuthenticated(true)
                            } else {
                                onAuthenticated?(result.centralJWT)
                            }
                        }
                    }
                } label: {
                    if viewModel.isSubmitting {
                        ProgressView()
                            .frame(maxWidth: .infinity)
                    } else {
                        Text("Sign in")
                    }
                }
                .buttonStyle(AUPrimaryButtonStyle())
                .disabled(viewModel.canSubmit == false)
                .accessibilityIdentifier("submit-sign-in-button")

                Divider()
                    .padding(.vertical, 8)

                Text("No account yet? Accept an invite from your agent.")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
                    .frame(maxWidth: .infinity)
                    .padding(.horizontal, 4)

                Button("← Back to Welcome") {
                    dismiss()
                }
                .font(.footnote)
                .foregroundStyle(Color.auEmerald)
                .padding(.top, 4)

                Button("Setting up? Get a claim key →") {
                    dismiss()
                    onOpenSignUp()
                }
                .font(.footnote.weight(.semibold))
                .foregroundStyle(Color.auEmerald)
                .padding(.top, 2)
            }
            .padding(.horizontal, 20)
            .padding(.top, 20)
            .padding(.bottom, 32)
        }
        .background(Color.auGrouped.ignoresSafeArea())
        .navigationTitle("Sign In")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarLeading) {
                Button("Close") {
                    dismiss()
                }
            }
        }
    }
}

@MainActor
private final class SignInViewModel: ObservableObject {
    struct Result {
        let centralJWT: String
    }

    @Published var workspaceURL: String = ""
    @Published var email: String = ""
    @Published var password: String = ""
    @Published var isSubmitting: Bool = false
    @Published var errorMessage: String?

    private let sessionStore: AppSessionStore
    private let mode: SignInScene.Mode

    init(sessionStore: AppSessionStore, mode: SignInScene.Mode) {
        self.sessionStore = sessionStore
        self.mode = mode
    }

    var canSubmit: Bool {
        let hasWorkspace = mode == .central || workspaceURL.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty == false
        return hasWorkspace
            && email.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty == false
            && password.isEmpty == false
            && isSubmitting == false
    }

    func signIn() async -> Result? {
        errorMessage = nil
        isSubmitting = true
        defer { isSubmitting = false }

        if mode == .central {
            do {
                let client = CentralAPIClient()
                let authResp = try await client.login(
                    email: email.trimmingCharacters(in: .whitespacesAndNewlines),
                    password: password
                )
                let keychain = KeychainHelper()
                try? keychain.storeJWT(authResp.token, for: "au.central.jwt")
                return Result(centralJWT: authResp.token)
            } catch let e as CentralAPIError {
                if case .invalidCredentials = e {
                    errorMessage = "Incorrect email or password."
                } else {
                    errorMessage = "Sign in failed. Please try again."
                }
                return nil
            } catch {
                errorMessage = "Sign in failed. Please try again."
                return nil
            }
        }

        guard let url = URL(string: workspaceURL.trimmingCharacters(in: .whitespacesAndNewlines)) else {
            errorMessage = "Enter a valid workspace URL."
            return nil
        }

        do {
            let client = LiveAUAPIClient(instanceURL: url)
            let response = try await client.login(
                email: email.trimmingCharacters(in: .whitespacesAndNewlines),
                password: password
            )

            let fallbackDisplayName = email
                .split(separator: "@")
                .first
                .map(String.init)
                .flatMap { $0.isEmpty ? nil : $0 }
                ?? "User"

            try sessionStore.persistInviteSession(
                instanceURL: url,
                userID: response.userID,
                email: email.trimmingCharacters(in: .whitespacesAndNewlines),
                displayName: fallbackDisplayName,
                token: response.token,
                expiresAt: response.expiresAt,
                userType: "human"
            )
            return Result(centralJWT: response.token)
        } catch {
            errorMessage = "Sign in failed. Check workspace URL, email, and password."
            return nil
        }
    }
}

private struct InviteAcceptScene: View {
    enum Field: Hashable {
        case displayName
        case password
        case confirmPassword
    }

    @EnvironmentObject private var coordinator: AppCoordinator
    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel: InviteAcceptGateViewModel
    @State private var isPresentingSignIn = false
    @State private var isPresentingSignUp = false

    init(pendingInvite: AppCoordinator.PendingInvite?, sessionStore: AppSessionStore) {
        _viewModel = StateObject(wrappedValue: InviteAcceptGateViewModel(pendingInvite: pendingInvite, sessionStore: sessionStore))
    }

    var body: some View {
        Group {
            if viewModel.hasPendingInvite {
                inviteGateBody
            } else {
                noTokenHoldingBody
            }
        }
        .navigationTitle("Join Workspace")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarLeading) {
                Button("Cancel") { dismiss() }
                    .foregroundStyle(Color.auEmerald)
            }
        }
        .task {
            await viewModel.loadInvite()
            await viewModel.tryAutoAcceptWithStoredCentralJWT(coordinator: coordinator)
        }
        .fullScreenCover(isPresented: $isPresentingSignIn) {
            NavigationStack {
                SignInScene(
                    sessionStore: viewModel.sessionStore,
                    mode: .central,
                    onOpenSignUp: {
                        isPresentingSignIn = false
                        Task { try? await Task.sleep(nanoseconds: 200_000_000); isPresentingSignUp = true }
                    },
                    onAuthenticated: { jwt in
                        isPresentingSignIn = false
                        Task { await viewModel.acceptInvite(withCentralJWT: jwt, coordinator: coordinator) }
                    }
                )
            }
            .tint(.auEmerald)
        }
        .fullScreenCover(isPresented: $isPresentingSignUp) {
            NavigationStack {
                SignUpScene(
                    onOpenSignIn: {
                        isPresentingSignUp = false
                        Task { try? await Task.sleep(nanoseconds: 200_000_000); isPresentingSignIn = true }
                    },
                    onRegistered: { _, jwt in
                        isPresentingSignUp = false
                        Task { await viewModel.acceptInvite(withCentralJWT: jwt, coordinator: coordinator) }
                    }
                )
            }
            .tint(.auEmerald)
        }
    }

    // MARK: - Sub-views

    private var noTokenHoldingBody: some View {
        VStack(alignment: .leading, spacing: 20) {
            Text("Check your invite")
                .font(.title.bold())

            Text("Your agent will send you an invite link via message or email. Open that link to join their workspace.")
                .font(.body)
                .foregroundStyle(.secondary)

            Spacer(minLength: 0)

            Button {
                dismiss()
            } label: {
                Label("Back", systemImage: "chevron.left")
                    .frame(maxWidth: .infinity)
            }
            .buttonStyle(AUSecondaryButtonStyle())
        }
        .padding(.horizontal, 20)
        .padding(.top, 20)
        .padding(.bottom, 24)
    }

    private var inviteGateBody: some View {
        VStack(alignment: .leading, spacing: 20) {
            Text("Join Workspace")
                .font(.title.bold())

            if let details = viewModel.inviteDetails {
                Text("\(details.email) invited by \(details.inviter ?? "your agent")")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }

            if viewModel.isSubmitting {
                HStack(spacing: 10) {
                    ProgressView()
                    Text("Connecting your account…")
                        .font(.body)
                }
            }

            if let error = viewModel.errorMessage {
                Text(error)
                    .font(.footnote)
                    .foregroundStyle(.red)
            }

            Button("Sign in to Agent United") {
                isPresentingSignIn = true
            }
            .buttonStyle(AUPrimaryButtonStyle())

            Button("Create an account") {
                isPresentingSignUp = true
            }
            .buttonStyle(AUGhostButtonStyle())

            Spacer(minLength: 0)
        }
        .padding(.horizontal, 20)
        .padding(.top, 20)
    }
}

@MainActor
private final class InviteAcceptGateViewModel: ObservableObject {
    @Published var inviteDetails: OnboardingInviteDetails?
    @Published var isSubmitting = false
    @Published var errorMessage: String?

    let sessionStore: AppSessionStore
    private let pendingInvite: AppCoordinator.PendingInvite?
    private let keychain = KeychainHelper()

    /// True only when we arrived from a real deep link with an invite token.
    var hasPendingInvite: Bool {
        guard let token = pendingInvite?.token else { return false }
        return !token.isEmpty
    }

    init(pendingInvite: AppCoordinator.PendingInvite?, sessionStore: AppSessionStore) {
        self.pendingInvite = pendingInvite
        self.sessionStore = sessionStore
    }

    func loadInvite() async {
        guard let pendingInvite, let instanceURL = pendingInvite.instanceURL else { return }
        do {
            let response = try await LiveOnboardingService().validateInvite(token: pendingInvite.token, instanceURL: instanceURL)
            inviteDetails = OnboardingInviteDetails(email: response.email, inviter: response.inviter, suggestedDisplayName: response.displayName)
        } catch {
            errorMessage = "Unable to validate invite."
        }
    }

    func tryAutoAcceptWithStoredCentralJWT(coordinator: AppCoordinator) async {
        if let jwt = try? keychain.readJWT(for: "au.central.jwt"), !jwt.isEmpty {
            await acceptInvite(withCentralJWT: jwt, coordinator: coordinator)
        }
    }

    func acceptInvite(withCentralJWT jwt: String, coordinator: AppCoordinator) async {
        guard let pendingInvite, let instanceURL = pendingInvite.instanceURL else { return }
        isSubmitting = true
        errorMessage = nil
        defer { isSubmitting = false }
        do {
            try keychain.storeJWT(jwt, for: "au.central.jwt")
            let client = LiveAUAPIClient(instanceURL: instanceURL)
            let response = try await client.acceptInvite(token: pendingInvite.token, centralJWT: jwt)
            try sessionStore.persistInviteSession(
                instanceURL: instanceURL,
                userID: response.userID,
                email: inviteDetails?.email ?? "user@agentunited.ai",
                displayName: inviteDetails?.email.components(separatedBy: "@").first ?? "User",
                token: response.token,
                expiresAt: response.expiresAt,
                userType: "human"
            )
            coordinator.setAuthenticated(true)
        } catch {
            errorMessage = "Failed to join workspace."
        }
    }
}

private struct InviteAcceptScreen: View {
    @EnvironmentObject private var coordinator: AppCoordinator

    let phase: OnboardingViewModel.Phase
    let inviteDetails: OnboardingInviteDetails?
    @Binding var displayName: String
    @Binding var password: String
    @Binding var confirmPassword: String
    let displayNameError: String?
    let passwordError: String?
    let confirmPasswordError: String?
    let passwordCountText: String
    let canSubmit: Bool
    let isSubmitting: Bool
    let submissionErrorMessage: String?
    let focusedField: FocusState<InviteAcceptScene.Field?>.Binding
    let onLoad: () async -> Void
    let onRetry: () async -> Void
    let onConfirmBlur: () -> Void
    let onSubmit: () async -> Void

    var body: some View {
        Group {
            switch phase {
            case .idle:
                inviteIdleState
            case .loading:
                loadingState
            case .valid:
                validState
            case let .error(errorState):
                errorStateView(errorState)
            }
        }
        .navigationTitle("Join Workspace")
        .navigationBarTitleDisplayMode(.inline)
        .task {
            await onLoad()
        }
    }

    private var inviteIdleState: some View {
        ContentUnavailableView(
            "Open Your Invite Link",
            systemImage: "link.badge.plus",
            description: Text("Tap the invite link your agent sent to continue.")
        )
        .padding(.horizontal, 24)
    }

    private var loadingState: some View {
        VStack(spacing: 16) {
            ProgressView()
                .controlSize(.large)
            Text("Validating your invite…")
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color.auGrouped)
    }

    private var validState: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                VStack(spacing: 16) {
                    ZStack {
                        Circle()
                            .fill(Color.auEmeraldLight)
                            .frame(width: 76, height: 76)

                        Text(inviterInitial)
                            .font(.system(size: 30, weight: .semibold))
                            .foregroundStyle(Color.auEmerald)
                    }
                    .accessibilityHidden(true)

                    VStack(spacing: 4) {
                        Text(inviteDetails?.email ?? "")
                            .font(.subheadline.weight(.medium))
                            .foregroundStyle(Color.auSecondaryLabel)
                        Text("\(inviteDetails?.inviter ?? "Your agent") invited you to join")
                            .font(.caption)
                            .foregroundStyle(Color.auSecondaryLabel)
                    }
                }
                .frame(maxWidth: .infinity)
                .padding(.top, 16)

                VStack(alignment: .leading, spacing: 16) {
                    formField(
                        title: "Display name",
                        prompt: "How should we call you?",
                        text: $displayName,
                        error: displayNameError,
                        accessibilityLabel: "Display name",
                        textContentType: .name,
                        submitLabel: .next,
                        isSecure: false,
                        field: .displayName
                    )

                    VStack(alignment: .leading, spacing: 8) {
                        HStack {
                            Text("Password")
                            Spacer()
                            Text(passwordCountText)
                                .font(.caption.monospacedDigit())
                                .foregroundStyle(password.count >= 12 ? Color.auEmerald : Color.auSecondaryLabel)
                        }
                        .foregroundStyle(Color.auSecondaryLabel)

                        formField(
                            title: nil,
                            prompt: "Minimum 12 characters",
                            text: $password,
                            error: passwordError,
                            accessibilityLabel: "Password",
                            textContentType: .newPassword,
                            submitLabel: .next,
                            isSecure: true,
                            field: .password
                        )
                    }

                    formField(
                        title: "Confirm password",
                        prompt: "Repeat your password",
                        text: $confirmPassword,
                        error: confirmPasswordError,
                        accessibilityLabel: "Confirm password",
                        textContentType: .newPassword,
                        submitLabel: canSubmit ? .join : .done,
                        isSecure: true,
                        field: .confirmPassword
                    )
                }

                if let submissionErrorMessage, submissionErrorMessage.isEmpty == false {
                    Text(submissionErrorMessage)
                        .font(.caption)
                        .foregroundStyle(Color.red)
                        .accessibilityLabel("Join workspace error: \(submissionErrorMessage)")
                }

                Button {
                    Task {
                        await onSubmit()
                    }
                } label: {
                    if isSubmitting {
                        ProgressView()
                            .frame(maxWidth: .infinity)
                    } else {
                        Text("Join workspace")
                    }
                }
                .buttonStyle(AUPrimaryButtonStyle())
                .disabled(canSubmit == false)
                .accessibilityLabel("Join workspace")
            }
            .padding(24)
        }
        .background(Color.auGrouped.ignoresSafeArea())
    }

    private func errorStateView(_ error: OnboardingViewModel.ErrorState) -> some View {
        VStack(spacing: 16) {
            Spacer()

            Image(systemName: "exclamationmark.triangle.fill")
                .font(.system(size: 36))
                .foregroundStyle(Color.orange)
                .accessibilityHidden(true)

            switch error {
            case .invalidInvite:
                Text("This invite link has expired or already been used.")
                    .font(.body.weight(.semibold))
                    .multilineTextAlignment(.center)
                Text("Ask your agent for a new invite link to join this workspace.")
                    .font(.subheadline)
                    .foregroundStyle(Color.auSecondaryLabel)
                    .multilineTextAlignment(.center)

                Button("Contact your agent") {
                    coordinator.isPresentingInvite = false
                }
                .buttonStyle(AUSecondaryButtonStyle())
                .accessibilityLabel("Contact your agent")

            case let .network(message):
                Text("We couldn’t validate this invite.")
                    .font(.body.weight(.semibold))
                    .multilineTextAlignment(.center)
                Text(message)
                    .font(.subheadline)
                    .foregroundStyle(Color.auSecondaryLabel)
                    .multilineTextAlignment(.center)

                Button("Retry") {
                    Task {
                        await onRetry()
                    }
                }
                .buttonStyle(AUPrimaryButtonStyle())
                .accessibilityLabel("Retry invite validation")
            }

            Spacer()
        }
        .padding(24)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color.auGrouped)
    }

    private var inviterInitial: String {
        let source = inviteDetails?.inviter?.trimmingCharacters(in: .whitespacesAndNewlines)
        return source?.first.map { String($0).uppercased() } ?? "A"
    }

    @ViewBuilder
    private func formField(
        title: String?,
        prompt: String,
        text: Binding<String>,
        error: String?,
        accessibilityLabel: String,
        textContentType: UITextContentType?,
        submitLabel: SubmitLabel,
        isSecure: Bool,
        field: InviteAcceptScene.Field
    ) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            if let title {
                Text(title)
                    .font(.subheadline.weight(.medium))
                    .foregroundStyle(Color.auSecondaryLabel)
            }

            Group {
                if isSecure {
                    SecureField(prompt, text: text)
                } else {
                    TextField(prompt, text: text)
                        .textInputAutocapitalization(.words)
                }
            }
            .textContentType(textContentType)
            .submitLabel(submitLabel)
            .focused(focusedField, equals: field)
            .onSubmit {
                switch field {
                case .displayName:
                    focusedField.wrappedValue = .password
                case .password:
                    focusedField.wrappedValue = .confirmPassword
                case .confirmPassword:
                    onConfirmBlur()
                    if canSubmit {
                        Task {
                            await onSubmit()
                        }
                    }
                }
            }
            .padding(.horizontal, 14)
            .frame(height: 50)
            .background(Color.auSecondary, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
            .overlay {
                RoundedRectangle(cornerRadius: 12, style: .continuous)
                    .stroke(error == nil ? Color.auSeparator.opacity(0.55) : Color.red.opacity(0.45), lineWidth: 1)
            }
            .accessibilityLabel(accessibilityLabel)

            if let error {
                Text(error)
                    .font(.caption)
                    .foregroundStyle(Color.red)
                    .accessibilityLabel("\(accessibilityLabel) error: \(error)")
            }
        }
    }
}

struct WelcomePattern: View {
    var body: some View {
        GeometryReader { geometry in
            let width = geometry.size.width
            let height = geometry.size.height

            ZStack {
                Circle()
                    .fill(Color.white.opacity(0.06))
                    .frame(width: width * 0.72)
                    .blur(radius: 24)
                    .offset(x: width * 0.28, y: -height * 0.22)

                RoundedRectangle(cornerRadius: 32, style: .continuous)
                    .fill(Color.auEmerald.opacity(0.12))
                    .frame(width: width * 0.84, height: 220)
                    .rotationEffect(.degrees(-14))
                    .offset(x: -width * 0.16, y: -height * 0.08)

                RoundedRectangle(cornerRadius: 28, style: .continuous)
                    .stroke(Color.white.opacity(0.08), lineWidth: 1)
                    .frame(width: width * 0.7, height: 180)
                    .rotationEffect(.degrees(18))
                    .offset(x: width * 0.18, y: height * 0.2)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .allowsHitTesting(false)
    }
}

private struct SignInFieldSection<Content: View>: View {
    let title: String
    let content: Content

    init(title: String, @ViewBuilder content: () -> Content) {
        self.title = title
        self.content = content()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.subheadline.weight(.medium))
                .foregroundStyle(Color.auSecondaryLabel)

            content
                .padding(.horizontal, 14)
                .frame(height: 50)
                .background(Color.auSecondary, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
        }
    }
}
