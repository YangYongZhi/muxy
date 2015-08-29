require 'formula'

class PackerWindowsPlugins < Formula
  homepage "https://github.com/packer-community/packer-windows-plugins"
  version "1.0.0"

  if Hardware.is_64_bit?
    url "https://github.com/mefellows/muxy/releases/download/v0.0.1/darwin_amd64.zip"
    sha1 "e09e2ab6f7fc39237b5b41f53aa5b2a815428cfc"
  else
    url "https://github.com/packer-community/packer-windows-plugins/releases/download/v0.0.1/darwin_386.zip"
    sha1 '64535683e7f261091629c5a96236263dc0856c63'
  end

  depends_on :arch => :intel

  def install
    bin.install Dir['*']
  end

  test do
    system "muxy" "--version"
  end
end
