package release

import (
	"context"
	"path/filepath"
	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/sourcelock"
)

func Acquire(ctx context.Context, root, snapshot, cache string) (sourcelock.Snapshot, error) {
	lock, e := sourcelock.ReadSnapshot(filepath.Join(root, "internal/wowapi/inputs/client_sources.json"))
	if e != nil {
		return sourcelock.Snapshot{}, e
	}
	if e = projectVersion(root, lock.Identity()); e != nil {
		return sourcelock.Snapshot{}, e
	}
	if exists(snapshot) {
		return VerifySnapshot(snapshot, &lock)
	}
	return Stage(ctx, root, snapshot, cache, lock.Identity(), &lock)
}
func StageSelected(ctx context.Context, root, snapshot, cache, region string) (sourcelock.Snapshot, error) {
	selected, e := consumer.ReadVersion(root)
	if e != nil {
		return sourcelock.Snapshot{}, e
	}
	if exists(snapshot) {
		record, e := VerifySnapshot(snapshot, nil)
		if e != nil {
			return sourcelock.Snapshot{}, e
		}
		if e = selected.Require(record.Version, record.Build); e != nil {
			return sourcelock.Snapshot{}, e
		}
		return record, nil
	}
	if region == "" {
		lock, e := sourcelock.ReadSnapshot(filepath.Join(root, "internal/wowapi/inputs/client_sources.json"))
		if e != nil {
			return sourcelock.Snapshot{}, e
		}
		region = lock.Region
	}
	identity, e := ResolveBuild(ctx, region, selected.Build)
	if e != nil {
		return sourcelock.Snapshot{}, e
	}
	if e = selected.Require(identity.Version, identity.Build); e != nil {
		return sourcelock.Snapshot{}, e
	}
	return Stage(ctx, root, snapshot, cache, identity, nil)
}
