package release

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi"
)

// clientSelection is the client build a snapshot must contain. A locked
// selection also pins the exact snapshot record. identity is consulted only
// when the snapshot has to be extracted.
type clientSelection struct {
	want     consumer.Version
	lock     *sourcelock.Snapshot
	identity func(context.Context) (sourcelock.ClientIdentity, error)
}

func lockedSelection(lock sourcelock.Snapshot) clientSelection {
	return clientSelection{
		want:     consumer.Version{Full: lock.Version, Build: lock.Build},
		lock:     &lock,
		identity: func(context.Context) (sourcelock.ClientIdentity, error) { return lock.Identity(), nil },
	}
}

// offeredSelection resolves the client identity from the build Battle.net
// currently offers in region, which must be the wanted one.
func offeredSelection(want consumer.Version, f fetcher, region string) clientSelection {
	return clientSelection{want: want, identity: func(ctx context.Context) (sourcelock.ClientIdentity, error) {
		identity, e := resolveBuild(ctx, f, region, want.Build)
		if e != nil {
			return sourcelock.ClientIdentity{}, e
		}
		return identity, want.Require(identity.Version, identity.Build)
	}}
}

func (s clientSelection) verify(snapshot string) (sourcelock.Snapshot, error) {
	record, e := VerifySnapshot(snapshot, s.lock)
	if e != nil {
		return sourcelock.Snapshot{}, e
	}
	return record, s.want.Require(record.Version, record.Build)
}

type extractFunc func(ctx context.Context, output string, identity sourcelock.ClientIdentity, expected *sourcelock.Snapshot) (sourcelock.Snapshot, error)

func extractor(root, cache string) extractFunc {
	return func(ctx context.Context, output string, identity sourcelock.ClientIdentity, expected *sourcelock.Snapshot) (sourcelock.Snapshot, error) {
		return extractSnapshot(ctx, root, output, cache, identity, expected)
	}
}

// obtainSnapshot verifies the snapshot at path, extracting it there first when absent.
func obtainSnapshot(ctx context.Context, path string, s clientSelection, extract extractFunc) (sourcelock.Snapshot, error) {
	if exists(path) {
		record, e := s.verify(path)
		if e != nil {
			return sourcelock.Snapshot{}, fmt.Errorf("client snapshot %s: %w", path, e)
		}
		return record, nil
	}
	identity, e := s.identity(ctx)
	if e != nil {
		return sourcelock.Snapshot{}, e
	}
	record, e := extract(ctx, path, identity, s.lock)
	if e != nil {
		return sourcelock.Snapshot{}, fmt.Errorf("client snapshot %s: %w", path, e)
	}
	return record, nil
}

// ClientSnapshotPath is the conventional location of a user-provided client
// snapshot for build.
func ClientSnapshotPath(root string, build int) string {
	return filepath.Join(root, "build/wow-client", strconv.Itoa(build))
}

// resolveSnapshot prefers a user-provided build/wow-client/<build> snapshot and
// otherwise uses the update cache. Cache entries are keyed by full client
// version and current extractor source, so an extractor change re-extracts
// instead of trusting output of an older build. Locked and .wow-version
// selections of one build share an entry.
func resolveSnapshot(ctx context.Context, root string, s clientSelection, extract extractFunc) (string, error) {
	provided := ClientSnapshotPath(root, s.want.Build)
	if exists(provided) {
		if _, e := s.verify(provided); e != nil {
			return "", fmt.Errorf("client snapshot %s: %w", provided, e)
		}
		return provided, nil
	}
	extractorSHA, e := extractorSourceSHA(root)
	if e != nil {
		return "", e
	}
	cached := filepath.Join(root, "build/wow-update/snapshots", s.want.Full, extractorSHA)
	if _, e := obtainSnapshot(ctx, cached, s, extract); e != nil {
		return "", e
	}
	return cached, nil
}

func Acquire(ctx context.Context, root, snapshot, cache string) (sourcelock.Snapshot, error) {
	lock, e := sourcelock.ReadSnapshot(wowapi.NewProject(root).ClientSourcesPath())
	if e != nil {
		return sourcelock.Snapshot{}, e
	}
	if e = projectVersion(root, lock.Identity()); e != nil {
		return sourcelock.Snapshot{}, e
	}
	return obtainSnapshot(ctx, snapshot, lockedSelection(lock), extractor(root, cache))
}
func StageSelected(ctx context.Context, root, snapshot, cache, region string) (sourcelock.Snapshot, error) {
	selected, e := consumer.ReadVersion(root)
	if e != nil {
		return sourcelock.Snapshot{}, e
	}
	if region == "" {
		lock, e := sourcelock.ReadSnapshot(wowapi.NewProject(root).ClientSourcesPath())
		if e != nil {
			return sourcelock.Snapshot{}, e
		}
		region = lock.Region
	}
	return obtainSnapshot(ctx, snapshot, offeredSelection(selected, newFetcher(), region), extractor(root, cache))
}
