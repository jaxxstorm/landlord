package restate

import (
	"testing"

	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/assert"
)

func TestMapInvocationSubStateWaiting(t *testing.T) {
	subState := mapInvocationSubState("suspended")
	assert.Equal(t, workflow.SubStateWaiting, subState)
}

func TestMapInvocationSubStateBackingOff(t *testing.T) {
	subState := mapInvocationSubState("backoff")
	assert.Equal(t, workflow.SubStateBackingOff, subState)
















}

func TestMapInvocationSubStateFailed(t *testing.T) {
	subState := mapInvocationSubState("error")
	assert.Equal(t, workflow.SubStateFailed, subState)
}

func TestMapInvocationSubStateRunning(t *testing.T) {
	subState := mapInvocationSubState("running")
	assert.Equal(t, workflow.SubStateRunning, subState)
}

func TestMapInvocationSubStateDefault(t *testing.T) {
	subState := mapInvocationSubState("unknown")
	assert.Equal(t, workflow.SubStateRunning, subState)
}
