package config

// DefaultAtlasBaseURL is the default used when the ATLAS_BASE_URL variable is
// missing.
const DefaultAtlasBaseURL = "https://cloud.mongodb.com/api/atlas/v1.0"

// AtlasConfig contains the different configuration settings available for
// Atlas clients.
type AtlasConfig struct {
	BaseURL    string
	GroupID    string
	PublicKey  string
	PrivateKey string
}

// ParseAtlasConfig will parse environment variables generate an AtlasConfig.
func ParseAtlasConfig() (*AtlasConfig, error) {
	url := getVarWithDefault("ATLAS_BASE_URL", DefaultAtlasBaseURL)

	group, err := getVar("ATLAS_GROUP_ID")
	if err != nil {
		return nil, err
	}

	publicKey, err := getVar("ATLAS_PUBLIC_KEY")
	if err != nil {
		return nil, err
	}

	privateKey, err := getVar("ATLAS_PRIVATE_KEY")
	if err != nil {
		return nil, err
	}

	return &AtlasConfig{
		BaseURL:    url,
		GroupID:    group,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}
