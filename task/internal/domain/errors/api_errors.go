package errors

type ErrApiBadRequest struct { // 400
	S string `json:"error" example:"invalid json body"`
}

func (e ErrApiBadRequest) Error() string {
	return e.S
}

type ErrApiAuthFailed struct { // 403
	S string `json:"error" example:"authorization failed, wrong token"`
}

func (e ErrApiAuthFailed) Error() string {
	return e.S
}

type ErrApiNotFound struct { // 404
	S string `json:"error" example:"task id or approval login not found. please check variables"`
}

func (e ErrApiNotFound) Error() string {
	return e.S
}

type ErrApiInternalServerError struct { // 500
	S string `json:"error" example:"rpc error: code = Unavailable desc = connection error: desc = transport: Error while dialing dial tcp [::1]:4000: connectex: No connection could be made because the target machine actively refused it."`
}

func (e ErrApiInternalServerError) Error() string {
	return e.S
}
