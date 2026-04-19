class J2z < Formula
  desc "Convert Jekyll markdown posts to Zola markdown posts"
  homepage "https://github.com/en9inerd/j2z"
  license "MIT"
  version "VERSION_PLACEHOLDER"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/en9inerd/j2z/releases/download/vVERSION_PLACEHOLDER/j2z-darwin-arm64"
      sha256 "SHA256_MACOS_ARM64"
    else
      url "https://github.com/en9inerd/j2z/releases/download/vVERSION_PLACEHOLDER/j2z-darwin-amd64"
      sha256 "SHA256_MACOS_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/en9inerd/j2z/releases/download/vVERSION_PLACEHOLDER/j2z-linux-arm64"
      sha256 "SHA256_LINUX_ARM64"
    else
      url "https://github.com/en9inerd/j2z/releases/download/vVERSION_PLACEHOLDER/j2z-linux-amd64"
      sha256 "SHA256_LINUX_AMD64"
    end
  end

  def install
    binary = Dir["j2z*"].first
    bin.install binary => "j2z"
  end

  test do
    assert_match "j2z version", shell_output("#{bin}/j2z --version")
  end
end
