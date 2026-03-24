import Foundation

protocol AUAPIClient {
    var instanceURL: URL { get }
    var authToken: String? { get set }

    func get<T: Decodable>(_ path: String) async throws -> T
    func post<T: Decodable, Body: Encodable>(_ path: String, body: Body) async throws -> T
    func put<T: Decodable, Body: Encodable>(_ path: String, body: Body) async throws -> T
    func patch<T: Decodable, Body: Encodable>(_ path: String, body: Body) async throws -> T

    func login(email: String, password: String) async throws -> LoginResponse
    func validateInvite(token: String) async throws -> InviteValidationResponse
    func acceptInvite(token: String, displayName: String, password: String) async throws -> InviteAcceptResponse
    func acceptInvite(token: String, centralJWT: String) async throws -> InviteAcceptResponse
    func listMessages(channelId: String, before: String?) async throws -> [MessageResponse]
    func sendMessage(channelId: String, text: String, replyToID: String?) async throws -> MessageResponse
    func listDMs() async throws -> [ConversationResponse]
    func listChannels() async throws -> [ConversationResponse]
    func getChannel(id: String) async throws -> ChannelDetailResponse
    func getChannelMembers(channelId: String) async throws -> [ChannelMemberResponse]
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
    let expiresAt: Date?

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: FlexibleCodingKeys.self)
        token = try container.decodeOne(of: ["token", "jwt_token"])

        if let directUserID: String = try container.decodeIfPresentOne(of: ["user_id", "userID", "id"]) {
            userID = directUserID
        } else if let userObject = try container.decodeIfPresent(LoginUserObject.self, forKey: FlexibleCodingKeys("user")) {
            userID = userObject.id
        } else {
            throw DecodingError.keyNotFound(
                FlexibleCodingKeys("user_id"),
                DecodingError.Context(codingPath: container.codingPath, debugDescription: "Missing user id in login response")
            )
        }

        expiresAt = try container.decodeIfPresentOne(of: ["expires_at", "expiresAt"])
    }
}

private struct LoginUserObject: Decodable {
    let id: String
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
    let topic: String?
    let lastMessageText: String?
    let lastMessageAt: Date?
    let unreadCount: Int
    let memberCount: Int?
    let participantIDs: [String]

    private enum CodingKeys: String, CodingKey {
        case id
        case name
        case type
        case topic
        case description
        case lastMessageText = "last_message_text"
        case lastMessageAt = "last_message_at"
        case unreadCount = "unread_count"
        case memberCount = "member_count"
        case participantIDs = "participant_ids"
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(String.self, forKey: .id)
        name = try container.decode(String.self, forKey: .name)
        type = try container.decodeIfPresent(String.self, forKey: .type) ?? "channel"
        topic = try container.decodeIfPresent(String.self, forKey: .topic)
            ?? container.decodeIfPresent(String.self, forKey: .description)
        lastMessageText = try container.decodeIfPresent(String.self, forKey: .lastMessageText)
        lastMessageAt = try container.decodeIfPresent(Date.self, forKey: .lastMessageAt)
        unreadCount = try container.decodeIfPresent(Int.self, forKey: .unreadCount) ?? 0
        memberCount = try container.decodeIfPresent(Int.self, forKey: .memberCount)
        participantIDs = try container.decodeIfPresent([String].self, forKey: .participantIDs) ?? []
    }
}

struct ChannelDetailResponse: Decodable {
    let id: String
    let name: String
    let topic: String?
    let memberCount: Int
    let members: [ChannelMemberResponse]

    private enum CodingKeys: String, CodingKey {
        case id
        case name
        case topic
        case description
        case memberCount = "member_count"
        case members
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(String.self, forKey: .id)
        name = try container.decode(String.self, forKey: .name)
        topic = try container.decodeIfPresent(String.self, forKey: .topic)
            ?? container.decodeIfPresent(String.self, forKey: .description)
        members = try container.decodeIfPresent([ChannelMemberResponse].self, forKey: .members) ?? []
        memberCount = try container.decodeIfPresent(Int.self, forKey: .memberCount) ?? members.count
    }
}

struct ChannelMemberResponse: Decodable, Identifiable {
    let id: String
    let email: String
    let role: String
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
