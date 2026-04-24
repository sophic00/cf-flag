package flagapi

type createUserRequest struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Country string `json:"country"`
}

type createFlagRequest struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Country    *string `json:"country,omitempty"`
	Percentage *int    `json:"percentage,omitempty"`
}

type createUserResponse struct {
	User UserRecord `json:"user"`
}

type createFlagResponse struct {
	Flag FlagRecord `json:"flag"`
}

type flagStatusResponse struct {
	FlagID string `json:"flagId"`
	UserID string `json:"userId"`
	Rule   string `json:"rule"`
	Active bool   `json:"active"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type UserRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Country string `json:"country"`
}

type FlagRecord struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Rule string `json:"rule"`
}

type RuleType string

const (
	RuleTypeCountry    RuleType = "country"
	RuleTypePercentage RuleType = "percentage"
)

type FlagRule struct {
	TypeName   RuleType
	Country    string
	Percentage int
}
