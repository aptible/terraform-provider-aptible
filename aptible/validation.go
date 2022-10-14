package aptible

import (
	"fmt"
	"net/url"
)

// Modified validation.IsURLWithScheme to simply check for a URL that has any scheme and a host.
//
// Ignore linter rule complaining about reimplementing validation.StringNotInSlice.
// lintignore:V013
func validateURL(i interface{}, k string) (_ []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return
	}

	if v == "" {
		errors = append(errors, fmt.Errorf("expected %q url to not be empty", k))
		return
	}

	u, err := url.Parse(v)
	if err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a valid url, got %v: %+v", k, v, err))
		return
	}

	if u.Scheme == "" {
		errors = append(errors, fmt.Errorf("expected %q to have a scheme, got %v", k, v))
		return
	}

	if u.Host == "" {
		errors = append(errors, fmt.Errorf("expected %q to have a host, got %v", k, v))
		return
	}

	return
}

func errorsToWarnings(validator func(i interface{}, k string) ([]string, []error)) func(i interface{}, k string) ([]string, []error) {
	return func(i interface{}, k string) ([]string, []error) {
		warns, errs := validator(i, k)
		for _, err := range errs {
			if err != nil {
				warns = append(warns, err.Error())
			}
		}

		return warns, nil
	}
}
