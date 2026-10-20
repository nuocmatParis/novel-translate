package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
	"github.com/voocel/ainovel-cli/internal/translation/state"
	"github.com/voocel/ainovel-cli/internal/translation/workspace"
)

func providerPath(c workspace.Config) string {
	if c.ProviderConfig != "" {
		return c.ReadPath(c.ProviderConfig)
	}
	return bootstrap.DefaultConfigPath()
}

func (m *translationTUI) saveModelSetup(values []string) func(context.Context, io.Writer, io.Writer) error {
	root := m.config.Root
	return func(ctx context.Context, out, progress io.Writer) error {
		unlock, err := (state.Store{Root: root}).Lock(ctx)
		if err != nil {
			return err
		}
		defer unlock()
		c, err := workspace.Load(root)
		if err != nil {
			return err
		}
		protocol, modelName, baseURL := strings.TrimSpace(values[0]), strings.TrimSpace(values[1]), strings.TrimSpace(values[2])
		switch protocol {
		case "openai", "gemini", "anthropic", "openrouter", "ollama":
		default:
			return errors.New("giao thức: openai, gemini, anthropic, openrouter hoặc ollama")
		}
		if modelName == "" {
			return errors.New("cần tên model")
		}
		window, err := strconv.Atoi(strings.TrimSpace(values[5]))
		if err != nil || window < 0 {
			return errors.New("context tokens phải là số >= 0")
		}
		keyName := "NOVEL_TRANSLATE_API_KEY"
		old, _ := bootstrap.ReadConfigFile(providerPath(c))
		pc := old.DefaultProviderConfig()
		if pc.APIKeyEnv != "" {
			keyName = pc.APIKeyEnv
		}
		key := strings.TrimSpace(values[3])
		if key == "" {
			key = pc.APIKey
		}
		if key == "" {
			key, err = bootstrap.LookupAPIKey(keyName, root, filepath.Dir(providerPath(c)))
			if err != nil {
				return err
			}
		}
		if key == "" && protocol != "ollama" {
			return errors.New("nhập API key hoặc lưu key vào .env trước")
		}
		pc = bootstrap.ProviderConfig{Type: protocol, BaseURL: baseURL, Models: []bootstrap.ModelConfig{{Name: modelName, ContextWindow: window}}}
		if protocol == "openai" {
			pc.API = strings.TrimSpace(values[4])
		}
		if key != "" {
			pc.APIKeyEnv = keyName
		}
		cfg := bootstrap.Config{Provider: "translation-service", ModelName: modelName, Providers: map[string]bootstrap.ProviderConfig{"translation-service": pc}, Budget: old.Budget}
		// Validate using a private copy, never persist the resolved key in JSON.
		validated := bootstrap.CloneConfig(cfg)
		target := validated.Providers[cfg.Provider]
		target.APIKey = key
		validated.Providers[cfg.Provider] = target
		if err := validated.ValidateBase(); err != nil {
			return err
		}
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}
		configPath, err := c.WritePath(filepath.Join(".translation_work", "provider.json"))
		if err != nil {
			return err
		}
		envPath, err := c.WritePath(".env")
		if err != nil {
			return err
		}
		writes := []localWrite{}
		if key != "" {
			values, err := bootstrap.ReadDotEnv(envPath)
			if err != nil {
				return err
			}
			values[keyName] = key
			text, err := godotenv.Marshal(values)
			if err != nil {
				return errors.New("không thể tạo .env")
			}
			writes = append(writes, localWrite{path: envPath, data: []byte(text + "\n")})
		}
		writes = append(writes, localWrite{path: configPath, data: append(data, '\n')})
		c.ProviderConfig = filepath.ToSlash(filepath.Join(".translation_work", "provider.json"))
		if err := c.Validate(); err != nil {
			return err
		}
		if err := saveLocalFiles(writes, c.Save); err != nil {
			return err
		}
		fmt.Fprintln(out, "Đã lưu model trong workspace; lần mở sau tự đọc key từ .env.")
		if os.Getenv(keyName) != "" {
			fmt.Fprintln(out, "Terminal đang đặt key cùng tên và được ưu tiên hơn .env.")
		}
		return nil
	}
}

type localWrite struct {
	path string
	data []byte
}

// Roll back local setup files if updating workspace.yaml fails midway.
func saveLocalFiles(writes []localWrite, commit func() error) error {
	backups := make([]localWrite, 0, len(writes))
	existed := map[string]bool{}
	for _, w := range writes {
		data, err := os.ReadFile(w.path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		existed[w.path] = err == nil
		backups = append(backups, localWrite{w.path, data})
	}
	applied := 0
	var err error
	for _, w := range writes {
		if err = state.Atomic(w.path, w.data); err != nil {
			break
		}
		applied++
	}
	if err == nil {
		err = commit()
	}
	if err == nil {
		return nil
	}
	for i := applied - 1; i >= 0; i-- {
		var rollback error
		if existed[backups[i].path] {
			rollback = state.Atomic(backups[i].path, backups[i].data)
		} else {
			rollback = os.Remove(backups[i].path)
		}
		err = errors.Join(err, rollback)
	}
	return err
}

func (m *translationTUI) saveRunSettings(values []string) func(context.Context, io.Writer, io.Writer) error {
	root := m.config.Root
	return func(ctx context.Context, out, progress io.Writer) error {
		unlock, err := (state.Store{Root: root}).Lock(ctx)
		if err != nil {
			return err
		}
		defer unlock()
		c, err := workspace.Load(root)
		if err != nil {
			return err
		}
		c.Run.MaxRequests, err = strconv.Atoi(values[0])
		if err != nil {
			return errors.New("số requests phải là số nguyên")
		}
		c.Run.RequestTimeoutSeconds, err = strconv.Atoi(values[1])
		if err != nil {
			return errors.New("timeout phải là số nguyên")
		}
		c.Output.EPUBCheck, c.Output.FontDirectory = strings.Trim(values[2], "\"'"), strings.Trim(values[3], "\"'")
		return c.Save()
	}
}
