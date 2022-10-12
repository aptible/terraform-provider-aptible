package aptible

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/url"
)

type ResourceDiff struct {
	*schema.ResourceDiff
}

// Modified IsURLWithScheme to simply check for a URL
// https://github.com/hashicorp/terraform-plugin-sdk/blob/v1.17.2/helper/validation/web.go#L22
func ValidateURL(i interface{}, k string) (_ []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return
	}

	if v == "" {
		errors = append(errors, fmt.Errorf("expected %q url to not be empty, got %v", k, i))
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

func (d *ResourceDiff) IsProvided(attr string) bool {
	_, ok := d.GetOkExists(attr)
	return ok || !d.NewValueKnown(attr)
}
