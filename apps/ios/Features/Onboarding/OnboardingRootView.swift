import SwiftUI

struct OnboardingRootView: View {
    @EnvironmentObject private var coordinator: AppCoordinator
    @EnvironmentObject private var sessionStore: AppSessionStore

    private var pendingInvite: AppCoordinator.PendingInvite? {
        if case let .invite(invite) = coordinator.pendingRoute {
            return invite
        }
        return nil
    }

    var body: some View {
        NavigationStack {
            WelcomeScreen(
                isInvitePresented: $coordinator.isPresentingInvite,
                sessionStore: sessionStore
            )
            .navigationDestination(isPresented: $coordinator.isPresentingInvite) {
                InviteAcceptScene(
                    pendingInvite: pendingInvite,
                    sessionStore: sessionStore
                )
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color(hex: 0x0F172A).ignoresSafeArea())
        .onAppear {
            if pendingInvite != nil {
                coordinator.isPresentingInvite = true
            }
        }
        .onChange(of: coordinator.pendingRoute) { _, newValue in
            if case .invite = newValue {
                coordinator.isPresentingInvite = true
            }
        }
    }
}

private struct WelcomeScreen: View {
    @Binding var isInvitePresented: Bool
    let sessionStore: AppSessionStore

    var body: some View {
        ZStack {
            LinearGradient(
                colors: [Color(hex: 0x064E3B), Color(hex: 0x0F172A)],
                startPoint: .top,
                endPoint: .bottom
            )
            .ignoresSafeArea()

            VStack(spacing: 0) {
                Spacer()

                VStack(spacing: 16) {
                    ZStack {
                        Circle()
                            .fill(Color.auEmerald.opacity(0.18))
                            .frame(width: 88, height: 88)

                        Image(systemName: "message.badge.waveform.fill")
                            .font(.system(size: 38, weight: .semibold))
                            .foregroundStyle(Color.auEmerald)
                    }
                    .accessibilityHidden(true)

                    VStack(spacing: 8) {
                        Text("Agent United")
                            .font(.system(size: 30, weight: .bold))
                            .foregroundStyle(Color.white)

                        Text("Your workspace, everywhere.")
                            .font(.system(size: 16))
                            .foregroundStyle(Color.white.opacity(0.6))
                    }
                }

                Spacer()

                VStack(spacing: 12) {
                    Button("Accept an invite") {
                        isInvitePresented = true
                    }
                    .buttonStyle(AUPrimaryButtonStyle())
                    .accessibilityIdentifier("accept-invite-button")

                    NavigationLink {
                        SignInScene(sessionStore: sessionStore)
                    } label: {
                        Text("Sign in")
                    }
                    .buttonStyle(AUGhostButtonStyle())
                    .accessibilityIdentifier("sign-in-button")
                }
                .padding(.horizontal, 24)
                .padding(.bottom, 40)
            }
        }
        .navigationBarHidden(true)
    }
}

private struct SignInScene: View {
    @EnvironmentObject private var coordinator: AppCoordinator
    @StateObject private var viewModel: SignInViewModel

    init(sessionStore: AppSessionStore) {
        _viewModel = StateObject(wrappedValue: SignInViewModel(sessionStore: sessionStore))
    }

    var body: some View {
        Form {
            Section("Workspace") {
                TextField("https://workspace.example.com", text: $viewModel.workspaceURL)
                    .keyboardType(.URL)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled(true)
                    .accessibilityIdentifier("workspace-url-field")
            }

            Section("Credentials") {
                TextField("Email", text: $viewModel.email)
                    .keyboardType(.emailAddress)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled(true)
                    .accessibilityIdentifier("email-field")

                SecureField("Password", text: $viewModel.password)
                    .accessibilityIdentifier("password-field")
            }

            if let errorMessage = viewModel.errorMessage {
                Section {
                    Text(errorMessage)
                        .foregroundStyle(.red)
                        .font(.footnote)
                }
            }

            Section {
                Button {
                    Task {
                        let success = await viewModel.signIn()
                        if success {
                            coordinator.setAuthenticated(true)
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
                .disabled(viewModel.canSubmit == false)
                .accessibilityIdentifier("submit-sign-in-button")
            }
        }
        .navigationTitle("Sign In")
        .navigationBarTitleDisplayMode(.large)
        .toolbarColorScheme(.dark, for: .navigationBar)
        .toolbarBackground(Color(hex: 0x064E3B), for: .navigationBar)
        .toolbarBackground(.visible, for: .navigationBar)
        .tint(Color.auEmerald)
    }
}

@MainActor
private final class SignInViewModel: ObservableObject {
    @Published var workspaceURL: String = ""
    @Published var email: String = ""
    @Published var password: String = ""
    @Published var isSubmitting: Bool = false
    @Published var errorMessage: String?

    private let sessionStore: AppSessionStore

    init(sessionStore: AppSessionStore) {
        self.sessionStore = sessionStore
    }

    var canSubmit: Bool {
        workspaceURL.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty == false
            && email.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty == false
            && password.isEmpty == false
            && isSubmitting == false
    }

    func signIn() async -> Bool {
        errorMessage = nil

        guard let url = URL(string: workspaceURL.trimmingCharacters(in: .whitespacesAndNewlines)) else {
            errorMessage = "Enter a valid workspace URL."
            return false
        }

        isSubmitting = true
        defer { isSubmitting = false }

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
            return true
        } catch {
#if DEBUG
            NSLog("[AU][signin] failed: %@", String(describing: error))
#endif
            errorMessage = "Sign in failed. Check workspace URL, email, and password."
            return false
        }
    }
}

private struct InviteAcceptScene: View {
    @EnvironmentObject private var coordinator: AppCoordinator
    @StateObject private var viewModel: OnboardingViewModel
    @FocusState private var focusedField: Field?

    enum Field: Hashable {
        case displayName
        case password
        case confirmPassword
    }

    init(pendingInvite: AppCoordinator.PendingInvite?, sessionStore: AppSessionStore) {
        _viewModel = StateObject(
            wrappedValue: OnboardingViewModel(
                pendingInvite: pendingInvite,
                service: LiveOnboardingService(),
                sessionStore: sessionStore
            )
        )
    }

    var body: some View {
        InviteAcceptScreen(
            phase: viewModel.phase,
            inviteDetails: viewModel.inviteDetails,
            displayName: $viewModel.displayName,
            password: $viewModel.password,
            confirmPassword: $viewModel.confirmPassword,
            displayNameError: viewModel.displayNameError,
            passwordError: viewModel.passwordError,
            confirmPasswordError: viewModel.confirmPasswordError,
            passwordCountText: viewModel.passwordCountText,
            canSubmit: viewModel.canSubmit,
            isSubmitting: viewModel.isSubmitting,
            submissionErrorMessage: viewModel.submissionErrorMessage,
            focusedField: $focusedField,
            onLoad: {
                await viewModel.loadInviteIfNeeded()
            },
            onRetry: {
                await viewModel.retry()
            },
            onConfirmBlur: {
                viewModel.confirmPasswordBlurred()
            },
            onSubmit: {
                await viewModel.submit()
            }
        )
        .environmentObject(coordinator)
        .onAppear {
            viewModel.attachCoordinator(coordinator)
        }
        .onChange(of: focusedField) { oldValue, newValue in
            if oldValue == .confirmPassword, newValue != .confirmPassword {
                viewModel.confirmPasswordBlurred()
            }
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
                VStack(spacing: 14) {
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
                            .font(.footnote)
                            .foregroundStyle(Color.auSecondaryLabel)
                    }
                }
                .frame(maxWidth: .infinity)
                .padding(.top, 12)

                VStack(alignment: .leading, spacing: 18) {
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
                                .font(.footnote.monospacedDigit())
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
                        .font(.footnote)
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
                    .font(.title3.weight(.semibold))
                    .multilineTextAlignment(.center)
                Text("Ask your agent for a new invite link to join this workspace.")
                    .foregroundStyle(Color.auSecondaryLabel)
                    .multilineTextAlignment(.center)

                Button("Contact your agent") {
                    coordinator.isPresentingInvite = false
                }
                .buttonStyle(AUSecondaryButtonStyle())
                .accessibilityLabel("Contact your agent")

            case let .network(message):
                Text("We couldn’t validate this invite.")
                    .font(.title3.weight(.semibold))
                    .multilineTextAlignment(.center)
                Text(message)
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
            .padding(.vertical, 14)
            .background(Color.auBackground, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
            .overlay {
                RoundedRectangle(cornerRadius: 14, style: .continuous)
                    .stroke(error == nil ? Color.auSeparator.opacity(0.55) : Color.red.opacity(0.45), lineWidth: 1)
            }
            .accessibilityLabel(accessibilityLabel)

            if let error {
                Text(error)
                    .font(.footnote)
                    .foregroundStyle(Color.red)
                    .accessibilityLabel("\(accessibilityLabel) error: \(error)")
            }
        }
    }
}
