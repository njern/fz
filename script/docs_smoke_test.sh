#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
README_PATH="$ROOT_DIR/README.md"

if [ ! -f "$README_PATH" ]; then
  echo "README.md not found at $README_PATH" >&2
  exit 1
fi

tmp_dir="$(mktemp -d)"
trap 'chmod -R u+w "$tmp_dir" 2>/dev/null || true; rm -rf "$tmp_dir"' EXIT

export HOME="$tmp_dir/home"
export XDG_CONFIG_HOME="$tmp_dir/xdg"
export PATH="$ROOT_DIR:$PATH"
mkdir -p "$HOME" "$XDG_CONFIG_HOME"

block_count=0
line_no=0
block_start=0
in_block=0
smoke_block=0
block_shell=""
block_content=""

run_block() {
  local shell_name="$1"
  local content="$2"
  local start_line="$3"
  local script_path="$tmp_dir/block_${block_count}.${shell_name}"

  printf '%s' "$content" >"$script_path"
  echo "=== README smoke block $block_count ($shell_name, line $start_line) ==="

  case "$shell_name" in
    sh | bash)
      bash "$script_path" </dev/null 3<&-
      ;;
    zsh)
      if ! command -v zsh >/dev/null 2>&1; then
        echo "Skipping zsh block; zsh is not installed."
        return 0
      fi
      local wrapper_path="$tmp_dir/block_${block_count}.wrapper.zsh"
      {
        echo "set -euo pipefail"
        echo "autoload -Uz compinit"
        echo "compinit"
        cat "$script_path"
      } >"$wrapper_path"
      zsh "$wrapper_path" </dev/null 3<&-
      ;;
    powershell | pwsh)
      local ps_bin=""
      if command -v pwsh >/dev/null 2>&1; then
        ps_bin="pwsh"
      elif command -v powershell >/dev/null 2>&1; then
        ps_bin="powershell"
      else
        echo "Skipping PowerShell block; PowerShell is not installed."
        return 0
      fi
      local wrapper_path="$tmp_dir/block_${block_count}.ps1"
      {
        echo "\$ErrorActionPreference = 'Stop'"
        cat "$script_path"
      } >"$wrapper_path"
      "$ps_bin" -NoLogo -NoProfile -NonInteractive -File "$wrapper_path" </dev/null 3<&-
      ;;
    *)
      echo "Unsupported smoke-docs shell '$shell_name' at README line $start_line" >&2
      return 1
      ;;
  esac
}

exec 3<"$README_PATH"

while :; do
  line=""
  if ! IFS= read -r line <&3; then
    if [ -z "$line" ]; then
      break
    fi
  fi
  line_no=$((line_no + 1))

  if [ "$in_block" -eq 0 ]; then
    if [[ "$line" == \`\`\`* ]]; then
      in_block=1
      smoke_block=0
      block_shell=""
      block_content=""
      block_start=$line_no

      info_string="${line#\`\`\`}"
      read -r -a tokens <<<"$info_string"
      for token in "${tokens[@]}"; do
        case "$token" in
          sh | bash | zsh | powershell | pwsh)
            block_shell="$token"
            ;;
          smoke-docs)
            smoke_block=1
            ;;
        esac
      done
    fi
    continue
  fi

  if [[ "$line" == '```' ]]; then
    if [ "$smoke_block" -eq 1 ]; then
      if [ -z "$block_shell" ]; then
        echo "smoke-docs block starting at README line $block_start is missing a shell" >&2
        exit 1
      fi
      block_count=$((block_count + 1))
      run_block "$block_shell" "$block_content" "$block_start"
    fi
    in_block=0
    smoke_block=0
    block_shell=""
    block_content=""
    continue
  fi

  block_content+="$line"$'\n'
done

if [ "$in_block" -eq 1 ]; then
  echo "README.md ended while a fenced code block was still open" >&2
  exit 1
fi

if [ "$block_count" -eq 0 ]; then
  echo "No smoke-docs blocks found in README.md" >&2
  exit 1
fi

echo "Executed $block_count README smoke block(s)."
