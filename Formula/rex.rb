class Rex < Formula
  desc "Zero-config universal project runner. Run anything, know nothing."
  homepage "https://rexrun.dev"
  url "https://github.com/rexrun-dev/rex/archive/refs/tags/v0.1.0.tar.gz"
  license "Apache-2.0"
  head "https://github.com/rexrun-dev/rex.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/rex"
  end

  test do
    assert_match "rex", shell_output("#{bin}/rex --help 2>&1", 0)
  end
end
