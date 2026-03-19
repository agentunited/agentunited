import Foundation

struct LiveAUAPIClient: AUAPIClient {
    let instanceURL: URL
    var authToken: String?

    private let session: URLSession
    private let encoder: JSONEncoder
    private let decoder: JSONDecoder

    init(instanceURL: URL, authToken: String? = nil, session: URLSession = .shared) {
        self.instanceURL = instanceURL
        self.authToken = authToken
        self.session = session

        let encoder = JSONEncoder()
        encoder.keyEncodingStrategy = .convertToSnakeCase
        self.encoder = encoder

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        self.decoder = decoder
    }

    func get<T: Decodable>(_ path: String) async throws -> T {
        let request = try makeRequest(path: path, method: "GET", body: Optional<EmptyRequestBody>.none)
        return try await perform(request, expecting: T.self)
    }

    func post<T: Decodable, Body: Encodable>(_ path: String, body: Body) async throws -> T {
        let request = try makeRequest(path: path, method: "POST", body: body)
        return try await perform(request, expecting: T.self)
    }

    func patch<T: Decodable, Body: Encodable>(_ path: String, body: Body) async throws -> T {
        let request = try makeRequest(path: path, method: "PATCH", body: body)
        return try await perform(request, expecting: T.self)
    }

    func login(email: String, password: String) async throws -> LoginResponse {
        try await post("auth/login", body: LoginRequest(email: email, password: password))
    }

    func validateInvite(token: String) async throws -> InviteValidationResponse {
        try await get("invite?token=\(token.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? token)")
    }

    func acceptInvite(token: String, displayName: String, password: String) async throws -> InviteAcceptResponse {
        try await post(
            "invite/accept",
            body: AcceptInviteRequest(token: token, displayName: displayName, password: password)
        )
    }

    func listMessages(channelId: String, before: String? = nil) async throws -> [MessageResponse] {
        var path = "channels/\(channelId)/messages"
        if let before, !before.isEmpty {
            path += "?before=\(before)"
        }
        return try await get(path)
    }

    func sendMessage(channelId: String, text: String, replyToID: String? = nil) async throws -> MessageResponse {
        try await post("channels/\(channelId)/messages", body: SendMessageRequest(text: text, replyToID: replyToID))
    }

    func listDMs() async throws -> [ConversationResponse] {
        let request = try makeRequest(path: "dm", method: "GET", body: Optional<EmptyRequestBody>.none)
        let data = try await performRequest(request)
        if let direct = try? decoder.decode([ConversationResponse].self, from: data) {
            return direct
        }
        if let wrapped = try? decoder.decode(ConversationListEnvelope.self, from: data) {
            return wrapped.conversations
        }
        throw AUAPIError.decodingError(
            DecodingError.dataCorrupted(.init(codingPath: [], debugDescription: "Unsupported DM list response"))
        )
    }

    func listChannels() async throws -> [ConversationResponse] {
        let request = try makeRequest(path: "channels", method: "GET", body: Optional<EmptyRequestBody>.none)
        let data = try await performRequest(request)
        if let direct = try? decoder.decode([ConversationResponse].self, from: data) {
            return direct
        }
        if let wrapped = try? decoder.decode(ChannelListEnvelope.self, from: data) {
            return wrapped.channels
        }
        if let wrapped = try? decoder.decode(ConversationListEnvelope.self, from: data) {
            return wrapped.conversations
        }
        throw AUAPIError.decodingError(
            DecodingError.dataCorrupted(.init(codingPath: [], debugDescription: "Unsupported channel list response"))
        )
    }

    func getChannel(id: String) async throws -> ChannelDetailResponse {
        let request = try makeRequest(path: "channels/\(id)", method: "GET", body: Optional<EmptyRequestBody>.none)
        let data = try await performRequest(request)
        if let direct = try? decoder.decode(ChannelDetailResponse.self, from: data) {
            return direct
        }
        if let wrapped = try? decoder.decode(ChannelDetailEnvelope.self, from: data) {
            return wrapped.channel
        }
        throw AUAPIError.decodingError(
            DecodingError.dataCorrupted(.init(codingPath: [], debugDescription: "Unsupported channel detail response"))
        )
    }

    func getChannelMembers(channelId: String) async throws -> [ChannelMemberResponse] {
        let request = try makeRequest(path: "channels/\(channelId)/members", method: "GET", body: Optional<EmptyRequestBody>.none)
        let data = try await performRequest(request)
        if let direct = try? decoder.decode([ChannelMemberResponse].self, from: data) {
            return direct
        }
        if let wrapped = try? decoder.decode(ChannelMembersEnvelope.self, from: data) {
            return wrapped.members
        }
        throw AUAPIError.decodingError(
            DecodingError.dataCorrupted(.init(codingPath: [], debugDescription: "Unsupported channel members response"))
        )
    }

    func createDM(targetUserID: String) async throws -> ConversationResponse {
        try await post("dm", body: CreateDMRequest(targetUserID: targetUserID))
    }

    func listUsers() async throws -> [UserResponse] {
        try await get("users")
    }

    func updateProfile(displayName: String) async throws -> UserResponse {
        try await patch("profile", body: UpdateProfileRequest(displayName: displayName))
    }

    private func makeRequest<Body: Encodable>(
        path: String,
        method: String,
        body: Body? = nil
    ) throws -> URLRequest {
        guard let url = URL(string: "/api/v1/\(path)", relativeTo: instanceURL) else {
            throw AUAPIError.invalidURL
        }

        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Accept")

        if let authToken {
            request.setValue("Bearer \(authToken)", forHTTPHeaderField: "Authorization")
        }

        if let body {
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
            request.httpBody = try encoder.encode(AnyEncodable(body))
        }

        return request
    }

    private func perform<T: Decodable>(_ request: URLRequest, expecting type: T.Type) async throws -> T {
        let data = try await performRequest(request)
        do {
            return try decoder.decode(T.self, from: data)
        } catch {
            throw AUAPIError.decodingError(error)
        }
    }

    private func performRequest(_ request: URLRequest) async throws -> Data {
        do {
            let (data, response) = try await session.data(for: request)
            guard let httpResponse = response as? HTTPURLResponse else {
                throw AUAPIError.unexpectedStatus(-1)
            }

            switch httpResponse.statusCode {
            case 200 ..< 300:
                return data
            case 401:
                throw AUAPIError.unauthorized
            case 404:
                throw AUAPIError.notFound
            case 500 ... 599:
                throw AUAPIError.serverError(httpResponse.statusCode)
            default:
                throw AUAPIError.unexpectedStatus(httpResponse.statusCode)
            }
        } catch let error as AUAPIError {
            throw error
        } catch {
            throw AUAPIError.networkError(error)
        }
    }

}

private struct EmptyRequestBody: Encodable {}

private struct ConversationListEnvelope: Decodable {
    let conversations: [ConversationResponse]
}

private struct ChannelListEnvelope: Decodable {
    let channels: [ConversationResponse]
}

private struct ChannelDetailEnvelope: Decodable {
    let channel: ChannelDetailResponse
}

private struct ChannelMembersEnvelope: Decodable {
    let members: [ChannelMemberResponse]
}

private struct LoginRequest: Encodable {
    let email: String
    let password: String
}

private struct AcceptInviteRequest: Encodable {
    let token: String
    let displayName: String
    let password: String
}

private struct SendMessageRequest: Encodable {
    let text: String
    let replyToID: String?
}

private struct CreateDMRequest: Encodable {
    let targetUserID: String
}

private struct UpdateProfileRequest: Encodable {
    let displayName: String
}

private struct AnyEncodable: Encodable {
    private let encodeImpl: (Encoder) throws -> Void

    init<T: Encodable>(_ value: T) {
        encodeImpl = value.encode(to:)
    }

    func encode(to encoder: Encoder) throws {
        try encodeImpl(encoder)
    }
}
