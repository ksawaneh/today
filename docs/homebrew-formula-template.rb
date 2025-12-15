# This is a reference template for the Homebrew formula
# When using GoReleaser with the homebrew-tap integration,
# this file will be auto-generated in your tap repository.
#
# To set up automated Homebrew releases:
# 1. Create a GitHub repo: homebrew-tap
# 2. Add HOMEBREW_TAP_GITHUB_TOKEN secret to this repo's GitHub Actions
# 3. Uncomment the 'brews' section in .goreleaser.yaml
# 4. GoReleaser will automatically update the formula on each release

class Today < Formula
  desc "A unified productivity dashboard for your terminal"
  homepage "https://github.com/yourusername/today"
  version "1.0.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yourusername/today/releases/download/v1.0.0/today_darwin_arm64.tar.gz"
      sha256 "PUT_SHA256_HERE"
    else
      url "https://github.com/yourusername/today/releases/download/v1.0.0/today_darwin_amd64.tar.gz"
      sha256 "PUT_SHA256_HERE"
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/yourusername/today/releases/download/v1.0.0/today_linux_arm64.tar.gz"
      sha256 "PUT_SHA256_HERE"
    else
      url "https://github.com/yourusername/today/releases/download/v1.0.0/today_linux_amd64.tar.gz"
      sha256 "PUT_SHA256_HERE"
    end
  end

  license "MIT"

  def install
    bin.install "today"
    man1.install "docs/today.1"
  end

  def caveats
    <<~EOS
      Run 'today' to start the app.
      Data is stored in ~/.today/
      Config (optional): ~/.config/today/config.yaml

      Press '?' in the app for help.
    EOS
  end

  test do
    assert_match "today version", shell_output("#{bin}/today --version")
  end
end
