package compute

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// SchemaValidationError captures JSON schema validation issues.
type SchemaValidationError struct {
	Details []string
}

func (e *SchemaValidationError) Error() string {
	if len(e.Details) == 0 {
		return "schema validation failed"
	}
	return fmt.Sprintf("schema validation failed: %s", e.Details[0])
}

// ValidateConfigAgainstSchema validates provider config against provider schema.
func ValidateConfigAgainstSchema(provider Provider, config []byte) error {
	if provider == nil {
		return fmt.Errorf("compute provider not configured")
	}

	schema := provider.ConfigSchema()
	if len(schema) == 0 {
		return fmt.Errorf("compute provider schema not available")
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", bytes.NewReader(schema)); err != nil {
		return fmt.Errorf("load schema: %w", err)
	}
	compiled, err := compiler.Compile("schema.json")
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}

	var payload interface{}
	if err := json.Unmarshal(config, &payload); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	if err := compiled.Validate(payload); err != nil {
		if vErr, ok := err.(*jsonschema.ValidationError); ok {
			return &SchemaValidationError{Details: flattenValidationErrors(vErr)}
		}
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}

func flattenValidationErrors(err *jsonschema.ValidationError) []string {
	var details []string
	var walk func(e *jsonschema.ValidationError)
	walk = func(e *jsonschema.ValidationError) {
		location := e.InstanceLocation
		if location == "" {
			location = "/"
		}
		details = append(details, fmt.Sprintf("%s: %s", location, e.Message))
		for _, cause := range e.Causes {
			walk(cause)
		}
	}
	walk(err)
	return details
}
