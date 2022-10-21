package aptible

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestValidateURL(t *testing.T) {
	type args struct {
		i interface{}
		k string
	}
	testAttr := "test_url"
	tests := []struct {
		name       string
		args       args
		want       []string
		wantErrors []error
	}{
		{
			name:       "returns an error when given a non-string",
			args:       args{i: 1, k: "other_url"},
			wantErrors: []error{fmt.Errorf("expected type of \"other_url\" to be string")},
		},
		{
			name:       "returns an error when given an empty string",
			args:       args{i: "", k: testAttr},
			wantErrors: []error{fmt.Errorf("expected %q url to not be empty", testAttr)},
		},
		{
			name:       "returns an error when given an un-parsable url",
			args:       args{i: ":bad_scheme", k: testAttr},
			wantErrors: []error{fmt.Errorf("expected %q to be a valid url, got :bad_scheme: parse \":bad_scheme\": missing protocol scheme", testAttr)},
		},
		{
			name:       "returns an error when the url scheme is missing",
			args:       args{i: "no_scheme", k: testAttr},
			wantErrors: []error{fmt.Errorf("expected %q to have a scheme, got %v", testAttr, "no_scheme")},
		},
		{
			name:       "returns an error when the url host is missing",
			args:       args{i: "scheme://", k: testAttr},
			wantErrors: []error{fmt.Errorf("expected %q to have a host, got %v", testAttr, "scheme://")},
		},
		{
			name: "returns no errors when the url is valid",
			args: args{i: "scheme://host", k: testAttr},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErrors := validateURL(tt.args.i, tt.args.k)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateURL() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(gotErrors, tt.wantErrors) {
				t.Errorf("validateURL() gotErrors = %v, want %v", gotErrors, tt.wantErrors)
			}
		})
	}
}

func Test_errorsToWarnings(t *testing.T) {
	type args struct {
		// nolint:staticcheck
		validator schema.SchemaValidateFunc
	}
	tests := []struct {
		name       string
		args       args
		want       []string
		wantErrors []error
	}{
		{
			name: "appends errors to warnings",
			args: args{
				validator: func(_ interface{}, _ string) ([]string, []error) {
					return []string{"warning 1", "warning 2"}, []error{fmt.Errorf("error 1"), fmt.Errorf("error 2")}
				},
			},
			want: []string{"warning 1", "warning 2", "error 1", "error 2"},
		},
		{
			name: "skips nil errors",
			args: args{
				validator: func(_ interface{}, _ string) ([]string, []error) {
					return []string{"warning 1", "warning 2"}, []error{fmt.Errorf("error 1"), nil, fmt.Errorf("error 2")}
				},
			},
			want: []string{"warning 1", "warning 2", "error 1", "error 2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWarnings, gotErrors := errorsToWarnings(tt.args.validator)(nil, "don't worry about it")
			if !reflect.DeepEqual(gotWarnings, tt.want) {
				t.Errorf("warnings = %v, want %v", gotWarnings, tt.want)
			}
			if !reflect.DeepEqual(gotErrors, tt.wantErrors) {
				t.Errorf("errors = %v, want %v", gotErrors, tt.wantErrors)
			}
		})
	}
}
