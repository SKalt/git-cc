#!/usr/bin/env sh
resp=/tmp/latest
checksums=/tmp/checksums
repo=skalt/git-cc
usage_message="$0 [apk|deb|rpm|tar.gz|exe]
download a release of git-cc for your OS and chip.
"
usage() { echo "$usage_message"; }
is_installed() { command -v sha256sum  1>/dev/null 2>&1; }
dl_url() { echo "https://github.com/$repo/releases/download/$1/$2"; }
json_values() { cat - | grep -e "$1" | awk -F '"' '{ print $4 }'; }
download_metadata() {
  curl -s https://api.github.com/repos/$repo/releases/latest | tee $resp \
  | json_values "browser_download_url" \
  | grep "checksums.txt"               \
  | xargs curl -sL | tee $checksums
}
get_arch() {
  case "$(arch)" in
  x86_64|x64|amd64) echo "amd64" ;;
  i386|386)         echo "386"   ;;
  arm64)            echo "arm64" ;;
  *) return 1;;
  esac
}
get_os() {
  os="$(uname --operating-system | tr '[:upper:]' '[:lower:]')"
  case "$os" in
  *linux) echo "linux";;
  *darwin) echo "darwin";;
  *windows) echo "windows";; # somehow
  *) echo "unprocessable os '$os'" >/dev/stderr && return 1;; 
  esac
}
get_fmt() {
  case "$1" in
    '') echo ".tar.gz";;
    .tar.gz|tar.gz|gz|tar) echo ".tar.gz";;
    .apk|apk) echo ".apk";;
    .deb|deb) echo ".deb";;
    .rpm|rpm) echo ".rpm";;
    .exe|exe) echo ".exe";;
  esac
}
check_sha256() {
  if is_installed sha256sum; then sha256sum     --ignore-missing -c $checksums;
  elif is_installed shasum;  then shasum -a 256 --ignore-missing -c $checksums;
  else return 127; fi
}
main() {
  case "$1" in
  -h|--help) usage && return 0;
  esac
  set -eu
  os="$(get_os)";        echo "os=$os"
  arch="$(get_arch)";    echo "arch=$arch"
  fmt="$(get_fmt "${1:-}")"; echo "format=$fmt"
  name="$(
    download_metadata      \
      | awk '{ print $2 }' \
      | grep    "$arch"    \
      | grep -i "$os"      \
      | grep    "$fmt"     \
      | tail -1
  )"
  version="$(json_values 'tag_name' < $resp)";
  echo "version=$version"
  echo "name=$name"

  curl -sLO "$(dl_url "$version" "$name")"
  check_sha256
}

main "$@"