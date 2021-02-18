#!/usr/bin/env bash

{ # this ensures the entire script is downloaded #

  BINARY_DEST=${2:-/usr/local/bin}
  BINARY_NAME="rhoas"
  SRC_ORG="bf3fc6c"
  SRC_REPO="cli"
  OS_TYPE="linux"
  OS_LONG_BIT=""
  RHOAS_VERSION="0.16.0"
  API_BASE_URL="https://api.github.com"
  API_RELEASE_URL="${API_BASE_URL}/repos/${SRC_ORG}/${SRC_REPO}/releases/latest"
  DOWNLOAD_DIR="/tmp"
  USER_SHELL=$(ps -p $$ | tail -1 | awk '{print $NF}')

  has() {
    type "$1" > /dev/null 2>&1
  }

  update_profile() {
    [ -f "$1" ] || return 1

    grep -F "$source_line" "$1" > /dev/null 2>&1
    if [ $? -ne 0 ]; then
      echo -e "\n$source_line" >> "$1"
    fi
  }

  # get OS type
  if [[ "$OSTYPE" == "darwin"* ]]; then
    OS_TYPE = "macOS"
  fi

  MACHINE_TYPE=`uname -m`
  if [ ${MACHINE_TYPE} == 'x86_64' ]; then
    OS_LONG_BIT="64"
  else
    OS_LONG_BIT="386"
  fi

  ASSET_NAME="${BINARY_NAME}_${RHOAS_VERSION}_${OS_TYPE}_amd${OS_LONG_BIT}"
  ASSET_NAME_COMPRESSED="${ASSET_NAME}.tar.gz"

  if ! has "curl"; then
    echo "curl is required to download this binary"
    exit 1
  fi

  DOWNLOAD_URL=$(curl -s "${API_RELEASE_URL}" \
  | grep "browser.download_url.*${ASSET_NAME_COMPRESSED}" \
  | cut -d '"' -f 4)

  cd "$DOWNLOAD_DIR"

  # wget is faster, use it to download the release if available
  if has "wget"; then
    wget "$DOWNLOAD_URL"
  else
    curl -L "$DOWNLOAD_URL" --output ${ASSET_NAME_COMPRESSED}
  fi

  # unpack and place the binary in the users PATH
  tar xvf $ASSET_NAME_COMPRESSED
  rm -rf "${ASSET_NAME_COMPRESSED}"
  cp "${ASSET_NAME}/bin/${BINARY_NAME}" "${BINARY_DEST}/${BINARY_NAME}"
  rm -rf "${ASSET_NAME}"


  if [ "$USER_SHELL" == "zsh" ]; then
    rhoas completion zsh > "${fpath[1]}/_rhoas"
    
    update_profile "$HOME/.zshrc"
  elif [ "$(uname)" == "Linux" ]; then
    rhoas completion bash > /usr/local/etc/bash_completion.d/rhoas

    update_profile "$HOME/.bashrc" || update_profile "$HOME/.bash_profile"
  elif [ "$(uname)" == "Darwin" ]; then
    rhoas completion bash > /etc/bash_completion.d/rhoas

    update_profile "$HOME/.profile" || update_profile "$HOME/.bash_profile"
  fi
} # this ensures the entire script is downloaded #