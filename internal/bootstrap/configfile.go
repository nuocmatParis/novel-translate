package bootstrap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DefaultConfigPath is a read-only credential fallback. Workspace configuration
// takes precedence and never depends on the command's current directory.
func DefaultConfigPath() string {
	home, e := os.UserHomeDir()
	if e != nil {
		return ""
	}
	return filepath.Join(home, ".novel-translate", "config.json")
}
func LoadConfigFile(path string) (Config, error) {
	var c Config
	b, e := os.ReadFile(path)
	if e != nil {
		return c, e
	}
	d := json.NewDecoder(bytes.NewReader(stripJSONComments(b)))
	d.DisallowUnknownFields()
	if e = d.Decode(&c); e != nil {
		return c, fmt.Errorf("provider config: %w", e)
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		return c, fmt.Errorf("provider config cần một JSON object")
	}
	for name, p := range c.Providers {
		if p.APIKeyEnv != "" {
			if p.APIKey != "" {
				return c, fmt.Errorf("provider %s chỉ dùng api_key hoặc api_key_env", name)
			}
			p.APIKey = os.Getenv(p.APIKeyEnv)
			if p.APIKey == "" {
				return c, fmt.Errorf("thiếu biến môi trường %s cho provider %s", p.APIKeyEnv, name)
			}
			c.Providers[name] = p
		}
	}
	return c, nil
}
func cloneMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	copy := map[string]any{}
	for k, v := range m {
		copy[k] = v
	}
	return copy
}
func CloneConfig(c Config) Config {
	copy := c
	copy.Providers = map[string]ProviderConfig{}
	for k, p := range c.Providers {
		p.Models = append([]ModelConfig(nil), p.Models...)
		p.Extra = cloneMap(p.Extra)
		p.ExtraBody = cloneMap(p.ExtraBody)
		copy.Providers[k] = p
	}
	copy.Roles = map[string]RoleConfig{}
	for k, r := range c.Roles {
		r.Fallbacks = append([]ModelRef(nil), r.Fallbacks...)
		copy.Roles[k] = r
	}
	return copy
}
func stripJSONComments(data []byte) []byte {
	out := []byte{}
	quoted, escape := false, false
	for i := 0; i < len(data); i++ {
		b := data[i]
		if escape {
			out = append(out, b)
			escape = false
			continue
		}
		if quoted {
			out = append(out, b)
			if b == '\\' {
				escape = true
			} else if b == '"' {
				quoted = false
			}
			continue
		}
		if b == '"' {
			quoted = true
			out = append(out, b)
			continue
		}
		if b == '/' && i+1 < len(data) && data[i+1] == '/' {
			for i < len(data) && data[i] != '\n' {
				i++
			}
			if i < len(data) {
				out = append(out, '\n')
			}
			continue
		}
		out = append(out, b)
	}
	return out
}
