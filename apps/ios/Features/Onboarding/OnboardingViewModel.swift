import Foundation
import UIKit

@MainActor
protocol OnboardingRouting: AnyObject {
    func completeOnboarding(welcomeConversationID: String?)
}

extension AppCoordinator: OnboardingRouting {}

struct OnboardingInviteDetails: Equatable {
    let email: String
    let inviter: String?
    let suggestedDisplayName: String?
}

@MainActor
final class OnboardingViewModel: ObservableObject {
    enum Phase: Equatable {
        case idle
        case loading
        case valid
        case error(ErrorState)
    }

    enum ErrorState: Equatable {
        case invalidInvite
        case network(String)
    }

    @Published private(set) var phase: Phase = .idle
    @Published private(set) var inviteDetails: OnboardingInviteDetails?
    @Published var displayName: String = ""
    @Published var password: String = ""
    @Published var confirmPassword: String = ""
    @Published private(set) var confirmPasswordDidBlur = false
    @Published private(set) var isSubmitting = false
    @Published private(set) var submissionErrorMessage: String?

    private let pendingInvite: AppCoordinator.PendingInvite?
    private let service: OnboardingServicing
    private let sessionStore: SessionPersisting
    private weak var coordinator: OnboardingRouting?
    private var hasLoadedInvite = false

    init(
        pendingInvite: AppCoordinator.PendingInvite?,
        service: OnboardingServicing,
        sessionStore: SessionPersisting
    ) {
        self.pendingInvite = pendingInvite
        self.service = service
        self.sessionStore = sessionStore
    }

    func attachCoordinator(_ coordinator: OnboardingRouting) {
        self.coordinator = coordinator
    }

    var displayNameError: String? {
        Self.validateDisplayName(displayName)
    }

    var passwordError: String? {
        Self.validatePassword(password)
    }

    var confirmPasswordError: String? {
        guard confirmPasswordDidBlur else {
            return nil
        }
        return Self.validateConfirmPassword(password: password, confirmPassword: confirmPassword)
    }

    var passwordCountText: String {
        "\(password.count)/12"
    }

    var canSubmit: Bool {
        guard phase == .valid,
              isSubmitting == false,
              displayNameError == nil,
              passwordError == nil,
              Self.validateConfirmPassword(password: password, confirmPassword: confirmPassword) == nil
        else {
            return false
        }

        return true
    }

    func loadInviteIfNeeded() async {
        guard hasLoadedInvite == false else {
            return
        }

        hasLoadedInvite = true
        await loadInvite()
    }

    func retry() async {
        submissionErrorMessage = nil
        await loadInvite()
    }

    func confirmPasswordBlurred() {
        confirmPasswordDidBlur = true
    }

    func submit() async {
        guard canSubmit,
              let pendingInvite,
              let instanceURL = pendingInvite.instanceURL,
              let inviteDetails
        else {
            return
        }

        isSubmitting = true
        submissionErrorMessage = nil

        do {
            let response = try await service.acceptInvite(
                token: pendingInvite.token,
                displayName: displayName.trimmingCharacters(in: .whitespacesAndNewlines),
                password: password,
                instanceURL: instanceURL
            )

            try sessionStore.persistInviteSession(
                instanceURL: instanceURL,
                userID: response.userID,
                email: inviteDetails.email,
                displayName: displayName.trimmingCharacters(in: .whitespacesAndNewlines),
                token: response.token,
                expiresAt: response.expiresAt,
                userType: "human"
            )

            UINotificationFeedbackGenerator().notificationOccurred(.success)
            coordinator?.completeOnboarding(welcomeConversationID: response.onboarding?.welcomeDMID)
        } catch let error as OnboardingServiceError {
            switch error {
            case .invalidInvite:
                phase = .error(.invalidInvite)
            case let .network(message):
                submissionErrorMessage = message
            }
        } catch {
            submissionErrorMessage = error.localizedDescription
        }

        isSubmitting = false
    }

    static func validateDisplayName(_ value: String) -> String? {
        let trimmed = value.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.isEmpty {
            return "Display name is required."
        }
        if trimmed.count > 255 {
            return "Display name must be 255 characters or fewer."
        }
        return nil
    }

    static func validatePassword(_ value: String) -> String? {
        if value.count < 12 {
            return "Password must be at least 12 characters."
        }
        return nil
    }

    static func validateConfirmPassword(password: String, confirmPassword: String) -> String? {
        if confirmPassword.isEmpty {
            return "Confirm your password."
        }
        if password != confirmPassword {
            return "Passwords do not match."
        }
        return nil
    }

    private func loadInvite() async {
        guard let pendingInvite else {
            phase = .idle
            return
        }

        guard let instanceURL = pendingInvite.instanceURL else {
            phase = .error(.network("This invite link is missing its workspace address."))
            return
        }

        phase = .loading

        do {
            let response = try await service.validateInvite(token: pendingInvite.token, instanceURL: instanceURL)
            inviteDetails = OnboardingInviteDetails(
                email: response.email,
                inviter: response.inviter,
                suggestedDisplayName: response.displayName
            )

            if displayName.isEmpty, let suggestedDisplayName = response.displayName, suggestedDisplayName.isEmpty == false {
                displayName = suggestedDisplayName
            }

            phase = .valid
        } catch let error as OnboardingServiceError {
            switch error {
            case .invalidInvite:
                phase = .error(.invalidInvite)
            case let .network(message):
                phase = .error(.network(message))
            }
        } catch {
            phase = .error(.network(error.localizedDescription))
        }
    }
}
