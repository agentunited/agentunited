import Foundation

struct CentralAPIClient {
    private let baseURL = URL(string: "https://agent-united-central-961985674922.us-central1.run.app")!
    private let authToken: String?
    private let session: URLSession
    private let encoder: JSONEncoder
    private let decoder: JSONDecoder

    init(authToken: String? = nil) {
        self.authToken = authToken

        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 30
        config.timeoutIntervalForResource = 30
        self.session = URLSession(configuration: config)

        let encoder = JSONEncoder()
        encoder.keyEncodingStrategy = .convertToSnakeCase
        self.encoder = encoder

        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        decoder.dateDecodingStrategy = .iso8601
        self.decoder = decoder
    }

    func register(email: String, password: String, displayName: String) async throws -> CentralAuthResponse {
        try await post(
            "/api/v1/users/register",
            body: RegisterRequest(email: email, password: password, displayName: displayName),
            requiresAuth: false
        )
    }

    func login(email: String, password: String) async throws -> CentralAuthResponse {
        try await post(
            "/api/v1/users/login",
            body: LoginRequest(email: email, password: password),
            requiresAuth: false
        )
    }

    func generateClaimKey() async throws -> CentralClaimKeyResponse {
        try await post("/api/v1/claim/generate", body: EmptyRequestBody(), requiresAuth: true)
    }

    func listWorkspaces() async throws -> [CentralWorkspace] {
        let response: CentralWorkspacesResponse = try await get("/api/v1/workspaces", requiresAuth: true)
        return response.workspaces
    }

    private func get<T: Decodable>(_ path: String, requiresAuth: Bool) async throws -> T {
        let request = try makeRequest(path: path, method: "GET", requiresAuth: requiresAuth)
        return try await perform(request, expecting: T.self)
    }

    private func post<T: Decodable, Body: Encodable>(_ path: String, body: Body, requiresAuth: Bool) async throws -> T {
        let request = try makeRequest(path: path, method: "POST", body: body, requiresAuth: requiresAuth)
        return try await perform(request, expecting: T.self)
    }

    private func makeRequest(
        path: String,
        method: String,
        body: (any Encodable)? = nil,
        requiresAuth: Bool
    ) throws -> URLRequest {
        guard let url = URL(string: path, relativeTo: baseURL) else {
            throw CentralAPIError.invalidURL
        }

        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Accept")

        if requiresAuth {
            guard let authToken, authToken.isEmpty == false else {
                throw CentralAPIError.unauthorized
            }
            request.setValue("Bearer \(authToken)", forHTTPHeaderField: "Authorization")
        }

        if let body {
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
            request.httpBody = try encoder.encode(body)
        }

        return request
    }

    private func perform<T: Decodable>(_ request: URLRequest, expecting type: T.Type) async throws -> T {
        let data = try await performRequest(request)
        do {
            return try decoder.decode(T.self, from: data)
        } catch {
            throw CentralAPIError.decodingError(error)
        }
    }

    private func performRequest(_ request: URLRequest) async throws -> Data {
        do {
            let (data, response) = try await session.data(for: request)
            guard let httpResponse = response as? HTTPURLResponse else {
                throw CentralAPIError.serverError(-1)
            }

            switch httpResponse.statusCode {
            case 200 ..< 300:
                return data
            case 401:
                if let errorCode = try? decoder.decode(CentralErrorResponse.self, from: data).error,
                   errorCode == "invalid_credentials" {
                    throw CentralAPIError.invalidCredentials
                }
                throw CentralAPIError.unauthorized
            case 409:
                if let errorCode = try? decoder.decode(CentralErrorResponse.self, from: data).error,
                   errorCode == "email_already_registered" {
                    throw CentralAPIError.emailAlreadyRegistered
                }
                throw CentralAPIError.serverError(httpResponse.statusCode)
            default:
                throw CentralAPIError.serverError(httpResponse.statusCode)
            }
        } catch let error as CentralAPIError {
            throw error
        } catch {
            throw CentralAPIError.networkError(error)
        }
    }
}

struct CentralAuthResponse: Decodable {
    let token: String
    let userID: String
    let email: String?
    let displayName: String?
    let plan: String?

    // convertFromSnakeCase maps user_id → userId, not userID.
    // Explicit CodingKeys avoid the mismatch.
    private enum CodingKeys: String, CodingKey {
        case token
        case userID = "user_id"
        case email
        case displayName = "display_name"
        case plan
    }
}

struct CentralClaimKeyResponse: Decodable {
    let claimKey: String
    let expiresAt: Date
    let deepLink: String?

    private enum CodingKeys: String, CodingKey {
        case claimKey = "claim_key"
        case expiresAt = "expires_at"
        case deepLink = "deep_link"
    }
}

struct CentralWorkspace: Decodable, Identifiable {
    var id: String { workspaceID }
    let workspaceID: String
    let name: String
    let relayURL: String
    let joinedAt: Date?
    let status: String?

    // relay_url → relayUrl (not relayURL) under convertFromSnakeCase.
    private enum CodingKeys: String, CodingKey {
        case workspaceID = "workspace_id"
        case name
        case relayURL = "relay_url"
        case joinedAt = "joined_at"
        case status
    }
}

struct CentralWorkspacesResponse: Decodable {
    let workspaces: [CentralWorkspace]
}

enum CentralAPIError: Error {
    case invalidURL
    case unauthorized
    case emailAlreadyRegistered
    case invalidCredentials
    case serverError(Int)
    case networkError(Error)
    case decodingError(Error)
}

extension CentralAPIError: LocalizedError {
    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "Something went wrong. Please try again."
        case .unauthorized:
            return "You need to sign in again."
        case .emailAlreadyRegistered:
            return "That email is already registered."
        case .invalidCredentials:
            return "Incorrect email or password."
        case let .serverError(status) where status >= 500:
            return "Our servers are having trouble. Please try again in a moment."
        case .serverError:
            return "Something went wrong. Please try again."
        case .networkError:
            return "Check your connection and try again."
        case .decodingError:
            return "Something went wrong. Please try again."
        }
    }
}

private struct RegisterRequest: Encodable {
    let email: String
    let password: String
    let displayName: String
}

private struct LoginRequest: Encodable {
    let email: String
    let password: String
}

private struct CentralErrorResponse: Decodable {
    let error: String
}

private struct EmptyRequestBody: Encodable {}
