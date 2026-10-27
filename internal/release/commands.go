package release

import (
	"context"
	"fmt"
	"path/filepath"
	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/wowapi"
)

func Acquire(ctx context.Context, root, snapshot, cache string) (M, error) {
	lock, e := readObject(filepath.Join(root, "scripts/wow_api/client_sources.json"))
	if e != nil {
		return nil, e
	}
	if e = projectVersion(root, lock); e != nil {
		return nil, e
	}
	if exists(snapshot) {
		return VerifySnapshot(snapshot, lock)
	}
	return Stage(ctx, root, snapshot, cache, lock, lock)
}
func Probe(ctx context.Context, root, command, output string) (M, error) {
	p := wowapi.NewProject(root)
	switch command {
	case "plan":
		return p.ProbePlan(ctx)
	case "build":
		return p.BuildProbe(ctx, output, "")
	default:
		return nil, fmt.Errorf("unknown probe command: %s", command)
	}
}

func StageSelected(ctx context.Context, root, snapshot, cache, region string) (M, error) {
	selected, e := consumer.ReadVersion(root)
	if e != nil {
		return nil, e
	}
	if exists(snapshot) {
		record, e := VerifySnapshot(snapshot, nil)
		if e != nil {
			return nil, e
		}
		if e = selected.Require(record); e != nil {
			return nil, e
		}
		return record, nil
	}
	if region == "" {
		lock, e := readObject(filepath.Join(root, "scripts/wow_api/client_sources.json"))
		if e != nil {
			return nil, e
		}
		region = s(lock["region"])
	}
	identity, e := ResolveBuild(ctx, region, selected.Build)
	if e != nil {
		return nil, e
	}
	if e = selected.Require(identity); e != nil {
		return nil, e
	}
	return Stage(ctx, root, snapshot, cache, identity, nil)
}
