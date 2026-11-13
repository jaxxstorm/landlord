#!/usr/bin/env bash
set -euo pipefail

aws=(aws --endpoint-url "$AWS_ENDPOINT_URL")

state_machine_arn=$("${aws[@]}" stepfunctions list-state-machines \
  --query "stateMachines[?name=='${FLOCI_STATE_MACHINE_NAME}'].stateMachineArn | [0]" \
  --output text 2>/dev/null || true)

if [[ -n "$state_machine_arn" && "$state_machine_arn" != "None" ]]; then
  executions=$("${aws[@]}" stepfunctions list-executions \
    --state-machine-arn "$state_machine_arn" \
    --status-filter RUNNING \
    --query 'executions[].executionArn' --output text 2>/dev/null || true)
  for execution in $executions; do
    "${aws[@]}" stepfunctions stop-execution --execution-arn "$execution" --error StackCleanup >/dev/null || true
  done
  "${aws[@]}" stepfunctions delete-state-machine --state-machine-arn "$state_machine_arn" >/dev/null || true
fi

"${aws[@]}" lambda delete-function --function-name "$FLOCI_LAMBDA_NAME" >/dev/null || true
"${aws[@]}" iam delete-role --role-name "$FLOCI_STATE_MACHINE_ROLE_NAME" >/dev/null || true
"${aws[@]}" iam delete-role --role-name "$FLOCI_LAMBDA_ROLE_NAME" >/dev/null || true
