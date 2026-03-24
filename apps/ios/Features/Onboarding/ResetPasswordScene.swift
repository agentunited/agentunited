import SwiftUI

struct ResetPasswordScene: View {
    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel: ResetPasswordViewModel

    let onSuccess: () -> Void

    init(token: String, onSuccess: @escaping () -> Void) {
        self.onSuccess = onSuccess
        _viewModel = StateObject(wrappedValue: ResetPasswordViewModel(token: token))
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
                        Text("Set new password")
                            .font(.title.bold())
                            .foregroundStyle(.white)
                        Text("Choose a strong password for your Agent United account.")
                            .font(.subheadline)
                            .foregroundStyle(Color.white.opacity(0.75))
                    }

                    VStack(alignment: .leading, spacing: 20) {
                        darkField(title: "New password") {
                            SecureField("Minimum 8 characters", text: $viewModel.password)
                                .textContentType(.newPassword)
                                .accessibilityIdentifier("reset-password-field")
                        }

                        VStack(alignment: .leading, spacing: 6) {
                            darkField(title: "Confirm password") {
                                SecureField("Confirm password", text: $viewModel.confirmPassword)
                                    .textContentType(.newPassword)
                                    .accessibilityIdentifier("reset-confirm-password-field")
                            }
                            if viewModel.showMismatch {
                                Text("Passwords don't match.")
                                    .font(.caption)
                                    .foregroundStyle(Color.red.opacity(0.9))
                                    .padding(.leading, 4)
                            }
                        }

                        if let error = viewModel.errorMessage {
                            Text(error)
                                .font(.subheadline)
                                .foregroundStyle(Color.red.opacity(0.9))
                        }

                        Button {
                            Task {
                                if await viewModel.submit() {
                                    dismiss()
                                    onSuccess()
                                }
                            }
                        } label: {
                            if viewModel.isSubmitting {
                                ProgressView()
                                    .tint(.black)
                                    .frame(maxWidth: .infinity)
                            } else {
                                Text("Reset password")
                                    .frame(maxWidth: .infinity)
                            }
                        }
                        .buttonStyle(AUPrimaryButtonStyle())
                        .disabled(!viewModel.canSubmit)
                        .accessibilityIdentifier("reset-password-submit-button")
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

    // MARK: - Dark field helper

    private func darkField<Content: View>(title: String, @ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(title)
                .font(.footnote.weight(.semibold))
                .foregroundStyle(Color.white.opacity(0.65))
            content()
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
        }
    }
}

// MARK: - ViewModel

@MainActor
private final class ResetPasswordViewModel: ObservableObject {
    @Published var password = ""
    @Published var confirmPassword = ""
    @Published var isSubmitting = false
    @Published var errorMessage: String?

    private let token: String

    init(token: String) {
        self.token = token
    }

    var showMismatch: Bool {
        !confirmPassword.isEmpty && confirmPassword != password
    }

    var canSubmit: Bool {
        password.count >= 8 &&
        confirmPassword == password &&
        !isSubmitting
    }

    /// Returns true on success.
    func submit() async -> Bool {
        errorMessage = nil
        isSubmitting = true
        defer { isSubmitting = false }
        do {
            try await CentralAPIClient().resetPassword(token: token, newPassword: password)
            return true
        } catch let e as CentralAPIError {
            errorMessage = e.localizedDescription
            return false
        } catch {
            errorMessage = "Something went wrong. Please try again."
            return false
        }
    }
}
