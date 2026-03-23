import SwiftData
import SwiftUI

@main
struct AgentUnitedApp: App {
    private let sharedModelContainer: ModelContainer = AgentUnitedApp.makeModelContainer()
    @StateObject private var coordinator = AppCoordinator()
    @StateObject private var sessionStore: AppSessionStore

    init() {
        let modelContext = sharedModelContainer.mainContext
        _sessionStore = StateObject(wrappedValue: AppSessionStore(modelContext: modelContext))
    }

    var body: some Scene {
        WindowGroup {
            RootView()
                .environmentObject(coordinator)
                .environmentObject(sessionStore)
                .onOpenURL { url in
                    coordinator.handleIncomingURL(url)
                }
        }
        .modelContainer(sharedModelContainer)
    }

    private static func makeModelContainer() -> ModelContainer {
        do {
            return try ModelContainer(for: WorkspaceCredential.self, Conversation.self, Message.self, User.self)
        } catch {
            fatalError("Failed to create model container: \(error.localizedDescription)")
        }
    }
}

private struct RootView: View {
    @EnvironmentObject private var coordinator: AppCoordinator
    @EnvironmentObject private var sessionStore: AppSessionStore

    var body: some View {
        ZStack {
            Color.clear.ignoresSafeArea()

            switch coordinator.authState {
            case .authenticated:
                MainTabView()
            case .unauthenticated, .launching:
                OnboardingRootView()
            }
        }
        .preferredColorScheme(.light) // AU v1 — light mode only; moved inside RootView so it doesn't wrap the safe-area context
        .task {
            if coordinator.authState == .launching {
                coordinator.finishLaunch(isAuthenticated: sessionStore.restoreAuthentication())
            }
        }
    }
}
