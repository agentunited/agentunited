import SwiftUI

struct AUPrimaryButtonStyle: ButtonStyle {
    @Environment(\.isEnabled) private var isEnabled

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.headline)
            .frame(maxWidth: .infinity)
            .frame(height: 50)
            .padding(.horizontal, 16)
            .background(
                Color.auEmerald.opacity(
                    isEnabled
                        ? (configuration.isPressed ? 0.85 : 1.0)
                        : 0.45
                )
            )
            .foregroundStyle(.white)
            .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
            .shadow(color: Color.auEmerald.opacity(isEnabled ? 0.18 : 0), radius: 12, y: 8)
    }
}

struct AUGhostButtonStyle: ButtonStyle {
    @Environment(\.isEnabled) private var isEnabled

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.headline)
            .frame(maxWidth: .infinity)
            .frame(height: 50)
            .padding(.horizontal, 16)
            .background(
                RoundedRectangle(cornerRadius: 12, style: .continuous)
                    .fill(Color.white.opacity(isEnabled ? (configuration.isPressed ? 0.08 : 0.02) : 0.01))
            )
            .foregroundStyle(Color.white.opacity(isEnabled ? 1.0 : 0.6))
            .overlay {
                RoundedRectangle(cornerRadius: 12, style: .continuous)
                    .stroke(Color.white.opacity(isEnabled ? (configuration.isPressed ? 0.72 : 0.84) : 0.32), lineWidth: 1)
            }
    }
}

struct AUSecondaryButtonStyle: ButtonStyle {
    @Environment(\.isEnabled) private var isEnabled

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.headline)
            .frame(maxWidth: .infinity)
            .frame(height: 50)
            .padding(.horizontal, 16)
            .background(Color.auSecondary.opacity(isEnabled ? (configuration.isPressed ? 0.7 : 1.0) : 0.55))
            .foregroundStyle(Color.auLabel)
            .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
            .overlay {
                RoundedRectangle(cornerRadius: 12, style: .continuous)
                    .stroke(Color.auSeparator.opacity(0.28), lineWidth: 1)
            }
    }
}
