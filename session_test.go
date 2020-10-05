package vrm_test

import (
	"github.com/rs/zerolog/log"

	"victron/vrm"
)

func ExampleLogin() {
	session, err := vrm.Login("username", "password", vrm.WithSMSToken("123456"))
	if err != nil {
		log.Error().Err(err).Msg("login failed")
	}

	_ = session.Logout()
}
