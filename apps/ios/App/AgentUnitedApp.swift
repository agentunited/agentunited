import Foundation
import OSLog
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
        let schema = Schema([WorkspaceCredential.self, Conversation.self, Message.self, User.self])
        let config = ModelConfiguration(schema: schema, isStoredInMemoryOnly: false)

        do {
            return try ModelContainer(for: schema, configurations: [config])
        } catch {
            // First failure: wipe store and retry (handles schema migrations)
            os_log("ModelContainer failed (attempt 1), wiping store: %@", error.localizedDescription)
            try? FileManager.default.removeItem(at: config.url)

            do {
                return try ModelContainer(for: schema, configurations: [config])
            } catch {
                // Second failure: fall back to in-memory store — app works, data doesn't persist
                os_log("ModelContainer unrecoverable on disk, using in-memory fallback: %@", error.localizedDescription)
                let memConfig = ModelConfiguration(schema: schema, isStoredInMemoryOnly: true)
                // swiftlint:disable:next force_try
                return try! ModelContainer(for: schema, configurations: [memConfig])
            }
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
