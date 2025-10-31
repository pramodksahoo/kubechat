# typed: false
# frozen_string_literal: true

cask "kubechat" do
  version "<VERSION>"
  sha256 "<SHA256>"

  url "https://github.com/pramodksahoo/kubechat/releases/download/v#{version}/kubechat_Darwin_all.tar.gz"
  name "Kubechat"
  desc "Natural language Kubernetes management console"
  homepage "https://github.com/pramodksahoo/kubechat"

  depends_on macos: ">= :monterey"

  binary "kubechat", target: "kubechat"

  postflight do
    system_command "/usr/bin/xattr",
                   args: ["-dr", "com.apple.quarantine", "#{staged_path}/kubechat"],
                   must_succeed: false
  end

  caveats <<~EOS
    To run kubechat CLI simply run:
      kubechat

    To run kubechat on a specific IP and port:
      kubechat --listen :7080
  EOS
end
