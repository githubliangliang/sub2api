//go:build unit

package repository

import (
	"encoding/json"
	"fmt"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
	"runtime"
	"strings"
	"testing"
)

func scheduler024Fixture() service.Account {
	account := service.Account{ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth, Credentials: map[string]any{"model_mapping": map[string]any{"public": "gpt-6.2-sol"}, "access_token": "fixture"}}
	for i := 1; i <= 40; i++ {
		g := &service.Group{ID: int64(i), Name: fmt.Sprint(i), Description: strings.Repeat("group configuration", 256)}
		account.GroupIDs = append(account.GroupIDs, g.ID)
		account.Groups = append(account.Groups, g)
		account.AccountGroups = append(account.AccountGroups, service.AccountGroup{AccountID: 1, GroupID: g.ID, Priority: i, Group: g, Account: &service.Account{ID: 1}})
	}
	return account
}
func scheduler024Projection(account service.Account) ([]byte, []byte, error) {
	return marshalSchedulerCacheAccount(account)
}
func TestScheduler024ProjectionPreservesOwnership(t *testing.T) {
	account := scheduler024Fixture()
	before, err := json.Marshal(account)
	require.NoError(t, err)
	full, meta, err := scheduler024Projection(account)
	require.NoError(t, err)
	var decoded service.Account
	require.NoError(t, json.Unmarshal(full, &decoded))
	require.Empty(t, decoded.Groups)
	require.Equal(t, account.GroupIDs, decoded.GroupIDs)
	require.Equal(t, filterSchedulerAccountGroups(account.AccountGroups), decoded.AccountGroups)
	require.Equal(t, "gpt-6.2-sol", decoded.GetMappedModel("public"))
	require.Equal(t, "fixture", decoded.GetCredential("access_token"))
	after, err := json.Marshal(account)
	require.NoError(t, err)
	require.Equal(t, before, after)
	expected, err := json.Marshal(buildSchedulerMetadataAccount(account))
	require.NoError(t, err)
	require.Equal(t, expected, meta)
	t.Logf("expanded=%d projected=%d", len(before), len(full))
}
func BenchmarkScheduler024Projection(b *testing.B) {
	account := scheduler024Fixture()
	for _, tc := range []struct {
		name string
		fn   func(service.Account) ([]byte, []byte, error)
	}{{"before", legacy024MarshalSchedulerCacheAccount}, {"after", scheduler024Projection}} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				full, meta, err := tc.fn(account)
				if err != nil {
					b.Fatal(err)
				}
				runtime.KeepAlive(full)
				runtime.KeepAlive(meta)
			}
		})
	}
}

func legacy024MarshalSchedulerCacheAccount(account service.Account) ([]byte, []byte, error) {
	full, err := json.Marshal(account)
	if err != nil {
		return nil, nil, err
	}
	meta, err := json.Marshal(buildSchedulerMetadataAccount(account))
	return full, meta, err
}
