import Foundation
import SwiftUI

@MainActor
final class AppCoordinator: ObservableObject {
    enum AuthState: Equatable {
        case launching
        case unauthenticated
        case authenticated
    }

    enum ActiveTab: Hashable {
        case messages
        case channels
        case profile
    }

    enum Route: Equatable {
        case invite(PendingInvite)
        case claim(String)
        case dm(String)
        case channel(String)
    }

    struct PendingInvite: Equatable {
        let token: String
        let instanceURL: URL?
    }

    @Published var authState: AuthState = .launching
    @Published var selectedTab: ActiveTab = .messages
    @Published var pendingRoute: Route?
    @Published var isPresentingInvite: Bool = false
    @Published var messagesPath: [MessagesRoute] = []
    @Published var channelsPath: [ChannelsRoute] = []

    func finishLaunch(isAuthenticated: Bool) {
        authState = isAuthenticated ? .authenticated : .unauthenticated
        activatePendingRouteIfNeeded()
    }

    func setAuthenticated(_ authenticated: Bool) {
        authState = authenticated ? .authenticated : .unauthenticated
        if !authenticated {
            selectedTab = .messages
            messagesPath.removeAll()
            channelsPath.removeAll()
            isPresentingInvite = false
        } else {
            activatePendingRouteIfNeeded()
        }
    }

    func handleIncomingURL(_ url: URL) {
        guard url.scheme?.lowercased() == "agentunited" else {
            return
        }

        let pathComponents = url.pathComponents.filter { $0 != "/" }

        switch url.host?.lowercased() {
        case "invite":
            let components = URLComponents(url: url, resolvingAgainstBaseURL: false)
            let token = components?.queryItems?.first(where: { $0.name == "token" })?.value
            let instance = components?.queryItems?.first(where: { $0.name == "instance" })?.value
            guard let token, !token.isEmpty else {
                return
            }

            pendingRoute = .invite(
                PendingInvite(
                    token: token,
                    instanceURL: instance.flatMap(URL.init(string:))
                )
            )
            isPresentingInvite = authState != .authenticated

        case "claim":
            let components = URLComponents(url: url, resolvingAgainstBaseURL: false)
            guard let key = components?.queryItems?.first(where: { $0.name == "key" })?.value,
                  !key.isEmpty else {
                return
            }
            pendingRoute = .claim(key)

        case "dm":
            guard let id = pathComponents.first, !id.isEmpty else {
                return
            }
            pendingRoute = .dm(id)
            selectedTab = .messages
            activatePendingRouteIfNeeded()

        case "channel":
            guard let id = pathComponents.first, !id.isEmpty else {
                return
            }
            pendingRoute = .channel(id)
            selectedTab = .channels
            activatePendingRouteIfNeeded()

        default:
            break
        }
    }

    func consumePendingRoute() -> Route? {
        defer { pendingRoute = nil }
        return pendingRoute
    }

    func completeOnboarding(welcomeConversationID: String?) {
        pendingRoute = welcomeConversationID.map(Route.dm)
        isPresentingInvite = false
        setAuthenticated(true)
    }

    private func activatePendingRouteIfNeeded() {
        guard authState == .authenticated, let route = pendingRoute else {
            return
        }

        switch route {
        case .invite, .claim:
            break
        case let .dm(id):
            selectedTab = .messages
            messagesPath = [.conversation(id)]
            pendingRoute = nil
        case let .channel(id):
            selectedTab = .channels
            channelsPath = [.channel(id)]
            pendingRoute = nil
        }
    }
}
