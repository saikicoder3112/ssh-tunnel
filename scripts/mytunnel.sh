# Source this file to get the `mytunnel` command:
#   source ~/.ssh/mytunnel.sh

export MYTUNNEL_CONFIG="${MYTUNNEL_CONFIG:-$HOME/.ssh/mytunnel.yaml}"

mytunnel() {
  if [ -x "$HOME/.ssh/mytunnel" ]; then
    "$HOME/.ssh/mytunnel" "$@"
    return
  fi
  echo "mytunnel: binary missing at ~/.ssh/mytunnel — rebuild with go build -o ~/.ssh/mytunnel" >&2
  return 1
}
