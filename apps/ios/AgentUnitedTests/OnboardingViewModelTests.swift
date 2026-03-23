import Combine
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

    func testCriticalOnboardingPathCompletesAndLandsMessagesTab() async {
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
        let coordinator = AppCoordinator()
        let viewModel = OnboardingViewModel(
            pendingInvite: .init(
                token: "inv_123",
                instanceURL: URL(string: "https://workspace.example.com")
            ),
            service: service,
            sessionStore: sessionStore
        )
        viewModel.attachCoordinator(coordinator)

        await viewModel.loadInviteIfNeeded()
        viewModel.displayName = "Alex"
        viewModel.password = "long-enough-password"
        viewModel.confirmPassword = "long-enough-password"

        await viewModel.submit()

        XCTAssertEqual(coordinator.authState, .authenticated)
        XCTAssertEqual(coordinator.selectedTab, .messages)
        XCTAssertEqual(coordinator.messagesPath, [.conversation("dm_987")])
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

@MainActor
final class ConversationViewModelTests: XCTestCase {
    func testSendSuccessClearsPendingAndRequestsPushPermission() async {
        let store = MockConversationStore()
        let notifications = MockNotificationRequester()
        let viewModel = ConversationViewModel(
            conversationID: "dm_123",
            store: store,
            notificationRequester: notifications,
            networkMonitor: MockNetworkStatusProvider()
        )
        viewModel.inputText = "Hello there"

        await viewModel.sendCurrentMessage()

        XCTAssertEqual(store.insertCalls.count, 1)
        XCTAssertEqual(store.commitCalls.count, 1)
        XCTAssertEqual(viewModel.messages.count, 1)
        XCTAssertEqual(viewModel.messages.first?.id, "srv_1")
        XCTAssertEqual(viewModel.messages.first?.isPending, false)
        XCTAssertNil(viewModel.failureBanner)
        await Task.yield()
        XCTAssertGreaterThanOrEqual(notifications.requestCount, 1)
    }

    func testSendFailureMarksOptimisticMessageFailed() async {
        let store = MockConversationStore()
        store.commitResult = .failure(.sendFailed("Offline"))
        let viewModel = ConversationViewModel(
            conversationID: "dm_123",
            store: store,
            notificationRequester: MockNotificationRequester(),
            networkMonitor: MockNetworkStatusProvider()
        )
        viewModel.inputText = "Fail me"

        await viewModel.sendCurrentMessage()

        XCTAssertEqual(viewModel.messages.count, 1)
        XCTAssertEqual(viewModel.messages.first?.isFailed, true)
        XCTAssertEqual(viewModel.failureBanner?.text, "Offline")
    }

    func testSendInsertsOptimisticMessageBeforeCommitFinishes() async {
        let store = MockConversationStore()
        store.commitDelayNanoseconds = 200_000_000
        let viewModel = ConversationViewModel(
            conversationID: "dm_123",
            store: store,
            notificationRequester: MockNotificationRequester(),
            networkMonitor: MockNetworkStatusProvider()
        )
        viewModel.inputText = "Pending"

        let task = Task {
            await viewModel.sendCurrentMessage()
        }

        await Task.yield()

        XCTAssertEqual(viewModel.messages.count, 1)
        XCTAssertEqual(viewModel.messages.first?.isPending, true)

        await task.value
        XCTAssertEqual(viewModel.messages.first?.isPending, false)
    }

    func testRetryLogicReplacesFailedMessage() async {
        let store = MockConversationStore()
        store.commitResult = .failure(.sendFailed("Offline"))
        store.retryResult = .success(
            MessageItem(
                id: "srv_retry",
                text: "Retry me",
                authorID: "me",
                authorName: "Alex",
                authorType: "human",
                timestamp: .now,
                replyToID: nil,
                retryCount: 1,
                isPending: false,
                isFailed: false,
                isOutgoing: true
            )
        )
        let viewModel = ConversationViewModel(
            conversationID: "dm_123",
            store: store,
            notificationRequester: MockNotificationRequester(),
            networkMonitor: MockNetworkStatusProvider()
        )
        viewModel.inputText = "Retry me"

        await viewModel.sendCurrentMessage()
        await viewModel.retryFailedMessage()

        XCTAssertEqual(store.retryCalls.count, 1)
        XCTAssertEqual(viewModel.messages.first?.id, "srv_retry")
        XCTAssertEqual(viewModel.messages.first?.isFailed, false)
        XCTAssertNil(viewModel.failureBanner)
    }

    func testChannelSendSuccessCommitsMessage() async {
        let store = MockConversationStore()
        let viewModel = ConversationViewModel(
            conversationID: "chn_123",
            mode: .channel,
            store: store,
            notificationRequester: MockNotificationRequester(),
            networkMonitor: MockNetworkStatusProvider()
        )
        viewModel.inputText = "Hello channel"

        await viewModel.sendCurrentMessage()

        XCTAssertEqual(store.insertCalls.first?.conversationID, "chn_123")
        XCTAssertEqual(store.commitCalls.first?.conversationID, "chn_123")
        XCTAssertEqual(viewModel.messages.first?.id, "srv_1")
        XCTAssertNil(viewModel.failureBanner)
    }

    func testChannelSendFailureShowsFailureBanner() async {
        let store = MockConversationStore()
        store.commitResult = .failure(.sendFailed("Channel offline"))
        let viewModel = ConversationViewModel(
            conversationID: "chn_123",
            mode: .channel,
            store: store,
            notificationRequester: MockNotificationRequester(),
            networkMonitor: MockNetworkStatusProvider()
        )
        viewModel.inputText = "Hello channel"

        await viewModel.sendCurrentMessage()

        XCTAssertEqual(viewModel.messages.first?.isFailed, true)
        XCTAssertEqual(viewModel.failureBanner?.text, "Channel offline")
    }
}

@MainActor
final class ChannelListViewModelTests: XCTestCase {
    func testInitialStateIsIdle() {
        let viewModel = ChannelListViewModel(store: MockChannelListStore())

        XCTAssertEqual(viewModel.phase, .idle)
        XCTAssertTrue(viewModel.channels.isEmpty)
    }

    func testLoadTransitionsFromLoadingToLoadedWithChannels() async {
        let store = MockChannelListStore()
        store.syncResult = .success([.channelFixture(id: "chn_1", name: "general")])
        let viewModel = ChannelListViewModel(store: store)

        let task = Task {
            await viewModel.loadIfNeeded()
        }

        await Task.yield()
        XCTAssertEqual(viewModel.phase, .loading)

        await task.value

        XCTAssertEqual(viewModel.phase, .loaded)
        XCTAssertEqual(viewModel.channels.count, 1)
    }

    func testLoadTransitionsToLoadedWithEmptyStateData() async {
        let store = MockChannelListStore()
        store.syncResult = .success([])
        let viewModel = ChannelListViewModel(store: store)

        await viewModel.loadIfNeeded()

        XCTAssertEqual(viewModel.phase, .loaded)
        XCTAssertTrue(viewModel.channels.isEmpty)
    }

    func testLoadTransitionsToErrorAndFallsBackToLocalChannels() async {
        let store = MockChannelListStore()
        store.syncResult = .failure(.network("offline"))
        store.localChannels = [.channelFixture(id: "chn_local", name: "ops")]
        let viewModel = ChannelListViewModel(store: store)

        await viewModel.loadIfNeeded()

        XCTAssertEqual(viewModel.phase, .error(.network("offline")))
        XCTAssertEqual(viewModel.channels.map(\.id), ["chn_local"])
    }
}

@MainActor
final class TypingIndicatorTimeoutTests: XCTestCase {
    func testTypingIndicatorExpiresAfterThreeSeconds() async {
        let manager = AUWebSocketManager()

        manager.handleTypingEvent(conversationID: "dm_123", userID: "agent_1")
        XCTAssertEqual(manager.typingUsers["dm_123"]?.keys.contains("agent_1"), true)

        try? await Task.sleep(nanoseconds: 3_200_000_000)

        XCTAssertNil(manager.typingUsers["dm_123"]?["agent_1"])
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

@MainActor
private final class MockConversationStore: ConversationStoreProtocol {
    var insertCalls: [(conversationID: String, text: String, replyToID: String?)] = []
    var commitCalls: [(conversationID: String, text: String, replyToID: String?, optimisticID: String)] = []
    var retryCalls: [(messageID: String, conversationID: String)] = []
    var commitDelayNanoseconds: UInt64 = 0
    var commitResult: Result<MessageItem, MessagesFeatureError> = .success(
        MessageItem(
            id: "srv_1",
            text: "Hello there",
            authorID: "me",
            authorName: "Alex",
            authorType: "human",
            timestamp: .now,
            replyToID: nil,
            retryCount: 0,
            isPending: false,
            isFailed: false,
            isOutgoing: true
        )
    )
    var retryResult: Result<MessageItem, MessagesFeatureError> = .success(
        MessageItem(
            id: "srv_retry",
            text: "Retry",
            authorID: "me",
            authorName: "Alex",
            authorType: "human",
            timestamp: .now,
            replyToID: nil,
            retryCount: 1,
            isPending: false,
            isFailed: false,
            isOutgoing: true
        )
    )
    private var storedMessages: [String: MessageItem] = [:]

    func loadConversationScreen(conversationID: String) async throws -> ConversationScreenData {
        ConversationScreenData(header: .placeholder, messages: Array(storedMessages.values))
    }

    func loadChannelDetails(channelID: String) async throws -> ChannelDetails? {
        nil
    }

    func loadChannelMembers(channelID: String) async throws -> [ChannelMemberItem] {
        []
    }

    func insertOptimisticMessage(conversationID: String, text: String, replyToID: String?) throws -> MessageItem {
        insertCalls.append((conversationID, text, replyToID))
        let item = MessageItem(
            id: "local_1",
            text: text,
            authorID: "me",
            authorName: "Alex",
            authorType: "human",
            timestamp: .now,
            replyToID: replyToID,
            retryCount: 0,
            isPending: true,
            isFailed: false,
            isOutgoing: true
        )
        storedMessages[item.id] = item
        return item
    }

    func commitMessage(conversationID: String, text: String, replyToID: String?, optimisticID: String) async throws -> MessageItem {
        commitCalls.append((conversationID, text, replyToID, optimisticID))
        if commitDelayNanoseconds > 0 {
            try? await Task.sleep(nanoseconds: commitDelayNanoseconds)
        }
        let item = try commitResult.get()
        storedMessages.removeValue(forKey: optimisticID)
        storedMessages[item.id] = item
        return item
    }

    func markMessageFailed(id: String) throws -> MessageItem {
        let failed = MessageItem(
            id: id,
            text: storedMessages[id]?.text ?? "Failed",
            authorID: "me",
            authorName: "Alex",
            authorType: "human",
            timestamp: .now,
            replyToID: nil,
            retryCount: 1,
            isPending: false,
            isFailed: true,
            isOutgoing: true
        )
        storedMessages[id] = failed
        return failed
    }

    func retryFailedMessage(messageID: String, conversationID: String) async throws -> MessageItem {
        retryCalls.append((messageID, conversationID))
        let item = try retryResult.get()
        storedMessages.removeValue(forKey: messageID)
        storedMessages[item.id] = item
        return item
    }

    func deleteMessage(id: String) throws {
        storedMessages.removeValue(forKey: id)
    }

    func typingDisplayNames(conversationID: String, userIDs: [String]) throws -> [String] {
        []
    }
}

private final class MockNotificationRequester: NotificationPermissionRequesting {
    private(set) var requestCount = 0

    func requestAfterFirstMessageSend() async {
        requestCount += 1
    }
}

@MainActor
private final class MockNetworkStatusProvider: NetworkStatusProviding {
    private let subject = CurrentValueSubject<Bool, Never>(true)

    var statusPublisher: AnyPublisher<Bool, Never> {
        subject.eraseToAnyPublisher()
    }

    func start() {
    }
}

@MainActor
private final class MockChannelListStore: ChannelListStoreProtocol {
    var syncResult: Result<[ConversationListItem], MessagesFeatureError> = .success([])
    var localChannels: [ConversationListItem] = []

    func syncChannelList() async throws -> [ConversationListItem] {
        try await Task.sleep(nanoseconds: 50_000_000)
        return try syncResult.get()
    }

    func localChannelList() throws -> [ConversationListItem] {
        localChannels
    }

    func toggleMuteConversation(id: String) throws -> [ConversationListItem] {
        localChannels = localChannels.map { item in
            guard item.id == id else {
                return item
            }
            return ConversationListItem(
                kind: item.kind,
                id: item.id,
                displayName: item.displayName,
                isAgent: item.isAgent,
                isOnline: item.isOnline,
                unreadCount: item.unreadCount,
                lastMessagePreview: item.lastMessagePreview,
                lastMessageAt: item.lastMessageAt,
                initials: item.initials,
                isMuted: !item.isMuted,
                memberCount: item.memberCount
            )
        }
        return localChannels
    }
}

private extension ConversationListItem {
    static func channelFixture(id: String, name: String) -> ConversationListItem {
        ConversationListItem(
            kind: .channel,
            id: id,
            displayName: name,
            isAgent: false,
            isOnline: false,
            unreadCount: 0,
            lastMessagePreview: "Latest message",
            lastMessageAt: nil,
            initials: String(name.prefix(1)).uppercased(),
            isMuted: false,
            memberCount: 3
        )
    }
}

final class AUAPIClientTests: XCTestCase {
    override class func tearDown() {
        super.tearDown()
        MockURLProtocol.requestHandler = nil
    }

    func testLoginDecodesAuthResponse() async throws {
        let session = makeSession()
        var capturedRequest: URLRequest?
        MockURLProtocol.requestHandler = { request in
            capturedRequest = request
            let json = """
            {
              "token": "jwt_123",
              "user_id": "usr_1"
            }
            """
            return (HTTPURLResponse(url: request.url ?? URL(string: "https://workspace.example.com")!, statusCode: 200, httpVersion: nil, headerFields: nil)!, Data(json.utf8))
        }

        let client = LiveAUAPIClient(instanceURL: URL(string: "https://workspace.example.com")!, session: session)
        let response = try await client.login(email: "alice@example.com", password: "password-1234")

        XCTAssertEqual(capturedRequest?.httpMethod, "POST")
        XCTAssertEqual(response.token, "jwt_123")
        XCTAssertEqual(response.userID, "usr_1")
        XCTAssertNil(response.expiresAt)
    }

    func testLoginDecodesNestedUserObjectWhenUserIDIsNotTopLevel() async throws {
        let session = makeSession()
        MockURLProtocol.requestHandler = { request in
            let json = """
            {
              "token": "jwt_nested",
              "user": {
                "id": "usr_nested"
              }
            }
            """
            return (HTTPURLResponse(url: request.url ?? URL(string: "https://workspace.example.com")!, statusCode: 200, httpVersion: nil, headerFields: nil)!, Data(json.utf8))
        }

        let client = LiveAUAPIClient(instanceURL: URL(string: "https://workspace.example.com")!, session: session)
        let response = try await client.login(email: "alice@example.com", password: "password-1234")

        XCTAssertEqual(response.token, "jwt_nested")
        XCTAssertEqual(response.userID, "usr_nested")
    }

    func testLoginFailsWithExactErrorWhenUserIDMissingEverywhere() async throws {
        let session = makeSession()
        MockURLProtocol.requestHandler = { request in
            let json = """
            {
              "token": "jwt_only"
            }
            """
            return (HTTPURLResponse(url: request.url ?? URL(string: "https://workspace.example.com")!, statusCode: 200, httpVersion: nil, headerFields: nil)!, Data(json.utf8))
        }

        let client = LiveAUAPIClient(instanceURL: URL(string: "https://workspace.example.com")!, session: session)

        do {
            _ = try await client.login(email: "alice@example.com", password: "password-1234")
            XCTFail("Expected decode failure")
        } catch AUAPIError.decodingError(let error) {
            XCTAssertTrue(String(describing: error).contains("Missing user id in login response"))
        } catch {
            XCTFail("Unexpected error: \(error)")
        }
    }

    func testSendMessagePostsChannelPayloadAndDecodesMessage() async throws {
        let session = makeSession()
        var capturedRequest: URLRequest?
        MockURLProtocol.requestHandler = { request in
            capturedRequest = request
            let json = """
            {
              "id": "msg_1",
              "conversation_id": "chn_1",
              "text": "hello from test",
              "author_id": "usr_1",
              "author_name": "Alice",
              "author_type": "human",
              "timestamp": "2026-03-20T01:00:00Z",
              "reply_to_id": "msg_parent"
            }
            """
            return (HTTPURLResponse(url: request.url ?? URL(string: "https://workspace.example.com")!, statusCode: 200, httpVersion: nil, headerFields: nil)!, Data(json.utf8))
        }

        let client = LiveAUAPIClient(
            instanceURL: URL(string: "https://workspace.example.com")!,
            authToken: "jwt_abc",
            session: session
        )

        let response = try await client.sendMessage(channelId: "chn_1", text: "hello from test", replyToID: "msg_parent")

        XCTAssertEqual(capturedRequest?.httpMethod, "POST")
        XCTAssertEqual(capturedRequest?.value(forHTTPHeaderField: "Authorization"), "Bearer jwt_abc")
        XCTAssertEqual(response.id, "msg_1")
        XCTAssertEqual(response.conversationID, "chn_1")
        XCTAssertEqual(response.replyToID, "msg_parent")
    }

    private func makeSession() -> URLSession {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [MockURLProtocol.self]
        return URLSession(configuration: configuration)
    }
}

private final class MockURLProtocol: URLProtocol {
    nonisolated(unsafe) static var requestHandler: ((URLRequest) throws -> (HTTPURLResponse, Data))?

    override class func canInit(with request: URLRequest) -> Bool { true }
    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }

    override func startLoading() {
        guard let handler = Self.requestHandler else {
            XCTFail("Missing request handler")
            return
        }

        do {
            let (response, data) = try handler(request)
            client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
            client?.urlProtocol(self, didLoad: data)
            client?.urlProtocolDidFinishLoading(self)
        } catch {
            client?.urlProtocol(self, didFailWithError: error)
        }
    }

    override func stopLoading() {}
}
