package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/goware/emailx"

	"github.com/pavelmemory/faceit-users/internal"
)

type validationProperty string

func (vp validationProperty) String() string {
	return string(vp)
}

const (
	propertyFirstName = validationProperty("FirstName")
	propertyLastName  = validationProperty("LastName")
	propertyNickname  = validationProperty("Nickname")
	propertyEmail     = validationProperty("Email")
	propertyCountry   = validationProperty("Country")
	propertyPassword  = validationProperty("Password")
)

// ValidationError encapsulates in it validation failure details.
type ValidationError struct {
	// Cause should be one of pre-defined standard errors.
	Cause error
	// Details any additional details about validation failure.
	Details map[string]interface{}
}

func (ve ValidationError) Error() string {
	d, _ := json.Marshal(struct {
		Cause   string                 `json:"cause"`
		Details map[string]interface{} `json:"details"`
	}{
		Cause:   ve.Cause.Error(),
		Details: ve.Details,
	})
	return string(d)
}

func (ve ValidationError) Is(err error) bool {
	return errors.Is(ve.Cause, err)
}

func validateBlankOrEmptyWithMaxLen(value, field string, maxlength int) func() error {
	return func() error {
		if strings.TrimSpace(value) == "" {
			return ValidationError{
				Cause:   internal.ErrBadInput,
				Details: map[string]interface{}{field: "blank or empty"},
			}
		}

		length := utf8.RuneCountInString(value)
		if length > maxlength {
			return ValidationError{
				Cause:   internal.ErrBadInput,
				Details: map[string]interface{}{field: fmt.Sprintf("exceeds max length: %d", maxlength)},
			}
		}

		return nil
	}
}

func validateEmailFormat(value, filed string) func() error {
	return func() error {
		if err := emailx.ValidateFast(value); err != nil {
			return ValidationError{
				Cause:   internal.ErrBadInput,
				Details: map[string]interface{}{filed: err.Error()},
			}
		}

		return nil
	}
}
