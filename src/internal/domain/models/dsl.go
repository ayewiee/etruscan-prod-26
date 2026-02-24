package models

type DSLTransactionProps struct {
	Amount     int
	Currency   string
	MerchantID *string
	IPAddress  *string
	DeviceID   *string
	User       DSLUserProps
}

type DSLUserProps struct {
	Age    *int
	Region *string
}

type DSLError struct {
	Code     string
	Message  string
	Position *int
	Near     *string
}

type DSLValidationResult struct {
	IsValid              bool
	NormalizedExpression *string
	Errors               []*DSLError
}
