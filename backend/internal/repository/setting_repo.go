package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/setting"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type settingRepository struct {
	client *ent.Client
}

func NewSettingRepository(client *ent.Client) service.SettingRepository {
	return &settingRepository{client: client}
}

func (r *settingRepository) Get(ctx context.Context, key string) (*service.Setting, error) {
	m, err := r.client.Setting.Query().Where(setting.KeyEQ(key)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, service.ErrSettingNotFound
		}
		return nil, err
	}
	return &service.Setting{
		ID:        m.ID,
		Key:       m.Key,
		Value:     m.Value,
		UpdatedAt: m.UpdatedAt,
	}, nil
}

func (r *settingRepository) GetValue(ctx context.Context, key string) (string, error) {
	setting, err := r.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (r *settingRepository) Set(ctx context.Context, key, value string) error {
	now := time.Now()
	return r.client.Setting.
		Create().
		SetKey(key).
		SetValue(value).
		SetUpdatedAt(now).
		OnConflictColumns(setting.FieldKey).
		UpdateNewValues().
		Exec(ctx)
}

func (r *settingRepository) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}
	settings, err := r.client.Setting.Query().Where(setting.KeyIn(keys...)).All(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

func (r *settingRepository) SetMultiple(ctx context.Context, settings map[string]string) error {
	if len(settings) == 0 {
		return nil
	}

	now := time.Now()
	builders := make([]*ent.SettingCreate, 0, len(settings))
	for key, value := range settings {
		builders = append(builders, r.client.Setting.Create().SetKey(key).SetValue(value).SetUpdatedAt(now))
	}
	return r.client.Setting.
		CreateBulk(builders...).
		OnConflictColumns(setting.FieldKey).
		UpdateNewValues().
		Exec(ctx)
}

func (r *settingRepository) GetAll(ctx context.Context) (map[string]string, error) {
	settings, err := r.client.Setting.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

func (r *settingRepository) Delete(ctx context.Context, key string) error {
	_, err := r.client.Setting.Delete().Where(setting.KeyEQ(key)).Exec(ctx)
	return err
}

// SetContentModerationConfig serializes scope validation and persistence with
// group deletion using SQLite's write transaction, including an initially absent setting.
func (r *settingRepository) SetContentModerationConfig(ctx context.Context, value string, groupIDs []int64) error {
	_, err := withRepositoryTransaction(ctx, r.client, func(txCtx context.Context, client *ent.Client) error {
		// Acquire the writer before reading groups; a deferred read transaction
		// otherwise cannot safely upgrade after another writer commits.
		if _, err := client.ExecContext(txCtx, "UPDATE groups SET id = id WHERE 0"); err != nil {
			return err
		}
		unique := make(map[int64]struct{}, len(groupIDs))
		ids := make([]int64, 0, len(groupIDs))
		for _, id := range groupIDs {
			if id <= 0 {
				return service.ErrGroupNotFound
			}
			if _, exists := unique[id]; !exists {
				unique[id] = struct{}{}
				ids = append(ids, id)
			}
		}
		if len(ids) > 0 {
			count, err := client.Group.Query().Where(group.IDIn(ids...)).Count(txCtx)
			if err != nil {
				return err
			}
			if count != len(ids) {
				return service.ErrGroupNotFound
			}
		}
		return client.Setting.Create().SetKey(service.SettingKeyContentModerationConfig).
			SetValue(value).SetUpdatedAt(time.Now()).OnConflictColumns(setting.FieldKey).UpdateNewValues().Exec(txCtx)
	})
	return err
}

// removeContentModerationGroup runs inside the group deletion transaction.
// Preserve legacy malformed settings and all unrelated JSON values.
func removeContentModerationGroup(ctx context.Context, client *ent.Client, groupID int64) error {
	current, err := client.Setting.Query().Where(setting.KeyEQ(service.SettingKeyContentModerationConfig)).Only(ctx)
	if ent.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var document map[string]json.RawMessage
	if json.Unmarshal([]byte(current.Value), &document) != nil || document == nil {
		return nil
	}
	var entries []json.RawMessage
	if json.Unmarshal(document["group_ids"], &entries) != nil || entries == nil {
		return nil
	}
	kept := make([]json.RawMessage, 0, len(entries))
	changed := false
	for _, entry := range entries {
		var id int64
		if json.Unmarshal(entry, &id) != nil || id <= 0 {
			return nil
		}
		if id == groupID {
			changed = true
		} else {
			kept = append(kept, entry)
		}
	}
	if !changed {
		return nil
	}
	document["group_ids"], err = json.Marshal(kept)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(document)
	if err != nil {
		return err
	}
	return client.Setting.UpdateOneID(current.ID).SetValue(string(raw)).SetUpdatedAt(time.Now()).Exec(ctx)
}
