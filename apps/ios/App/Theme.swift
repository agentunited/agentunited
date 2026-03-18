import SwiftUI

extension Color {
    static let auEmerald = Color(hex: 0x10B981)
    static let auEmeraldLight = Color(hex: 0xD1FAE5)
    static let auBackground = Color(.systemBackground)
    static let auSecondary = Color(.secondarySystemBackground)
    static let auGrouped = Color(.systemGroupedBackground)
    static let auLabel = Color(.label)
    static let auSecondaryLabel = Color(.secondaryLabel)
    static let auSeparator = Color(.separator)
    static let auBubbleOutgoing = Color.auEmerald
    static let auBubbleIncoming = Color(.systemBackground)
    static let auPresenceOnline = Color.auEmerald
    static let auPresenceOffline = Color(.systemGray3)

    init(hex: UInt32, alpha: Double = 1.0) {
        self.init(
            .sRGB,
            red: Double((hex >> 16) & 0xFF) / 255.0,
            green: Double((hex >> 8) & 0xFF) / 255.0,
            blue: Double(hex & 0xFF) / 255.0,
            opacity: alpha
        )
    }
}
