package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newMiniredisEmailCache(t *testing.T) (service.EmailCache, *miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return NewEmailCache(rdb), mr, rdb
}

func TestEmailCache_ConcurrentWrongCodesCannotExceedAttemptCap(t *testing.T) {
	cache, _, _ := newMiniredisEmailCache(t)
	ctx := context.Background()
	email := "user@example.com"

	svc := service.NewEmailService(nil, cache)
	require.NoError(t, cache.SetVerificationCode(ctx, email, &service.VerificationCodeData{
		Code:      "123456",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}, 15*time.Minute))

	const workers = 50
	var invalid, maxed atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := svc.VerifyCode(ctx, email, "000000")
			switch {
			case errors.Is(err, service.ErrInvalidVerifyCode):
				invalid.Add(1)
			case errors.Is(err, service.ErrVerifyCodeMaxAttempts):
				maxed.Add(1)
			default:
				t.Errorf("unexpected result: %v", err)
			}
		}()
	}
	wg.Wait()

	// Only attempts 1..4 may return "invalid"; every other guess is rejected by the cap.
	require.LessOrEqual(t, int(invalid.Load()), 4)
	require.Equal(t, workers, int(invalid.Load()+maxed.Load()))

	// Even the correct code is now rejected.
	require.ErrorIs(t, svc.VerifyCode(ctx, email, "123456"), service.ErrVerifyCodeMaxAttempts)

	data, err := cache.GetVerificationCode(ctx, email)
	require.NoError(t, err)
	require.GreaterOrEqual(t, data.Attempts, 5)
}

func TestEmailCache_AttemptsResetOnNewCodeAndTTLFollowsCode(t *testing.T) {
	cache, mr, _ := newMiniredisEmailCache(t)
	ctx := context.Background()
	email := "User@Example.com"

	require.NoError(t, cache.SetVerificationCode(ctx, email, &service.VerificationCodeData{Code: "1"}, time.Minute))
	n, err := cache.IncrVerificationCodeAttempts(ctx, email)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Greater(t, mr.TTL(verifyCodeKey(email)+attemptsKeySuffix), time.Duration(0))

	require.NoError(t, cache.SetVerificationCode(ctx, email, &service.VerificationCodeData{Code: "2"}, time.Minute))
	data, err := cache.GetVerificationCode(ctx, email)
	require.NoError(t, err)
	require.Equal(t, 0, data.Attempts)

	require.NoError(t, cache.DeleteVerificationCode(ctx, email))
	_, err = cache.IncrVerificationCodeAttempts(ctx, email)
	require.Error(t, err)
	require.False(t, mr.Exists(verifyCodeKey(email)+attemptsKeySuffix))
}

func TestEmailCache_PasswordResetTokenHashedAndSingleUse(t *testing.T) {
	cache, mr, _ := newMiniredisEmailCache(t)
	ctx := context.Background()
	email := "reset@example.com"

	svc := service.NewEmailService(nil, cache)

	// Seed the token the same way SendPasswordResetEmail does (hash only).
	token, err := svc.GeneratePasswordResetToken()
	require.NoError(t, err)
	sum := sha256.Sum256([]byte(token))
	require.NoError(t, cache.SetPasswordResetToken(ctx, email, &service.PasswordResetTokenData{
		Token: hex.EncodeToString(sum[:]), CreatedAt: time.Now(),
	}, 30*time.Minute))

	raw, err := mr.Get(passwordResetKey(email))
	require.NoError(t, err)
	require.False(t, strings.Contains(raw, token), "plaintext token must not be stored")

	require.NoError(t, svc.VerifyPasswordResetToken(ctx, email, token))
	require.ErrorIs(t, svc.ConsumePasswordResetToken(ctx, email, "wrong"), service.ErrInvalidResetToken)

	const workers = 30
	var ok atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if svc.ConsumePasswordResetToken(ctx, email, token) == nil {
				ok.Add(1)
			}
		}()
	}
	wg.Wait()
	require.Equal(t, int32(1), ok.Load())
	require.False(t, mr.Exists(passwordResetKey(email)))
}

func TestEmailCache_ConsumePasswordResetTokenMismatchKeepsToken(t *testing.T) {
	cache, mr, _ := newMiniredisEmailCache(t)
	ctx := context.Background()
	email := "keep@example.com"
	require.NoError(t, cache.SetPasswordResetToken(ctx, email, &service.PasswordResetTokenData{Token: "abc"}, time.Minute))

	ok, err := cache.ConsumePasswordResetToken(ctx, email, "xyz")
	require.NoError(t, err)
	require.False(t, ok)
	require.True(t, mr.Exists(passwordResetKey(email)))

	ok, err = cache.ConsumePasswordResetToken(ctx, email, "abc")
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = cache.ConsumePasswordResetToken(ctx, email, "abc")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestEmailCache_LegacyAttemptsAreNotResetAfterUpgrade(t *testing.T) {
	cache, mr, rdb := newMiniredisEmailCache(t)
	ctx := context.Background()
	email := "legacy@example.com"
	// Previous releases kept attempts only inside JSON.
	require.NoError(t, rdb.Set(ctx, verifyCodeKey(email), `{"Code":"123456","Attempts":4}`, time.Minute).Err())
	svc := service.NewEmailService(nil, cache)
	require.ErrorIs(t, svc.VerifyCode(ctx, email, "wrong"), service.ErrVerifyCodeMaxAttempts)
	require.ErrorIs(t, svc.VerifyCode(ctx, email, "123456"), service.ErrVerifyCodeMaxAttempts)
	require.LessOrEqual(t, mr.TTL(verifyCodeKey(email)+attemptsKeySuffix), mr.TTL(verifyCodeKey(email)))
}

func TestEmailCache_NotifyAttemptsExpireAndReset(t *testing.T) {
	cache, mr, _ := newMiniredisEmailCache(t)
	ctx := context.Background()
	email := "Notify@Example.com"
	require.NoError(t, cache.SetNotifyVerifyCode(ctx, email, &service.VerificationCodeData{Code: "123456"}, time.Minute))
	const workers = 20
	var wg sync.WaitGroup
	results := make(chan int, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, err := cache.IncrNotifyVerifyCodeAttempts(ctx, email)
			if err != nil {
				t.Error(err)
				return
			}
			results <- n
		}()
	}
	wg.Wait()
	close(results)
	seen := map[int]bool{}
	for n := range results {
		seen[n] = true
	}
	require.Len(t, seen, workers)
	data, err := cache.GetNotifyVerifyCode(ctx, email)
	require.NoError(t, err)
	require.Equal(t, workers, data.Attempts)
	require.Equal(t, mr.TTL(notifyVerifyKey(email)), mr.TTL(notifyVerifyKey(email)+attemptsKeySuffix))
	mr.FastForward(30 * time.Second)
	require.NoError(t, cache.SetNotifyVerifyCode(ctx, email, &service.VerificationCodeData{Code: "654321"}, time.Minute))
	data, err = cache.GetNotifyVerifyCode(ctx, email)
	require.NoError(t, err)
	require.Zero(t, data.Attempts)
	n, err := cache.IncrNotifyVerifyCodeAttempts(ctx, email)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	mr.FastForward(time.Minute)
	_, err = cache.IncrNotifyVerifyCodeAttempts(ctx, email)
	require.ErrorIs(t, err, redis.Nil)
	require.False(t, mr.Exists(notifyVerifyKey(email)+attemptsKeySuffix))
	require.NoError(t, cache.SetNotifyVerifyCode(ctx, email, &service.VerificationCodeData{Code: "1"}, time.Minute))
	_, err = cache.IncrNotifyVerifyCodeAttempts(ctx, email)
	require.NoError(t, err)
	require.NoError(t, cache.DeleteNotifyVerifyCode(ctx, email))
	require.False(t, mr.Exists(notifyVerifyKey(email)))
	require.False(t, mr.Exists(notifyVerifyKey(email)+attemptsKeySuffix))
}

func TestEmailCache_ExpiredOrBrokenCacheCannotValidate(t *testing.T) {
	cache, mr, rdb := newMiniredisEmailCache(t)
	ctx := context.Background()
	email := "expiry@example.com"
	svc := service.NewEmailService(nil, cache)
	require.NoError(t, cache.SetVerificationCode(ctx, email, &service.VerificationCodeData{Code: "123456"}, time.Minute))
	mr.FastForward(time.Minute)
	require.ErrorIs(t, svc.VerifyCode(ctx, email, "123456"), service.ErrInvalidVerifyCode)
	require.NoError(t, cache.SetVerificationCode(ctx, email, &service.VerificationCodeData{Code: "123456"}, time.Minute))
	require.NoError(t, rdb.Set(ctx, verifyCodeKey(email)+attemptsKeySuffix, "broken", time.Minute).Err())
	require.ErrorIs(t, svc.VerifyCode(ctx, email, "123456"), service.ErrInvalidVerifyCode)
	require.NoError(t, rdb.Close())
	require.ErrorIs(t, svc.ConsumePasswordResetToken(ctx, email, "token"), service.ErrInvalidResetToken)
}
