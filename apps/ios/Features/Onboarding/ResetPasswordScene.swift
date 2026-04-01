import SwiftUI

struct ResetPasswordScene: View {
    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel: ResetPasswordViewModel
    @FocusState private var focus: Field?
    @State private var showNewPassword = false
    @State private var showConfirmPassword = false

    let onSuccess: () -> Void

    private enum Field { case newPassword, confirmPassword }

    init(token: String, email: String, onSuccess: @escaping () -> Void) {
        self.onSuccess = onSuccess
        _viewModel = StateObject(wrappedValue: ResetPasswordViewModel(token: token, email: email))
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
                        // Hidden username field for iOS password manager account association.
                        // .frame(1,1) + near-zero opacity: invisible visually but visible to
                        // UIKit's credential system. Must NOT be accessibilityHidden(true).
                        TextField("", text: .constant(viewModel.email))
                            .textContentType(.username)
                            .textInputAutocapitalization(.never)
                            .autocorrectionDisabled(true)
                            .frame(width: 1, height: 1)
                            .opacity(0.001)
                            .accessibilityHidden(false)

                        darkField(title: "New password") {
                            HStack {
                                Group {
                                    if showNewPassword {
                                        TextField("Minimum 8 characters", text: $viewModel.password)
                                            .textContentType(.newPassword)
                                    } else {
                                        SecureField("Minimum 8 characters", text: $viewModel.password)
                                            .textContentType(.newPassword)
                                    }
                                }
                                .textInputAutocapitalization(.never)
                                .autocorrectionDisabled(true)
                                .focused($focus, equals: .newPassword)
                                .submitLabel(.next)
                                .onSubmit { focus = .confirmPassword }
                                .accessibilityIdentifier("reset-password-field")

                                Button {
                                    showNewPassword.toggle()
                                } label: {
                                    Image(systemName: showNewPassword ? "eye.slash" : "eye")
                                        .foregroundStyle(Color.white.opacity(0.6))
                                }
                                .accessibilityLabel(showNewPassword ? "Hide password" : "Show password")
                            }
                        }

                        VStack(alignment: .leading, spacing: 6) {
                            darkField(title: "Confirm password") {
                                HStack {
                                    Group {
                                        if showConfirmPassword {
                                            TextField("Confirm password", text: $viewModel.confirmPassword)
                                                .textContentType(.newPassword)
                                        } else {
                                            SecureField("Confirm password", text: $viewModel.confirmPassword)
                                                .textContentType(.newPassword)
                                        }
                                    }
                                    .textInputAutocapitalization(.never)
                                    .autocorrectionDisabled(true)
                                    .focused($focus, equals: .confirmPassword)
                                    .submitLabel(.done)
                                    .onSubmit {
                                        if viewModel.canSubmit {
                                            Task {
                                                if await viewModel.submit() {
                                                    // Brief pause: lets iOS credential-save sheet
                                                    // appear before the view leaves the responder chain.
                                                    try? await Task.sleep(nanoseconds: 150_000_000)
                                                    dismiss()
                                                    onSuccess()
                                                }
                                            }
                                        }
                                    }
                                    .accessibilityIdentifier("reset-confirm-password-field")

                                    Button {
                                        showConfirmPassword.toggle()
                                    } label: {
                                        Image(systemName: showConfirmPassword ? "eye.slash" : "eye")
                                            .foregroundStyle(Color.white.opacity(0.6))
                                    }
                                    .accessibilityLabel(showConfirmPassword ? "Hide password" : "Show password")
                                }
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
                                    // Brief pause: lets iOS credential-save sheet
                                    // appear before the view leaves the responder chain.
                                    try? await Task.sleep(nanoseconds: 150_000_000)
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

    let email: String
    private let token: String

    init(token: String, email: String) {
        self.token = token
        self.email = email
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
