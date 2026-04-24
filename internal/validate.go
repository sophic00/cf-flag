package flagapi

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var countryCodePattern = regexp.MustCompile(`^[A-Z]{2}$`)

func normalizeUserInput(in createUserRequest) (UserRecord, error) {
	user := UserRecord{
		ID:      strings.TrimSpace(in.ID),
		Name:    strings.TrimSpace(in.Name),
		Email:   strings.ToLower(strings.TrimSpace(in.Email)),
		Country: strings.ToUpper(strings.TrimSpace(in.Country)),
	}

	if user.ID == "" {
		return UserRecord{}, errors.New("id is required")
	}
	if user.Name == "" {
		return UserRecord{}, errors.New("name is required")
	}
	if user.Email == "" {
		return UserRecord{}, errors.New("email is required")
	}
	if !strings.Contains(user.Email, "@") {
		return UserRecord{}, errors.New("email is invalid")
	}
	if !isCountryCode(user.Country) {
		return UserRecord{}, errors.New("country must be an ISO-2 code")
	}

	return user, nil
}

func normalizeFlagInput(in createFlagRequest) (FlagRecord, error) {
	flag := FlagRecord{
		ID:   strings.TrimSpace(in.ID),
		Name: strings.TrimSpace(in.Name),
	}

	if flag.ID == "" {
		return FlagRecord{}, errors.New("id is required")
	}
	if flag.Name == "" {
		return FlagRecord{}, errors.New("name is required")
	}

	if (in.Country == nil) == (in.Percentage == nil) {
		return FlagRecord{}, errors.New("exactly one of country or percentage is required")
	}

	if in.Country != nil {
		country := strings.ToUpper(strings.TrimSpace(*in.Country))
		if !isCountryCode(country) {
			return FlagRecord{}, errors.New("country must be an ISO-2 code")
		}
		flag.Rule = "country:" + country
		return flag, nil
	}

	pct := *in.Percentage
	if pct < 0 || pct > 100 {
		return FlagRecord{}, errors.New("percentage must be in range 0..100")
	}
	flag.Rule = fmt.Sprintf("pct:%d", pct)
	return flag, nil
}

func isCountryCode(v string) bool {
	return countryCodePattern.MatchString(v)
}
