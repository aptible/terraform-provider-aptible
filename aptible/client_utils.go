package aptible

import (
	"encoding/json"
	"fmt"

	"github.com/aptible/go-deploy/models"
)

type clientError struct {
	Payload *models.InlineResponseDefault
}

func generateErrorFromClientError(abstractedError interface{}) error {
	/**
	Warning - we are using encoding/decoding to extract values from generated go-swagger client code pointers

	This is done because the go-swagger client generates a pointer, which results in a hard-to-read message
	https://github.com/go-swagger/go-swagger/issues/1007
	https://github.com/go-swagger/go-swagger/issues/2590
	*/
	data, err := json.Marshal(&abstractedError)
	if err != nil {
		return fmt.Errorf("Unable to properly decode error in marshal from client - %s\n", err.Error())
	}

	var out clientError
	if err := json.Unmarshal(data, &out); err != nil {
		return fmt.Errorf("Unable to properly decode error in unmarshal from client - %s\n", err.Error())
	}

	if out.Payload.Code == nil || out.Payload.Error == nil {
		var errString string
		if err != nil {
			errString = fmt.Sprintf("- %s", err.Error())
		}
		if out.Payload.Code != nil && *out.Payload.Code >= 400 {
			errString += fmt.Sprintf(" status code (%d)", *out.Payload.Code)
		}
		if out.Payload.Error != nil && len(*out.Payload.Error) > 0 {
			errString += fmt.Sprintf(" error (%s)", *out.Payload.Error)
		}
		if out.Payload.Message != nil && len(*out.Payload.Message) > 0 {
			errString += fmt.Sprintf(" error message (%s)", *out.Payload.Message)
		}
		return fmt.Errorf(
			"unable to properly decode error (missing fields to properly generate error)%s",
			errString,
		)
	}

	return fmt.Errorf("Error with status code: %d. %s - %s\n", *out.Payload.Code, *out.Payload.Error, *out.Payload.Message)
}
