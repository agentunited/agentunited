import Foundation
import SwiftData

// MARK: - SwiftData Models
// JWT tokens are stored in Keychain (see Core/Keychain/KeychainHelper.swift), never here.

@Model
final class WorkspaceCredential {
    @Attribute(.unique) var userID: String
    var instanceURL: URL
    var displayName: String
    var email: String
    var userType: String   // "human" | "agent"
    var jwtExpiresAt: Date
    var createdAt: Date

    init(
        userID: String,
        instanceURL: URL,
        displayName: String,
        email: String,
        userType: String,
        jwtExpiresAt: Date,
        createdAt: Date = .now
    ) {
        self.userID = userID
        self.instanceURL = instanceURL
        self.displayName = displayName
        self.email = email
        self.userType = userType
        self.jwtExpiresAt = jwtExpiresAt
        self.createdAt = createdAt
    }
}

@Model
final class Conversation {
    @Attribute(.unique) var id: String
    var name: String
    var type: String           // "dm" | "channel"
    var topic: String?
    var lastMessageText: String?
    var lastMessageAt: Date?
    var unreadCount: Int
    var isMuted: Bool
    var memberCount: Int
    var participantIDs: [String]
    @Relationship(deleteRule: .cascade, inverse: \Message.conversation)
    var messages: [Message]

    init(
        id: String,
        name: String,
        type: String,
        topic: String? = nil,
        lastMessageText: String? = nil,
        lastMessageAt: Date? = nil,
        unreadCount: Int = 0,
        isMuted: Bool = false,
        memberCount: Int = 0,
        participantIDs: [String] = [],
        messages: [Message] = []
    ) {
        self.id = id
        self.name = name
        self.type = type
        self.topic = topic
        self.lastMessageText = lastMessageText
        self.lastMessageAt = lastMessageAt
        self.unreadCount = unreadCount
        self.isMuted = isMuted
        self.memberCount = memberCount
        self.participantIDs = participantIDs
        self.messages = messages
    }
}

@Model
final class Message {
    @Attribute(.unique) var id: String
    var conversationID: String
    var text: String
    var authorID: String
    var authorName: String
    var authorType: String   // "human" | "agent"
    var timestamp: Date
    var replyToID: String?
    var retryCount: Int
    var isPending: Bool      // optimistic send — true until server confirms
    var isFailed: Bool       // send failed — show retry UI
    var conversation: Conversation?

    init(
        id: String,
        conversationID: String,
        text: String,
        authorID: String,
        authorName: String,
        authorType: String,
        timestamp: Date,
        replyToID: String? = nil,
        retryCount: Int = 0,
        isPending: Bool = false,
        isFailed: Bool = false,
        conversation: Conversation? = nil
    ) {
        self.id = id
        self.conversationID = conversationID
        self.text = text
        self.authorID = authorID
        self.authorName = authorName
        self.authorType = authorType
        self.timestamp = timestamp
        self.replyToID = replyToID
        self.retryCount = retryCount
        self.isPending = isPending
        self.isFailed = isFailed
        self.conversation = conversation
    }
}

@Model
final class User {
    @Attribute(.unique) var id: String
    var displayName: String
    var email: String
    var userType: String   // "human" | "agent"
    var avatarURL: URL?
    var isOnline: Bool

    init(
        id: String,
        displayName: String,
        email: String,
        userType: String,
        avatarURL: URL? = nil,
        isOnline: Bool = false
    ) {
        self.id = id
        self.displayName = displayName
        self.email = email
        self.userType = userType
        self.avatarURL = avatarURL
        self.isOnline = isOnline
    }
}
