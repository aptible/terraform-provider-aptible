package aptible

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func diagnosticsToError(ds diag.Diagnostics) error {
	var err error

	for _, d := range ds {
		if d.Severity == diag.Error {
			err = multierror.Append(err, fmt.Errorf("%s: %s", d.Summary, d.Detail))
		}
	}

	return err
}
