// Package ec2runner provides EC2-backed tenant runner instances.
package ec2runner

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// EC2Client is the EC2 API surface required by the runner provider. Keeping
// this narrow allows lifecycle behavior to be tested without AWS credentials.
type EC2Client interface {
	DescribeInstances(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	RunInstances(context.Context, *ec2.RunInstancesInput, ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	TerminateInstances(context.Context, *ec2.TerminateInstancesInput, ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
}
