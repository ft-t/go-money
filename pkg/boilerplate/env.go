package boilerplate

import (
	"os"
	"strings"
)

type Environment int32

const (
	Dev  Environment = 0
	Prod Environment = 1
	Ci   Environment = 2
)

func GetCurrentEnvironment() Environment {
	val := os.Getenv("ENVIRONMENT")
	
	switch strings.ToLower(val) {
	case "prod":
		return Prod
	case "ci":
		return Ci
	default:
		return Dev
	}
}
