package service

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFingerprintVersionOverrideIdentityConsistency(t *testing.T) {
	if os.Getenv("SUB2API_TEST_CLI_OVERRIDE_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestFingerprintVersionOverrideIdentityConsistency$")
		cmd.Env = append(os.Environ(), "SUB2API_TEST_CLI_OVERRIDE_CHILD=1", "SUB2API_CLAUDE_CLI_VERSION=2.2.0")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, string(output))
		return
	}
	for _, cachedUA := range []string{"", "claude-cli/2.1.220 (external, desktop)", "claude-cli/999.0.0-local"} {
		t.Run(cachedUA, func(t *testing.T) {
			cache := &stubIdentityCache{}
			if cachedUA != "" {
				cache.fingerprint = &Fingerprint{UserAgent: cachedUA, ClientID: "stable-id", StainlessOS: "Linux", UpdatedAt: time.Now().Unix()}
			}
			svc := NewIdentityService(cache)
			fp, err := svc.GetOrCreateFingerprint(context.Background(), 1, headersWithUA("claude-cli/2.1.75 (external, desktop)"))
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "https://example.test", nil)
			require.NoError(t, err)
			svc.ApplyFingerprint(req, fp)
			require.Equal(t, "claude-cli/2.2.0 (external, desktop)", req.Header.Get("User-Agent"))
			body := syncBillingHeaderVersion([]byte(`{"system":[{"type":"text","text":"x-anthropic-billing-header: cc_version=2.1.75.abc;"}]}`), fp.UserAgent)
			// SHA256("59cf53e54c78" + "000" + "2.2.0") starts with 3d5.
			require.Contains(t, string(body), "cc_version=2.2.0.3d5")
			if cachedUA != "" {
				require.Equal(t, "stable-id", fp.ClientID)
				require.Equal(t, "Linux", fp.StainlessOS)
			}
		})
	}
}

// GetOrCreateFingerprint 集成行为，直接对应线上故障：缓存指纹停留在历史版本
// 2.1.220（生产 Redis 中账号 147 的实际值），客户端送来更旧的 2.1.75。
// 修复前：isNewerVersion 不触发、UA 形态合法不触发自愈，旧指纹被原样返回并
// 近乎永不过期——上游按指纹 UA 做客户端版本闸门（Fable 5.1 要求 >= 2.1.251），
// 仅升 CLICurrentVersion 对存量账号完全无效。
func TestGetOrCreateFingerprintFloorsStaleCachedUserAgent(t *testing.T) {
	cache := &stubIdentityCache{fingerprint: &Fingerprint{
		UserAgent:               "claude-cli/2.1.220 (external, cli)",
		ClientID:                "cid-1",
		StainlessPackageVersion: "0.91.1",
		UpdatedAt:               time.Now().Unix(),
	}}
	svc := NewIdentityService(cache)

	fp, err := svc.GetOrCreateFingerprint(
		context.Background(), 147,
		headersWithUA("claude-cli/2.1.75 (external, cli)"),
	)

	require.NoError(t, err)
	require.Equal(t, "claude-cli/"+"2.1.258"+" (external, cli)", fp.UserAgent)
	require.Equal(t, 1, cache.setCalls, "下限抬升必须持久化写回缓存")
	require.Equal(t, "claude-cli/"+"2.1.258"+" (external, cli)", cache.lastSet.UserAgent)
	// X-Stainless-* 维持既有 merge 语义：客户端未携带时保留缓存中的真实值，不被下限逻辑覆盖。
	require.Equal(t, "0.91.1", fp.StainlessPackageVersion)
}

// 缓存版本高于下限：不得被降级，也不得触发多余写入。
func TestGetOrCreateFingerprintDoesNotTouchCacheAboveFloor(t *testing.T) {
	ua := "claude-cli/2.9.0 (external, cli)"
	cache := &stubIdentityCache{fingerprint: &Fingerprint{
		UserAgent:               ua,
		ClientID:                "cid-1",
		StainlessPackageVersion: "0.91.1",
		UpdatedAt:               time.Now().Unix(),
	}}
	svc := NewIdentityService(cache)

	fp, err := svc.GetOrCreateFingerprint(
		context.Background(), 1,
		headersWithUA("claude-cli/2.1.75 (external, cli)"),
	)

	require.NoError(t, err)
	require.Equal(t, ua, fp.UserAgent, "高于下限的缓存版本不得被降级")
	require.Zero(t, cache.setCalls)
}

// 客户端送来高于下限的更新版本：常规升级照常生效，且不被下限压回。
func TestGetOrCreateFingerprintClientNewerThanFloorStillWins(t *testing.T) {
	newUA := "claude-cli/2.9.0 (external, cli)"
	cache := &stubIdentityCache{fingerprint: &Fingerprint{
		UserAgent: "claude-cli/2.1.220 (external, cli)",
		ClientID:  "cid-1",
		UpdatedAt: time.Now().Unix(),
	}}
	svc := NewIdentityService(cache)

	fp, err := svc.GetOrCreateFingerprint(context.Background(), 1, headersWithUA(newUA))

	require.NoError(t, err)
	require.Equal(t, newUA, fp.UserAgent, "客户端更新版本优先于下限，不得被压回")
	require.Equal(t, 1, cache.setCalls)
}

// 客户端版本高于缓存但低于下限：merge 升级后仍被抬到下限（二者取更新者）。
func TestGetOrCreateFingerprintFloorWinsOverStaleClientUpgrade(t *testing.T) {
	cache := &stubIdentityCache{fingerprint: &Fingerprint{
		UserAgent: "claude-cli/2.1.22 (external, cli)",
		ClientID:  "cid-1",
		UpdatedAt: time.Now().Unix(),
	}}
	svc := NewIdentityService(cache)

	fp, err := svc.GetOrCreateFingerprint(
		context.Background(), 1,
		headersWithUA("claude-cli/2.1.223 (external, cli)"),
	)

	require.NoError(t, err)
	require.Equal(t, "claude-cli/"+"2.1.258"+" (external, cli)", fp.UserAgent)
	require.Equal(t, 1, cache.setCalls)
}

// 首次创建路径同样受下限约束：合法但过旧的 claude-cli UA 落库时必须抬到 CLICurrentVersion。
func TestCreateFingerprintFromHeadersFloorsOldClientUserAgent(t *testing.T) {
	cache := &stubIdentityCache{}
	svc := NewIdentityService(cache)

	fp, err := svc.GetOrCreateFingerprint(
		context.Background(), 1,
		headersWithUA("claude-cli/2.1.75 (external, claude-desktop-3p)"),
	)

	require.NoError(t, err)
	require.Equal(t, "claude-cli/"+"2.1.258"+" (external, claude-desktop-3p)", fp.UserAgent)
	require.Equal(t, 1, cache.setCalls)
}
