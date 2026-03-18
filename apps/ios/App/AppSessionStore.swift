import Foundation
import SwiftData

@MainActor
protocol SessionPersisting: AnyObject {
    func restoreAuthentication() -> Bool
    func persistInviteSession(
        instanceURL: URL,
        userID: String,
        email: String,
        displayName: String,
        token: String,
        expiresAt: Date?,
        userType: String
    ) throws
    func signOut() throws
}

@MainActor
final class AppSessionStore: ObservableObject, SessionPersisting {
    private let modelContext: ModelContext
    private let keychainHelper: KeychainHelper

    init(modelContext: ModelContext, keychainHelper: KeychainHelper = KeychainHelper()) {
        self.modelContext = modelContext
        self.keychainHelper = keychainHelper
    }

    func restoreAuthentication() -> Bool {
        let descriptor = FetchDescriptor<WorkspaceCredential>(
            sortBy: [SortDescriptor(\WorkspaceCredential.createdAt, order: .reverse)]
        )

        guard let credential = try? modelContext.fetch(descriptor).first,
              let token = try? keychainHelper.readJWT(for: credential.userID),
              !token.isEmpty
        else {
            return false
        }

        return true
    }

    func persistInviteSession(
        instanceURL: URL,
        userID: String,
        email: String,
        displayName: String,
        token: String,
        expiresAt: Date?,
        userType: String = "human"
    ) throws {
        let descriptor = FetchDescriptor<WorkspaceCredential>()
        let existingCredentials = try modelContext.fetch(descriptor)
        let credential = existingCredentials.first(where: { $0.userID == userID })
            ?? WorkspaceCredential(
                userID: userID,
                instanceURL: instanceURL,
                displayName: displayName,
                email: email,
                userType: userType,
                jwtExpiresAt: expiresAt ?? Date.now.addingTimeInterval(60 * 60 * 24 * 30)
            )

        credential.instanceURL = instanceURL
        credential.displayName = displayName
        credential.email = email
        credential.userType = userType
        credential.jwtExpiresAt = expiresAt ?? Date.now.addingTimeInterval(60 * 60 * 24 * 30)

        if existingCredentials.contains(where: { $0.userID == userID }) == false {
            modelContext.insert(credential)
        }

        for existing in existingCredentials where existing.userID != userID {
            try? keychainHelper.deleteJWT(for: existing.userID)
            modelContext.delete(existing)
        }

        try keychainHelper.storeJWT(token, for: userID)
        try modelContext.save()
    }

    func signOut() throws {
        let credentials = try modelContext.fetch(FetchDescriptor<WorkspaceCredential>())
        for credential in credentials {
            try? keychainHelper.deleteJWT(for: credential.userID)
            modelContext.delete(credential)
        }

        for conversation in try modelContext.fetch(FetchDescriptor<Conversation>()) {
            modelContext.delete(conversation)
        }

        for message in try modelContext.fetch(FetchDescriptor<Message>()) {
            modelContext.delete(message)
        }

        for user in try modelContext.fetch(FetchDescriptor<User>()) {
            modelContext.delete(user)
        }

        try modelContext.save()
    }
}
