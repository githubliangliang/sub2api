//go:build unit

package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type backupConfigFailEncryptor struct{ plainEncryptor }

func (backupConfigFailEncryptor) Encrypt(string) (string, error) {
	return "", errors.New("encryption fixture failure")
}

type backupConfigFailWriter struct {
	SettingRepository
	attempted string
}

func (r *backupConfigFailWriter) Set(_ context.Context, _ string, value string) error {
	r.attempted = value
	return errors.New("write fixture failure")
}

func TestBackupS3ConfigRetainedSecretBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, stored                  string
		fixed, encryptFail, writeFail bool
	}{
		{"ciphertext", "ENC:fixture-secret", true, false, false},
		{"legacy plaintext", "fixture-secret", true, false, false},
		{"retained ciphertext requires fixed key", "ENC:fixture-secret", false, false, false},
		{"retained plaintext requires fixed key", "fixture-secret", false, false, false},
		{"retained encryption failure", "ENC:fixture-secret", true, true, false},
		{"retained write failure", "ENC:fixture-secret", true, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			repo := newMockSettingRepo()
			original, err := json.Marshal(BackupS3Config{Bucket: "old-bucket", SecretAccessKey: tc.stored})
			require.NoError(t, err)
			require.NoError(t, repo.Set(ctx, settingKeyBackupS3Config, string(original)))
			var encryptor SecretEncryptor = &plainEncryptor{}
			if tc.encryptFail {
				encryptor = &backupConfigFailEncryptor{}
			}
			var writer SettingRepository = repo
			failWriter := &backupConfigFailWriter{SettingRepository: repo}
			if tc.writeFail {
				writer = failWriter
			}
			storeCalls := 0
			svc := NewBackupService(writer, &config.Config{Totp: config.TotpConfig{EncryptionKeyConfigured: tc.fixed}}, encryptor, func(context.Context, *BackupS3Config) (BackupObjectStore, error) {
				storeCalls++
				return nil, errors.New("must not create a store")
			}, nil)
			t.Cleanup(svc.Stop)
			result, err := svc.UpdateS3Config(ctx, BackupS3Config{Bucket: "new-bucket"})
			stored, readErr := repo.GetValue(ctx, settingKeyBackupS3Config)
			require.NoError(t, readErr)
			if !tc.fixed || tc.encryptFail || tc.writeFail {
				require.Error(t, err)
				require.Nil(t, result)
				require.Equal(t, string(original), stored)
				if !tc.fixed {
					require.ErrorIs(t, err, ErrSecretEncryptionKeyNotConfigured)
				}
				if tc.writeFail {
					var attempted BackupS3Config
					require.NoError(t, json.Unmarshal([]byte(failWriter.attempted), &attempted))
					require.Equal(t, "ENC:fixture-secret", attempted.SecretAccessKey)
				}
			} else {
				require.NoError(t, err)
				require.Empty(t, result.SecretAccessKey)
				for i := 0; i < 2; i++ {
					var persisted BackupS3Config
					require.NoError(t, json.Unmarshal([]byte(stored), &persisted))
					require.Equal(t, "ENC:fixture-secret", persisted.SecretAccessKey)
					loaded, loadErr := svc.loadS3Config(ctx)
					require.NoError(t, loadErr)
					require.Equal(t, "fixture-secret", loaded.SecretAccessKey)
					redacted, getErr := svc.GetS3Config(ctx)
					require.NoError(t, getErr)
					require.Empty(t, redacted.SecretAccessKey)
					_, saveErr := svc.UpdateS3Config(ctx, BackupS3Config{Bucket: "third-bucket"})
					require.NoError(t, saveErr)
					stored, readErr = repo.GetValue(ctx, settingKeyBackupS3Config)
					require.NoError(t, readErr)
				}
			}
			require.Zero(t, storeCalls, "config save must not create an object store")
		})
	}
}
