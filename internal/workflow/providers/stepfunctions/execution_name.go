package stepfunctions

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/jaxxstorm/landlord/internal/tenant"
	"github.com/jaxxstorm/landlord/internal/workflow"
)

const (
	maxExecutionNameLength  = 80
	executionRevisionLength = 16
)

func executionName(request *workflow.ProvisionRequest) (string, error) {
	tenantID := request.TenantUUID
	if tenantID == "" {
		tenantID = request.TenantID
	}
	if tenantID == "" {
		return "", fmt.Errorf("tenant identifier is required")
	}

	operation := normalizeExecutionNamePart(request.Operation)
	if operation == "" {
		operation = "provision"
	}
	operation = truncateExecutionNamePart(operation, 24)

	revision, err := executionRevision(request)
	if err != nil {
		return "", err
	}

	tenantPart := normalizeExecutionNamePart(tenantID)
	if tenantPart == "" {
		return "", fmt.Errorf("tenant identifier contains no valid execution-name characters")
	}

	suffix := "-" + operation + "-" + revision
	tenantPart = truncateExecutionNamePart(tenantPart, maxExecutionNameLength-len("tenant-")-len(suffix))
	return "tenant-" + tenantPart + suffix, nil
}

func executionRevision(request *workflow.ProvisionRequest) (string, error) {
	revision := ""
	if request.Metadata != nil {
		revision = request.Metadata["config_hash"]
	}
	if revision == "" {
		var err error
		revision, err = tenant.ComputeConfigHash(request.DesiredConfig)
		if err != nil {
			return "", fmt.Errorf("compute desired-state revision: %w", err)
		}
	}
	if revision == "" {
		revision = "empty"
	}

	digest := sha256.Sum256([]byte(revision))
	return hex.EncodeToString(digest[:])[:executionRevisionLength], nil
}

func normalizeExecutionNamePart(value string) string {
	var normalized strings.Builder
	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z', char >= '0' && char <= '9', char == '-', char == '_':
			normalized.WriteRune(char)
		case char >= 'A' && char <= 'Z':
			normalized.WriteRune(char + ('a' - 'A'))
		default:
			normalized.WriteByte('-')
		}
	}
	return strings.Trim(normalized.String(), "-_")
}

func truncateExecutionNamePart(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return strings.TrimRight(value[:limit], "-_")
}
