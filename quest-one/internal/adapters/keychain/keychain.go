// Package keychain provides access to OS-level credential storage.
// On macOS, it uses the security CLI; on other platforms it falls back
// to environment variables.
package keychain

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// SecretStore provides read/write access to OS-level credential storage.
type SecretStore interface {
	Get(service, account string) (string, error)
	Set(service, account, secret string) error
	Delete(service, account string) error
}

// New returns a SecretStore backed by the OS keychain on macOS,
// or an env-var reader on other platforms.
func New() SecretStore {
	if runtime.GOOS == "darwin" {
		return &darwinKeychain{}
	}
	return &envKeychain{}
}

// darwinKeychain uses the macOS `security` command-line tool.
type darwinKeychain struct{}

func (d *darwinKeychain) Get(service, account string) (string, error) {
	out, err := exec.Command(
		"security", "find-generic-password",
		"-s", service, "-a", account, "-w",
	).Output()
	if err != nil {
		return "", fmt.Errorf("keychain: get %s/%s: %w", service, account, err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func (d *darwinKeychain) Set(service, account, secret string) error {
	// Delete first (ignore error), then add fresh.
	_ = exec.Command(
		"security", "delete-generic-password",
		"-s", service, "-a", account,
	).Run()
	err := exec.Command(
		"security", "add-generic-password",
		"-s", service, "-a", account, "-w", secret,
	).Run()
	if err != nil {
		return fmt.Errorf("keychain: set %s/%s: %w", service, account, err)
	}
	return nil
}

func (d *darwinKeychain) Delete(service, account string) error {
	if err := exec.Command(
		"security", "delete-generic-password",
		"-s", service, "-a", account,
	).Run(); err != nil {
		return fmt.Errorf("keychain: delete %s/%s: %w", service, account, err)
	}
	return nil
}

// envKeychain reads secrets from environment variables.
// Convention: QUEST_ONE_<ACCOUNT_UPPER> where '/' and '-' are replaced with '_'.
type envKeychain struct{}

func envVarName(account string) string {
	upper := strings.ToUpper(account)
	upper = strings.ReplaceAll(upper, "-", "_")
	upper = strings.ReplaceAll(upper, "/", "_")
	upper = strings.ReplaceAll(upper, ".", "_")
	return "QUEST_ONE_" + upper
}

func (e *envKeychain) Get(_, account string) (string, error) {
	key := envVarName(account)
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("secret not found: set env var %s", key)
	}
	return v, nil
}

func (e *envKeychain) Set(_, _, _ string) error {
	return fmt.Errorf("env keychain is read-only; set the environment variable manually")
}

func (e *envKeychain) Delete(_, _ string) error {
	return fmt.Errorf("env keychain is read-only; unset the environment variable manually")
}
