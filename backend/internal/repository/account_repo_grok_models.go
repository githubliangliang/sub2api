package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

var errGrokCatalogInvalidJSON = errors.New("grok model catalog: invalid json")

var _ service.GrokCatalogSnapshotRepository = (*accountRepository)(nil)

// accountTransactionScope keeps native SQL and all ORM association reads in the
// caller's transaction and prevents publication of uncommitted scheduler state.
func (r *accountRepository) accountTransactionScope(ctx context.Context) *accountRepository {
	tx := dbent.TxFromContext(ctx)
	if tx == nil || r.client == tx.Client() {
		return r
	}
	scoped := *r
	scoped.client = tx.Client()
	scoped.sql = tx.Client()
	scoped.schedulerCache = nil
	return &scoped
}

func (r *accountRepository) CompareAndSwapGrokObservedModels(ctx context.Context, id int64, expected, next json.RawMessage) (bool, error) {
	expectedArg, err := grokCatalogJSONArg(expected)
	if err != nil {
		return false, err
	}
	nextArg, err := grokCatalogJSONArg(next)
	if err != nil {
		return false, err
	}
	applied := false
	committed, err := withRepositoryTransaction(ctx, r.client, func(txCtx context.Context, client *dbent.Client) error {
		// SQLite has no row locks. Acquire its writer lock before reading the
		// expected value so ordinary edits and observations cannot interleave.
		result, err := client.ExecContext(txCtx, `UPDATE accounts SET updated_at = updated_at
			WHERE id = $1 AND deleted_at IS NULL AND platform = 'grok'
			AND (extra IS NULL OR json_type(extra) IN ('object', 'null'))`, id)
		if err != nil {
			return err
		}
		count, err := result.RowsAffected()
		if err != nil || count == 0 {
			return err
		}
		rows, err := client.QueryContext(txCtx, `SELECT COALESCE(extra -> 'grok_observed_models', 'null')
			FROM accounts WHERE id = $1 AND deleted_at IS NULL AND platform = 'grok'`, id)
		if err != nil {
			return err
		}
		var current string
		if !rows.Next() {
			err := rows.Err()
			_ = rows.Close()
			return err
		}
		err = rows.Scan(&current)
		closeErr := rows.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		if canonicalJSON(current) != canonicalJSON(expectedArg) {
			return nil
		}
		result, err = client.ExecContext(txCtx, `UPDATE accounts
			SET extra = json_set(CASE WHEN json_type(extra) = 'object' THEN extra ELSE '{}' END, '$.grok_observed_models', json($2)), updated_at = $3
			WHERE id = $1 AND deleted_at IS NULL AND platform = 'grok'`, id, nextArg, time.Now())
		if err != nil {
			return err
		}
		count, err = result.RowsAffected()
		applied = err == nil && count > 0
		return err
	})
	if err != nil {
		return false, err
	}
	if committed && applied {
		r.syncSchedulerAccountSnapshot(ctx, id)
	}
	return applied, nil
}

func grokCatalogJSONArg(raw json.RawMessage) (string, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return "null", nil
	}
	if !json.Valid(raw) {
		return "", errGrokCatalogInvalidJSON
	}
	return string(raw), nil
}
