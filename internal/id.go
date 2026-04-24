package flagapi

import (
	"github.com/google/uuid"
)

func newUserID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return "usr_" + id.String(), nil
}

func newFlagID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return "flg_" + id.String(), nil
}
