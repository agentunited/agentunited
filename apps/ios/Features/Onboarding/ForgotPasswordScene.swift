import SwiftUI

struct ForgotPasswordScene: View {
    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel = ForgotPasswordViewModel()

    let onSuccessEmailCaptured: (String) -> Void

    init(onSuccessEmailCaptured: ((String) -> Void)? = nil) {
        self.onSuccessEmailCaptured = onSuccessEmailCaptured ?? { _ in }
    }

    var body: some View {
        ZStack {
            LinearGradient(
                colors: [Color(hex: 0x064E3B), Color(hex: 0x0F172A)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
            .ignoresSafeArea()

            ScrollView {
                VStack(alignment: .leading, spacing: 28) {
                    Spacer(minLength: 0)

                    VStack(alignment: .leading, spacing: 10) {
                        Text("Forgot password")
                            .font(.title.bold())
                            .foregroundStyle(.white)
                        Text("Enter your email and we'll send you a reset link.")
                            .font(.subheadline)
                            .foregroundStyle(Color.white.opacity(0.75))
                    }

                    if viewModel.didSend {
                        successView
                    } else {
                        formView
                    }
                }
                .padding(.horizontal, 24)
                .padding(.top, 40)
                .padding(.bottom, 40)
            }
        }
        .navigationBarHidden(true)
        .overlay(alignment: .topLeading) {
            Button {
                dismiss()
            } label: {
                Image(systemName: "xmark")
                    .font(.body.weight(.semibold))
                    .foregroundStyle(.white)
                    .padding(12)
                    .background(Color.white.opacity(0.12))
                    .clipShape(Circle())
            }
            .padding(.top, 16)
            .padding(.leading, 20)
            .accessibilityLabel("Close")
        }
    }

    // MARK: - Form

    private var formView: some View {
        VStack(alignment: .leading, spacing: 20) {
            VStack(alignment: .leading, spacing: 6) {
                Text("Email")
                    .font(.footnote.weight(.semibold))
                    .foregroundStyle(Color.white.opacity(0.65))
                TextField("Email", text: $viewModel.email)
                    .keyboardType(.emailAddress)
                    .textContentType(.emailAddress)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled(true)
                    .font(.body)
                    .padding(.horizontal, 14)
                    .frame(height: 50)
                    .background(Color.white.opacity(0.10))
                    .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    .overlay(
                        RoundedRectangle(cornerRadius: 12, style: .continuous)
                            .stroke(Color.white.opacity(0.18), lineWidth: 1)
                    )
                    .foregroundStyle(.white)
                    .tint(.auEmerald)
                    .accessibilityIdentifier("forgot-password-email-field")
            }

            if let error = viewModel.errorMessage {
                Text(error)
                    .font(.subheadline)
                    .foregroundStyle(Color.red.opacity(0.9))
            }

            Button {
                Task {
                    await viewModel.send()
                    if viewModel.didSend {
                        onSuccessEmailCaptured(viewModel.normalizedEmail)
                    }
                }
            } label: {
                if viewModel.isSubmitting {
                    ProgressView()
                        .tint(.black)
                        .frame(maxWidth: .infinity)
                } else {
                    Text("Send reset link")
                        .frame(maxWidth: .infinity)
                }
            }
            .buttonStyle(AUPrimaryButtonStyle())
            .disabled(!viewModel.canSubmit)
            .accessibilityIdentifier("send-reset-link-button")

            Button {
                dismiss()
            } label: {
                Text("← Back to sign in")
                    .font(.subheadline)
                    .foregroundStyle(Color.white.opacity(0.65))
            }
            .frame(maxWidth: .infinity)
            .accessibilityIdentifier("back-to-sign-in-button")
        }
    }

    // MARK: - Success state

    private var successView: some View {
        VStack(alignment: .leading, spacing: 24) {
            HStack(spacing: 12) {
                Image(systemName: "envelope.badge.fill")
                    .font(.title2)
                    .foregroundStyle(Color.auEmerald)
                Text("Check your email.")
                    .font(.headline)
                    .foregroundStyle(.white)
            }

            Text("If that address is registered, a reset link is on the way.")
                .font(.subheadline)
                .foregroundStyle(Color.white.opacity(0.75))
                .fixedSize(horizontal: false, vertical: true)

            Button {
                dismiss()
            } label: {
                Text("Back to sign in")
                    .frame(maxWidth: .infinity)
            }
            .buttonStyle(AUGhostButtonStyle())
            .accessibilityIdentifier("back-to-sign-in-success-button")
        }
        .padding(.top, 8)
    }
}

// MARK: - ViewModel

@MainActor
private final class ForgotPasswordViewModel: ObservableObject {
    @Published var email = ""
    @Published var isSubmitting = false
    @Published var errorMessage: String?
    @Published var didSend = false

    var normalizedEmail: String {
        email.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    var canSubmit: Bool {
        normalizedEmail.contains("@") && !isSubmitting
    }

    func send() async {
        errorMessage = nil
        isSubmitting = true
        defer { isSubmitting = false }
        do {
            try await CentralAPIClient().forgotPassword(email: normalizedEmail)
            didSend = true
        } catch {
            // Server is always silent on unknown email — surface only network/server errors
            if case CentralAPIError.networkError = error as? CentralAPIError ?? .networkError(error) {
                errorMessage = "Check your connection and try again."
            } else {
                // Show success anyway — don't leak whether the email is registered
                didSend = true
            }
        }
    }
}
