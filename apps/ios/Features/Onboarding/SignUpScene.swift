import SwiftUI

struct SignUpScene: View {
    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel = SignUpViewModel()
    let onOpenSignIn: () -> Void
    let onRegistered: (String, String) -> Void

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
                    fieldLabel("Display name")
                    textField("Your name", text: $viewModel.displayName)
                        .accessibilityIdentifier("signup-display-name-field")

                    fieldLabel("Email")
                    textField("you@example.com", text: $viewModel.email)
                        .keyboardType(.emailAddress)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled(true)
                        .accessibilityIdentifier("signup-email-field")

                    fieldLabel("Password")
                    SecureField("Minimum 12 characters", text: $viewModel.password)
                        .font(.body)
                        .padding(.horizontal, 14)
                        .frame(height: 50)
                        .background(Color.auSecondary)
                        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                        .accessibilityIdentifier("signup-password-field")
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

    private func textField(_ placeholder: String, text: Binding<String>) -> some View {
        TextField(placeholder, text: text)
            .font(.body)
            .padding(.horizontal, 14)
            .frame(height: 50)
            .background(Color.auSecondary)
            .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
    }
}

@MainActor
final class SignUpViewModel: ObservableObject {
    struct Result {
        let claimKey: String
        let centralJWT: String
    }

    @Published var displayName = ""
    @Published var email = ""
    @Published var password = ""
    @Published var isSubmitting = false
    @Published var errorMessage: String?

    var canSubmit: Bool {
        displayName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty == false &&
        email.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty == false &&
        password.count >= 12 &&
        isSubmitting == false
    }

    func register() async -> Result? {
        errorMessage = nil
        guard canSubmit else {
            errorMessage = "Please complete all fields. Password must be at least 12 characters."
            return nil
        }

        isSubmitting = true
        defer { isSubmitting = false }

        // Mock-first behavior while backend stabilizes.
        try? await Task.sleep(nanoseconds: 900_000_000)

        let key = "au_claim_\(String(UUID().uuidString.prefix(8)))"
        return Result(claimKey: key, centralJWT: "central_mock_jwt")
    }
}
