package errors

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidJsonBody                  = errors.New("{\"error\": \"invalid json body\"}")
	ErrIdNotFound                       = errors.New("{\"error\": \"id not found\"}")
	ErrLoginNotFoundInApprovals         = errors.New("{\"error\": \"login not found in approvals\"}")
	ErrAuthFailed                       = errors.New("{\"error\": \"authorization failed, wrong token\"}")
	ErrTokenLoginNotEqualInitiatorLogin = errors.New("{\"error\": \"token login not equal initiator login\"}")
	ErrNotFound                         = errors.New("{\"error\": \"task id or approval login not found. please check variables\"}")
	ErrNothingToChange                  = errors.New("name and text can't be empty strings at same time")
	ErrApprovalHasBeenDone              = errors.New("approval has already been done")
	ErrTaskNotAvailableForApproval      = errors.New("task is not available for approval")
)

type JsonErrWrapper struct {
	E string
}

func (j JsonErrWrapper) Error() string {
	return fmt.Sprintf("{\"error\": \"%s\"}", j.E)
}
