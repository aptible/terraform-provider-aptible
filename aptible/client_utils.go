package aptible

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

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
	var errorString string

	defer func() {
		// encoding/json library can potentially error with a panic, causing poorly handled errors
		if err := recover(); err != nil {
			errorString = "panic occurred in marshal or unmarshalling of json client"
			log.Println(fmt.Sprintf("[ERROR] %s", errorString), err)
		}
	}()

	log.Println("[ERROR] error received and being processed", abstractedError)
	data, err := json.Marshal(&abstractedError)
	if err != nil {
		return fmt.Errorf("Unable to properly decode error in marshal from client - %s\n", err.Error())
	}

	var out clientError
	if err := json.Unmarshal(data, &out); err != nil {
		return fmt.Errorf("Unable to properly decode error in unmarshal from client - %s\n", err.Error())
	}

	// payload is a pointer and can be nil, but we don't have all our information from a http code, so we construct it
	if out.Payload == nil {
		errorString = fmt.Sprint(
			"Error without a valid payload: ", abstractedError, "\n",
		)
	} else if out.Payload.Code == nil || out.Payload.Error == nil {
		if out.Payload.Code != nil && *out.Payload.Code >= 400 {
			errorString += fmt.Sprintf(" status code (%d)", *out.Payload.Code)
		}

		if out.Payload.Error != nil && len(*out.Payload.Error) > 0 {
			errorString += fmt.Sprintf(" error (%s)", *out.Payload.Error)
		}

		if out.Payload.Message != nil && len(*out.Payload.Message) > 0 {
			errorString += fmt.Sprintf(" error message (%s)", *out.Payload.Message)
		}

		return fmt.Errorf(
			"unable to properly decode error (missing fields to properly generate error) - %s\n",
			errorString,
		)
	} else {
		// payload is not nil, and fully populated
		errorString = fmt.Sprintf(
			"Error with status code: %d. %s - %s\n",
			*out.Payload.Code,
			*out.Payload.Error,
			*out.Payload.Message,
		)
	}

	return errors.New(errorString)
}
