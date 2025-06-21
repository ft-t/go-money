package main

import (
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/rs/zerolog/log"
)

func main() {
	keyGenerator := auth.NewKeyGenerator()

	key := keyGenerator.Generate()

	log.Info().Msg("New key generated")
	log.Info().Msgf("%s", keyGenerator.Serialize(key))

	if err := keyGenerator.Save(key, "private.key"); err != nil {
		panic(err)
	}
}
