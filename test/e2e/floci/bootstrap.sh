#!/usr/bin/env bash
set -euo pipefail

aws=(aws --endpoint-url "$AWS_ENDPOINT_URL")

lambda_role_arn="arn:aws:iam::${FLOCI_ACCOUNT_ID}:role/${FLOCI_LAMBDA_ROLE_NAME}"
state_machine_role_arn="arn:aws:iam::${FLOCI_ACCOUNT_ID}:role/${FLOCI_STATE_MACHINE_ROLE_NAME}"

lambda_trust='{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"lambda.amazonaws.com"},"Action":"sts:AssumeRole"}]}'
state_machine_trust='{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"states.amazonaws.com"},"Action":"sts:AssumeRole"}]}'

"${aws[@]}" iam create-role \
  --role-name "$FLOCI_LAMBDA_ROLE_NAME" \
  --assume-role-policy-document "$lambda_trust" >/dev/null
"${aws[@]}" iam create-role \
  --role-name "$FLOCI_STATE_MACHINE_ROLE_NAME" \
  --assume-role-policy-document "$state_machine_trust" >/dev/null

lambda_arn=$("${aws[@]}" lambda create-function \
  --function-name "$FLOCI_LAMBDA_NAME" \
  --runtime provided.al2023 \
  --handler bootstrap \
  --role "$lambda_role_arn" \
  --timeout 900 \
  --memory-size 512 \
  --environment "Variables={LANDLORD_CONFIG=/var/task/config.yaml,AWS_ENDPOINT_URL=http://floci:4566,AWS_ACCESS_KEY_ID=${FLOCI_ACCOUNT_ID},AWS_SECRET_ACCESS_KEY=test,AWS_DEFAULT_REGION=us-east-1,AWS_REGION=us-east-1,AWS_EC2_METADATA_DISABLED=true}" \
  --zip-file fileb:///artifacts/landlord-stepfunctions.zip \
  --query 'FunctionArn' --output text)

definition=$(sed "s|\${LifecycleLambdaArn}|${lambda_arn}|g" /bootstrap/lifecycle.asl.json)
"${aws[@]}" stepfunctions create-state-machine \
  --name "$FLOCI_STATE_MACHINE_NAME" \
  --definition "$definition" \
  --role-arn "$state_machine_role_arn" \
  --type STANDARD >/dev/null
