package aptible

import (
	"errors"
	"fmt"
	"testing"

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
			errorBody: "unable to properly decode error (missing fields to properly generate error) -  error (not_found) error message (resource not found)\n",
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
			errorBody: "unable to properly decode error (missing fields to properly generate error) -  status code (400) error message (resource not found)\n",
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
			name: "return a unmarshalable error (but unmarshalable into expected type)",
			args: args{
				abstractedError: "{",
			},
			wantErr:   true,
			errorBody: "Unable to properly decode error in unmarshal from client - json: cannot unmarshal string into Go value of type aptible.clientError\n",
		},
		{
			name: "return a unmarshalable error (but unmarshalable into expected type)",
			args: args{
				abstractedError: errors.New("any old error that is not json type"),
			},
			wantErr:   true,
			errorBody: "Error without a valid payload: any old error that is not json type\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := generateErrorFromClientError(tt.args.abstractedError); (err != nil) == tt.wantErr && tt.errorBody != err.Error() {
				t.Errorf("generateErrorFromClientError() error = %v, wantErr %v, errorBody %s", err, tt.wantErr, tt.errorBody)
			}
		})
	}
}
