package caddy_vault_storage_test

import (
	"context"
	. "fmt"
	"github.com/caddyserver/certmagic"
	"github.com/mywordpress-io/caddy-vault-storage"
	"github.com/mywordpress-io/certmagic-vault-storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	t = &testing.T{}

	storage certmagic.Storage
	keys    = []string{
		"foo.bar.com",
		"foo.bar.baz",
		"production/test1.baz.com",
		"production/test2.baz.com",
		"production/test3.baz.com",
		"staging/abc123/test3.whatever.com",
		"staging/abc456/test1.whatever.com",
		"staging/abc456/test3.whatever.com",
		"staging/test3.baz.com",
		"staging/test3.quux.org",
	}

	approleRoleId   = os.Getenv("VAULT_APPROLE_ROLE_ID")
	approleSecretId = os.Getenv("VAULT_APPROLE_SECRET_ID")
)

func TestVaultStorageSuite(test *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(test, "Storage")
}

var _ = BeforeSuite(func() {
	customLockTimeout := certmagic_vault_storage.Duration(15 * time.Second)
	customLockPollingDuration := certmagic_vault_storage.Duration(5 * time.Second)
	caddyVaultStorage := &caddy_vault_storage.Storage{
		URL:                 certmagic_vault_storage.MustParseURL("http://localhost:8200"),
		Token:               "dead-beef",
		SecretsPath:         "secrets",
		PathPrefix:          "certificates",
		LockTimeout:         &customLockTimeout,
		LockPollingInterval: &customLockPollingDuration,
		InsecureSkipVerify:  false,
	}
	Expect(caddyVaultStorage).ShouldNot(BeNil())
	caddyVaultStorage.SetLogger(SetupLogger(zap.NewAtomicLevelAt(zap.InfoLevel), "caddy-vault-storage"))

	storage = certmagic_vault_storage.NewStorage(caddyVaultStorage)
	for _, key := range keys {
		err := storage.Store(context.Background(), key, []byte(Sprintf("This is some long text we want to store for '%s'", key)))
		Expect(err).ShouldNot(HaveOccurred())
	}
})

func SetupLogger(level zap.AtomicLevel, plugin string) *zap.SugaredLogger {
	return zap.Must(zap.Config{
		Level:            level,
		Encoding:         "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    "function",
			MessageKey:     "message",
			StacktraceKey:  zapcore.OmitKey,
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		InitialFields: map[string]interface{}{
			"storage_plugin": plugin,
		},
	}.Build()).Sugar()
}

var _ = AfterSuite(func() {
	// Delete test keys, ignoring errors
	for _, key := range keys {
		storage.Delete(context.Background(), key) //nolint:errcheck
	}
})
