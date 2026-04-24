package flagapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
)

func ParseFlagRule(raw string) (FlagRule, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return FlagRule{}, errors.New("rule is empty")
	}

	if after, ok := strings.CutPrefix(raw, "country:"); ok {
		country := strings.ToUpper(strings.TrimSpace(after))
		if !isCountryCode(country) {
			return FlagRule{}, errors.New("invalid country rule")
		}
		return FlagRule{TypeName: RuleTypeCountry, Country: country}, nil
	}

	if after, ok := strings.CutPrefix(raw, "pct:"); ok {
		pctText := strings.TrimSpace(after)
		pct, err := strconv.Atoi(pctText)
		if err != nil {
			return FlagRule{}, errors.New("invalid percentage rule")
		}
		if pct < 0 || pct > 100 {
			return FlagRule{}, errors.New("invalid percentage rule")
		}
		return FlagRule{TypeName: RuleTypePercentage, Percentage: pct}, nil
	}

	return FlagRule{}, errors.New("unsupported rule")
}

func PercentageEnabled(flagID, userID string, percentage int, hashKey []byte) bool {
	if percentage <= 0 {
		return false
	}
	if percentage >= 100 {
		return true
	}

	mac := hmac.New(sha256.New, hashKey)
	mac.Write([]byte(flagID))
	mac.Write([]byte{':'})
	mac.Write([]byte(userID))
	sum := mac.Sum(nil)

	bucket := binary.BigEndian.Uint64(sum[:8]) % 10000
	threshold := uint64(percentage * 100)
	return bucket < threshold
}
