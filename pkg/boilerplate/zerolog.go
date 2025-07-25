package boilerplate

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"strings"
)

var (
	version   = ""
	commitSHA = ""
)

func GetVersion() string {
	return version
}

func GetCommit() string {
	return commitSHA
}

func SetupZeroLog() {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		sp := strings.Split(file, "/")

		segments := 4

		if len(sp) == 0 { // just in case
			segments = 0
		}

		if segments > 0 && segments > len(sp) {
			segments = len(sp) - 1
		}

		var final strings.Builder

		for _, ss := range sp[segments:] {
			if strings.Contains(ss, "@") { // git repos with version
				continue
			}

			final.WriteString("/")
			final.WriteString(ss)
		}

		return fmt.Sprintf("%s:%v", final.String(), line)
	}

	log.Logger = log.Logger.With().
		Str("version", version).
		Str("commit_sha", commitSHA).
		CallerWithSkipFrameCount(2).Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.DefaultContextLogger = &log.Logger
}
