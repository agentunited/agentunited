import SwiftUI

struct WorkspaceListScene: View {
    let centralJWT: String
    @EnvironmentObject private var coordinator: AppCoordinator
    @EnvironmentObject private var sessionStore: AppSessionStore
    @StateObject private var viewModel = WorkspaceListViewModel()
    @State private var isPresentingGetStarted = false

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading && viewModel.workspaces.isEmpty {
                    ProgressView("Loading workspaces…")
                        .frame(maxWidth: .infinity, maxHeight: .infinity)
                } else if viewModel.workspaces.isEmpty {
                    emptyState
                } else {
                    workspaceList
                }
            }
            .navigationTitle("Workspaces")
            .navigationBarTitleDisplayMode(.large)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button {
                        isPresentingGetStarted = true
                    } label: {
                        Image(systemName: "plus")
                    }
                    .accessibilityLabel("Add workspace")
                }
            }
            .sheet(isPresented: $isPresentingGetStarted) {
                NavigationStack {
                    GetStartedScene(sessionStore: sessionStore)
                }
                .tint(.auEmerald)
            }
        }
        .task {
            await viewModel.startPolling(centralJWT: centralJWT)
        }
        .safeAreaInset(edge: .bottom) {
            freePlanBanner
        }
    }

    // MARK: - Sub-views

    private var workspaceList: some View {
        List {
            ForEach(viewModel.workspaces) { workspace in
                Button {
                    enter(workspace)
                } label: {
                    WorkspaceRow(workspace: workspace)
                }
                .accessibilityLabel("Open workspace \(workspace.name)")
            }
        }
        .listStyle(.insetGrouped)
        .accessibilityIdentifier("workspace-list")
    }

    private var emptyState: some View {
        VStack(spacing: 24) {
            Spacer()
            Image(systemName: "server.rack")
                .font(.system(size: 52))
                .foregroundStyle(Color.auEmerald)
                .accessibilityHidden(true)
            VStack(spacing: 8) {
                Text("No workspaces yet")
                    .font(.title3.bold())
                Text("Your agent needs to set up a workspace first,\nor you can connect one yourself.")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
            }
            Button("Set up a workspace →") {
                isPresentingGetStarted = true
            }
            .buttonStyle(AUPrimaryButtonStyle())
            .padding(.horizontal, 32)
            Spacer()
        }
        .padding(.horizontal, 24)
        .accessibilityIdentifier("workspace-list-empty")
    }

    private var freePlanBanner: some View {
        HStack(spacing: 8) {
            Image(systemName: "bolt.circle.fill")
                .foregroundStyle(Color.auEmerald)
                .accessibilityHidden(true)
            Text("Free plan · 1 workspace · 3 entities · 5 GB/mo")
                .font(.caption)
                .foregroundStyle(.secondary)
            Spacer(minLength: 4)
            Button("Upgrade →") {
                // TODO: relay settings / paywall
            }
            .font(.caption.weight(.semibold))
            .foregroundStyle(Color.auEmerald)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 10)
        .background(.regularMaterial)
        .overlay(alignment: .top) { Divider() }
    }

    // MARK: - Actions

    private func enter(_ workspace: CentralWorkspace) {
        guard let url = URL(string: workspace.relayURL) else { return }
        let payload = centralJWT.decodedJWTPayload()
        let userID = payload?["sub"] as? String ?? "central-user"
        let email = payload?["email"] as? String ?? ""
        let displayName = payload?["display_name"] as? String ?? "User"
        try? sessionStore.persistInviteSession(
            instanceURL: url,
            userID: userID,
            email: email,
            displayName: displayName,
            token: centralJWT,
            expiresAt: nil,
            userType: "human"
        )
        coordinator.setAuthenticated(true)
    }
}

// MARK: - Row

private struct WorkspaceRow: View {
    let workspace: CentralWorkspace

    var body: some View {
        HStack(spacing: 12) {
            ZStack {
                Circle()
                    .fill(Color.auEmerald.opacity(0.15))
                    .frame(width: 40, height: 40)
                Image(systemName: "network")
                    .font(.system(size: 16, weight: .medium))
                    .foregroundStyle(Color.auEmerald)
            }
            .accessibilityHidden(true)

            VStack(alignment: .leading, spacing: 2) {
                Text(workspace.name)
                    .font(.body.weight(.medium))
                Text(workspace.relayURL)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .lineLimit(1)
            }

            Spacer()

            Image(systemName: "chevron.right")
                .font(.caption.weight(.semibold))
                .foregroundStyle(Color(.tertiaryLabel))
                .accessibilityHidden(true)
        }
        .padding(.vertical, 4)
        .contentShape(Rectangle())
    }
}

// MARK: - ViewModel

@MainActor
final class WorkspaceListViewModel: ObservableObject {
    @Published var workspaces: [CentralWorkspace] = []
    @Published var isLoading = false

    private var pollingTask: Task<Void, Never>?

    func startPolling(centralJWT: String) async {
        let client = CentralAPIClient(authToken: centralJWT)
        isLoading = true
        workspaces = (try? await client.listWorkspaces()) ?? []
        isLoading = false

        pollingTask?.cancel()
        pollingTask = Task {
            while !Task.isCancelled {
                try? await Task.sleep(nanoseconds: 30_000_000_000)
                guard !Task.isCancelled else { break }
                if let fresh = try? await client.listWorkspaces() {
                    workspaces = fresh
                }
            }
        }
    }

    deinit { pollingTask?.cancel() }
}
