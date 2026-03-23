import Foundation
import Security

enum KeychainHelperError: LocalizedError {
    case unexpectedData
    case unhandledStatus(OSStatus)

    var errorDescription: String? {
        switch self {
        case .unexpectedData:
            return "Keychain returned data in an unexpected format."
        case let .unhandledStatus(status):
            return SecCopyErrorMessageString(status, nil) as String? ?? "Keychain error \(status)"
        }
    }
}

struct KeychainHelper {
    private let service: String
    private let accessGroup: CFString = kSecAttrAccessibleWhenUnlockedThisDeviceOnly

    init(service: String = Bundle.main.bundleIdentifier ?? "ai.agentunited.app") {
        self.service = service
    }

    func storeJWT(_ token: String, for userID: String) throws {
        #if targetEnvironment(simulator)
        // Keychain requires entitlements not present in unsigned simulator builds.
        // Use UserDefaults as a safe fallback in the simulator.
        UserDefaults.standard.set(token, forKey: "jwt.\(userID)")
        return
        #else
        let encoded = Data(token.utf8)
        let query = baseQuery(for: userID)

        let attributes: [String: Any] = [
            kSecValueData as String: encoded,
            kSecAttrAccessible as String: accessGroup,
        ]

        let updateStatus = SecItemUpdate(query as CFDictionary, attributes as CFDictionary)
        if updateStatus == errSecSuccess {
            return
        }

        if updateStatus != errSecItemNotFound {
            throw KeychainHelperError.unhandledStatus(updateStatus)
        }

        var addQuery = query
        addQuery[kSecValueData as String] = encoded
        addQuery[kSecAttrAccessible as String] = accessGroup

        let addStatus = SecItemAdd(addQuery as CFDictionary, nil)
        guard addStatus == errSecSuccess else {
            throw KeychainHelperError.unhandledStatus(addStatus)
        }
        #endif
    }

    func readJWT(for userID: String) throws -> String? {
        #if targetEnvironment(simulator)
        return UserDefaults.standard.string(forKey: "jwt.\(userID)")
        #else
        var query = baseQuery(for: userID)
        query[kSecReturnData as String] = kCFBooleanTrue
        query[kSecMatchLimit as String] = kSecMatchLimitOne

        var result: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        switch status {
        case errSecSuccess:
            guard let data = result as? Data, let token = String(data: data, encoding: .utf8) else {
                throw KeychainHelperError.unexpectedData
            }
            return token
        case errSecItemNotFound:
            return nil
        default:
            throw KeychainHelperError.unhandledStatus(status)
        }
        #endif
    }

    func deleteJWT(for userID: String) throws {
        #if targetEnvironment(simulator)
        UserDefaults.standard.removeObject(forKey: "jwt.\(userID)")
        #else
        let status = SecItemDelete(baseQuery(for: userID) as CFDictionary)
        guard status == errSecSuccess || status == errSecItemNotFound else {
            throw KeychainHelperError.unhandledStatus(status)
        }
        #endif
    }

    private func baseQuery(for userID: String) -> [String: Any] {
        [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: "jwt.\(userID)",
        ]
    }
}
