#!/bin/sh
set -e

GH_REPO="loops-so/cli"
GH_ASSETS_URL="https://github.com/${GH_REPO}/releases/download"

#
# functions from https://github.com/client9/shlib
#
echoerr() {
  echo "$@" 1>&2
}

github_release() {
  owner_repo=$1
  version=$2
  header=$3
  test -z "$version" && version="latest"
  if [ "$version" = "latest" ]; then
    giturl="https://api.github.com/repos/${owner_repo}/releases/latest"
  else
    giturl="https://api.github.com/repos/${owner_repo}/releases/tags/${version}"
  fi
  json=$(http_copy "$giturl" "$header")
  test -z "$json" && return 1
  version=$(echo "$json" | tr -s '\n' ' ' | sed 's/.*"tag_name": *"//' | sed 's/".*//')
  test -z "$version" && return 1
  echo "$version"
}

http_download_curl() {
  local_file=$1
  source_url=$2
  header=$3
  if [ -z "$header" ]; then
    code=$(curl -w '%{http_code}' -sL -o "$local_file" "$source_url")
  else
    code=$(curl -w '%{http_code}' -sL -H "$header" -o "$local_file" "$source_url")
  fi
  if [ "$code" != "200" ]; then
    log_debug "http_download_curl received HTTP status $code"
    return 1
  fi
  return 0
}

http_download_wget() {
  local_file=$1
  source_url=$2
  header=$3
  if [ -z "$header" ]; then
    wget -q -O "$local_file" "$source_url"
  else
    wget -q --header "$header" -O "$local_file" "$source_url"
  fi
}

http_download() {
  log_debug "http_download $2"
  if is_command curl; then
    http_download_curl "$@"
    return
  elif is_command wget; then
    http_download_wget "$@"
    return
  fi
  log_crit "http_download unable to find wget or curl"
  return 1
}

http_copy() {
  tmp=$(mktemp)
  http_download "${tmp}" "$1" "$2" || return 1
  body=$(cat "$tmp")
  rm -f "${tmp}"
  echo "$body"
}

is_command() {
  command -v "$1" >/dev/null
}

log_prefix() {
  echo "$0"
}

# default priority
_logp=6

log_set_priority() {
  _logp="$1"
}

log_priority() {
  if test -z "$1"; then
    echo "$_logp"
    return
  fi
  [ "$1" -le "$_logp" ]
}

log_tag() {
  case $1 in
    0) echo "emerg" ;;
    1) echo "alert" ;;
    2) echo "crit" ;;
    3) echo "err" ;;
    4) echo "warning" ;;
    5) echo "notice" ;;
    6) echo "info" ;;
    7) echo "debug" ;;
    *) echo "$1" ;;
  esac
}

log_debug() {
  log_priority 7 || return 0
  echoerr "$(log_prefix)" "$(log_tag 7)" "$@"
}

log_info() {
  log_priority 6 || return 0
  echoerr "$(log_prefix)" "$(log_tag 6)" "$@"
}

log_err() {
  log_priority 3 || return 0
  echoerr "$(log_prefix)" "$(log_tag 3)" "$@"
}

log_crit() {
  log_priority 2 || return 0
  echoerr "$(log_prefix)" "$(log_tag 2)" "$@"
}

uname_arch() {
  arch=$(uname -m)
  case $arch in
    x86) arch="i386" ;;
    i686) arch="i386" ;;
    aarch64) arch="arm64" ;;
    armv5*) arch="armv5" ;;
    armv6*) arch="armv6" ;;
    armv7*) arch="armv7" ;;
  esac
  echo "${arch}"
}

uname_os() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')

  case "$os" in
    msys*) os="windows" ;;
    mingw*) os="windows" ;;
    cygwin*) os="windows" ;;
    win*) os="windows" ;;
  esac

  echo "$os"
}

untar() {
  tarball=$1
  case "${tarball}" in
    *.tar.gz | *.tgz) tar -xzf "${tarball}" ;;
    *.tar) tar -xf "${tarball}" ;;
    *.zip) unzip "${tarball}" ;;
    *)
      log_err "untar unknown archive format for ${tarball}"
      return 1
      ;;
  esac
}
#
# end functions from https://github.com/client9/shlib
#

OS=$(uname_os)
ARCH=$(uname_arch)

TAG="${1-latest}"
INSTALL_DIR="${2-./bin}"
PROJ_NAME="loops_cli"
SHORT_BIN_NAME="loops"
AUTH_HEADER=""
if [ -n "$GITHUB_TOKEN" ]; then
  AUTH_HEADER="Authorization: Bearer ${GITHUB_TOKEN}"
fi
GH_RELEASE=$(github_release "$GH_REPO" "$TAG" "$AUTH_HEADER")
if [ "$OS" = "windows" ]; then
  GH_RELEASE_FILENAME="${PROJ_NAME}_${OS}_${ARCH}.zip"
else
  GH_RELEASE_FILENAME="${PROJ_NAME}_${OS}_${ARCH}.tar.gz"
fi
DOWNLOAD_URL="${GH_ASSETS_URL}/${GH_RELEASE}/${GH_RELEASE_FILENAME}"

execute() {
  TMPDIR=$(mktemp -d)
  if ! http_download "${TMPDIR}/${GH_RELEASE_FILENAME}" "$DOWNLOAD_URL" "$AUTH_HEADER"; then
    log_err "Failed to download $DOWNLOAD_URL"
    return 1
  fi
  (cd "$TMPDIR" && untar "$GH_RELEASE_FILENAME")
  mkdir -p "$INSTALL_DIR"

  if [ "$OS" = "windows" ]; then
    SHORT_BIN_NAME="${SHORT_BIN_NAME}.exe"
  fi
  install "${TMPDIR}/${SHORT_BIN_NAME}" "${INSTALL_DIR}/${SHORT_BIN_NAME}"

  rm -rf "$TMPDIR"
}

echo "Installing ${PROJ_NAME} $GH_RELEASE for $OS $ARCH... "
execute
echo "Done!"
echo "Installed to ${INSTALL_DIR}/${SHORT_BIN_NAME}"
