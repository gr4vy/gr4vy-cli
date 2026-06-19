package clierr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gr4vy/gr4vy-go/models/apierrors"
)

// apiError is the normalized view of a gr4vy-go API error.
type apiError struct {
	status  int
	code    string
	message string
}

// classifyAPIError extracts a normalized view from any gr4vy-go API error type.
func classifyAPIError(err error) (apiError, bool) {
	// Generic transport-level error carries the status directly.
	var generic *apierrors.APIError
	if errors.As(err, &generic) {
		return apiError{status: generic.StatusCode, message: generic.Message}, true
	}

	// Typed, documented errors. Each encodes its status in the type.
	var (
		e400 *apierrors.Error400
		e401 *apierrors.Error401
		e403 *apierrors.Error403
		e404 *apierrors.Error404
		e405 *apierrors.Error405
		e409 *apierrors.Error409
		e425 *apierrors.Error425
		e429 *apierrors.Error429
		e500 *apierrors.Error500
		e502 *apierrors.Error502
		e504 *apierrors.Error504
		eVal *apierrors.HTTPValidationError
	)
	switch {
	case errors.As(err, &e400):
		return apiError{400, strp(e400.Code), strp(e400.Message)}, true
	case errors.As(err, &e401):
		return apiError{401, strp(e401.Code), strp(e401.Message)}, true
	case errors.As(err, &e403):
		return apiError{403, strp(e403.Code), strp(e403.Message)}, true
	case errors.As(err, &e404):
		return apiError{404, strp(e404.Code), strp(e404.Message)}, true
	case errors.As(err, &e405):
		return apiError{405, strp(e405.Code), strp(e405.Message)}, true
	case errors.As(err, &e409):
		return apiError{409, strp(e409.Code), strp(e409.Message)}, true
	case errors.As(err, &e425):
		return apiError{425, strp(e425.Code), strp(e425.Message)}, true
	case errors.As(err, &e429):
		return apiError{429, strp(e429.Code), strp(e429.Message)}, true
	case errors.As(err, &e500):
		return apiError{500, strp(e500.Code), strp(e500.Message)}, true
	case errors.As(err, &e502):
		return apiError{502, strp(e502.Code), strp(e502.Message)}, true
	case errors.As(err, &e504):
		return apiError{504, strp(e504.Code), strp(e504.Message)}, true
	case errors.As(err, &eVal):
		return apiError{422, "validation_error", validationSummary(eVal)}, true
	}
	return apiError{}, false
}

// apiExitCode maps an API error to a process exit code.
func apiExitCode(err error) int {
	ae, ok := classifyAPIError(err)
	if !ok {
		return ExitGeneric
	}
	switch {
	case ae.status == 429:
		return ExitRateLimit
	case ae.status >= 500:
		return ExitAPIServer
	case ae.status >= 400:
		return ExitAPIClient
	default:
		return ExitGeneric
	}
}

// FormatError renders a user-facing error line. API errors become
// "404 not_found — message"; everything else uses the plain message.
func FormatError(err error) string {
	if ae, ok := classifyAPIError(err); ok {
		parts := []string{fmt.Sprintf("%d", ae.status)}
		if ae.code != "" {
			parts = append(parts, ae.code)
		}
		line := strings.Join(parts, " ")
		if ae.message != "" {
			line += " — " + ae.message
		}
		return line
	}
	return err.Error()
}

func validationSummary(e *apierrors.HTTPValidationError) string {
	if e == nil || len(e.Detail) == 0 {
		return "request validation failed"
	}
	var msgs []string
	for _, d := range e.Detail {
		msgs = append(msgs, d.Msg)
	}
	return strings.Join(msgs, "; ")
}

func strp(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
