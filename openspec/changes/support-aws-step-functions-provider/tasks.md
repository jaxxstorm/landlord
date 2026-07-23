## 1. Configuration and AWS Client Setup

- [x] 1.1 Extend Step Functions configuration with a required `state_machine_arn` and explicit optional caller assume-role settings.
- [x] 1.2 Update Viper environment bindings and configuration validation for the Step Functions provider.
- [x] 1.3 Reuse `internal/cloud/awsconfig` to load Step Functions caller credentials and add injectable SFN/STS client seams.
- [x] 1.4 Update configuration and provider documentation to use `step-functions` and document the new required settings.
- [x] 1.5 Add configuration validation and constructor unit tests for valid, incomplete, and assume-role configurations.

## 2. Step Functions Provider

- [x] 2.1 Replace the simulated provider constructor and ARN generation with a configured state-machine ARN and real AWS SDK clients.
- [x] 2.2 Implement deterministic Step Functions execution-name generation from tenant UUID, operation, and desired-state revision.
- [x] 2.3 Implement StartExecution with serialized provisioning input and duplicate-execution recovery.
- [x] 2.4 Implement DescribeExecution status mapping, output/error extraction, and paginated execution-history mapping.
- [x] 2.5 Implement idempotent StopExecution and AWS error translation for missing and terminal executions.
- [x] 2.6 Define behavior for unsupported state-machine create/delete interface methods when using a pre-provisioned machine.
- [x] 2.7 Add mocked-SFN provider tests for start, duplicate start, status, history, failures, stop, and context cancellation.

## 3. Reconciliation Integration

- [x] 3.1 Update controller Step Functions invocation to use the shared state machine while preserving tenant lifecycle revision metadata.
- [x] 3.2 Update reconciler status tests for Step Functions succeeded, failed, timed-out, aborted, and missing execution responses.
- [x] 3.3 Verify changed desired configuration produces a distinct execution identity while retrying the same revision does not.

## 4. Shared Lifecycle Executor

- [x] 4.1 Extract payload validation, operation dispatch, compute-provider resolution, and result construction from the Temporal worker into a shared lifecycle executor.
- [x] 4.2 Adapt the Temporal worker to call the shared executor without changing its workflow/activity contract.
- [x] 4.3 Add unit tests that exercise provision, update, delete, invalid payload, and unknown compute provider through the shared executor.
- [x] 4.4 Preserve tracked compute execution IDs and provider callbacks when lifecycle work runs through the shared executor.

## 5. Step Functions Lambda Target

- [x] 5.1 Add a Lambda handler that initializes configured compute providers and invokes the shared lifecycle executor from the Step Functions payload.
- [x] 5.2 Pass the Step Functions execution ARN to the Lambda handler and include it in lifecycle logging and tracked compute calls.
- [x] 5.3 Add cooperative cancellation checks before mutable lifecycle operation boundaries.
- [x] 5.4 Add Lambda handler unit tests for successful lifecycle dispatch, validation errors, cancellation, and compute failures.

## 6. AWS Deployment Assets and Documentation

- [x] 6.1 Add a versioned Standard ASL definition that invokes the Lambda handler synchronously and defines bounded retry and catch behavior.
- [x] 6.2 Add an AWS deployment reference template for the state machine, Lambda function, logging, and required parameters.
- [x] 6.3 Document distinct least-privilege IAM policies for the Landlord caller, state-machine role, and Lambda compute role.
- [x] 6.4 Document deployment, rollback, observability, CloudWatch log correlation, and the cooperative cancellation limitation.

## 7. End-to-End Verification

- [x] 7.1 Add ASL definition validation tests and deployment-template static checks.
- [x] 7.2 Add LocalStack or AWS-compatible integration tests for start, describe, stop, terminal status, and duplicate execution behavior.
- [x] 7.3 Add an end-to-end lifecycle test using mock compute to verify Step Functions payload execution and reconciler state transitions.
- [x] 7.4 Run the focused Go test suites, full Go test suite, formatting, and OpenSpec validation; record any environment-dependent integration test prerequisites.
