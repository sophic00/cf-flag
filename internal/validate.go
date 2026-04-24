package flagapi

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var countryCodePattern = regexp.MustCompile(`^[A-Z]{2}$`)
var emailPattern = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)
var errIDGeneration = errors.New("id generation failed")

func normalizeUserInput(in createUserRequest) (UserRecord, error) {
	name := strings.TrimSpace(in.Name)
	email := strings.ToLower(strings.TrimSpace(in.Email))
	country := strings.ToUpper(strings.TrimSpace(in.Country))

	if name == "" {
		return UserRecord{}, errors.New("name is required")
	}
	if email == "" {
		return UserRecord{}, errors.New("email is required")
	}
	if !emailPattern.MatchString(email) {
		return UserRecord{}, errors.New("email is invalid")
	}
	if !isCountryCode(country) {
		return UserRecord{}, errors.New("country must be an ISO-2 code")
	}

	id, err := newUserID()
	if err != nil {
		return UserRecord{}, fmt.Errorf("%w: %v", errIDGeneration, err)
	}

	return UserRecord{
		ID:      id,
		Name:    name,
		Email:   email,
		Country: country,
	}, nil
}

func normalizeFlagInput(in createFlagRequest) (FlagRecord, error) {
	name := strings.TrimSpace(in.Name)

	if name == "" {
		return FlagRecord{}, errors.New("name is required")
	}

	id, err := newFlagID()
	if err != nil {
		return FlagRecord{}, fmt.Errorf("%w: %v", errIDGeneration, err)
	}

	flag := FlagRecord{ID: id, Name: name}

	if (in.Country == nil) == (in.Percentage == nil) {
		return FlagRecord{}, errors.New("exactly one of country or percentage is required")
	}

	if in.Country != nil {
		country := strings.ToUpper(strings.TrimSpace(*in.Country))
		if !isCountryCode(country) {
			return FlagRecord{}, errors.New("country must be an ISO-2 code")
		}
		flag.Rule = "country:" + country
	} else {
		pct := *in.Percentage
		if pct < 0 || pct > 100 {
			return FlagRecord{}, errors.New("percentage must be in range 0..100")
		}
		flag.Rule = fmt.Sprintf("pct:%d", pct)
	}

	return flag, nil
}

func isCountryCode(v string) bool {
	return countryCodePattern.MatchString(v)
}
