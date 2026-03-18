import Combine
import Network
import SwiftData
import SwiftUI
import UserNotifications
import UIKit

struct MessagesRootView: View {
    @Environment(\.modelContext) private var modelContext
    @EnvironmentObject private var coordinator: AppCoordinator
    @StateObject private var webSocketManager = AUWebSocketManager()
    @StateObject private var listViewModel = MessagesListViewModel()

    var body: some View {
        NavigationStack(path: $coordinator.messagesPath) {
            ConversationListView(
                viewModel: listViewModel,
                onSelectConversation: { conversationID in
                    coordinator.messagesPath.append(.conversation(conversationID))
                },
                onCompose: {
                    coordinator.messagesPath.append(.newMessage)
                }
            )
            .navigationDestination(for: MessagesRoute.self) { route in
                switch route {
                case .newMessage:
                    NewMessageView(
                        viewModel: NewMessageViewModel(),
                        onOpenConversation: { conversationID in
                            if coordinator.messagesPath.last == .newMessage {
                                coordinator.messagesPath.removeLast()
                            }
                            coordinator.messagesPath.append(.conversation(conversationID))
                        }
                    )
                    .environmentObject(webSocketManager)
                case let .conversation(id):
                    ConversationView(conversationID: id)
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

enum MessagesRoute: Hashable {
    case newMessage
    case conversation(String)
}

@MainActor
final class MessagesListViewModel: ObservableObject {
    enum Phase: Equatable {
        case idle
        case loading
        case loaded
        case error(MessagesFeatureError)
    }

    @Published private(set) var phase: Phase = .idle
    @Published private(set) var conversations: [ConversationListItem] = []

    private var store: LiveMessagesStore?
    private var cancellables = Set<AnyCancellable>()
    private var hasLoaded = false

    func configure(modelContext: ModelContext, webSocketManager: AUWebSocketManager) {
        guard store == nil else {
            return
        }

        store = LiveMessagesStore(modelContext: modelContext)
        webSocketManager.$lastEvent
            .receive(on: RunLoop.main)
            .sink { [weak self] event in
                guard event != nil else {
                    return
                }
                Task { [weak self] in
                    await self?.reloadFromLocal()
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
            conversations = try await store.syncConversationList()
            phase = .loaded
        } catch let error as MessagesFeatureError {
            phase = .error(error)
            conversations = (try? store.localConversationList()) ?? []
        } catch {
            phase = .error(.generic(error.localizedDescription))
        }
    }

    func deleteConversation(id: String) {
        guard let store else {
            return
        }

        do {
            try store.deleteConversation(id: id)
            conversations.removeAll { $0.id == id }
            phase = .loaded
        } catch let error as MessagesFeatureError {
            phase = .error(error)
        } catch {
            phase = .error(.generic(error.localizedDescription))
        }
    }

    func muteConversation(id: String) {
        guard let store else {
            return
        }

        do {
            conversations = try store.toggleMuteConversation(id: id)
            phase = .loaded
        } catch let error as MessagesFeatureError {
            phase = .error(error)
        } catch {
            phase = .error(.generic(error.localizedDescription))
        }
    }

    func reloadFromLocal() async {
        guard let store else {
            return
        }

        conversations = (try? store.localConversationList()) ?? conversations
        if case .idle = phase {
            phase = .loaded
        }
    }
}

@MainActor
final class NewMessageViewModel: ObservableObject {
    enum Phase: Equatable {
        case idle
        case loading
        case loaded
        case error(MessagesFeatureError)
    }

    @Published var searchText = ""
    @Published private(set) var phase: Phase = .idle
    @Published private(set) var users: [UserListItem] = []

    private var store: LiveMessagesStore?

    var filteredUsers: [UserListItem] {
        guard searchText.isEmpty == false else {
            return users
        }

        return users.filter {
            $0.displayName.localizedCaseInsensitiveContains(searchText)
                || $0.email.localizedCaseInsensitiveContains(searchText)
        }
    }

    func configure(modelContext: ModelContext) {
        if store == nil {
            store = LiveMessagesStore(modelContext: modelContext)
        }
    }

    func loadIfNeeded() async {
        guard case .idle = phase, let store else {
            return
        }

        phase = .loading
        do {
            users = try await store.syncUsers()
            phase = .loaded
        } catch let error as MessagesFeatureError {
            phase = .error(error)
        } catch {
            phase = .error(.generic(error.localizedDescription))
        }
    }

    func openConversation(for userID: String) async throws -> String {
        guard let store else {
            throw MessagesFeatureError.missingSession
        }

        let conversation = try await store.openOrCreateConversation(targetUserID: userID)
        return conversation.id
    }
}

@MainActor
final class ConversationViewModel: ObservableObject {
    enum Phase: Equatable {
        case idle
        case loading
        case loaded
        case error(MessagesFeatureError)
    }

    @Published var inputText = ""
    @Published private(set) var phase: Phase = .idle
    @Published private(set) var header = ConversationHeader.placeholder
    @Published private(set) var messages: [MessageItem] = []
    @Published private(set) var replyContext: MessageItem?
    @Published private(set) var failureBanner: SendFailureBanner?
    @Published private(set) var isOffline = false
    @Published private(set) var isShowingTypingIndicator = false
    @Published private(set) var typingIndicatorText: String?

    private let conversationID: String
    private let mode: ConversationKind
    private let notificationRequester: NotificationPermissionRequesting
    private let networkMonitor: NetworkStatusProviding
    private var store: ConversationStoreProtocol?
    private weak var webSocketManager: AUWebSocketManager?
    private var cancellables = Set<AnyCancellable>()
    private var typingDebounceTask: Task<Void, Never>?
    private var agentThinkingTask: Task<Void, Never>?
    private var hasRequestedPushPermission = false
    private var isAgentThinking = false

    init(
        conversationID: String,
        mode: ConversationKind = .dm,
        store: ConversationStoreProtocol? = nil,
        notificationRequester: NotificationPermissionRequesting = PushPermissionRequester(),
        networkMonitor: NetworkStatusProviding = NetworkStatusMonitor()
    ) {
        self.conversationID = conversationID
        self.mode = mode
        self.store = store
        self.notificationRequester = notificationRequester
        self.networkMonitor = networkMonitor
        header = mode == .channel ? .channelPlaceholder : .placeholder
    }

    func configure(modelContext: ModelContext, webSocketManager: AUWebSocketManager) {
        if store == nil {
            store = LiveMessagesStore(modelContext: modelContext)
        }
        if self.webSocketManager == nil {
            self.webSocketManager = webSocketManager
            bindWebSocket(webSocketManager)
            bindNetworkMonitor()
        }
    }

    func loadIfNeeded() async {
        guard let store else {
            return
        }
        guard case .idle = phase else {
            return
        }

        phase = .loading
        do {
            let screen = try await store.loadConversationScreen(conversationID: conversationID)
            header = screen.header
            messages = screen.messages
            phase = .loaded
            refreshTypingIndicator()
        } catch let error as MessagesFeatureError {
            phase = .error(error)
        } catch {
            phase = .error(.generic(error.localizedDescription))
        }
    }

    func refresh() async {
        guard let store else {
            return
        }

        do {
            let screen = try await store.loadConversationScreen(conversationID: conversationID)
            header = screen.header
            messages = screen.messages
            phase = .loaded
            refreshTypingIndicator()
        } catch let error as MessagesFeatureError {
            phase = .error(error)
        } catch {
            phase = .error(.generic(error.localizedDescription))
        }
    }

    func onInputChanged(_ text: String) {
        inputText = text
        typingDebounceTask?.cancel()
        typingDebounceTask = Task { [weak self] in
            try? await Task.sleep(for: .milliseconds(500))
            await self?.webSocketManager?.sendTyping(channelId: self?.conversationID ?? "")
        }
    }

    func sendCurrentMessage() async {
        let trimmedText = inputText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard trimmedText.isEmpty == false, let store else {
            return
        }

        let replyToID = replyContext?.id
        inputText = ""
        replyContext = nil
        failureBanner = nil
        var optimisticMessageID: String?

        do {
            let optimisticMessage = try store.insertOptimisticMessage(
                conversationID: conversationID,
                text: trimmedText,
                replyToID: replyToID
            )
            optimisticMessageID = optimisticMessage.id
            messages.append(optimisticMessage)
            phase = .loaded
            startAgentThinking()

            let sentMessage = try await store.commitMessage(
                conversationID: conversationID,
                text: trimmedText,
                replyToID: replyToID,
                optimisticID: optimisticMessage.id
            )
            replaceMessage(sentMessage)
            if hasRequestedPushPermission == false {
                hasRequestedPushPermission = true
                Task { [notificationRequester] in
                    await notificationRequester.requestAfterFirstMessageSend()
                }
            }
        } catch let error as MessagesFeatureError {
            if let optimisticMessageID,
               let failedMessage = try? store.markMessageFailed(id: optimisticMessageID) {
                replaceMessage(failedMessage)
                failureBanner = SendFailureBanner(messageID: failedMessage.id, text: error.errorDescription ?? "Message failed to send.")
            }
            phase = .loaded
        } catch {
            phase = .error(.generic(error.localizedDescription))
        }
    }

    func retryFailedMessage() async {
        guard let failedMessageID = failureBanner?.messageID, let store else {
            return
        }

        do {
            let sentMessage = try await store.retryFailedMessage(
                messageID: failedMessageID,
                conversationID: conversationID
            )
            replaceMessage(sentMessage)
            failureBanner = nil
            startAgentThinking()
        } catch let error as MessagesFeatureError {
            failureBanner = SendFailureBanner(messageID: failedMessageID, text: error.errorDescription ?? "Retry failed.")
        } catch {
            failureBanner = SendFailureBanner(messageID: failedMessageID, text: error.localizedDescription)
        }
    }

    func setReplyContext(_ message: MessageItem) {
        replyContext = message
    }

    func clearReplyContext() {
        replyContext = nil
    }

    func deleteOwnMessage(_ message: MessageItem) {
        guard let store else {
            return
        }

        do {
            try store.deleteMessage(id: message.id)
            messages.removeAll { $0.id == message.id }
        } catch {
            failureBanner = SendFailureBanner(messageID: message.id, text: error.localizedDescription)
        }
    }

    private func replaceMessage(_ message: MessageItem) {
        if let index = messages.firstIndex(where: { $0.id == message.id || ($0.isPending && $0.text == message.text && $0.timestamp == message.timestamp) }) {
            messages[index] = message
        } else if let index = messages.firstIndex(where: { $0.isPending && $0.text == message.text }) {
            messages[index] = message
        } else {
            messages.append(message)
        }
        messages.sort { $0.timestamp < $1.timestamp }
    }

    private func bindWebSocket(_ webSocketManager: AUWebSocketManager) {
        webSocketManager.$lastEvent
            .receive(on: RunLoop.main)
            .sink { [weak self] event in
                guard let self, let event else {
                    return
                }

                switch event {
                case let .message(conversationID, _):
                    guard conversationID == self.conversationID else {
                        return
                    }
                    Task { [weak self] in
                        self?.agentThinkingTask?.cancel()
                        self?.agentThinkingTask = nil
                        self?.isAgentThinking = false
                        self?.isShowingTypingIndicator = false
                        await self?.refresh()
                    }
                case let .typing(conversationID, _):
                    guard conversationID == self.conversationID else {
                        return
                    }
                    self.refreshTypingIndicator()
                case .presence:
                    Task { [weak self] in
                        await self?.refresh()
                    }
                }
            }
            .store(in: &cancellables)

        webSocketManager.$typingUsers
            .receive(on: RunLoop.main)
            .sink { [weak self] _ in
                self?.refreshTypingIndicator()
            }
            .store(in: &cancellables)
    }

    private func bindNetworkMonitor() {
        networkMonitor.statusPublisher
            .receive(on: RunLoop.main)
            .sink { [weak self] isConnected in
                self?.isOffline = !isConnected
            }
            .store(in: &cancellables)
        networkMonitor.start()
    }

    private func startAgentThinking() {
        agentThinkingTask?.cancel()
        isAgentThinking = true
        refreshTypingIndicator()
        agentThinkingTask = Task { [weak self] in
            try? await Task.sleep(for: .seconds(30))
            guard let self else {
                return
            }
            self.isAgentThinking = false
            self.agentThinkingTask = nil
            self.refreshTypingIndicator()
        }
    }

    private func refreshTypingIndicator() {
        let typingUserIDs = (webSocketManager?.typingUsers[conversationID] ?? [:]).keys.sorted()
        let resolvedNames = (try? store?.typingDisplayNames(conversationID: conversationID, userIDs: typingUserIDs)) ?? []
        let names = resolvedNames.filter { $0.isEmpty == false }

        if mode == .channel {
            if names.count >= 2 {
                typingIndicatorText = "\(names[0]) and \(names[1]) are typing…"
            } else if let first = names.first {
                typingIndicatorText = "\(first) is typing…"
            } else if isAgentThinking {
                typingIndicatorText = "Someone is typing…"
            } else {
                typingIndicatorText = nil
            }
        } else if let first = names.first {
            typingIndicatorText = "\(first) is typing…"
        } else if isAgentThinking {
            typingIndicatorText = "\(header.displayName) is typing…"
        } else {
            typingIndicatorText = nil
        }

        isShowingTypingIndicator = typingIndicatorText != nil
    }
}

private struct ConversationListView: View {
    @ObservedObject var viewModel: MessagesListViewModel
    let onSelectConversation: (String) -> Void
    let onCompose: () -> Void

    var body: some View {
        Group {
            switch viewModel.phase {
            case .idle:
                ProgressView("Loading messages")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            case .loading:
                List(0..<3, id: \.self) { _ in
                    ConversationRowSkeleton()
                        .listRowSeparator(.hidden)
                }
                .listStyle(.plain)
                .redacted(reason: .placeholder)
            case .loaded:
                if viewModel.conversations.isEmpty {
                    ContentUnavailableView(
                        "No messages yet.",
                        systemImage: "bubble.left.and.bubble.right",
                        description: Text("Start a conversation to talk with an agent or teammate.")
                    )
                    .overlay(alignment: .bottom) {
                        Button("Start a conversation", action: onCompose)
                            .buttonStyle(.borderedProminent)
                            .tint(.auEmerald)
                            .padding(.bottom, 24)
                            .accessibilityLabel("Start a conversation")
                    }
                } else {
                    conversationList
                }
            case let .error(error):
                VStack(spacing: 16) {
                    errorBanner(text: error.errorDescription ?? "Couldn't load conversations. Pull to refresh.")
                    if viewModel.conversations.isEmpty {
                        ContentUnavailableView("Messages unavailable", systemImage: "wifi.exclamationmark")
                    } else {
                        conversationList
                    }
                }
            }
        }
        .navigationTitle("Messages")
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Button(action: onCompose) {
                    Image(systemName: "square.and.pencil")
                        .foregroundStyle(Color.auEmerald)
                }
                .accessibilityLabel("Compose a new message")
            }
        }
        .refreshable {
            await viewModel.refresh()
        }
    }

    private var conversationList: some View {
        List(viewModel.conversations) { conversation in
            Button {
                onSelectConversation(conversation.id)
            } label: {
                ConversationRow(item: conversation)
            }
            .buttonStyle(.plain)
            .swipeActions(edge: .trailing, allowsFullSwipe: false) {
                Button(role: .destructive) {
                    viewModel.deleteConversation(id: conversation.id)
                } label: {
                    Label("Delete", systemImage: "trash")
                }
                .accessibilityLabel("Delete conversation with \(conversation.displayName)")

                Button {
                    viewModel.muteConversation(id: conversation.id)
                } label: {
                    Label(conversation.isMuted ? "Unmute" : "Mute", systemImage: "bell.slash")
                }
                .tint(.orange)
                .accessibilityLabel("\(conversation.isMuted ? "Unmute" : "Mute") conversation with \(conversation.displayName)")
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

private struct NewMessageView: View {
    @Environment(\.modelContext) private var modelContext
    @ObservedObject var viewModel: NewMessageViewModel
    let onOpenConversation: (String) -> Void

    var body: some View {
        Group {
            switch viewModel.phase {
            case .idle, .loading:
                List(0..<5, id: \.self) { _ in
                    HStack {
                        Circle().frame(width: 44, height: 44)
                        VStack(alignment: .leading) {
                            RoundedRectangle(cornerRadius: 4).frame(width: 120, height: 14)
                            RoundedRectangle(cornerRadius: 4).frame(width: 180, height: 12)
                        }
                    }
                    .redacted(reason: .placeholder)
                    .listRowSeparator(.hidden)
                }
                .listStyle(.plain)
            case .loaded:
                if viewModel.filteredUsers.isEmpty {
                    ContentUnavailableView(
                        "No one here yet.",
                        systemImage: "person.2.slash",
                        description: Text("Your agent will appear once they're online.")
                    )
                } else {
                    List(viewModel.filteredUsers) { user in
                        Button {
                            Task {
                                if let conversationID = try? await viewModel.openConversation(for: user.id) {
                                    onOpenConversation(conversationID)
                                }
                            }
                        } label: {
                            HStack(spacing: 12) {
                                AvatarView(initials: user.initials, isAgent: user.isAgent)
                                VStack(alignment: .leading, spacing: 4) {
                                    HStack {
                                        Text(user.displayName)
                                            .foregroundStyle(.primary)
                                        if user.isAgent {
                                            AgentBadge()
                                        }
                                    }
                                    Text(user.email)
                                        .font(.subheadline)
                                        .foregroundStyle(.secondary)
                                }
                                Spacer()
                                PresenceDot(isOnline: user.isOnline)
                            }
                        }
                        .buttonStyle(.plain)
                        .listRowSeparator(.hidden)
                        .accessibilityLabel("Message \(user.displayName)")
                    }
                    .listStyle(.plain)
                }
            case let .error(error):
                ContentUnavailableView(
                    "Couldn't load people",
                    systemImage: "exclamationmark.triangle",
                    description: Text(error.errorDescription ?? "Try again.")
                )
            }
        }
        .navigationTitle("New Message")
        .searchable(text: $viewModel.searchText, prompt: "Search agents and people")
        .task {
            viewModel.configure(modelContext: modelContext)
            await viewModel.loadIfNeeded()
        }
    }
}

struct ConversationView: View {
    @Environment(\.modelContext) private var modelContext
    @EnvironmentObject private var webSocketManager: AUWebSocketManager
    @StateObject private var viewModel: ConversationViewModel

    init(conversationID: String, mode: ConversationKind = .dm) {
        _viewModel = StateObject(wrappedValue: ConversationViewModel(conversationID: conversationID, mode: mode))
    }

    var body: some View {
        VStack(spacing: 0) {
            if viewModel.isOffline {
                statusBanner(text: "No internet — messages will send when connected", color: .orange)
            }
            if let failureBanner = viewModel.failureBanner {
                HStack {
                    Text(failureBanner.text)
                        .font(.subheadline)
                    Spacer()
                    Button("Retry") {
                        Task {
                            await viewModel.retryFailedMessage()
                        }
                    }
                    .foregroundStyle(Color.auEmerald)
                    .accessibilityLabel("Retry sending failed message")
                }
                .padding(.horizontal, 16)
                .padding(.vertical, 12)
                .background(Color.red.opacity(0.08))
            }

            content
        }
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .principal) {
                VStack(spacing: 2) {
                    Text(viewModel.header.displayName)
                        .font(.headline)
                    Text(viewModel.header.subtitle)
                        .font(.system(size: 13))
                        .foregroundStyle(.gray)
                }
            }
        }
        .safeAreaInset(edge: .bottom) {
            composeBar
        }
        .task {
            viewModel.configure(modelContext: modelContext, webSocketManager: webSocketManager)
            await viewModel.loadIfNeeded()
        }
    }

    @ViewBuilder
    private var content: some View {
        switch viewModel.phase {
        case .idle, .loading:
            ProgressView("Loading conversation")
                .frame(maxWidth: .infinity, maxHeight: .infinity)
        case .error(let error):
            ContentUnavailableView(
                "Conversation unavailable",
                systemImage: "exclamationmark.bubble",
                description: Text(error.errorDescription ?? "Try again later.")
            )
        case .loaded:
            List {
                ForEach(groupedMessages) { section in
                    Section {
                        ForEach(section.messages) { message in
                            MessageRow(
                                message: message,
                                mode: viewModel.header.kind,
                                replyPreview: viewModel.messages.first(where: { $0.id == message.replyToID }),
                                onReply: { viewModel.setReplyContext(message) },
                                onDelete: { viewModel.deleteOwnMessage(message) }
                            )
                            .listRowInsets(EdgeInsets(top: 4, leading: 12, bottom: 4, trailing: 12))
                            .listRowSeparator(.hidden)
                        }

                        if section.id == groupedMessages.last?.id, viewModel.isShowingTypingIndicator {
                            TypingIndicatorRow(text: viewModel.typingIndicatorText ?? "Typing…")
                                .listRowSeparator(.hidden)
                        }
                    } header: {
                        Text(section.title)
                            .font(.caption.smallCaps())
                            .foregroundStyle(.secondary)
                            .frame(maxWidth: .infinity)
                    }
                }
            }
            .listStyle(.plain)
            .refreshable {
                await viewModel.refresh()
            }
        }
    }

    private var composeBar: some View {
        VStack(spacing: 8) {
            if let replyContext = viewModel.replyContext {
                HStack {
                    VStack(alignment: .leading, spacing: 2) {
                        Text("Replying to \(replyContext.authorName)")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        Text(replyContext.text)
                            .font(.subheadline)
                            .lineLimit(1)
                    }
                    Spacer()
                    Button {
                        viewModel.clearReplyContext()
                    } label: {
                        Image(systemName: "xmark.circle.fill")
                            .foregroundStyle(.secondary)
                    }
                    .accessibilityLabel("Cancel reply")
                }
                .padding(12)
                .background(Color.auSecondary)
                .clipShape(RoundedRectangle(cornerRadius: 14))
                .padding(.horizontal, 12)
            }

            HStack(alignment: .bottom, spacing: 12) {
                TextField("Message", text: Binding(
                    get: { viewModel.inputText },
                    set: { viewModel.onInputChanged($0) }
                ), axis: .vertical)
                .textFieldStyle(.roundedBorder)
                .lineLimit(1...5)
                .accessibilityLabel("Message input")

                Button {
                    Task {
                        await viewModel.sendCurrentMessage()
                    }
                } label: {
                    Image(systemName: "chevron.up.circle.fill")
                        .font(.system(size: 30))
                        .foregroundStyle(viewModel.inputText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? .gray : Color.auEmerald)
                }
                .disabled(viewModel.inputText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                .accessibilityLabel("Send message")
            }
            .padding(.horizontal, 12)
            .padding(.top, 8)
            .padding(.bottom, 12)
            .background(.regularMaterial)
        }
        .background(Color.auBackground)
    }

    private var groupedMessages: [MessageSection] {
        let calendar = Calendar.current
        let grouped = Dictionary(grouping: viewModel.messages) { message in
            calendar.startOfDay(for: message.timestamp)
        }

        return grouped.keys.sorted().map { day in
            MessageSection(
                id: day,
                title: day.sectionTitle,
                messages: grouped[day]?.sorted(by: { $0.timestamp < $1.timestamp }) ?? []
            )
        }
    }

    private func statusBanner(text: String, color: Color) -> some View {
        Text(text)
            .font(.subheadline)
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.horizontal, 16)
            .padding(.vertical, 12)
            .background(color.opacity(0.12))
    }
}

struct ConversationRow: View {
    let item: ConversationListItem

    var body: some View {
        HStack(spacing: 12) {
            if item.kind == .channel {
                ChannelAvatarView(initials: "#")
            } else {
                ZStack(alignment: .bottomTrailing) {
                    AvatarView(initials: item.initials, isAgent: item.isAgent)
                    PresenceDot(isOnline: item.isOnline)
                }
            }

            VStack(alignment: .leading, spacing: 4) {
                HStack(spacing: 8) {
                    Text(item.kind == .channel ? "#\(item.displayName)" : item.displayName)
                        .font(.headline)
                        .foregroundStyle(.primary)
                    if item.kind == .dm && item.isAgent {
                        AgentBadge()
                    }
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
                    unreadView
                }
            }
        }
        .padding(.vertical, 6)
    }

    @ViewBuilder
    private var unreadView: some View {
        if item.unreadCount > 0 {
            if item.unreadCount > 9 {
                Text("9+")
                    .font(.caption2.bold())
                    .foregroundStyle(.white)
                    .padding(.horizontal, 6)
                    .padding(.vertical, 3)
                    .background(Color.blue)
                    .clipShape(Capsule())
            } else {
                Circle()
                    .fill(Color.blue)
                    .frame(width: 12, height: 12)
            }
        }
    }
}

private struct MessageRow: View {
    let message: MessageItem
    let mode: ConversationKind
    let replyPreview: MessageItem?
    let onReply: () -> Void
    let onDelete: () -> Void

    var body: some View {
        HStack {
            if message.isOutgoing {
                Spacer(minLength: 48)
            }

            VStack(alignment: message.isOutgoing ? .trailing : .leading, spacing: 4) {
                if mode == .channel && message.isOutgoing == false {
                    Text(message.authorName)
                        .font(.caption.weight(.semibold))
                        .foregroundStyle(.secondary)
                }

                Text(message.timestamp.chatTimestamp)
                    .font(.caption2)
                    .foregroundStyle(.secondary)

                if let replyPreview {
                    VStack(alignment: .leading, spacing: 2) {
                        Text("Replying to \(replyPreview.authorName)")
                            .font(.caption2)
                            .foregroundStyle(.secondary)
                        Text(replyPreview.text)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                            .lineLimit(2)
                    }
                    .padding(8)
                    .background(Color.auSecondary)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }

                Text(message.text)
                    .padding(.horizontal, 14)
                    .padding(.vertical, 10)
                    .foregroundStyle(message.isOutgoing ? Color.white : Color.auLabel)
                    .background(
                        BubbleShape(isOutgoing: message.isOutgoing)
                            .fill(message.isOutgoing ? Color.auBubbleOutgoing : Color.auBubbleIncoming)
                    )
                    .overlay {
                        if message.isOutgoing == false {
                            BubbleShape(isOutgoing: false)
                            .stroke(Color(.systemGray5), lineWidth: 1)
                        }
                    }
                    .clipShape(BubbleShape(isOutgoing: message.isOutgoing))
                    .frame(maxWidth: UIScreen.main.bounds.width * 0.75, alignment: message.isOutgoing ? .trailing : .leading)
                    .opacity(message.isPending ? 0.72 : 1)

                if message.isFailed {
                    Text("Failed to send")
                        .font(.caption2)
                        .foregroundStyle(.red)
                }
            }
            .contextMenu {
                Button("Copy") {
                    UIPasteboard.general.string = message.text
                }
                Button("Reply") {
                    onReply()
                }
                if message.isOutgoing {
                    Button("Delete", role: .destructive) {
                        onDelete()
                    }
                }
            }
            .swipeActions(edge: .leading, allowsFullSwipe: true) {
                if message.isOutgoing == false {
                    Button {
                        onReply()
                    } label: {
                        Label("Reply", systemImage: "arrowshape.turn.up.left")
                    }
                    .tint(.auEmerald)
                }
            }

            if message.isOutgoing == false {
                Spacer(minLength: 48)
            }
        }
        .accessibilityElement(children: .combine)
        .accessibilityLabel("\(message.authorName): \(message.text)")
    }
}

private struct TypingIndicatorRow: View {
    let text: String

    var body: some View {
        HStack {
            TypingIndicatorView()
            Text(text)
                .font(.system(size: 13))
                .foregroundStyle(.secondary)
            Spacer()
        }
    }
}

private struct ChannelAvatarView: View {
    let initials: String

    var body: some View {
        RoundedRectangle(cornerRadius: 12)
            .fill(Color.auEmerald.opacity(0.14))
            .frame(width: 44, height: 44)
            .overlay {
                Text(initials)
                    .font(.headline)
                    .foregroundStyle(Color.auEmerald)
            }
    }
}

private struct TypingIndicatorView: View {
    @State private var animate = false

    var body: some View {
        HStack(spacing: 4) {
            ForEach(0..<3, id: \.self) { index in
                Circle()
                    .fill(Color(.systemGray3))
                    .frame(width: 8, height: 8)
                    .offset(y: animate ? -4 : 0)
                    .animation(
                        .easeInOut(duration: 0.4)
                            .repeatForever()
                            .delay(Double(index) * 0.15),
                        value: animate
                    )
            }
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 10)
        .background(Color.auBubbleIncoming)
        .clipShape(RoundedRectangle(cornerRadius: 18))
        .overlay {
            RoundedRectangle(cornerRadius: 18)
                .stroke(Color(.systemGray5), lineWidth: 1)
        }
        .onAppear {
            animate = true
        }
    }
}

private struct AvatarView: View {
    let initials: String
    let isAgent: Bool

    var body: some View {
        Circle()
            .fill(isAgent ? Color.auEmerald.opacity(0.18) : Color(.systemGray5))
            .frame(width: 44, height: 44)
            .overlay {
                Text(initials)
                    .font(.headline)
                    .foregroundStyle(isAgent ? Color.auEmerald : Color.auLabel)
            }
    }
}

private struct PresenceDot: View {
    let isOnline: Bool

    var body: some View {
        Circle()
            .fill(isOnline ? Color.auPresenceOnline : Color.auPresenceOffline)
            .frame(width: 10, height: 10)
            .overlay {
                Circle().stroke(Color.white, lineWidth: 2)
            }
    }
}

private struct AgentBadge: View {
    var body: some View {
        Text("agent")
            .font(.caption2.weight(.semibold))
            .foregroundStyle(Color.auEmerald)
            .padding(.horizontal, 6)
            .padding(.vertical, 2)
            .background(Color.auEmerald.opacity(0.12))
            .clipShape(Capsule())
    }
}

struct ConversationRowSkeleton: View {
    var body: some View {
        HStack(spacing: 12) {
            Circle().frame(width: 44, height: 44)
            VStack(alignment: .leading, spacing: 6) {
                RoundedRectangle(cornerRadius: 4).frame(width: 140, height: 14)
                RoundedRectangle(cornerRadius: 4).frame(width: 220, height: 12)
                RoundedRectangle(cornerRadius: 4).frame(width: 180, height: 12)
            }
        }
        .foregroundStyle(Color(.systemGray5))
        .padding(.vertical, 6)
    }
}

enum MessagesFeatureError: LocalizedError, Equatable {
    case missingSession
    case conversationNotFound
    case sendFailed(String)
    case network(String)
    case generic(String)

    var errorDescription: String? {
        switch self {
        case .missingSession:
            return "Your session is missing. Sign in again."
        case .conversationNotFound:
            return "That conversation could not be found."
        case let .sendFailed(message), let .network(message), let .generic(message):
            return message
        }
    }
}

struct ConversationListItem: Identifiable, Equatable {
    let kind: ConversationKind
    let id: String
    let displayName: String
    let isAgent: Bool
    let isOnline: Bool
    let unreadCount: Int
    let lastMessagePreview: String
    let lastMessageAt: Date?
    let initials: String
    let isMuted: Bool
    let memberCount: Int

    var relativeTimestamp: String {
        guard let lastMessageAt else {
            return ""
        }
        return lastMessageAt.listTimestamp
    }
}

enum ConversationKind: String, Equatable {
    case dm
    case channel
}

struct UserListItem: Identifiable, Equatable {
    let id: String
    let displayName: String
    let email: String
    let isAgent: Bool
    let isOnline: Bool
    let initials: String
}

struct MessageItem: Identifiable, Equatable {
    let id: String
    let text: String
    let authorID: String
    let authorName: String
    let authorType: String
    let timestamp: Date
    let replyToID: String?
    let retryCount: Int
    let isPending: Bool
    let isFailed: Bool
    let isOutgoing: Bool
}

struct ConversationHeader: Equatable {
    let kind: ConversationKind
    let displayName: String
    let subtitle: String
    let participantID: String?
    let memberCount: Int

    static let placeholder = ConversationHeader(
        kind: .dm,
        displayName: "Conversation",
        subtitle: "agent · offline",
        participantID: nil,
        memberCount: 0
    )

    static let channelPlaceholder = ConversationHeader(
        kind: .channel,
        displayName: "#channel",
        subtitle: "0 members",
        participantID: nil,
        memberCount: 0
    )
}

struct ConversationScreenData: Equatable {
    let header: ConversationHeader
    let messages: [MessageItem]
}

struct SendFailureBanner: Equatable {
    let messageID: String
    let text: String
}

struct MessageSection: Identifiable {
    let id: Date
    let title: String
    let messages: [MessageItem]
}

@MainActor
protocol ConversationStoreProtocol: AnyObject {
    func loadConversationScreen(conversationID: String) async throws -> ConversationScreenData
    func insertOptimisticMessage(conversationID: String, text: String, replyToID: String?) throws -> MessageItem
    func commitMessage(conversationID: String, text: String, replyToID: String?, optimisticID: String) async throws -> MessageItem
    func markMessageFailed(id: String) throws -> MessageItem
    func retryFailedMessage(messageID: String, conversationID: String) async throws -> MessageItem
    func deleteMessage(id: String) throws
    func typingDisplayNames(conversationID: String, userIDs: [String]) throws -> [String]
}

@MainActor
protocol NotificationPermissionRequesting {
    func requestAfterFirstMessageSend() async
}

struct PushPermissionRequester: NotificationPermissionRequesting {
    func requestAfterFirstMessageSend() async {
        _ = try? await UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .sound, .badge])
    }
}

@MainActor
protocol NetworkStatusProviding: AnyObject {
    var statusPublisher: AnyPublisher<Bool, Never> { get }
    func start()
}

@MainActor
final class NetworkStatusMonitor: NetworkStatusProviding {
    private let subject = CurrentValueSubject<Bool, Never>(true)
    private let monitor = NWPathMonitor()
    private let queue = DispatchQueue(label: "ai.agentunited.network.monitor")

    var statusPublisher: AnyPublisher<Bool, Never> {
        subject.eraseToAnyPublisher()
    }

    func start() {
        monitor.pathUpdateHandler = { [weak self] path in
            Task { @MainActor [weak self] in
                self?.subject.send(path.status == .satisfied)
            }
        }
        monitor.start(queue: queue)
    }
}

@MainActor
final class LiveMessagesStore: ConversationStoreProtocol {
    private let modelContext: ModelContext
    private let keychainHelper: KeychainHelper

    init(modelContext: ModelContext, keychainHelper: KeychainHelper = KeychainHelper()) {
        self.modelContext = modelContext
        self.keychainHelper = keychainHelper
    }

    func syncConversationList() async throws -> [ConversationListItem] {
        let client = try makeClient()
        let conversations = try await client.listDMs()
        try syncConversations(conversations)
        return try localConversationList()
    }

    func localConversationList() throws -> [ConversationListItem] {
        try localConversations(type: "dm")
    }

    func syncChannelList() async throws -> [ConversationListItem] {
        let client = try makeClient()
        let channels = try await client.listChannels()
        try syncConversations(channels)
        return try localChannelList()
    }

    func localChannelList() throws -> [ConversationListItem] {
        try localConversations(type: "channel")
    }

    func syncUsers() async throws -> [UserListItem] {
        let client = try makeClient()
        let users = try await client.listUsers()
        try upsertUsers(users)
        return try localUsers()
    }

    func localUsers() throws -> [UserListItem] {
        let currentUserID = try currentCredential().userID
        let users = try modelContext.fetch(FetchDescriptor<User>(sortBy: [SortDescriptor(\User.displayName)]))
        return users
            .filter { $0.id != currentUserID }
            .map(mapUser)
            .sorted {
                ($0.isAgent ? 0 : 1, $0.displayName.lowercased()) < ($1.isAgent ? 0 : 1, $1.displayName.lowercased())
            }
    }

    func openOrCreateConversation(targetUserID: String) async throws -> ConversationListItem {
        let currentUserID = try currentCredential().userID
        if let existing = try conversationForParticipants(currentUserID: currentUserID, otherUserID: targetUserID) {
            return try mapConversation(existing, currentUserID: currentUserID)
        }

        let client = try makeClient()
        let created = try await client.createDM(targetUserID: targetUserID)
        try syncConversations([created])
        guard let conversation = try fetchConversation(id: created.id) else {
            throw MessagesFeatureError.conversationNotFound
        }
        return try mapConversation(conversation, currentUserID: currentUserID)
    }

    func loadConversationScreen(conversationID: String) async throws -> ConversationScreenData {
        let client = try makeClient()
        let currentUserID = try currentCredential().userID
        let messages = try await client.listMessages(channelId: conversationID, before: nil)
        try syncMessages(messages, conversationID: conversationID)

        if try fetchConversation(id: conversationID) == nil {
            try syncConversations(try await client.listDMs())
            try syncConversations(try await client.listChannels())
            try syncMessages(messages, conversationID: conversationID)
        }

        guard let conversation = try fetchConversation(id: conversationID) else {
            throw MessagesFeatureError.conversationNotFound
        }

        return try ConversationScreenData(
            header: mapHeader(conversation, currentUserID: currentUserID),
            messages: localMessages(conversationID: conversationID, currentUserID: currentUserID)
        )
    }

    func insertOptimisticMessage(conversationID: String, text: String, replyToID: String?) throws -> MessageItem {
        let credential = try currentCredential()
        guard let conversation = try fetchConversation(id: conversationID) else {
            throw MessagesFeatureError.conversationNotFound
        }

        let optimisticID = "local-\(UUID().uuidString)"
        let message = Message(
            id: optimisticID,
            conversationID: conversationID,
            text: text,
            authorID: credential.userID,
            authorName: credential.displayName,
            authorType: credential.userType,
            timestamp: .now,
            replyToID: replyToID,
            retryCount: 0,
            isPending: true,
            isFailed: false,
            conversation: conversation
        )
        modelContext.insert(message)
        conversation.messages.append(message)
        conversation.lastMessageText = text
        conversation.lastMessageAt = message.timestamp
        try modelContext.save()
        return mapMessage(message, currentUserID: credential.userID)
    }

    func commitMessage(conversationID: String, text: String, replyToID: String?, optimisticID: String) async throws -> MessageItem {
        let credential = try currentCredential()
        let client = try makeClient()

        do {
            let response = try await client.sendMessage(channelId: conversationID, text: text, replyToID: replyToID)
            try replaceOptimisticMessage(optimisticID: optimisticID, with: response)
            guard let message = try fetchMessage(id: response.id) else {
                throw MessagesFeatureError.sendFailed("Message sent but could not be stored locally.")
            }
            return mapMessage(message, currentUserID: credential.userID)
        } catch let error as AUAPIError {
            throw MessagesFeatureError.sendFailed(error.errorDescription ?? "Message failed to send.")
        } catch {
            throw MessagesFeatureError.sendFailed(error.localizedDescription)
        }
    }

    func markMessageFailed(id: String) throws -> MessageItem {
        let credential = try currentCredential()
        guard let message = try fetchMessage(id: id) else {
            throw MessagesFeatureError.sendFailed("No pending message to retry.")
        }
        message.isPending = false
        message.isFailed = true
        message.retryCount += 1
        try modelContext.save()
        return mapMessage(message, currentUserID: credential.userID)
    }

    func retryFailedMessage(messageID: String, conversationID: String) async throws -> MessageItem {
        let credential = try currentCredential()
        guard let failedMessage = try fetchMessage(id: messageID) else {
            throw MessagesFeatureError.sendFailed("The failed message no longer exists.")
        }

        failedMessage.isPending = true
        failedMessage.isFailed = false
        try modelContext.save()

        let client = try makeClient()
        do {
            let response = try await client.sendMessage(channelId: conversationID, text: failedMessage.text, replyToID: failedMessage.replyToID)
            try replaceOptimisticMessage(optimisticID: messageID, with: response)
            guard let message = try fetchMessage(id: response.id) else {
                throw MessagesFeatureError.sendFailed("Retry succeeded but local state did not update.")
            }
            return mapMessage(message, currentUserID: credential.userID)
        } catch let error as AUAPIError {
            failedMessage.isPending = false
            failedMessage.isFailed = true
            failedMessage.retryCount += 1
            try modelContext.save()
            throw MessagesFeatureError.sendFailed(error.errorDescription ?? "Retry failed.")
        } catch {
            failedMessage.isPending = false
            failedMessage.isFailed = true
            failedMessage.retryCount += 1
            try modelContext.save()
            throw MessagesFeatureError.sendFailed(error.localizedDescription)
        }
    }

    func deleteConversation(id: String) throws {
        guard let conversation = try fetchConversation(id: id) else {
            return
        }
        modelContext.delete(conversation)
        try modelContext.save()
    }

    func toggleMuteConversation(id: String) throws -> [ConversationListItem] {
        guard let conversation = try fetchConversation(id: id) else {
            throw MessagesFeatureError.conversationNotFound
        }
        conversation.isMuted.toggle()
        try modelContext.save()
        return try conversation.type == "channel" ? localChannelList() : localConversationList()
    }

    func deleteMessage(id: String) throws {
        if let message = try fetchMessage(id: id) {
            modelContext.delete(message)
            try modelContext.save()
        }
    }

    func typingDisplayNames(conversationID: String, userIDs: [String]) throws -> [String] {
        let currentUserID = try currentCredential().userID
        return try userIDs
            .filter { $0 != currentUserID }
            .compactMap { try fetchUser(id: $0)?.displayName }
    }

    private func syncConversations(_ responses: [ConversationResponse]) throws {
        for response in responses {
            let conversation = try fetchConversation(id: response.id)
                ?? Conversation(id: response.id, name: response.name, type: response.type)

            if conversation.modelContext == nil {
                modelContext.insert(conversation)
            }

            conversation.name = response.name
            conversation.type = response.type
            conversation.lastMessageText = response.lastMessageText
            conversation.lastMessageAt = response.lastMessageAt
            conversation.unreadCount = response.unreadCount
            conversation.participantIDs = response.participantIDs
        }

        try modelContext.save()
    }

    private func syncMessages(_ responses: [MessageResponse], conversationID: String) throws {
        guard let conversation = try fetchConversation(id: conversationID) else {
            return
        }

        for response in responses {
            if let existing = try fetchMessage(id: response.id) {
                existing.text = response.text
                existing.timestamp = response.timestamp
                existing.replyToID = response.replyToID
                existing.isPending = false
                existing.isFailed = false
                continue
            }

            let message = Message(
                id: response.id,
                conversationID: response.conversationID,
                text: response.text,
                authorID: response.authorID,
                authorName: response.authorName,
                authorType: response.authorType,
                timestamp: response.timestamp,
                replyToID: response.replyToID,
                retryCount: 0,
                isPending: false,
                isFailed: false,
                conversation: conversation
            )
            modelContext.insert(message)
            conversation.messages.append(message)
        }

        try upsertUsers(
            responses.map {
                UserResponse(
                    id: $0.authorID,
                    displayName: $0.authorName,
                    email: "",
                    userType: $0.authorType,
                    avatarURL: nil,
                    isOnline: false
                )
            }
        )
        try modelContext.save()
    }

    private func upsertUsers(_ users: [UserResponse]) throws {
        for response in users {
            let user = try fetchUser(id: response.id)
                ?? User(
                    id: response.id,
                    displayName: response.displayName,
                    email: response.email,
                    userType: response.userType,
                    avatarURL: response.avatarURL,
                    isOnline: response.isOnline
                )

            if user.modelContext == nil {
                modelContext.insert(user)
            }

            user.displayName = response.displayName
            user.email = response.email
            user.userType = response.userType
            user.avatarURL = response.avatarURL
            user.isOnline = response.isOnline
        }

        try modelContext.save()
    }

    private func replaceOptimisticMessage(optimisticID: String, with response: MessageResponse) throws {
        if let optimistic = try fetchMessage(id: optimisticID) {
            optimistic.id = response.id
            optimistic.text = response.text
            optimistic.timestamp = response.timestamp
            optimistic.replyToID = response.replyToID
            optimistic.isPending = false
            optimistic.isFailed = false
            optimistic.authorID = response.authorID
            optimistic.authorName = response.authorName
            optimistic.authorType = response.authorType
            optimistic.conversationID = response.conversationID
        } else {
            guard let conversation = try fetchConversation(id: response.conversationID) else {
                throw MessagesFeatureError.conversationNotFound
            }
            let message = Message(
                id: response.id,
                conversationID: response.conversationID,
                text: response.text,
                authorID: response.authorID,
                authorName: response.authorName,
                authorType: response.authorType,
                timestamp: response.timestamp,
                replyToID: response.replyToID,
                retryCount: 0,
                isPending: false,
                isFailed: false,
                conversation: conversation
            )
            modelContext.insert(message)
            conversation.messages.append(message)
        }

        if let conversation = try fetchConversation(id: response.conversationID) {
            conversation.lastMessageText = response.text
            conversation.lastMessageAt = response.timestamp
        }

        try modelContext.save()
    }

    private func localMessages(conversationID: String, currentUserID: String) throws -> [MessageItem] {
        let descriptor = FetchDescriptor<Message>(
            predicate: #Predicate { $0.conversationID == conversationID },
            sortBy: [SortDescriptor(\Message.timestamp)]
        )
        return try modelContext.fetch(descriptor).map { mapMessage($0, currentUserID: currentUserID) }
    }

    private func localConversations(type: String) throws -> [ConversationListItem] {
        let currentUserID = try currentCredential().userID
        let conversations = try modelContext.fetch(
            FetchDescriptor<Conversation>(
                predicate: #Predicate { $0.type == type },
                sortBy: [SortDescriptor(\Conversation.lastMessageAt, order: .reverse)]
            )
        )
        return try conversations.map { try mapConversation($0, currentUserID: currentUserID) }
    }

    private func mapConversation(_ conversation: Conversation, currentUserID: String) throws -> ConversationListItem {
        let isChannel = conversation.type == "channel"
        let participant = isChannel ? nil : try otherParticipant(conversation: conversation, currentUserID: currentUserID)
        let name = isChannel ? conversation.name : (participant?.displayName ?? conversation.name)
        return ConversationListItem(
            kind: isChannel ? .channel : .dm,
            id: conversation.id,
            displayName: name,
            isAgent: (participant?.userType ?? "agent") == "agent",
            isOnline: participant?.isOnline ?? false,
            unreadCount: conversation.unreadCount,
            lastMessagePreview: conversation.lastMessageText ?? "No messages yet",
            lastMessageAt: conversation.lastMessageAt,
            initials: name.initials,
            isMuted: conversation.isMuted,
            memberCount: conversation.participantIDs.count
        )
    }

    private func mapHeader(_ conversation: Conversation, currentUserID: String) throws -> ConversationHeader {
        if conversation.type == "channel" {
            let memberCount = max(conversation.participantIDs.count, 1)
            return ConversationHeader(
                kind: .channel,
                displayName: "#\(conversation.name)",
                subtitle: "\(memberCount) \(memberCount == 1 ? "member" : "members")",
                participantID: nil,
                memberCount: memberCount
            )
        }

        let participant = try otherParticipant(conversation: conversation, currentUserID: currentUserID)
        let name = participant?.displayName ?? conversation.name
        let type = participant?.userType ?? "agent"
        let status = participant?.isOnline == true ? "online" : "offline"
        return ConversationHeader(
            kind: .dm,
            displayName: name,
            subtitle: "\(type) · \(status)",
            participantID: participant?.id,
            memberCount: 0
        )
    }

    private func mapUser(_ user: User) -> UserListItem {
        UserListItem(
            id: user.id,
            displayName: user.displayName,
            email: user.email,
            isAgent: user.userType == "agent",
            isOnline: user.isOnline,
            initials: user.displayName.initials
        )
    }

    private func mapMessage(_ message: Message, currentUserID: String) -> MessageItem {
        MessageItem(
            id: message.id,
            text: message.text,
            authorID: message.authorID,
            authorName: message.authorName,
            authorType: message.authorType,
            timestamp: message.timestamp,
            replyToID: message.replyToID,
            retryCount: message.retryCount,
            isPending: message.isPending,
            isFailed: message.isFailed,
            isOutgoing: message.authorID == currentUserID
        )
    }

    private func otherParticipant(conversation: Conversation, currentUserID: String) throws -> User? {
        let otherID = conversation.participantIDs.first(where: { $0 != currentUserID })
        guard let otherID else {
            return nil
        }
        return try fetchUser(id: otherID)
    }

    private func conversationForParticipants(currentUserID: String, otherUserID: String) throws -> Conversation? {
        let conversations = try modelContext.fetch(FetchDescriptor<Conversation>(predicate: #Predicate { $0.type == "dm" }))
        return conversations.first {
            $0.participantIDs.contains(currentUserID) && $0.participantIDs.contains(otherUserID)
        }
    }

    private func currentCredential() throws -> WorkspaceCredential {
        let descriptor = FetchDescriptor<WorkspaceCredential>(sortBy: [SortDescriptor(\WorkspaceCredential.createdAt, order: .reverse)])
        guard let credential = try modelContext.fetch(descriptor).first else {
            throw MessagesFeatureError.missingSession
        }
        return credential
    }

    private func makeClient() throws -> LiveAUAPIClient {
        let credential = try currentCredential()
        guard let token = try keychainHelper.readJWT(for: credential.userID), token.isEmpty == false else {
            throw MessagesFeatureError.missingSession
        }
        return LiveAUAPIClient(instanceURL: credential.instanceURL, authToken: token)
    }

    private func fetchConversation(id: String) throws -> Conversation? {
        try modelContext.fetch(FetchDescriptor<Conversation>(predicate: #Predicate { $0.id == id })).first
    }

    private func fetchMessage(id: String) throws -> Message? {
        try modelContext.fetch(FetchDescriptor<Message>(predicate: #Predicate { $0.id == id })).first
    }

    private func fetchUser(id: String) throws -> User? {
        try modelContext.fetch(FetchDescriptor<User>(predicate: #Predicate { $0.id == id })).first
    }
}

private extension Date {
    var listTimestamp: String {
        let calendar = Calendar.current
        if calendar.isDateInToday(self) {
            return formatted(date: .omitted, time: .shortened)
        }
        if calendar.isDateInYesterday(self) {
            return "Yesterday"
        }
        if calendar.isDate(self, equalTo: .now, toGranularity: .weekOfYear) {
            return formatted(.dateTime.weekday(.abbreviated))
        }
        return formatted(.dateTime.month(.abbreviated).day())
    }

    var sectionTitle: String {
        let calendar = Calendar.current
        if calendar.isDateInToday(self) {
            return "Today"
        }
        if calendar.isDateInYesterday(self) {
            return "Yesterday"
        }
        return formatted(.dateTime.weekday(.abbreviated).month(.abbreviated).day())
    }

    var chatTimestamp: String {
        formatted(date: .omitted, time: .shortened)
    }
}

private extension String {
    var initials: String {
        let parts = split(separator: " ").prefix(2)
        let letters = parts.compactMap { $0.first }
        if letters.isEmpty, let first = first {
            return String(first).uppercased()
        }
        return String(letters).uppercased()
    }
}

private struct BubbleShape: Shape {
    let isOutgoing: Bool

    func path(in rect: CGRect) -> Path {
        let shape = isOutgoing
            ? UnevenRoundedRectangle(
                topLeadingRadius: 18,
                bottomLeadingRadius: 18,
                bottomTrailingRadius: 4,
                topTrailingRadius: 18
            )
            : UnevenRoundedRectangle(
                topLeadingRadius: 18,
                bottomLeadingRadius: 4,
                bottomTrailingRadius: 18,
                topTrailingRadius: 18
            )
        return shape.path(in: rect)
    }
}
