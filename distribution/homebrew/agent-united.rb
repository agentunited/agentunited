cask "agent-united" do
  version "0.1.0"
  sha256 :no_check  # TODO: Add actual SHA256 after first release

  url "https://agentunited.ai/download/macos/Agent-United-#{version}.dmg"
  name "Agent United"
  desc "Communication infrastructure for autonomous AI agents"
  homepage "https://agentunited.ai"

  livecheck do
    url "https://agentunited.ai/api/latest-version"
    strategy :json do |json|
      json["version"]
    end
  end

  app "Agent United.app"

  postflight do
    # Register deep linking protocol
    system_command "/usr/bin/defaults",
                   args: ["write", "com.apple.LaunchServices/com.apple.launchservices.secure",
                          "LSHandlers", "-array-add",
                          "{LSHandlerContentType = 'x-web-navigation'; LSHandlerRoleAll = 'ai.agentunited.desktop';}"]
  end

  zap trash: [
    "~/Library/Application Support/Agent United",
    "~/Library/Caches/ai.agentunited.desktop",
    "~/Library/Preferences/ai.agentunited.desktop.plist",
    "~/Library/Saved Application State/ai.agentunited.desktop.savedState",
  ]
end
