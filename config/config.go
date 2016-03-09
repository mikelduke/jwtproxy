// Copyright 2015 CoreOS, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"os"
	"time"
)

// URL is a custom URL type that allows validation at configuration load time.
type URL struct {
	*url.URL
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for URLs.
func (u *URL) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	urlp, err := url.Parse(s)
	if err != nil {
		return err
	}
	u.URL = urlp
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface for URLs.
func (u URL) MarshalYAML() (interface{}, error) {
	if u.URL != nil {
		return u.String(), nil
	}
	return nil, nil
}

// Config is the global configuration
type configFile struct {
	HmacProxy *Config
}

type Config struct {
	Signer   *SignerConfig
	Verifier *VerifierConfig
}

// Configuration used to enable and configure the signing half of the proxy
type SignerConfig struct {
	ListenerAddr string
	Key          *HMACKey
}

type HMACKey struct {
	ID      string
	Secret  string
	Region  string
	Service string
}

// Configuration used to enable and configure the verifier half of the proxy
type VerifierConfig struct {
	ListenerAddr     string
	Upstream         URL
	MaxClockSkew     time.Duration
	TLS              *TLSConfig
	CredentialSource *CredentialSourceConfig
}

// Configuration which when specified enables TLS(SSL), and optionally requires
// the use of client certificates
type TLSConfig struct {
	CertFile                 string
	KeyFile                  string
	CAFile                   string
	RequireClientCertificate string
}

// Configuration options for a verifier credential source
type CredentialSourceConfig struct {
	Type    string
	Options map[string]interface{} `yaml:",inline"`
}

// DefaultConfig is a configuration that can be used as a fallback value.
var DefaultConfig = configFile{
	HmacProxy: &Config{
		Signer: &SignerConfig{
			ListenerAddr: ":8080",
		},
		Verifier: &VerifierConfig{
			MaxClockSkew: 1 * time.Minute,
			ListenerAddr: ":8081",
		},
	},
}

// Load is a shortcut to open a file, read it, and generate a Config.
// It supports relative and absolute paths.
func Load(path string) (config *Config, err error) {
	cFile := &DefaultConfig
	if path == "" {
		err = fmt.Errorf("A configuration file is required")
		return
	}

	f, err := os.Open(os.ExpandEnv(path))
	if err != nil {
		return
	}
	defer f.Close()

	d, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(d, cFile)
	if err != nil {
		return
	}

	config = cFile.HmacProxy
	return
}
