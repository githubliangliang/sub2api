//go:build unit

package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"io"
	"mime/quotedprintable"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type resetTokenCacheStub struct {
	emailCacheStub
	stored       *PasswordResetTokenData
	consumedHash string
}

func (s *resetTokenCacheStub) GetPasswordResetToken(context.Context, string) (*PasswordResetTokenData, error) {
	return s.stored, nil
}

func (s *resetTokenCacheStub) ConsumePasswordResetToken(_ context.Context, _ string, tokenHash string) (bool, error) {
	s.consumedHash = tokenHash
	if s.stored == nil || s.stored.Token != tokenHash {
		return false, nil
	}
	s.stored = nil
	return true, nil
}

func TestConsumePasswordResetToken_ComparesHashNotPlaintext(t *testing.T) {
	token := "deadbeef"
	sum := sha256.Sum256([]byte(token))
	hash := hex.EncodeToString(sum[:])
	require.Equal(t, hash, hashPasswordResetToken(token))
	require.NotEqual(t, token, hashPasswordResetToken(token))

	cache := &resetTokenCacheStub{stored: &PasswordResetTokenData{Token: hash}}
	svc := NewEmailService(nil, cache)

	require.NoError(t, svc.ConsumePasswordResetToken(context.Background(), "a@b.c", token))
	require.Equal(t, hash, cache.consumedHash)
	require.ErrorIs(t, svc.ConsumePasswordResetToken(context.Background(), "a@b.c", token), ErrInvalidResetToken)

	// A legacy plaintext value (issued before upgrade) no longer validates.
	cache.stored = &PasswordResetTokenData{Token: token}
	require.ErrorIs(t, svc.ConsumePasswordResetToken(context.Background(), "a@b.c", token), ErrInvalidResetToken)
}

func (s *resetTokenCacheStub) SetPasswordResetToken(_ context.Context, _ string, data *PasswordResetTokenData, _ time.Duration) error {
	copied := *data
	s.stored = &copied
	return nil
}

func TestSendPasswordResetEmail_HashesActualLinkAndInvalidatesPreviousLink(t *testing.T) {
	smtp, port := startFakeSMTPServer(t, false, false)
	cache := &resetTokenCacheStub{}
	svc := NewEmailService(&settingRepoStub{values: map[string]string{
		SettingKeySMTPHost: "127.0.0.1", SettingKeySMTPPort: strconv.Itoa(port), SettingKeySMTPFrom: "noreply@example.com",
	}}, cache)
	ctx := context.Background()
	var previous string
	for i := 0; i < 2; i++ {
		require.NoError(t, svc.SendPasswordResetEmail(ctx, "reset@example.com", "Test", "https://example.com/reset"))
		smtp.mu.Lock()
		raw := smtp.messages[len(smtp.messages)-1]
		smtp.mu.Unlock()
		message, err := mail.ReadMessage(strings.NewReader(raw))
		require.NoError(t, err)
		content, err := io.ReadAll(quotedprintable.NewReader(message.Body))
		require.NoError(t, err)
		match := regexp.MustCompile(`token=([a-f0-9]{64})`).FindSubmatch(content)
		require.Len(t, match, 2, string(content))
		token := string(match[1])
		require.NotEqual(t, token, cache.stored.Token)
		digest := sha256.Sum256([]byte(token))
		require.Equal(t, hex.EncodeToString(digest[:]), cache.stored.Token)
		require.NoError(t, svc.VerifyPasswordResetToken(ctx, "reset@example.com", token))
		if previous != "" {
			require.NotEqual(t, previous, token)
			require.ErrorIs(t, svc.VerifyPasswordResetToken(ctx, "reset@example.com", previous), ErrInvalidResetToken)
		}
		previous = token
	}
	require.NoError(t, svc.ConsumePasswordResetToken(ctx, "reset@example.com", previous))
	require.ErrorIs(t, svc.ConsumePasswordResetToken(ctx, "reset@example.com", previous), ErrInvalidResetToken)
}

func TestAuthService_ResetPasswordConsumesTokenBeforeUpdatingPassword(t *testing.T) {
	ctx := context.Background()
	token := "reset-token"
	sum := sha256.Sum256([]byte(token))
	cache := &resetTokenCacheStub{stored: &PasswordResetTokenData{Token: hex.EncodeToString(sum[:])}}
	repo := &userRepoStub{user: &User{ID: 1, Email: "reset@example.com", Status: StatusActive}}
	auth := newAuthService(repo, map[string]string{SettingKeyEmailVerifyEnabled: "true", SettingKeyPasswordResetEnabled: "true"}, nil, nil)
	auth.emailService = NewEmailService(nil, cache)
	require.NoError(t, auth.ResetPassword(ctx, "reset@example.com", token, "new-password"))
	require.Len(t, repo.updated, 1)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(repo.updated[0].PasswordHash), []byte("new-password")))
	require.ErrorIs(t, auth.ResetPassword(ctx, "reset@example.com", token, "second-password"), ErrInvalidResetToken)
	require.Len(t, repo.updated, 1)
}

type notifyAtomicCache struct {
	EmailCache
	attempts atomic.Int32
	ready    sync.WaitGroup
}

func (c *notifyAtomicCache) GetNotifyVerifyCode(context.Context, string) (*VerificationCodeData, error) {
	c.ready.Done()
	c.ready.Wait()
	return &VerificationCodeData{Code: "123456"}, nil
}
func (c *notifyAtomicCache) IncrNotifyVerifyCodeAttempts(context.Context, string) (int, error) {
	return int(c.attempts.Add(1)), nil
}
func TestVerifyNotifyCode_ConcurrentGuessesRespectAttemptLimit(t *testing.T) {
	const n = 20
	cache := &notifyAtomicCache{}
	cache.ready.Add(n)
	results := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() { results <- verifyNotifyCode(context.Background(), cache, "notify@example.com", "wrong") }()
	}
	invalid := 0
	maxed := 0
	for i := 0; i < n; i++ {
		err := <-results
		if errors.Is(err, ErrInvalidVerifyCode) {
			invalid++
		} else if errors.Is(err, ErrVerifyCodeMaxAttempts) {
			maxed++
		} else {
			t.Errorf("unexpected result: %v", err)
		}
	}
	require.Equal(t, 4, invalid)
	require.Equal(t, 16, maxed)
}
