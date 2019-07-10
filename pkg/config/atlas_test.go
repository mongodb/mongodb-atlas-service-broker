package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	BaseURL    = "BASE"
	GroupID    = "GROUP"
	PublicKey  = "PUB"
	PrivateKey = "PRIV"
)

func TestAtlasParse(t *testing.T) {
	os.Clearenv()
	os.Setenv("ATLAS_BASE_URL", BaseURL)
	os.Setenv("ATLAS_GROUP_ID", GroupID)
	os.Setenv("ATLAS_PUBLIC_KEY", PublicKey)
	os.Setenv("ATLAS_PRIVATE_KEY", PrivateKey)

	config, err := ParseAtlasConfig()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, BaseURL, config.BaseURL)
	assert.Equal(t, GroupID, config.GroupID)
	assert.Equal(t, PublicKey, config.PublicKey)
	assert.Equal(t, PrivateKey, config.PrivateKey)
}

func TestAtlasDefaults(t *testing.T) {
	os.Clearenv()
	os.Setenv("ATLAS_GROUP_ID", GroupID)
	os.Setenv("ATLAS_PUBLIC_KEY", PublicKey)
	os.Setenv("ATLAS_PRIVATE_KEY", PrivateKey)

	config, err := ParseAtlasConfig()
	if !assert.NoError(t, err) {
		return
	}

	// Ensure the base URL was populated by the default value.
	assert.Equal(t, DefaultAtlasBaseURL, config.BaseURL)
}

func TestMissingAtlasGroupID(t *testing.T) {
	os.Clearenv()
	os.Setenv("ATLAS_PUBLIC_KEY", PublicKey)
	os.Setenv("ATLAS_PRIVATE_KEY", PrivateKey)

	_, err := ParseAtlasConfig()
	assert.Error(t, err, "Expected missing group ID error")
}

func TestMissingAtlasPubKey(t *testing.T) {
	os.Clearenv()
	os.Setenv("ATLAS_GROUP_ID", GroupID)
	os.Setenv("ATLAS_PRIVATE_KEY", PrivateKey)

	_, err := ParseAtlasConfig()
	assert.Error(t, err, "Expected missing public key error")
}

func TestMissingAtlasPrivKey(t *testing.T) {
	os.Clearenv()
	os.Setenv("ATLAS_GROUP_ID", GroupID)
	os.Setenv("ATLAS_PUBLIC_KEY", PublicKey)

	_, err := ParseAtlasConfig()
	assert.Error(t, err, "Expected missing private key error")
}
