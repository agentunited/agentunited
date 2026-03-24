import SwiftUI

struct SignUpScene: View {
    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel: SignUpViewModel
    let onOpenSignIn: () -> Void
    let onRegistered: (String, String) -> Void

    /// Pass `needsClaimKey: false` in relay-first flows — just register, skip claim key generation.
    init(needsClaimKey: Bool = true, onOpenSignIn: @escaping () -> Void, onRegistered: @escaping (String, String) -> Void) {
        _viewModel = StateObject(wrappedValue: SignUpViewModel(needsClaimKey: needsClaimKey))
        self.onOpenSignIn = onOpenSignIn
        self.onRegistered = onRegistered
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Set up your workspace")
                        .font(.title.bold())
                    Text("Create your Agent United account to generate a claim key.")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                VStack(alignment: .leading, spacing: 16) {
                    // MARK: Name
                    fieldLabel("Display name")
                    TextField("Your name", text: $viewModel.displayName)
                        .textContentType(.name)
                        .textInputAutocapitalization(.words)
                        .autocorrectionDisabled(true)
                        .styled()
                        .accessibilityIdentifier("signup-display-name-field")

                    // MARK: Email
                    fieldLabel("Email")
                    TextField("Email", text: $viewModel.email)
                        .textContentType(.emailAddress)
                        .keyboardType(.emailAddress)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled(true)
                        .styled()
                        .accessibilityIdentifier("signup-email-field")

                    // MARK: Password
                    fieldLabel("Password")
                    SecureField("Minimum 8 characters", text: $viewModel.password)
                        .textContentType(.newPassword)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled(true)
                        .styled()
                        .accessibilityIdentifier("signup-password-field")

                    // MARK: Confirm password
                    fieldLabel("Confirm password")
                    SecureField("Confirm password", text: $viewModel.confirmPassword)
                        .textContentType(.newPassword)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled(true)
                        .styled()
                        .accessibilityIdentifier("signup-confirm-password-field")

                    if viewModel.showPasswordMismatch {
                        Text("Passwords don't match")
                            .font(.footnote)
                            .foregroundStyle(.red)
                    }
                }

                if let errorMessage = viewModel.errorMessage {
                    Text(errorMessage)
                        .font(.footnote)
                        .foregroundStyle(.red)
                }

                Button {
                    Task {
                        let result = await viewModel.register()
                        if let result {
                            onRegistered(result.claimKey, result.centralJWT)
                        }
                    }
                } label: {
                    if viewModel.isSubmitting {
                        ProgressView().frame(maxWidth: .infinity)
                    } else {
                        Text("Create account")
                    }
                }
                .buttonStyle(AUPrimaryButtonStyle())
                .disabled(!viewModel.canSubmit)
                .accessibilityIdentifier("signup-submit-button")

                Divider().padding(.vertical, 4)

                HStack(spacing: 6) {
                    Text("Already have an account?")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                    Button("Sign in") {
                        dismiss()
                        onOpenSignIn()
                    }
                    .font(.footnote.weight(.semibold))
                    .foregroundStyle(Color.auEmerald)
                }
            }
            .padding(.horizontal, 20)
            .padding(.top, 20)
            .padding(.bottom, 28)
        }
        .navigationTitle("Create Account")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarLeading) {
                Button("Cancel") { dismiss() }
                    .foregroundStyle(Color.auEmerald)
            }
        }
    }

    private func fieldLabel(_ text: String) -> some View {
        Text(text)
            .font(.subheadline.weight(.medium))
            .foregroundStyle(Color.auSecondaryLabel)
    }
}

// MARK: - Field style helper

private extension View {
    func styled() -> some View {
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
final class SignUpViewModel: ObservableObject {
    struct Result {
        let claimKey: String
        let centralJWT: String
    }

    @Published var displayName = ""
    @Published var email = ""
    @Published var password = ""
    @Published var confirmPassword = ""
    @Published var isSubmitting = false
    @Published var errorMessage: String?

    private let needsClaimKey: Bool

    init(needsClaimKey: Bool = true) {
        self.needsClaimKey = needsClaimKey
    }

    /// True only when confirmPassword is non-empty and doesn't match password.
    var showPasswordMismatch: Bool {
        !confirmPassword.isEmpty && confirmPassword != password
    }

    var canSubmit: Bool {
        !displayName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty &&
        email.trimmingCharacters(in: .whitespacesAndNewlines).contains("@") &&
        password.count >= 8 &&
        confirmPassword == password &&
        !isSubmitting
    }

    func register() async -> Result? {
        errorMessage = nil
        guard canSubmit else {
            errorMessage = "Please complete all fields correctly."
            return nil
        }

        isSubmitting = true
        defer { isSubmitting = false }

        do {
            let client = CentralAPIClient()
            let authResp = try await client.register(
                email: email.trimmingCharacters(in: .whitespacesAndNewlines),
                password: password,
                displayName: displayName.trimmingCharacters(in: .whitespacesAndNewlines)
            )

            let keychain = KeychainHelper()
            try? keychain.storeJWT(authResp.token, for: "au.central.jwt")

            guard needsClaimKey else {
                return Result(claimKey: "", centralJWT: authResp.token)
            }

            let claimClient = CentralAPIClient(authToken: authResp.token)
            let claimResp = try await claimClient.generateClaimKey()

            return Result(claimKey: claimResp.claimKey, centralJWT: authResp.token)
        } catch let e as CentralAPIError {
            errorMessage = e.localizedDescription
            return nil
        } catch {
            errorMessage = "Registration failed. Please try again."
            return nil
        }
    }
}
