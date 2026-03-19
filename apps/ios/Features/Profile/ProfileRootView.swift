import SwiftUI

struct ProfileRootView: View {
    @EnvironmentObject private var coordinator: AppCoordinator
    @EnvironmentObject private var sessionStore: AppSessionStore

    var body: some View {
        NavigationStack {
            List {
                Section("Account") {
                    LabeledContent("Display name", value: "Local User")
                    LabeledContent("Email", value: "user@example.com")
                }

                Section("Plan") {
                    LabeledContent("Plan", value: "Free")
                    LabeledContent("Entities", value: "0 of 3")
                }

                Section("Notifications") {
                    Toggle("Push notifications", isOn: .constant(false))
                    Toggle("Message preview", isOn: .constant(true))
                }

                Section {
                    Button("Sign out", role: .destructive) {
                        try? sessionStore.signOut()
                        coordinator.setAuthenticated(false)
                    }
                }
            }
            .navigationTitle("Profile")
        }
    }
}
