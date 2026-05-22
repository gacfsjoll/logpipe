// Package schema provides JSON schema validation for incoming log entries,
// ensuring that parsed log lines conform to a user-defined field contract
// before they are forwarded to downstream sinks.
package schema

import (
	"errors"
	"fmt"

	"github.com/logpipe/logpipe/internal/parser"
)

// FieldType enumerates the supported field types for schema validation.
type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeNumber  FieldType = "number"
	FieldTypeBoolean FieldType = "boolean"
)

// FieldRule describes a single field constraint.
type FieldRule struct {
	// Name is the JSON key to validate.
	Name string
	// Required indicates the field must be present.
	Required bool
	// Type, when non-empty, asserts the JSON value type.
	Type FieldType
}

// Validator checks log entries against a set of FieldRules.
type Validator struct {
	rules []FieldRule
}

// New creates a Validator from the provided rules.
// Returns an error if rules is empty or any rule has an empty Name.
func New(rules []FieldRule) (*Validator, error) {
	if len(rules) == 0 {
		return nil, errors.New("schema: at least one rule is required")
	}
	for i, r := range rules {
		if r.Name == "" {
			return nil, fmt.Errorf("schema: rule[%d] has empty Name", i)
		}
	}
	return &Validator{rules: rules}, nil
}

// Validate checks entry against all rules.
// It returns a non-nil error describing the first violation found.
func (v *Validator) Validate(entry parser.Entry) error {
	for _, rule := range v.rules {
		val, exists := entry.Fields[rule.Name]
		if !exists {
			if rule.Required {
				return fmt.Errorf("schema: required field %q is missing", rule.Name)
			}
			continue
		}
		if rule.Type == "" {
			continue
		}
		if err := assertType(rule.Name, rule.Type, val); err != nil {
			return err
		}
	}
	return nil
}

func assertType(name string, want FieldType, val any) error {
	switch want {
	case FieldTypeString:
		if _, ok := val.(string); !ok {
			return fmt.Errorf("schema: field %q must be a string, got %T", name, val)
		}
	case FieldTypeNumber:
		if _, ok := val.(float64); !ok {
			return fmt.Errorf("schema: field %q must be a number, got %T", name, val)
		}
	case FieldTypeBoolean:
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("schema: field %q must be a boolean, got %T", name, val)
		}
	}
	return nil
}
