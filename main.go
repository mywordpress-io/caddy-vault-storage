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
	certmagicVaultStorage.StorageConfig
}

// Provisions an instance of the storage provider in caddy
func (s *Storage) Provision(ctx caddy.Context) error {
	s.Storage = certmagicVaultStorage.NewStorage(s.StorageConfig)
	s.Storage.SetLogger(ctx.Logger(s).Sugar())
	return nil
}

// CaddyModule returns the Caddy module information.
func (s Storage) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "caddy.storage.vault",
		New: func() caddy.Module {
			return new(Storage)
		},
	}
}

// CertMagicStorage converts s to a certmagic.Storage instance.
func (s *Storage) CertMagicStorage() (certmagic.Storage, error) {
	return s, nil
}

// UnmarshalCaddyfile sets up the storage module from Caddyfile tokens. For syntax, review README.md
func (s *Storage) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		var err error
		if !d.NextArg() {
			return d.ArgErr()
		}
		s.StorageConfig.URL, err = certmagicVaultStorage.ParseURL(d.Val())
		if err != nil {
			return err
		}

		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "token":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.StorageConfig.Token = d.Val()
			case "approle_login_path":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.StorageConfig.ApproleLoginPath = d.Val()
			case "approle_logout_path":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.StorageConfig.ApproleLogoutPath = d.Val()
			case "approle_role_id":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.StorageConfig.ApproleRoleId = d.Val()
			case "approle_secret_id":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.StorageConfig.ApproleSecretId = d.Val()
			case "secrets_path":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.StorageConfig.SecretsPath = d.Val()
			case "path_prefix":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.StorageConfig.PathPrefix = d.Val()
			case "insecure_skip_verify":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.StorageConfig.InsecureSkipVerify = d.ScalarVal().(bool)
			case "lock_timeout":
				if !d.NextArg() {
					return d.ArgErr()
				}
				val, err := time.ParseDuration(d.Val())
				if err != nil {
					return err
				}
				lockTimeout := certmagicVaultStorage.Duration(val)
				s.StorageConfig.LockTimeout = &lockTimeout
			case "lock_polling_interval":
				if !d.NextArg() {
					return d.ArgErr()
				}
				val, err := time.ParseDuration(d.Val())
				if err != nil {
					return err
				}
				lockPollingInterval := certmagicVaultStorage.Duration(val)
				s.StorageConfig.LockPollingInterval = &lockPollingInterval
			default:
				return d.Errf("unrecognized parameter '%s'", d.Val())
			}
		}
	}

	// Make sure 'secrets_path' is non-empty
	if s.StorageConfig.SecretsPath == "" {
		return errors.New("secret_path is required")
	}

	// Make sure user has non-empty values for at least 'token' OR 'approle_role_id' / 'approle_secret_id'
	if s.StorageConfig.Token == "" && (s.StorageConfig.ApproleRoleId == "" || s.StorageConfig.ApproleSecretId == "") {
		return errors.New("you must define 'token' or 'approle_role_id' + 'approle_secret_id' in order to authenticate with Vault")
	}

	return nil
}

// Interface guards
var (
	_ caddy.StorageConverter = (*Storage)(nil)
	_ caddyfile.Unmarshaler  = (*Storage)(nil)
	_ caddy.Provisioner      = (*Storage)(nil)
)
