import Foundation

enum OnboardingServiceError: Error, Equatable {
    case invalidInvite
    case network(String)
}

@MainActor
protocol OnboardingServicing {
    func validateInvite(token: String, instanceURL: URL) async throws -> InviteValidationResponse
    func acceptInvite(
        token: String,
        displayName: String,
        password: String,
        instanceURL: URL
    ) async throws -> InviteAcceptResponse
}

@MainActor
struct LiveOnboardingService: OnboardingServicing {
    func validateInvite(token: String, instanceURL: URL) async throws -> InviteValidationResponse {
        do {
            let client = LiveAUAPIClient(instanceURL: instanceURL)
            return try await client.validateInvite(token: token)
        } catch let error as AUAPIError {
            throw map(error)
        } catch {
            throw OnboardingServiceError.network(error.localizedDescription)
        }
    }

    func acceptInvite(
        token: String,
        displayName: String,
        password: String,
        instanceURL: URL
    ) async throws -> InviteAcceptResponse {
        do {
            let client = LiveAUAPIClient(instanceURL: instanceURL)
            return try await client.acceptInvite(token: token, displayName: displayName, password: password)
        } catch let error as AUAPIError {
            throw map(error)
        } catch {
            throw OnboardingServiceError.network(error.localizedDescription)
        }
    }

    private func map(_ error: AUAPIError) -> OnboardingServiceError {
        switch error {
        case .notFound:
            return .invalidInvite
        default:
            return .network(error.localizedDescription)
        }
    }
}
