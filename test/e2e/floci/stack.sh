#!/usr/bin/env bash
set -euo pipefail

root=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
compose_file="$root/test/e2e/floci/compose.yml"
command=${1:-up}
shift || true

if [[ "$command" != "up" && "$command" != "verify" && "$command" != "down" && "$command" != "logs" ]]; then
  echo "usage: $0 {up|verify|down|logs}" >&2
  exit 2
fi

new_account_id() {
  local random
  random=$(od -An -N6 -tu8 /dev/urandom | tr -d ' ')
  printf '%012d\n' "$((100000000000 + random % 900000000000))"
}

id=${LANDLORD_FLOCI_E2E_ID:-}
if [[ -z "$id" && "$command" != "up" && "$command" != "verify" ]]; then
  echo "LANDLORD_FLOCI_E2E_ID is required for $command" >&2
  exit 2
fi

account_id=${LANDLORD_FLOCI_ACCOUNT_ID:-}
if [[ -z "$account_id" && ( "$command" == "up" || "$command" == "verify" ) ]]; then
  account_id=$(new_account_id)
fi
if [[ -z "$id" ]]; then
  id="floci-${account_id}"
fi
if [[ ! "$id" =~ ^[a-z0-9][a-z0-9_.-]*$ ]]; then
  echo "LANDLORD_FLOCI_E2E_ID must contain only lowercase letters, numbers, dots, underscores, or hyphens" >&2
  exit 2
fi

artifact_dir="$root/dist/floci-e2e/$id"
env_file="$artifact_dir/stack.env"
project="landlord-floci-e2e-${id}"

compose() {
  docker compose --project-name "$project" --env-file "$env_file" -f "$compose_file" "$@"
}

prepare() {
  command -v docker >/dev/null || { echo "Docker is required" >&2; exit 1; }
  command -v go >/dev/null || { echo "Go is required" >&2; exit 1; }
  command -v zip >/dev/null || { echo "zip is required" >&2; exit 1; }
  if [[ ! "$account_id" =~ ^[0-9]{12}$ ]]; then
    echo "LANDLORD_FLOCI_ACCOUNT_ID must be a 12-digit Floci account ID" >&2
    exit 2
  fi

  mkdir -p "$artifact_dir/lambda"
  cat >"$env_file" <<EOF
FLOCI_ACCOUNT_ID=$account_id
FLOCI_NETWORK_NAME=landlord-floci-$id
FLOCI_LAMBDA_NAME=landlord-lifecycle-$id
FLOCI_LAMBDA_ROLE_NAME=landlord-lambda-$id
FLOCI_STATE_MACHINE_NAME=landlord-lifecycle-$id
FLOCI_STATE_MACHINE_ROLE_NAME=landlord-sfn-$id
FLOCI_ARTIFACT_DIR=$artifact_dir
FLOCI_API_PORT=0
LANDLORD_API_PORT=0
EOF

  cat >"$artifact_dir/config.yaml" <<EOF
database:
  provider: postgres
  host: postgres
  port: 5432
  user: landlord
  password: landlord_password
  database: landlord
  ssl_mode: disable
compute:
  ec2_runner:
    region: us-east-1
    instance_type: t3.micro
    subnet_id: subnet-default-a
    security_group_ids:
      - sg-default
workflow:
  default_provider: step-functions
  step_functions:
    region: us-east-1
    state_machine_arn: arn:aws:states:us-east-1:$account_id:stateMachine:landlord-lifecycle-$id
controller:
  enabled: true
  reconciliation_interval: 1s
  status_poll_interval: 1s
  workers: 1
  workflow_trigger_timeout: 30s
  shutdown_timeout: 30s
EOF

  (
    cd "$root"
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "$artifact_dir/lambda/bootstrap" ./cmd/workers/stepfunctions
  )
  (
    cd "$artifact_dir/lambda"
    cp ../config.yaml config.yaml
    zip -q -FS ../landlord-stepfunctions.zip bootstrap config.yaml
  )
}

record_endpoints() {
  local floci_port landlord_port
  floci_port=$(compose port floci 4566 | tr -d '\r')
  landlord_port=$(compose port landlord 8080 | tr -d '\r')
  cat >>"$env_file" <<EOF
FLOCI_ENDPOINT=http://$floci_port
LANDLORD_API_URL=http://$landlord_port
EOF
  printf 'Floci account: %s\nLandlord API: http://%s\nState: %s\n' "$account_id" "$landlord_port" "$env_file"
}

teardown() {
  if [[ -f "$env_file" ]] && [[ -n "$(compose ps -q floci 2>/dev/null)" ]]; then
    compose run --rm --no-deps cleanup || true
  fi
  compose down --volumes --remove-orphans
}

case "$command" in
  up)
    prepare
    compose up --build --detach --wait --wait-timeout 180
    record_endpoints
    ;;
  verify)
    prepare
    if ! compose up --build --detach --wait --wait-timeout 180; then
      compose logs >&2 || true
      teardown
      exit 1
    fi
    record_endpoints
    api_url=$(awk -F= '/^LANDLORD_API_URL=/{print $2}' "$env_file")
    curl --fail --silent --show-error "$api_url/ready" >/dev/null
    teardown
    ;;
  down)
    teardown
    ;;
  logs)
    compose logs "$@"
    ;;
esac
