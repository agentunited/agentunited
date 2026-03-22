import Combine
import SwiftData
import SwiftUI

struct ChannelsRootView: View {
    @Environment(\.modelContext) private var modelContext
    @EnvironmentObject private var coordinator: AppCoordinator
    @StateObject private var webSocketManager = AUWebSocketManager()
    @StateObject private var listViewModel = ChannelListViewModel()

    var body: some View {
        NavigationStack(path: $coordinator.channelsPath) {
            ChannelListView(
                viewModel: listViewModel,
                onSelectChannel: { channelID in
                    coordinator.channelsPath.append(.channel(channelID))
                }
            )
            .navigationDestination(for: ChannelsRoute.self) { route in
                switch route {
                case let .channel(id):
                    ChannelConversationView(channelID: id)
                        .environmentObject(webSocketManager)
                }
            }
        }
        .task {
            webSocketManager.configure(modelContext: modelContext)
            await webSocketManager.connect()
            listViewModel.configure(modelContext: modelContext, webSocketManager: webSocketManager)
            await listViewModel.loadIfNeeded()
        }
        .onDisappear {
            webSocketManager.disconnect()
        }
    }
}

enum ChannelsRoute: Hashable {
    case channel(String)
}

@MainActor
final class ChannelListViewModel: ObservableObject {
    enum Phase: Equatable {
        case idle
        case loading
        case loaded
        case error(MessagesFeatureError)
    }

    @Published private(set) var phase: Phase = .idle
    @Published private(set) var channels: [ConversationListItem] = []

    private var store: ChannelListStoreProtocol?
    private var cancellables = Set<AnyCancellable>()
    private var hasLoaded = false

    init(store: ChannelListStoreProtocol? = nil) {
        self.store = store
    }

    func configure(modelContext: ModelContext, webSocketManager: AUWebSocketManager) {
        guard store == nil else {
            return
        }

        store = LiveMessagesStore(modelContext: modelContext)
        webSocketManager.$lastEvent
            .receive(on: RunLoop.main)
            .sink { [weak self] event in
                guard let self, let event else {
                    return
                }

                switch event {
                case let .message(conversationID, _):
                    Task { [weak self] in
                        guard let self else {
                            return
                        }
                        if self.channels.contains(where: { $0.id == conversationID }) {
                            await self.reloadFromLocal()
                        }
                    }
                case .typing, .presence:
                    break
                }
            }
            .store(in: &cancellables)
    }

    func loadIfNeeded() async {
        guard hasLoaded == false else {
            return
        }
        hasLoaded = true
        await refresh(showLoading: true)
    }

    func refresh(showLoading: Bool = false) async {
        guard let store else {
            return
        }

        if showLoading {
            phase = .loading
        }

        do {
            channels = try await store.syncChannelList()
            phase = .loaded
        } catch let error as MessagesFeatureError {
            phase = .error(error)
            channels = (try? store.localChannelList()) ?? []
        } catch {
            phase = .error(.generic(error.localizedDescription))
        }
    }

    func reloadFromLocal() async {
        guard let store else {
            return
        }

        channels = (try? store.localChannelList()) ?? channels
        if case .idle = phase {
            phase = .loaded
        }
    }

    func muteChannel(id: String) {
        guard let store else {
            return
        }

        do {
            channels = try store.toggleMuteConversation(id: id)
            phase = .loaded
        } catch let error as MessagesFeatureError {
            phase = .error(error)
        } catch {
            phase = .error(.generic(error.localizedDescription))
        }
    }
}

@MainActor
protocol ChannelListStoreProtocol: AnyObject {
    func syncChannelList() async throws -> [ConversationListItem]
    func localChannelList() throws -> [ConversationListItem]
    func toggleMuteConversation(id: String) throws -> [ConversationListItem]
}

extension LiveMessagesStore: ChannelListStoreProtocol {}

private struct ChannelListView: View {
    @ObservedObject var viewModel: ChannelListViewModel
    let onSelectChannel: (String) -> Void

    var body: some View {
        Group {
            switch viewModel.phase {
            case .idle:
                ProgressView("Loading channels")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            case .loading:
                List(0..<3, id: \.self) { _ in
                    ConversationRowSkeleton()
                        .listRowSeparator(.hidden)
                }
                .listStyle(.plain)
                .redacted(reason: .placeholder)
            case .loaded:
                if viewModel.channels.isEmpty {
                    ContentUnavailableView(
                        "No channels yet.",
                        systemImage: "number",
                        description: Text("Your agent will invite you to channels.")
                    )
                } else {
                    channelList
                }
            case let .error(error):
                VStack(spacing: 16) {
                    errorBanner(text: error.errorDescription ?? "Couldn't load channels. Pull to refresh.")
                    if viewModel.channels.isEmpty {
                        ContentUnavailableView(
                            "Channels unavailable",
                            systemImage: "wifi.exclamationmark",
                            description: Text("Pull to refresh and try again.")
                        )
                    } else {
                        channelList
                    }
                }
            }
        }
        .navigationTitle("Channels")
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Image(systemName: "plus")
                    .foregroundStyle(Color.auEmerald)
                    .accessibilityLabel("Create channel")
            }
        }
        .refreshable {
            await viewModel.refresh()
        }
    }

    private var channelList: some View {
        List(viewModel.channels) { channel in
            Button {
                onSelectChannel(channel.id)
            } label: {
                ChannelRow(item: channel)
            }
            .buttonStyle(.plain)
            .contextMenu {
                Button(channel.isMuted ? "Unmute" : "Mute") {
                    viewModel.muteChannel(id: channel.id)
                }
            }
            .listRowSeparator(.hidden)
        }
        .listStyle(.plain)
    }

    private func errorBanner(text: String) -> some View {
        Text(text)
            .font(.subheadline)
            .foregroundStyle(.primary)
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.horizontal, 16)
            .padding(.vertical, 12)
            .background(Color.red.opacity(0.08))
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .padding(.horizontal, 16)
    }
}

private struct ChannelConversationView: View {
    let channelID: String

    var body: some View {
        ConversationView(conversationID: channelID, mode: .channel)
    }
}

private struct ChannelRow: View {
    let item: ConversationListItem

    var body: some View {
        HStack(spacing: 12) {
            ZStack {
                RoundedRectangle(cornerRadius: 10)
                    .fill(Color.auEmerald.opacity(0.14))
                    .frame(width: 42, height: 42)
                Image(systemName: "number")
                    .foregroundStyle(Color.auEmerald)
            }

            VStack(alignment: .leading, spacing: 4) {
                HStack(spacing: 8) {
                    Text("#\(item.displayName)")
                        .font(.headline)
                        .foregroundStyle(.primary)
                    Spacer()
                    Text(item.relativeTimestamp)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }

                HStack(spacing: 8) {
                    Text(item.lastMessagePreview)
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .lineLimit(2)
                    Spacer()
                    if item.unreadCount > 0 {
                        Text(item.unreadCount > 9 ? "9+" : "\(item.unreadCount)")
                            .font(.caption2.bold())
                            .foregroundStyle(.white)
                            .padding(.horizontal, 6)
                            .padding(.vertical, 3)
                            .background(Color.blue)
                            .clipShape(Capsule())
                    }
                }

                Text("\(max(item.memberCount, 1)) members")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
        .padding(.vertical, 6)
        .accessibilityElement(children: .combine)
        .accessibilityLabel("Channel #\(item.displayName), \(item.unreadCount) unread")
    }
}
