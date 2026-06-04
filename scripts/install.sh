#!/usr/bin/env bash
set -euo pipefail

repo="SamyRai/juleson"
version="latest"
install_dir="${INSTALL_DIR:-/usr/local/bin}"

usage() {
	cat <<'EOF'
Install the latest Juleson release binaries.

Usage:
  install.sh [--version <tag|latest>] [--install-dir <path>] [--repo <owner/repo>]

Environment:
  INSTALL_DIR  Destination directory. Defaults to /usr/local/bin.

Examples:
  curl -L https://github.com/SamyRai/juleson/releases/latest/download/install.sh | bash
  INSTALL_DIR="$HOME/.local/bin" bash install.sh
  bash install.sh --version v1.0.0 --install-dir "$HOME/bin"
EOF
}

while [ "$#" -gt 0 ]; do
	case "$1" in
		--version)
			version="${2:-}"
			shift 2
			;;
		--install-dir)
			install_dir="${2:-}"
			shift 2
			;;
		--repo)
			repo="${2:-}"
			shift 2
			;;
		-h|--help)
			usage
			exit 0
			;;
		*)
			echo "Unknown argument: $1" >&2
			usage >&2
			exit 2
			;;
	esac
done

if [ -z "$version" ] || [ -z "$install_dir" ] || [ -z "$repo" ]; then
	echo "version, install directory, and repository must be non-empty" >&2
	exit 2
fi
if [ "$install_dir" != "/" ]; then
	install_dir="${install_dir%/}"
fi

require_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "Required command not found: $1" >&2
		exit 1
	fi
}

require_cmd curl
require_cmd tar
require_cmd install
require_cmd mktemp

case "$(uname -s)" in
	Linux)
		os="linux"
		;;
	Darwin)
		os="darwin"
		;;
	*)
		echo "Unsupported operating system: $(uname -s)" >&2
		exit 1
		;;
esac

case "$(uname -m)" in
	x86_64|amd64)
		arch="amd64"
		;;
	arm64|aarch64)
		arch="arm64"
		;;
	*)
		echo "Unsupported architecture: $(uname -m)" >&2
		exit 1
		;;
esac

if [ -n "${JULESON_INSTALL_BASE_URL:-}" ]; then
	base_url="${JULESON_INSTALL_BASE_URL%/}"
elif [ "$version" = "latest" ]; then
	base_url="https://github.com/${repo}/releases/latest/download"
else
	base_url="https://github.com/${repo}/releases/download/${version}"
fi

tmp_dir="$(mktemp -d)"
cleanup() {
	rm -rf "$tmp_dir"
}
trap cleanup EXIT

download_and_extract() {
	binary="$1"
	asset="${binary}-${os}-${arch}.tar.gz"
	archive="${tmp_dir}/${asset}"

	echo "Downloading ${asset}..."
	curl -fsSL "${base_url}/${asset}" -o "$archive"

	mkdir -p "${tmp_dir}/${binary}"
	tar -xzf "$archive" -C "${tmp_dir}/${binary}"
	if [ ! -f "${tmp_dir}/${binary}/${binary}" ]; then
		echo "Release asset ${asset} did not contain ${binary}" >&2
		exit 1
	fi
}

download_and_extract "juleson"
download_and_extract "jsn"

mkdir -p "$install_dir" 2>/dev/null || true

install_binary() {
	binary="$1"
	source="${tmp_dir}/${binary}/${binary}"
	target="${install_dir}/${binary}"

	if [ -w "$install_dir" ]; then
		install -m 0755 "$source" "$target"
	elif command -v sudo >/dev/null 2>&1; then
		sudo mkdir -p "$install_dir"
		sudo install -m 0755 "$source" "$target"
	else
		echo "Cannot write to ${install_dir}; rerun with a writable INSTALL_DIR or install sudo" >&2
		exit 1
	fi
}

install_binary "juleson"
install_binary "jsn"

echo "Installed juleson and jsn to ${install_dir}"
case ":$PATH:" in
	*":${install_dir}:"*) ;;
	*) echo "Add ${install_dir} to PATH before running juleson." ;;
esac
