import SwiftUI

struct OnboardingRootView: View {
    @EnvironmentObject private var coordinator: AppCoordinator
    @EnvironmentObject private var sessionStore: AppSessionStore
    @State private var isPresentingSignIn = false
    @State private var isPresentingSetup = false
    @State private var isPresentingClaimKey = false
    @State private var claimKey: String = ""
    @State private var centralJWT: String = ""

    private var pendingInvite: AppCoordinator.PendingInvite? {
        if case let .invite(invite) = coordinator.pendingRoute {
            return invite
        }
        return nil
    }

    var body: some View {
        WelcomeScreen(
            isInvitePresented: $coordinator.isPresentingInvite,
            isSignInPresented: $isPresentingSignIn,
            isSetupPresented: $isPresentingSetup
        )
        .fullScreenCover(isPresented: $isPresentingSignIn) {
            NavigationStack {
                SignInScene(
                    sessionStore: sessionStore,
                    onOpenSignUp: {
                        isPresentingSignIn = false
                        Task {
                            try? await Task.sleep(nanoseconds: 200_000_000)
                            isPresentingSetup = true
                        }
                    }
                )
            }
            .tint(.auEmerald)
        }
        .fullScreenCover(isPresented: $isPresentingSetup) {
            NavigationStack {
                SignUpScene(
                    onOpenSignIn: {
                        isPresentingSetup = false
                        Task {
                            try? await Task.sleep(nanoseconds: 200_000_000)
                            isPresentingSignIn = true
                        }
                    },
                    onRegistered: { key, jwt in
                        claimKey = key
                        centralJWT = jwt
                        isPresentingSetup = false
                        Task {
                            try? await Task.sleep(nanoseconds: 200_000_000)
                            isPresentingClaimKey = true
                        }
                    }
                )
            }
            .tint(.auEmerald)
        }
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
        .fullScreenCover(isPresented: $coordinator.isPresentingInvite) {
            NavigationStack {
                InviteAcceptScene(
                    pendingInvite: pendingInvite,
                    sessionStore: sessionStore
                )
            }
            .tint(.auEmerald)
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
        }
    }
}

private struct WelcomeScreen: View {
    @Binding var isInvitePresented: Bool
    @Binding var isSignInPresented: Bool
    @Binding var isSetupPresented: Bool

    var body: some View {
        ZStack {
            LinearGradient(
                colors: [Color(hex: 0x064E3B), Color(hex: 0x0F172A)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
            .ignoresSafeArea()

            WelcomePattern()
                .ignoresSafeArea()

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
            VStack(spacing: 12) {
                Button("Accept an invite") {
                    isInvitePresented = true
                }
                .buttonStyle(AUPrimaryButtonStyle())
                .accessibilityLabel("Accept an invite")
                .accessibilityIdentifier("accept-invite-button")

                Button("Sign in") {
                    isSignInPresented = true
                }
                .buttonStyle(AUGhostButtonStyle())
                .accessibilityLabel("Sign in")
                .accessibilityIdentifier("sign-in-button")

                Button("Set up a new workspace") {
                    isSetupPresented = true
                }
                .buttonStyle(AUGhostButtonStyle())
                .accessibilityIdentifier("setup-workspace-button")
            }
            .padding(.horizontal, 20)
            .padding(.top, 16)
            .padding(.bottom, 16)
            .background(Color.clear)
        }
        .navigationBarHidden(true)
    }

    private var logoMark: some View {
        ZStack {
            Circle()
                .fill(Color.auEmerald.opacity(0.26))
                .frame(width: 104, height: 104)
                .blur(radius: 16)

            Circle()
                .fill(
                    LinearGradient(
                        colors: [Color.white.opacity(0.2), Color.white.opacity(0.06)],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )
                .overlay {
                    Circle()
                        .stroke(Color.white.opacity(0.18), lineWidth: 1)
                }
                .frame(width: 96, height: 96)

            Image(systemName: "message.badge.waveform.fill")
                .font(.system(size: 38, weight: .semibold))
                .foregroundStyle(.white)
        }
        .accessibilityHidden(true)
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
                        TextField("name@workspace.com", text: $viewModel.email)
                            .keyboardType(.emailAddress)
                            .textInputAutocapitalization(.never)
                            .autocorrectionDisabled(true)
                            .accessibilityIdentifier("email-field")
                    }

                    SignInFieldSection(title: "Password") {
                        SecureField("Enter your password", text: $viewModel.password)
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

                VStack(spacing: 6) {
                    Text("Don't have an account?")
                        .font(.subheadline)
                        .fontWeight(.medium)
                        .foregroundStyle(.primary)

                    Text("Accounts are created when you accept an invite from your agent. Go back and tap \"Accept an invite\".")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                }
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
}

@MainActor
private final class InviteAcceptGateViewModel: ObservableObject {
    @Published var inviteDetails: OnboardingInviteDetails?
    @Published var isSubmitting = false
    @Published var errorMessage: String?

    let sessionStore: AppSessionStore
    private let pendingInvite: AppCoordinator.PendingInvite?
    private let keychain = KeychainHelper()

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

private struct WelcomePattern: View {
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
