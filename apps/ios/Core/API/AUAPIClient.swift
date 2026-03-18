import Foundation

protocol AUAPIClient {
    var instanceURL: URL { get }
    var authToken: String? { get set }

    func get<T: Decodable>(_ path: String) async throws -> T
    func post<T: Decodable, Body: Encodable>(_ path: String, body: Body) async throws -> T
    func patch<T: Decodable, Body: Encodable>(_ path: String, body: Body) async throws -> T

    func login(email: String, password: String) async throws -> LoginResponse
    func validateInvite(token: String) async throws -> InviteValidationResponse
    func acceptInvite(token: String, displayName: String, password: String) async throws -> InviteAcceptResponse
    func listMessages(channelId: String, before: String?) async throws -> [MessageResponse]
    func sendMessage(channelId: String, text: String, replyToID: String?) async throws -> MessageResponse
    func listDMs() async throws -> [ConversationResponse]
    func listChannels() async throws -> [ConversationResponse]
    func createDM(targetUserID: String) async throws -> ConversationResponse
    func listUsers() async throws -> [UserResponse]
    func updateProfile(displayName: String) async throws -> UserResponse
}

enum AUAPIError: LocalizedError {
    case invalidURL
    case unauthorized
    case notFound
    case serverError(Int)
    case networkError(Error)
    case decodingError(Error)
    case unexpectedStatus(Int)

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "The API request URL is invalid."
        case .unauthorized:
            return "Your session has expired."
        case .notFound:
            return "The requested resource was not found."
        case let .serverError(statusCode):
            return "The server returned an error (\(statusCode))."
        case let .networkError(error):
            return "Network error: \(error.localizedDescription)"
        case let .decodingError(error):
            return "Response decoding failed: \(error.localizedDescription)"
        case let .unexpectedStatus(statusCode):
            return "Unexpected response status: \(statusCode)"
        }
    }
}

struct LoginResponse: Decodable {
    let token: String
    let userID: String
    let expiresAt: Date
}

struct InviteAcceptResponse: Decodable {
    let token: String
    let userID: String
    let expiresAt: Date?
    let onboarding: OnboardingPayload?

    init(token: String, userID: String, expiresAt: Date?, onboarding: OnboardingPayload?) {
        self.token = token
        self.userID = userID
        self.expiresAt = expiresAt
        self.onboarding = onboarding
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: FlexibleCodingKeys.self)
        token = try container.decodeOne(of: ["token", "jwt_token"])
        userID = try container.decodeOne(of: ["user_id", "userID"])
        expiresAt = try container.decodeIfPresentOne(of: ["expires_at", "expiresAt"])
        onboarding = try container.decodeIfPresent(OnboardingPayload.self, forKey: FlexibleCodingKeys("onboarding"))
    }
}

struct OnboardingPayload: Decodable {
    let welcomeDMID: String?

    private enum CodingKeys: String, CodingKey {
        case welcomeDMID = "welcome_dm_id"
    }
}

struct InviteValidationResponse: Decodable, Equatable {
    let email: String
    let displayName: String?
    let role: String?
    let inviter: String?
}

struct MessageResponse: Decodable, Identifiable {
    let id: String
    let conversationID: String
    let text: String
    let authorID: String
    let authorName: String
    let authorType: String
    let timestamp: Date
    let replyToID: String?

    private enum CodingKeys: String, CodingKey {
        case id
        case conversationID = "conversation_id"
        case text
        case authorID = "author_id"
        case authorName = "author_name"
        case authorType = "author_type"
        case timestamp
        case replyToID = "reply_to_id"
    }
}

struct ConversationResponse: Decodable, Identifiable {
    let id: String
    let name: String
    let type: String
    let lastMessageText: String?
    let lastMessageAt: Date?
    let unreadCount: Int
    let participantIDs: [String]

    private enum CodingKeys: String, CodingKey {
        case id
        case name
        case type
        case lastMessageText = "last_message_text"
        case lastMessageAt = "last_message_at"
        case unreadCount = "unread_count"
        case participantIDs = "participant_ids"
    }
}

struct UserResponse: Decodable, Identifiable {
    let id: String
    let displayName: String
    let email: String
    let userType: String
    let avatarURL: URL?
    let isOnline: Bool

    private enum CodingKeys: String, CodingKey {
        case id
        case displayName = "display_name"
        case email
        case userType = "user_type"
        case avatarURL = "avatar_url"
        case isOnline = "is_online"
    }
}

private struct FlexibleCodingKeys: CodingKey {
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

private extension KeyedDecodingContainer where K == FlexibleCodingKeys {
    func decodeOne<T: Decodable>(of keys: [String]) throws -> T {
        for key in keys {
            if let value = try decodeIfPresent(T.self, forKey: FlexibleCodingKeys(key)) {
                return value
            }
        }

        throw DecodingError.keyNotFound(
            FlexibleCodingKeys(keys.first ?? "unknown"),
            DecodingError.Context(codingPath: codingPath, debugDescription: "Missing required value for keys \(keys)")
        )
    }

    func decodeIfPresentOne<T: Decodable>(of keys: [String]) throws -> T? {
        for key in keys {
            if let value = try decodeIfPresent(T.self, forKey: FlexibleCodingKeys(key)) {
                return value
            }
        }
        return nil
    }
}
