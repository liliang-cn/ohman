package errors

import "errors"

// Predefined errors
var (
	// ErrManNotFound man page not found
	ErrManNotFound = errors.New("man page not found")

	// ErrLLMUnavailable LLM service unavailable
	ErrLLMUnavailable = errors.New("LLM service unavailable")

	// ErrConfigNotFound configuration file not found
	ErrConfigNotFound = errors.New("configuration not found")

	// ErrAPIKeyMissing API key not configured
	ErrAPIKeyMissing = errors.New("API key not configured")

	// ErrShellNotSupported shell type not supported
	ErrShellNotSupported = errors.New("shell not supported")

	// ErrNoFailedCommand no failed command detected
	ErrNoFailedCommand = errors.New("no failed command detected")

	// ErrInvalidConfig invalid configuration
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrTimeout request timeout
	ErrTimeout = errors.New("request timeout")

	// ErrRateLimit rate limit exceeded
	ErrRateLimit = errors.New("rate limit exceeded")
)

// IsManNotFound checks if error is man not found
func IsManNotFound(err error) bool {
	return errors.Is(err, ErrManNotFound)
}

// IsAPIKeyMissing checks if error is API key missing
func IsAPIKeyMissing(err error) bool {
	return errors.Is(err, ErrAPIKeyMissing)
}

// IsNoFailedCommand checks if error is no failed command
func IsNoFailedCommand(err error) bool {
	return errors.Is(err, ErrNoFailedCommand)
}
