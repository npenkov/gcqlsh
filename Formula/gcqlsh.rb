class Gcqlsh < Formula
  desc "Cassandra command line shell written in Go"
  homepage "https://github.com/npenkov/gcqlsh"
  version "0.0.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/npenkov/gcqlsh/releases/download/v#{version}/gcqlsh_darwin_arm64.tar.gz"
      sha256 "" # Will be updated automatically by goreleaser
    else
      url "https://github.com/npenkov/gcqlsh/releases/download/v#{version}/gcqlsh_darwin_x86_64.tar.gz"
      sha256 "" # Will be updated automatically by goreleaser
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/npenkov/gcqlsh/releases/download/v#{version}/gcqlsh_linux_arm64.tar.gz"
        sha256 "" # Will be updated automatically by goreleaser
      else
        url "https://github.com/npenkov/gcqlsh/releases/download/v#{version}/gcqlsh_linux_armv7.tar.gz"
        sha256 "" # Will be updated automatically by goreleaser
      end
    else
      url "https://github.com/npenkov/gcqlsh/releases/download/v#{version}/gcqlsh_linux_x86_64.tar.gz"
      sha256 "" # Will be updated automatically by goreleaser
    end
  end

  def install
    bin.install "gcqlsh"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/gcqlsh -v")
  end
end
