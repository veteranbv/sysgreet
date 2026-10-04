package bootstrap

import (
	"errors"
	"fmt"
	"strings"
)

// PolicyValue represents the overwrite policy for configuration bootstrap.
type PolicyValue string

const (
	PolicyPrompt    PolicyValue = "prompt"
	PolicyKeep      PolicyValue = "keep"
	PolicyOverwrite PolicyValue = "overwrite"
)

// PolicySource identifies where the policy came from.
type PolicySource string

const (
	PolicySourceFlag    PolicySource = "flag"
	PolicySourceEnv     PolicySource = "env"
	PolicySourceDefault PolicySource = "default"
)

var (
	// ErrInvalidPolicy indicates an unsupported policy value was provided.
	ErrInvalidPolicy = errors.New("invalid config policy value")
)

// PolicyResolution captures the evaluated policy state.
type PolicyResolution struct {
	Value       PolicyValue
	Source      PolicySource
	Interactive bool
}

// ResolvePolicy evaluates the final policy according to precedence rules.
func ResolvePolicy(flagValue, envValue string, interactive bool) (PolicyResolution, error) {
	if flagValue != "" {
		value, err := ParsePolicy(flagValue)
		if err != nil {
			return PolicyResolution{}, err
		}
		return PolicyResolution{Value: value, Source: PolicySourceFlag, Interactive: interactive}, nil
	}

	if envValue != "" {
		value, err := ParsePolicy(envValue)
		if err != nil {
			return PolicyResolution{}, err
		}
		return PolicyResolution{Value: value, Source: PolicySourceEnv, Interactive: interactive}, nil
	}

	if !interactive {
		// Without a terminal there is no one to answer a prompt; leave any
		// existing config alone.
		return PolicyResolution{Value: PolicyKeep, Source: PolicySourceDefault, Interactive: interactive}, nil
	}

	return PolicyResolution{Value: PolicyPrompt, Source: PolicySourceDefault, Interactive: interactive}, nil
}

// ParsePolicy converts a raw string into a PolicyValue, returning an error for invalid inputs.
func ParsePolicy(input string) (PolicyValue, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case string(PolicyPrompt):
		return PolicyPrompt, nil
	case string(PolicyKeep):
		return PolicyKeep, nil
	case string(PolicyOverwrite):
		return PolicyOverwrite, nil
	default:
		return "", fmt.Errorf("%w %q (want prompt, keep, or overwrite)", ErrInvalidPolicy, input)
	}
}
