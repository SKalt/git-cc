#!/usr/bin/env sh
# shellcheck disable=SC2016
releases=/tmp/latest
checksums=/tmp/checksums
log_file=/tmp/git-cc-install.log
repo=skalt/git-cc

# global script state
fmt=
dry_run=false # set to 'true' to for a dry run

usage_message="$0 [-h|--help] [--dry-run] [FMT]
download a release of git-cc for your OS and instruction set architecture.

ARGS:
  -h|--help  print this message and exit
  --dry-run  print rather than follow the download url for the binary
  FMT        The download format. Valid values are .apk, .deb, .rpm, .tar.gz,
             and .exe; default .tar.gz.
"

usage() { echo "$usage_message"; }
is_installed() { command -v "$1"  1>/dev/null 2>&1; }

should_use_color() {
  test -t 1 && # stdout (device 1) is a tty
  test -z "${NO_COLOR:-}" && # the NO_COLOR variable isn't set
  is_installed tput
}

if should_use_color; then
  red="$(tput setaf 1)"
  orange="$(tput setaf 3)"
  blue="$(tput setaf 4)"
  gray="$(tput setaf 7)"
  reset="$(tput sgr0)"
else
  red=""
  orange=""
  blue=""
  gray=""
  reset=""
fi

log_message() {
  level="$1"; shift;
  color="$1"; shift;
  message="$*"
  printf "%s\t%s\t%s\n" "$level" "$(date '+%Y-%m-%dT%H:%M:%S%z')" "$message" |
  tee -a $log_file                                                         |
  sed "s/^/${color}/g; s/\t/\t${gray}/1; s/\t/${reset}\t/2;" >&2
}

log_info() { log_message "INFO" "$blue" "$*"; }
log_warning() { log_message "WARN" "$orange" "$*"; }
log_error() { log_message "ERROR" "$red" "$*"; }
fail() { log_error "$1" && exit "${2:-1}"; }

dl_url() { echo "https://github.com/$repo/releases/download/$1/$2"; }
json_values() { cat - | grep -e "$1" | awk -F '"' '{ print $4 }'; }

download_metadata() {
  url="https://api.github.com/repos/$repo/releases/latest"
  # log_info "downloading release metadata to $releases"
  curl -s "$url"                       |
    tee $releases                      |
    json_values "browser_download_url" |
    grep "checksums.txt"               |
    xargs curl -Ls                     |
    tee $checksums
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
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
  *linux) echo "linux";;
  *darwin) echo "darwin";;
  *windows) echo "windows";; # if running from git bash, I guess
  *) fail "unprocessable os '$os'";; 
  esac
}

get_fmt() {
  case "${1:-}" in
    '') echo ".tar.gz";;
    .tar.gz|tar.gz|gz|tar) echo ".tar.gz";;
    .apk|apk) echo ".apk";;
    .deb|deb) echo ".deb";;
    .rpm|rpm) echo ".rpm";;
    .exe|exe) echo ".exe";;
    *) fail "invalid format: $1"
  esac
}


get_name() {
  arch="$1"; os="$2"; fmt="$3";
  download_metadata    | # prints the checksum file in `sum name` format
    awk '{ print $2 }' | # extract the names
    grep    "$arch"    | # search for a compatible release
    grep -i "$os"      |
    grep    "$fmt"     |
    tail -1
}

check_sha256() {
  cmd=""
  if is_installed sha256sum; then
    cmd="sha256sum --ignore-missing -c $checksums";
  elif is_installed shasum; then
    cmd="shasum -a 256 --ignore-missing -c $checksums";
  else
    fail 'unable to find `sha256sum` or `shasum`' 127;
  fi

  if test "$dry_run" = "true"; then 
    log_info "would check shasums by running \`$cmd\` in /tmp"
  else
    log_info "checking shasums: running \`$cmd\` in /tmp"
    (cd /tmp; eval "$cmd")
    # need to cd into /tmp to run the check since the paths in the checksums file
    #are relative
  fi
}

download_git_cc() {
  version="$1"
  name="$2"
  url=; url="$(dl_url "$version" "$name")"
  if test "$dry_run" = "true"; then
    log_info "would download $name into /tmp/"
  else
    log_info "downloading $name into /tmp/"
    curl -sL  "$url" > "/tmp/$name"
  fi
}

install_git_cc() {
  version="$1"
  name="$2"
  cmd=
  case "$name" in
    *.tar.gz)
      # TODO: figure out a more idiomatic user-local location to place it?
      cmd="tar -xf /tmp/$name && sudo cp /tmp/git-cc /usr/local/bin/;"
      ;;
    *.apk) cmd="apk add /tmp/$name";; # TODO: verify this works
    *.deb) cmd="sudo apt-get install -y /tmp/$name";;
    *.rpm)
      if is_installed yum; then   cmd="sudo yum localinstall /tmp/$name"
      elif is_installed dnf; then cmd="sudo dnf localinstall /tmp/$name"
      elif is_installed rpm; then cmd="sudo rpm -i /tmp/$name"
      else fail 'neither `yum`, `dnf`, nor `rmp` found' 127
      fi
      ;;
    *.exe)
      log_warning "you'll need to install /tmp/$name manually"
      ;;
  esac
  if test "$dry_run" = "true"; then
    log_info "would install git-cc by running \`$cmd\`"
  else
    log_info "installing git-cc: running \`$cmd\`"
    eval "$cmd"
  fi
}

main() {
  set -eu
  while [ -n "${1:-}" ]; do
    case "$1" in
      -h|--help) usage && return 0;;
      --dry-run) export dry_run=true; shift;;
      *)
        if test -n "$fmt"; then
          log_error "FMT can only be passed once"
          usage
          exit 1
        else
          fmt="$(get_fmt "$1")"
          shift
        fi
      ;;
    esac
  done

  os="$(get_os)";
  arch="$(get_arch)";
  if test -z "${fmt:-}"; then  fmt="$(get_fmt "")"; fi 

  log_info "dry_run=$dry_run"
  log_info "os=$os"
  log_info "arch=$arch"
  log_info "format=$fmt"
  
  name="$(get_name "$arch" "$os" "$fmt")"
  log_info "name=$name"

  version="$(json_values 'tag_name' < $releases)";
  log_info "version=$version"

  if [ -z "${name:-}" ]; then
    fail "unfortunately, there's no prebuilt release for $fmt and $arch. " \
      'try `go get github.com/skalt/git-cc` to compile it yourself.'
  fi
  download_git_cc "$version" "$name"
  check_sha256
  install_git_cc "$version" "$name"
}

main "$@"
