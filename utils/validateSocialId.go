package utils

import (
	"errors"
	"challenge/graph/model"
)

func ValidateSocialId(data *model.SocialLoginRequestInput) error {

	if StringIsEmpty(data.SocialID) {
		return errors.New("socialId cannot be empty")
	}

	return nil
}
