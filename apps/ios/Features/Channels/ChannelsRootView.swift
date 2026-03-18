import SwiftUI

struct ChannelsRootView: View {
    @EnvironmentObject private var coordinator: AppCoordinator

    var body: some View {
        NavigationStack(path: $coordinator.channelsPath) {
            ContentUnavailableView(
                "No channels yet",
                systemImage: "number",
                description: Text("Joined channels will appear here.")
            )
            .navigationTitle("Channels")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Image(systemName: "plus")
                        .foregroundStyle(Color.auEmerald)
                }
            }
            .navigationDestination(for: ChannelsRoute.self) { route in
                switch route {
                case let .channel(id):
                    ChannelPlaceholderView(title: "#\(id)")
                }
            }
        }
    }
}

enum ChannelsRoute: Hashable {
    case channel(String)
}

private struct ChannelPlaceholderView: View {
    let title: String

    var body: some View {
        Text(title)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(Color.auGrouped)
    }
}
