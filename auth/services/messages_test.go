package services

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNutsAccessToken_FromMap(t *testing.T) {
	expected := NutsAccessToken{Name: "Foobar"}
	asJSON, _ := json.Marshal(&expected)
	var asMap map[string]interface{}
	err := json.Unmarshal(asJSON, &asMap)
	if !assert.NoError(t, err) {
		return
	}
	var actual NutsAccessToken
	err = actual.FromMap(asMap)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestNutsJwtBearerToken_FromMap(t *testing.T) {
	expected := NutsJwtBearerToken{KeyID: "kid"}
	m, _ := expected.AsMap()
	var actual NutsJwtBearerToken
	err := actual.FromMap(m)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
