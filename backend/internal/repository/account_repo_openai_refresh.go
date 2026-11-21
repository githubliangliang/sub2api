package repository

import (
	"context"
	"errors"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// UpdateOpenAIOAuthCredentialsIfUnchanged commits a rotation and its outbox event
// together. Taking SQLite's writer lock before reading makes the full document
// comparison atomic without relying on JSON key order in stored text.
func (r *accountRepository) UpdateOpenAIOAuthCredentialsIfUnchanged(ctx context.Context, id int64, expected map[string]any, expectedProxyID *int64, replacement map[string]any) (bool, error) {
	if r == nil || r.client == nil {
		return false, errors.New("account repository Ent client is not configured")
	}
	applied := false
	committed, err := withRepositoryTransaction(dbent.WithoutTx(ctx), r.client, func(txCtx context.Context, client *dbent.Client) error {
		if _, err := client.ExecContext(txCtx, "UPDATE accounts SET updated_at = updated_at WHERE id = $1 AND deleted_at IS NULL", id); err != nil {
			return err
		}
		current, err := client.Account.Get(txCtx, id)
		if dbent.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if current.DeletedAt != nil || current.Platform != service.PlatformOpenAI || current.Type != service.AccountTypeOAuth || current.ParentAccountID != nil ||
			!jsonMapsEqual(normalizeJSONMap(current.Credentials), normalizeJSONMap(expected)) || !nullableInt64Equal(current.ProxyID, expectedProxyID) {
			return nil
		}
		if _, err := client.Account.UpdateOneID(id).SetCredentials(normalizeJSONMap(replacement)).Save(txCtx); err != nil {
			return err
		}
		if err := enqueueSchedulerOutbox(txCtx, client, service.SchedulerOutboxEventAccountChanged, &id, nil, nil); err != nil {
			return err
		}
		applied = true
		return nil
	})
	if err != nil {
		return false, err
	}
	if applied && committed {
		r.syncSchedulerAccountSnapshotDetached(ctx, id)
	}
	return applied && committed, nil
}
