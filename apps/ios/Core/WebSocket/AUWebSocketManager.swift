import Combine
import Foundation
import SwiftData

@MainActor
final class AUWebSocketManager: ObservableObject {
    enum ConnectionState: Equatable {
        case disconnected
        case connecting
        case connected
        case reconnecting
    }

    enum Event: Equatable {
        case message(conversationID: String, messageID: String)
        case typing(conversationID: String, userID: String)
        case presence(userID: String, isOnline: Bool)
    }

    @Published private(set) var connectionState: ConnectionState = .disconnected
    @Published private(set) var lastEvent: Event?
    @Published private(set) var typingUsers: [String: [String: Date]] = [:]

    private var modelContext: ModelContext?
    private var keychainHelper = KeychainHelper()
    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder
    private var wsTask: URLSessionWebSocketTask?
    private var receiveTask: Task<Void, Never>?
    private var reconnectTask: Task<Void, Never>?
    private var reconnectAttempts = 0
    private let maxReconnectAttempts = 5

    init(session: URLSession = .shared) {
        self.session = session

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        self.decoder = decoder

        let encoder = JSONEncoder()
        encoder.keyEncodingStrategy = .convertToSnakeCase
        self.encoder = encoder
    }

    func configure(modelContext: ModelContext, keychainHelper: KeychainHelper = KeychainHelper()) {
        self.modelContext = modelContext
        self.keychainHelper = keychainHelper
    }

    func connect() async {
        guard connectionState == .disconnected || connectionState == .reconnecting else {
            return
        }

        guard let credentials = try? loadCredentials(), let url = makeWebSocketURL(instanceURL: credentials.instanceURL, token: credentials.token) else {
            connectionState = .disconnected
            return
        }

        reconnectTask?.cancel()
        receiveTask?.cancel()
        wsTask?.cancel(with: .goingAway, reason: nil)

        connectionState = .connecting
        let task = session.webSocketTask(with: url)
        wsTask = task
        task.resume()
        reconnectAttempts = 0
        connectionState = .connected

        receiveTask = Task { [weak self] in
            await self?.receiveLoop()
        }
    }

    func disconnect() {
        reconnectTask?.cancel()
        receiveTask?.cancel()
        wsTask?.cancel(with: .goingAway, reason: nil)
        wsTask = nil
        connectionState = .disconnected
    }

    func sendTyping(channelId: String) async {
        guard connectionState == .connected, let wsTask else {
            return
        }

        let payload = OutgoingTypingEvent(type: "typing", channelID: channelId)
        guard let data = try? encoder.encode(payload),
              let text = String(data: data, encoding: .utf8)
        else {
            return
        }

        do {
            try await wsTask.send(.string(text))
        } catch {
            await scheduleReconnect()
        }
    }

    func receiveLoop() async {
        guard let task = wsTask else {
            return
        }

        do {
            while !Task.isCancelled {
                let message = try await task.receive()
                try await handleMessage(message)
            }
        } catch {
            await scheduleReconnect()
        }
    }

    func handleTypingEvent(conversationID: String, userID: String) {
        var conversationTyping = typingUsers[conversationID, default: [:]]
        conversationTyping[userID] = Date()
        typingUsers[conversationID] = conversationTyping
        lastEvent = .typing(conversationID: conversationID, userID: userID)

        Task { [weak self] in
            try? await Task.sleep(for: .seconds(3))
            await self?.expireTypingIndicator(conversationID: conversationID, userID: userID)
        }
    }

    private func expireTypingIndicator(conversationID: String, userID: String) {
        guard var conversationTyping = typingUsers[conversationID],
              let lastSeen = conversationTyping[userID],
              Date().timeIntervalSince(lastSeen) >= 2.9
        else {
            return
        }

        conversationTyping.removeValue(forKey: userID)
        typingUsers[conversationID] = conversationTyping.isEmpty ? nil : conversationTyping
    }

    private func handleMessage(_ message: URLSessionWebSocketTask.Message) async throws {
        let data: Data

        switch message {
        case let .data(payload):
            data = payload
        case let .string(payload):
            guard let stringData = payload.data(using: .utf8) else {
                return
            }
            data = stringData
        @unknown default:
            return
        }

        let event = try decoder.decode(IncomingWebSocketEvent.self, from: data)

        switch event.type {
        case "message", "message.created", "conversation.message":
            if let message = event.message {
                try persist(message: message, conversationName: event.conversationName, participantIDs: event.participantIDs)
                lastEvent = .message(conversationID: message.conversationID, messageID: message.id)
            }
        case "typing":
            if let conversationID = event.conversationID, let userID = event.userID {
                handleTypingEvent(conversationID: conversationID, userID: userID)
            }
        case "presence":
            if let userID = event.userID, let isOnline = event.isOnline {
                try persistPresence(userID: userID, isOnline: isOnline)
                lastEvent = .presence(userID: userID, isOnline: isOnline)
            }
        default:
            break
        }
    }

    private func persist(message: MessageResponse, conversationName: String?, participantIDs: [String]?) throws {
        guard let modelContext else {
            return
        }

        let currentUserID = try? loadCredentials().userID

        let conversation = try fetchConversation(id: message.conversationID, in: modelContext)
            ?? Conversation(
                id: message.conversationID,
                name: conversationName ?? message.authorName,
                type: "dm",
                participantIDs: participantIDs ?? [message.authorID]
            )

        if conversation.modelContext == nil {
            modelContext.insert(conversation)
        }

        conversation.lastMessageText = message.text
        conversation.lastMessageAt = message.timestamp
        if currentUserID != message.authorID {
            conversation.unreadCount += 1
        }
        if let participantIDs {
            conversation.participantIDs = participantIDs
        }

        let existingMessage = try fetchMessage(id: message.id, in: modelContext)
        if existingMessage == nil {
            let record = Message(
                id: message.id,
                conversationID: message.conversationID,
                text: message.text,
                authorID: message.authorID,
                authorName: message.authorName,
                authorType: message.authorType,
                timestamp: message.timestamp,
                replyToID: message.replyToID,
                isPending: false,
                isFailed: false,
                conversation: conversation
            )
            modelContext.insert(record)
            conversation.messages.append(record)
        }

        let user = try fetchUser(id: message.authorID, in: modelContext)
            ?? User(
                id: message.authorID,
                displayName: message.authorName,
                email: "",
                userType: message.authorType,
                isOnline: true
            )

        if user.modelContext == nil {
            modelContext.insert(user)
        }
        user.displayName = message.authorName
        user.userType = message.authorType

        try modelContext.save()
    }

    private func persistPresence(userID: String, isOnline: Bool) throws {
        guard let modelContext else {
            return
        }

        if let user = try fetchUser(id: userID, in: modelContext) {
            user.isOnline = isOnline
            try modelContext.save()
        }
    }

    private func scheduleReconnect() async {
        guard reconnectAttempts < maxReconnectAttempts else {
            connectionState = .disconnected
            return
        }

        let attempt = reconnectAttempts
        reconnectAttempts += 1
        connectionState = .reconnecting

        reconnectTask?.cancel()
        reconnectTask = Task { [weak self] in
            let delay = min(pow(2.0, Double(attempt)), 30.0)
            try? await Task.sleep(for: .seconds(delay))
            await self?.connect()
        }
    }

    private func makeWebSocketURL(instanceURL: URL, token: String) -> URL? {
        guard var components = URLComponents(url: instanceURL.appendingPathComponent("ws"), resolvingAgainstBaseURL: false) else {
            return nil
        }

        components.scheme = instanceURL.scheme == "https" ? "wss" : "ws"
        components.queryItems = [URLQueryItem(name: "token", value: token)]
        return components.url
    }

    private func loadCredentials() throws -> (instanceURL: URL, token: String, userID: String) {
        guard let modelContext else {
            throw AUAPIError.unauthorized
        }

        let descriptor = FetchDescriptor<WorkspaceCredential>(
            sortBy: [SortDescriptor(\WorkspaceCredential.createdAt, order: .reverse)]
        )

        guard let credential = try modelContext.fetch(descriptor).first,
              let token = try keychainHelper.readJWT(for: credential.userID),
              !token.isEmpty
        else {
            throw AUAPIError.unauthorized
        }

        return (credential.instanceURL, token, credential.userID)
    }

    private func fetchConversation(id: String, in modelContext: ModelContext) throws -> Conversation? {
        let descriptor = FetchDescriptor<Conversation>(predicate: #Predicate { $0.id == id })
        return try modelContext.fetch(descriptor).first
    }

    private func fetchMessage(id: String, in modelContext: ModelContext) throws -> Message? {
        let descriptor = FetchDescriptor<Message>(predicate: #Predicate { $0.id == id })
        return try modelContext.fetch(descriptor).first
    }

    private func fetchUser(id: String, in modelContext: ModelContext) throws -> User? {
        let descriptor = FetchDescriptor<User>(predicate: #Predicate { $0.id == id })
        return try modelContext.fetch(descriptor).first
    }
}

private struct OutgoingTypingEvent: Encodable {
    let type: String
    let channelID: String
}

private struct IncomingWebSocketEvent: Decodable {
    let type: String
    let conversationID: String?
    let userID: String?
    let isOnline: Bool?
    let message: MessageResponse?
    let conversationName: String?
    let participantIDs: [String]?

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: WebSocketCodingKeys.self)
        type = try container.decodeOne(of: ["type", "event"])
        conversationID = try container.decodeIfPresentOne(of: ["conversation_id", "channel_id", "dm_id"])
        userID = try container.decodeIfPresentOne(of: ["user_id", "author_id"])
        isOnline = try container.decodeIfPresentOne(of: ["is_online", "online"])
        conversationName = try container.decodeIfPresentOne(of: ["conversation_name", "channel_name", "name"])
        participantIDs = try container.decodeIfPresent([String].self, forKey: WebSocketCodingKeys("participant_ids"))

        if let nestedMessage = try container.decodeIfPresent(MessageResponse.self, forKey: WebSocketCodingKeys("message")) {
            message = nestedMessage
        } else if let id: String = try container.decodeIfPresentOne(of: ["message_id", "id"]),
                  let conversationID: String = try container.decodeIfPresentOne(of: ["conversation_id", "channel_id", "dm_id"]),
                  let text: String = try container.decodeIfPresentOne(of: ["text", "body"]),
                  let authorID: String = try container.decodeIfPresentOne(of: ["author_id", "user_id"]),
                  let authorName: String = try container.decodeIfPresentOne(of: ["author_name", "display_name"]),
                  let authorType: String = try container.decodeIfPresentOne(of: ["author_type", "user_type"]),
                  let timestamp: Date = try container.decodeIfPresentOne(of: ["timestamp", "created_at"])
        {
            message = MessageResponse(
                id: id,
                conversationID: conversationID,
                text: text,
                authorID: authorID,
                authorName: authorName,
                authorType: authorType,
                timestamp: timestamp,
                replyToID: try container.decodeIfPresentOne(of: ["reply_to_id"])
            )
        } else {
            message = nil
        }
    }
}

private struct WebSocketCodingKeys: CodingKey {
    let stringValue: String
    let intValue: Int?

    init(_ stringValue: String) {
        self.stringValue = stringValue
        intValue = nil
    }

    init?(stringValue: String) {
        self.init(stringValue)
    }

    init?(intValue: Int) {
        return nil
    }
}

private extension KeyedDecodingContainer where K == WebSocketCodingKeys {
    func decodeOne<T: Decodable>(of keys: [String]) throws -> T {
        for key in keys {
            if let value = try decodeIfPresent(T.self, forKey: WebSocketCodingKeys(key)) {
                return value
            }
        }

        throw DecodingError.keyNotFound(
            WebSocketCodingKeys(keys.first ?? "unknown"),
            DecodingError.Context(codingPath: codingPath, debugDescription: "Missing required key from \(keys)")
        )
    }

    func decodeIfPresentOne<T: Decodable>(of keys: [String]) throws -> T? {
        for key in keys {
            if let value = try decodeIfPresent(T.self, forKey: WebSocketCodingKeys(key)) {
                return value
            }
        }

        return nil
    }
}
