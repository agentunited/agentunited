import SafariServices
import SwiftData
import SwiftUI

struct ProfileRootView: View {
    @Environment(\.modelContext) private var modelContext
    @EnvironmentObject private var coordinator: AppCoordinator
    @EnvironmentObject private var sessionStore: AppSessionStore
    @StateObject private var viewModel = ProfileViewModel()

    var body: some View {
        NavigationStack {
            List {
                Section {
                    HStack(spacing: 12) {
                        AvatarInitialsView(initials: viewModel.initials)
                        VStack(alignment: .leading, spacing: 4) {
                            Text(viewModel.displayName)
                                .font(.headline)
                            Text(viewModel.email)
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }
                    }
                    .padding(.vertical, 4)
                    .accessibilityElement(children: .combine)
                    .accessibilityLabel("Profile for \(viewModel.displayName), \(viewModel.email)")
                }

                Section("Account") {
                    LabeledContent("Display name", value: viewModel.displayName)
                    Button("Edit display name") {
                        viewModel.startEditingDisplayName()
                    }
                    .accessibilityLabel("Edit display name")
                }

                Section("Billing") {
                    NavigationLink {
                        BillingView(
                            planName: viewModel.planName,
                            billingPortalURL: viewModel.billingPortalURL
                        )
                    } label: {
                        LabeledContent("Current plan", value: viewModel.planName)
                    }
                    .accessibilityLabel("Open billing")
                }

                Section {
                    Button("Sign out", role: .destructive) {
                        try? sessionStore.signOut()
                        coordinator.setAuthenticated(false)
                    }
                    .accessibilityLabel("Sign out")
                }
            }
            .navigationTitle("Profile")
            .task {
                await viewModel.load(modelContext: modelContext)
            }
            .sheet(isPresented: $viewModel.isEditingDisplayName) {
                EditDisplayNameSheet(viewModel: viewModel)
            }
            .alert("Could not update profile", isPresented: $viewModel.showErrorAlert) {
                Button("OK", role: .cancel) {}
            } message: {
                Text(viewModel.errorMessage)
            }
        }
    }
}

private struct BillingView: View {
    let planName: String
    let billingPortalURL: URL?
    @State private var showSafari = false

    var body: some View {
        List {
            Section("Subscription") {
                LabeledContent("Plan", value: planName)
            }

            Section {
                Button("Manage Billing") {
                    showSafari = true
                }
                .foregroundStyle(Color.auEmerald)
                .disabled(billingPortalURL == nil)
                .accessibilityLabel("Manage billing")
            }
        }
        .navigationTitle("Billing")
        .sheet(isPresented: $showSafari) {
            if let billingPortalURL {
                SafariView(url: billingPortalURL)
            }
        }
    }
}

private struct EditDisplayNameSheet: View {
    @ObservedObject var viewModel: ProfileViewModel
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            Form {
                TextField("Display name", text: $viewModel.editingDisplayName)
                    .textInputAutocapitalization(.words)
                    .autocorrectionDisabled(true)
                    .accessibilityLabel("Display name")
            }
            .navigationTitle("Edit Name")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        Task {
                            let saved = await viewModel.saveDisplayName()
                            if saved {
                                dismiss()
                            }
                        }
                    }
                    .disabled(viewModel.canSaveDisplayName == false)
                }
            }
        }
        .presentationDetents([.medium])
    }
}

@MainActor
final class ProfileViewModel: ObservableObject {
    @Published var displayName = ""
    @Published var email = ""
    @Published var planName = "Free"
    @Published var billingPortalURL: URL?
    @Published var isEditingDisplayName = false
    @Published var editingDisplayName = ""
    @Published var showErrorAlert = false
    @Published var errorMessage = ""

    private var credential: WorkspaceCredential?
    private let keychainHelper = KeychainHelper()
    private var modelContext: ModelContext?

    var initials: String {
        displayName
            .split(separator: " ")
            .prefix(2)
            .compactMap { $0.first }
            .map(String.init)
            .joined()
            .uppercased()
            .isEmpty ? "AU" : displayName
                .split(separator: " ")
                .prefix(2)
                .compactMap { $0.first }
                .map(String.init)
                .joined()
                .uppercased()
    }

    var canSaveDisplayName: Bool {
        let trimmed = editingDisplayName.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty == false && trimmed != displayName
    }

    func load(modelContext: ModelContext) async {
        self.modelContext = modelContext
        do {
            let descriptor = FetchDescriptor<WorkspaceCredential>(
                sortBy: [SortDescriptor(\.createdAt, order: .reverse)]
            )
            if let credential = try modelContext.fetch(descriptor).first {
                self.credential = credential
                displayName = credential.displayName
                email = credential.email
                planName = "Free"
                billingPortalURL = credential.instanceURL.appending(path: "billing")
            }
        } catch {
            showError("Unable to load profile.")
        }
    }

    func startEditingDisplayName() {
        editingDisplayName = displayName
        isEditingDisplayName = true
    }

    func saveDisplayName() async -> Bool {
        guard canSaveDisplayName,
              let credential,
              let modelContext,
              let token = try? keychainHelper.readJWT(for: credential.userID),
              token.isEmpty == false
        else {
            showError("Missing active session.")
            return false
        }

        do {
            let client = LiveAUAPIClient(instanceURL: credential.instanceURL, authToken: token)
            let trimmed = editingDisplayName.trimmingCharacters(in: .whitespacesAndNewlines)
            _ = try await client.updateProfile(displayName: trimmed)
            credential.displayName = trimmed
            displayName = trimmed
            try modelContext.save()
            return true
        } catch {
            showError("Failed to update display name.")
            return false
        }
    }

    private func showError(_ message: String) {
        errorMessage = message
        showErrorAlert = true
    }
}

private struct AvatarInitialsView: View {
    let initials: String

    var body: some View {
        Circle()
            .fill(Color.auEmerald.opacity(0.18))
            .frame(width: 56, height: 56)
            .overlay {
                Text(initials)
                    .font(.headline)
                    .foregroundStyle(Color.auEmerald)
            }
    }
}

private struct SafariView: UIViewControllerRepresentable {
    let url: URL

    func makeUIViewController(context: Context) -> SFSafariViewController {
        SFSafariViewController(url: url)
    }

    func updateUIViewController(_ uiViewController: SFSafariViewController, context: Context) {}
}
