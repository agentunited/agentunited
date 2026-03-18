import SwiftUI

struct MainTabView: View {
    @EnvironmentObject private var coordinator: AppCoordinator

    var body: some View {
        TabView(selection: $coordinator.selectedTab) {
            MessagesRootView()
                .tabItem {
                    Label("Messages", systemImage: "person.2.fill")
                }
                .tag(AppCoordinator.ActiveTab.messages)

            ChannelsRootView()
                .tabItem {
                    Label("Channels", systemImage: "number")
                }
                .tag(AppCoordinator.ActiveTab.channels)

            ProfileRootView()
                .tabItem {
                    Label("Profile", systemImage: "person.crop.circle")
                }
                .tag(AppCoordinator.ActiveTab.profile)
        }
        .tint(.auEmerald)
    }
}
