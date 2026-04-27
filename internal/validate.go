package flagapi

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var countryCodePattern = regexp.MustCompile(`^[A-Z]{2}$`)
var errIDGeneration = errors.New("id generation failed")

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

	if in.Country == nil && in.Percentage == nil {
		return FlagRecord{}, errors.New("at least one of country or percentage is required")
	}

	hasCountry := in.Country != nil
	hasPercentage := in.Percentage != nil

	if hasCountry {
		country := strings.ToUpper(strings.TrimSpace(*in.Country))
		if !isCountryCode(country) {
			return FlagRecord{}, errors.New("country must be an ISO-2 code")
		}
		if hasPercentage {
			pct := *in.Percentage
			if pct < 0 || pct > 100 {
				return FlagRecord{}, errors.New("percentage must be in range 0..100")
			}
			flag.Rule = fmt.Sprintf("country_pct:%s:%d", country, pct)
			return flag, nil
		}
		flag.Rule = "country:" + country
		return flag, nil
	}

	if hasPercentage {
		pct := *in.Percentage
		if pct < 0 || pct > 100 {
			return FlagRecord{}, errors.New("percentage must be in range 0..100")
		}
		flag.Rule = fmt.Sprintf("pct:%d", pct)
		return flag, nil
	}

	return FlagRecord{}, errors.New("invalid rule payload")
}

func isCountryCode(v string) bool {
	return countryCodePattern.MatchString(v)
}
