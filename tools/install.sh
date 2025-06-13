#!/bin/sh
BIN=bin/tcli
CLI_HOME="$HOME/.tcli"
CLI_HOME_BACKUP="$CLI_HOME.backup"
# config
CONFIG_SRC="tools/config.yaml"
# modules
MODULES_SRC="tools/modules.yaml"
# data
DATA_SRC="tools/data"

set -e

# shellcheck disable=SC2046,SC2091
if ! $(cmp -s "$BIN" /usr/local/bin/tcli); then
  echo "updating tcli"
  sudo cp "$BIN" /usr/local/bin
else
  echo "tcli is already at the latest version. Skipping copy."
fi

# check if config dir exists
if [ -d "$CLI_HOME" ]; then
  # backup current config
  echo "backing up current config to $CLI_HOME_BACKUP"
  cp -r "$CLI_HOME" "$CLI_HOME_BACKUP"
else
  # ensure dir
  mkdir -p "$CLI_HOME"
fi

echo "copying config, modules and data"
cp -r "$CONFIG_SRC" "$MODULES_SRC" "$DATA_SRC" "$CLI_HOME"

echo "done!"
