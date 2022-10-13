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

func errorToDiagnostic(err error) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Error,
		Summary:  err.Error(),
	}
}

func errorToDiagnostics(err error) diag.Diagnostics {
	return diag.Diagnostics{errorToDiagnostic(err)}
}
