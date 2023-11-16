package caddy_vault_storage

import (
	. "fmt"
	"github.com/mywordpress-io/certmagic-vault-storage"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/certmagic"
)

var (
	defaultApproleLoginPath  = "auth/approle/login"
	defaultApproleLogoutPath = "auth/token/revoke-self"

	defaultLockTimeout         = certmagic_vault_storage.Duration(5 * time.Minute)
	defaultLockPollingInterval = certmagic_vault_storage.Duration(5 * time.Second)
)

func init() {
	caddy.RegisterModule(Storage{})
}

type Storage struct {
	// URL the URL for Vault without any API versions or paths like 'https://vault.example.org:8201'.
	URL *certmagic_vault_storage.URL `json:"address"`

	// Token, the static Vault token.  If 'Token' is set, we blindly use that 'Token' when making any calls to
	// the Vault API. Management of the token (create, revoke, renew, etc.) is up to the caller.
	Token string `json:"token"`

	// If 'Approle*', options are available, we log in to Vault to create a short-lived token, using that token to make
	// future calls into Vault, and once we are done automatically revoke it.  Note that we will "cache" that token for
	// up to its lifetime minus 5m so it can be re-used for future calls in to Vault by subsequent CertMagic Storage
	// operations.
	//
	// Approle settings are the recommended way to manage Vault authentication
	ApproleLoginPath  string `json:"approle_login_path"`
	ApproleLogoutPath string `json:"approle_logout_path"`
	ApproleRoleId     string `json:"approle_role_id"`
	ApproleSecretId   string `json:"approle_secret_id"`

	// SecretsPath is the path in Vault to the secrets engine
	SecretsPath string `json:"secrets_path"`

	// PathPrefix is the path in the secrets engine where certificates will be placed (default: 'certificates'), assuming:
	//           URL: https://vault.example.org:8201
	//       SecretsPath: secrets/production
	//        PathPrefix: engineering/certmagic/certificates
	//
	// You will end up with paths like this in vault:
	//     'data' path: https://vault.example.org:8201/v1/secrets/production/data/engineering/certmagic/certificates
	// 'metadata' path: https://vault.example.org:8201/v1/secrets/production/metadata/engineering/certmagic/certificates
	PathPrefix string `json:"path_prefix"`

	// InsecureSkipVerify ignore TLS errors when communicating with vault - Default: false
	InsecureSkipVerify bool `json:"insecure_skip_verify"`

	// Locking mechanism
	LockTimeout         *certmagic_vault_storage.Duration `json:"lock_timeout"`
	LockPollingInterval *certmagic_vault_storage.Duration `json:"lock_polling_interval"`

	// logger Zap sugared logger
	logger *zap.SugaredLogger

	// CertMagic storage backend for Vault
	certmagicStorage *certmagic_vault_storage.Storage
}

// Provisions an instance of the storage provider in caddy
func (s *Storage) Provision(ctx caddy.Context) error {
	s.logger = ctx.Logger().Sugar()
	s.certmagicStorage = certmagic_vault_storage.NewStorage(s)
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
	return s.certmagicStorage, nil
}

// UnmarshalCaddyfile sets up the storage module from Caddyfile tokens. For syntax, review README.md
func (s *Storage) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		var err error
		if !d.NextArg() {
			return d.ArgErr()
		}
		s.URL, err = certmagic_vault_storage.ParseURL(d.Val())
		if err != nil {
			return err
		}

		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "token":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.Token = d.Val()
			case "approle_login_path":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.ApproleLoginPath = d.Val()
			case "approle_logout_path":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.ApproleLogoutPath = d.Val()
			case "approle_role_id":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.ApproleRoleId = d.Val()
			case "approle_secret_id":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.ApproleSecretId = d.Val()
			case "secrets_path":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.SecretsPath = d.Val()
			case "path_prefix":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.PathPrefix = d.Val()
			case "insecure_skip_verify":
				if !d.NextArg() {
					return d.ArgErr()
				}
				s.InsecureSkipVerify = d.ScalarVal().(bool)
			case "lock_timeout":
				if !d.NextArg() {
					return d.ArgErr()
				}
				val, err := time.ParseDuration(d.Val())
				if err != nil {
					return err
				}
				lockTimeout := certmagic_vault_storage.Duration(val)
				s.LockTimeout = &lockTimeout
			case "lock_polling_interval":
				if !d.NextArg() {
					return d.ArgErr()
				}
				val, err := time.ParseDuration(d.Val())
				if err != nil {
					return err
				}
				lockPollingInterval := certmagic_vault_storage.Duration(val)
				s.LockPollingInterval = &lockPollingInterval
			default:
				return d.Errf("unrecognized parameter '%s'", d.Val())
			}
		}
	}

	// Make sure 'secrets_path' is non-empty
	if s.SecretsPath == "" {
		return errors.New("secret_path is required")
	}

	// Make sure user has non-empty values for at least 'token' OR 'approle_role_id' / 'approle_secret_id'
	if s.Token == "" && (s.ApproleRoleId == "" || s.ApproleSecretId == "") {
		return errors.New("you must define 'token' or 'approle_role_id' + 'approle_secret_id' in order to authenticate with Vault")
	}

	return nil
}

func (s *Storage) SetLogger(logger *zap.SugaredLogger) *Storage {
	s.logger = logger
	return s
}

func (s *Storage) GetLogger() *zap.SugaredLogger {
	return s.logger
}

func (s *Storage) GetVaultBaseUrl() string {
	return Sprintf("%s/v1/", s.URL)
}

func (s *Storage) GetToken() string {
	return s.Token
}

func (s *Storage) GetApproleLoginPath() string {
	if s.ApproleLoginPath != "" {
		return s.ApproleLoginPath
	}

	return defaultApproleLoginPath
}

func (s *Storage) GetApproleLogoutPath() string {
	if s.ApproleLogoutPath != "" {
		return s.ApproleLogoutPath
	}

	return defaultApproleLogoutPath
}

func (s *Storage) GetApproleRoleId() string {
	return s.ApproleRoleId
}

func (s *Storage) GetApproleSecretId() string {
	return s.ApproleSecretId
}

func (s *Storage) GetSecretsPath() string {
	return s.SecretsPath
}

func (s *Storage) GetPathPrefix() string {
	return s.PathPrefix
}

func (s *Storage) GetInsecureSkipVerify() bool {
	return s.InsecureSkipVerify
}

func (s *Storage) GetLockTimeout() certmagic_vault_storage.Duration {
	if s.LockTimeout != nil {
		return *s.LockTimeout
	}

	return defaultLockTimeout
}

func (s *Storage) GetLockPollingInterval() certmagic_vault_storage.Duration {
	if s.LockPollingInterval != nil {
		return *s.LockPollingInterval
	}

	return defaultLockPollingInterval
}

// Interface guards
var (
	_ caddy.StorageConverter = (*Storage)(nil)
	_ caddyfile.Unmarshaler  = (*Storage)(nil)
	_ caddy.Provisioner      = (*Storage)(nil)
)
