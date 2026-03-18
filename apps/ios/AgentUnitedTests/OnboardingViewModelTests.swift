import XCTest
@testable import AgentUnited

@MainActor
final class OnboardingViewModelTests: XCTestCase {
    func testIdleStateWithoutInvite() async {
        let viewModel = makeViewModel(pendingInvite: nil)

        await viewModel.loadInviteIfNeeded()

        XCTAssertEqual(viewModel.phase, .idle)
    }

    func testLoadInviteTransitionsFromLoadingToValid() async {
        let service = MockOnboardingService()
        service.validateResult = .success(
            InviteValidationResponse(
                email: "alex@example.com",
                displayName: "Alex",
                role: "member",
                inviter: "Empire"
            )
        )

        let viewModel = makeViewModel(service: service)

        let task = Task {
            await viewModel.loadInviteIfNeeded()
        }

        await Task.yield()
        XCTAssertEqual(viewModel.phase, .loading)

        await task.value

        XCTAssertEqual(viewModel.phase, .valid)
        XCTAssertEqual(viewModel.inviteDetails?.email, "alex@example.com")
        XCTAssertEqual(viewModel.displayName, "Alex")
    }

    func testLoadInviteTransitionsToInvalidError() async {
        let service = MockOnboardingService()
        service.validateResult = .failure(.invalidInvite)

        let viewModel = makeViewModel(service: service)

        await viewModel.loadInviteIfNeeded()

        XCTAssertEqual(viewModel.phase, .error(.invalidInvite))
    }

    func testLoadInviteTransitionsToNetworkError() async {
        let service = MockOnboardingService()
        service.validateResult = .failure(.network("offline"))

        let viewModel = makeViewModel(service: service)

        await viewModel.loadInviteIfNeeded()

        XCTAssertEqual(viewModel.phase, .error(.network("offline")))
    }

    func testConfirmPasswordValidationAppearsOnBlur() {
        let viewModel = makeViewModel()
        viewModel.displayName = "Alex"
        viewModel.password = "long-enough-password"
        viewModel.confirmPassword = "different-password"

        XCTAssertNil(viewModel.confirmPasswordError)

        viewModel.confirmPasswordBlurred()

        XCTAssertEqual(viewModel.confirmPasswordError, "Passwords do not match.")
    }

    func testSubmitPersistsSessionAndRoutesWelcomeDM() async {
        let service = MockOnboardingService()
        service.validateResult = .success(
            InviteValidationResponse(
                email: "alex@example.com",
                displayName: nil,
                role: "member",
                inviter: "Empire"
            )
        )
        service.acceptResult = .success(
            InviteAcceptResponse.fixture(
                token: "jwt-token",
                userID: "usr_123",
                expiresAt: Date(timeIntervalSince1970: 1_000),
                welcomeDMID: "dm_987"
            )
        )

        let sessionStore = MockSessionStore()
        let coordinator = MockOnboardingRouter()
        let viewModel = makeViewModel(service: service, sessionStore: sessionStore, coordinator: coordinator)

        await viewModel.loadInviteIfNeeded()
        viewModel.displayName = "Alex"
        viewModel.password = "long-enough-password"
        viewModel.confirmPassword = "long-enough-password"
        viewModel.confirmPasswordBlurred()

        await viewModel.submit()

        XCTAssertEqual(viewModel.phase, .valid)
        XCTAssertEqual(sessionStore.persistedUserID, "usr_123")
        XCTAssertEqual(sessionStore.persistedEmail, "alex@example.com")
        XCTAssertEqual(sessionStore.persistedDisplayName, "Alex")
        XCTAssertEqual(sessionStore.persistedToken, "jwt-token")
        XCTAssertEqual(coordinator.completedWelcomeConversationID, "dm_987")
    }

    private func makeViewModel(
        pendingInvite: AppCoordinator.PendingInvite? = .init(
            token: "inv_123",
            instanceURL: URL(string: "https://workspace.example.com")
        ),
        service: MockOnboardingService = MockOnboardingService(),
        sessionStore: MockSessionStore = MockSessionStore(),
        coordinator: MockOnboardingRouter = MockOnboardingRouter()
    ) -> OnboardingViewModel {
        let viewModel = OnboardingViewModel(
            pendingInvite: pendingInvite,
            service: service,
            sessionStore: sessionStore
        )
        viewModel.attachCoordinator(coordinator)
        return viewModel
    }
}

private final class MockOnboardingService: OnboardingServicing {
    var validateResult: Result<InviteValidationResponse, OnboardingServiceError> = .success(
        InviteValidationResponse(email: "alex@example.com", displayName: nil, role: nil, inviter: "Empire")
    )
    var acceptResult: Result<InviteAcceptResponse, OnboardingServiceError> = .success(
        InviteAcceptResponse.fixture(token: "jwt", userID: "usr", expiresAt: nil, welcomeDMID: nil)
    )

    func validateInvite(token: String, instanceURL: URL) async throws -> InviteValidationResponse {
        try await Task.sleep(nanoseconds: 50_000_000)
        return try validateResult.get()
    }

    func acceptInvite(
        token: String,
        displayName: String,
        password: String,
        instanceURL: URL
    ) async throws -> InviteAcceptResponse {
        return try acceptResult.get()
    }
}

@MainActor
private final class MockSessionStore: SessionPersisting {
    var persistedUserID: String?
    var persistedEmail: String?
    var persistedDisplayName: String?
    var persistedToken: String?

    func restoreAuthentication() -> Bool {
        false
    }

    func persistInviteSession(
        instanceURL: URL,
        userID: String,
        email: String,
        displayName: String,
        token: String,
        expiresAt: Date?,
        userType: String
    ) throws {
        persistedUserID = userID
        persistedEmail = email
        persistedDisplayName = displayName
        persistedToken = token
    }

    func signOut() throws {
    }
}

@MainActor
private final class MockOnboardingRouter: OnboardingRouting {
    var completedWelcomeConversationID: String?

    func completeOnboarding(welcomeConversationID: String?) {
        completedWelcomeConversationID = welcomeConversationID
    }
}

private extension InviteAcceptResponse {
    static func fixture(
        token: String,
        userID: String,
        expiresAt: Date?,
        welcomeDMID: String?
    ) -> InviteAcceptResponse {
        let onboarding = welcomeDMID.map { OnboardingPayload(welcomeDMID: $0) }
        return InviteAcceptResponse(token: token, userID: userID, expiresAt: expiresAt, onboarding: onboarding)
    }
}
