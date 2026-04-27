package flagapi

type checkFlagRequest struct {
	FlagID      string `json:"flagId"`
	UserID      string `json:"userId"`
	UserCountry string `json:"userCountry"`
}

type createFlagRequest struct {
	Name       string  `json:"name"`
	Country    *string `json:"country,omitempty"`
	Percentage *int    `json:"percentage,omitempty"`
}

type createFlagResponse struct {
	Flag FlagRecord `json:"flag"`
}

type listFlagsResponse struct {
	Flags []FlagRecord `json:"flags"`
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

type FlagRecord struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Rule string `json:"rule"`
}

type RuleType string

const (
	RuleTypeCountry           RuleType = "country"
	RuleTypePercentage        RuleType = "percentage"
	RuleTypeCountryPercentage RuleType = "country_percentage"
)

type FlagRule struct {
	TypeName   RuleType
	Country    string
	Percentage int
}
