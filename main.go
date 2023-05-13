package caddy_vault_storage

import (
	certmagicVaultStorage "github.com/mywordpress-io/certmagic-vault-storage"
	"github.com/pkg/errors"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/certmagic"
)

func init() {
	caddy.RegisterModule(Storage{})
}

type Storage struct {
	*certmagicVaultStorage.Storage
}

// CaddyModule returns the Caddy module information.
func (s Storage) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "caddy.storage.vault",
		New: func() caddy.Module {
			storage := new(Storage)
			storage.Storage = certmagicVaultStorage.NewStorage(certmagicVaultStorage.StorageConfig{})
			return storage
		},
	}
}

// CertMagicStorage converts s to a certmagic.Storage instance.
func (s *Storage) CertMagicStorage() (certmagic.Storage, error) {
	return s, nil
}

// UnmarshalCaddyfile sets up the storage module from Caddyfile tokens. For syntax, review README.md
func (s *Storage) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	cfg := certmagicVaultStorage.StorageConfig{}
	for d.Next() {
		var err error
		if !d.NextArg() {
			return d.ArgErr()
		}
		cfg.URL, err = certmagicVaultStorage.ParseURL(d.Val())
		if err != nil {
			return err
		}

		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "token":
				if !d.NextArg() {
					return d.ArgErr()
				}
				cfg.Token = d.Val()
			case "approle_login_path":
				if !d.NextArg() {
					return d.ArgErr()
				}
				cfg.ApproleLoginPath = d.Val()
			case "approle_logout_path":
				if !d.NextArg() {
					return d.ArgErr()
				}
				cfg.ApproleLogoutPath = d.Val()
			case "approle_role_id":
				if !d.NextArg() {
					return d.ArgErr()
				}
				cfg.ApproleRoleId = d.Val()
			case "approle_secret_id":
				if !d.NextArg() {
					return d.ArgErr()
				}
				cfg.ApproleSecretId = d.Val()
			case "secrets_path":
				if !d.NextArg() {
					return d.ArgErr()
				}
				cfg.SecretsPath = d.Val()
			case "path_prefix":
				if !d.NextArg() {
					return d.ArgErr()
				}
				cfg.PathPrefix = d.Val()
			case "insecure_skip_verify":
				if !d.NextArg() {
					return d.ArgErr()
				}
				cfg.InsecureSkipVerify = d.ScalarVal().(bool)
			case "lock_timeout":
				if !d.NextArg() {
					return d.ArgErr()
				}
				val, err := time.ParseDuration(d.Val())
				if err != nil {
					return err
				}
				lockTimeout := certmagicVaultStorage.Duration(val)
				cfg.LockTimeout = &lockTimeout
			case "lock_polling_interval":
				if !d.NextArg() {
					return d.ArgErr()
				}
				val, err := time.ParseDuration(d.Val())
				if err != nil {
					return err
				}
				lockPollingInterval := certmagicVaultStorage.Duration(val)
				cfg.LockPollingInterval = &lockPollingInterval
			case "log_level":
				if !d.NextArg() {
					return d.ArgErr()
				}
				cfg.LogLevel = d.Val()
			default:
				return d.Errf("unrecognized parameter '%s'", d.Val())
			}
		}
	}

	// Make sure 'secrets_path' is non-empty
	if cfg.SecretsPath == "" {
		return errors.New("secret_path is required")
	}

	// Make sure user has non-empty values for at least 'token' OR 'approle_role_id' / 'approle_secret_id'
	if cfg.Token == "" && (cfg.ApproleRoleId == "" || cfg.ApproleSecretId == "") {
		return errors.New("you must define 'token' or 'approle_role_id' + 'approle_secret_id' in order to authenticate with Vault")
	}

	// Initialize Storage
	s.Storage = certmagicVaultStorage.NewStorage(cfg)

	return nil
}

// Interface guards
var (
	_ caddy.StorageConverter = (*Storage)(nil)
	_ caddyfile.Unmarshaler  = (*Storage)(nil)
)
