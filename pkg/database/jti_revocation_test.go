package database_test

import (
	"testing"

	"github.com/ft-t/go-money/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestJtiRevocation_TableName(t *testing.T) {
	jti := database.JtiRevocation{}
	assert.Equal(t, "jti_revocations", jti.TableName())
}
