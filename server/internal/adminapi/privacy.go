package adminapi

import (
	"github.com/Alievs-corp/hypedomino/server/internal/privacy"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

func privacyDomain() configDomain {
	return configDomain{name: privacy.Domain, object: rbac.PrivacyConfig, accept: func(_, after []byte) error {
		_, err := privacy.Parse(after)
		return err
	}}
}
