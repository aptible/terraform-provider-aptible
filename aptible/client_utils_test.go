package aptible

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"

	"github.com/aptible/go-deploy/aptible"
	"github.com/aptible/go-deploy/client/operations"
	"github.com/aptible/go-deploy/models"
)

var (
	badRequestCode          = int64(400)
	notFoundText            = "not_found"
	resourceNotFoundMessage = "resource not found"
)

func TestGenerateErrorFromClientError(t *testing.T) {
	type args struct {
		abstractedError interface{}
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		errorBody string
	}{
		{
			name: "return a common client error call",
			args: args{
				abstractedError: operations.PatchDatabasesIDDefault{
					Payload: &models.InlineResponseDefault{
						Code:    &badRequestCode,
						Error:   &notFoundText,
						Message: &resourceNotFoundMessage,
					},
				},
			},
			wantErr: true,
			errorBody: fmt.Sprintf(
				"Error with status code: %d. %s - %s\n",
				badRequestCode,
				notFoundText,
				resourceNotFoundMessage,
			),
		},
		{
			name: "return a pre-baked error when error code not found",
			args: args{
				abstractedError: operations.PatchDatabasesIDDefault{
					Payload: &models.InlineResponseDefault{
						Code:    nil,
						Error:   &notFoundText,
						Message: &resourceNotFoundMessage,
					},
				},
			},
			wantErr:   true,
			errorBody: "unable to properly decode error (missing fields to properly generate error) -  error (not_found) error message (resource not found)",
		},
		{
			name: "return a pre-baked error when error code not found",
			args: args{
				abstractedError: operations.PatchDatabasesIDDefault{
					Payload: &models.InlineResponseDefault{
						Code:    &badRequestCode,
						Error:   nil,
						Message: &resourceNotFoundMessage,
					},
				},
			},
			wantErr:   true,
			errorBody: "unable to properly decode error (missing fields to properly generate error) -  status code (400) error message (resource not found)",
		},
		{
			name: "return a marshalable error with a nil payload should break early",
			args: args{
				abstractedError: nil,
			},
			wantErr:   true,
			errorBody: "Error without a valid payload: <nil>\n",
		},
		{
			name: "return a unmarshalable error (invalid json)",
			args: args{
				abstractedError: "{",
			},
			wantErr:   true,
			errorBody: "unable to properly decode error in unmarshal from client - json: cannot unmarshal string into Go value of type aptible.clientError",
		},
		{
			name: "return a marshalable error but without a payload (regular errors, non-swagger client)",
			args: args{
				abstractedError: errors.New("any old error that is not json type"),
			},
			wantErr:   true,
			errorBody: "Error without a valid payload: any old error that is not json type\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generateErrorFromClientError(tt.args.abstractedError)
			gotErr := err != nil
			if tt.wantErr != gotErr {
				t.Errorf("wanted an error (tt.wantErr), but did not get an error (gotErr) OR didn't want an error" +
					"and got an error!")
			}
			if gotErr && tt.errorBody != err.Error() {
				t.Errorf("generateErrorFromClientError() error = %v, wantErr %v, errorBody %s", err, tt.wantErr, tt.errorBody)
			}
		})
	}
}

func WithTestAccEnvironment(t *testing.T, f func(env aptible.Environment)) {
	if os.Getenv("TF_ACC") != "1" {
		return
	}

	client, err := aptible.SetUpClient()
	if err != nil {
		t.Fatalf("Failed to set up client for test environment - %s", err.Error())
		return
	}

	env, err := client.CreateEnvironment(testOrganizationId, int64(testStackId), aptible.EnvironmentCreateAttrs{
		Handle: acctest.RandString(10),
	})
	if err != nil {
		t.Fatalf("Failed to create test environment - %s", generateErrorFromClientError(err))
		return
	}

	defer func() {
		_ = client.DeleteEnvironment(env.ID)
	}()

	f(env)
}
