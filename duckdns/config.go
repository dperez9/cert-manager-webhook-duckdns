package duckdns

import (
	"encoding/json"
	"fmt"

	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// Config holds configuration provided in the Issuer resource.
type Config struct {
	APITokenSecretRef cmmeta.SecretKeySelector `json:"apiTokenSecretRef"`
}

// loadConfig decodes the config JSON into a typed struct.
func loadConfig(cfgJSON *extapi.JSON) (Config, error) {
	cfg := Config{}

	if cfgJSON == nil {
		return cfg, nil
	}

	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}
