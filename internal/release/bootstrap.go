package release

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi"
)

func parallel(ctx context.Context, items []sourceJob, fn func(context.Context, sourceJob) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	work := make(chan sourceJob)
	var once sync.Once
	var first error
	var wg sync.WaitGroup
	for range min(8, len(items)) {
		wg.Go(func() {
			for item := range work {
				if ctx.Err() != nil {
					continue
				}
				if e := fn(ctx, item); e != nil {
					once.Do(func() { first = e; cancel() })
				}
			}
		})
	}
	for _, item := range items {
		if ctx.Err() != nil {
			break
		}
		select {
		case work <- item:
		case <-ctx.Done():
		}
	}
	close(work)
	wg.Wait()
	if first != nil {
		return first
	}
	return ctx.Err()
}

type sourceJob struct {
	group, repo, commit, upstream, name, blob string
	size                                      int
}

func escapePath(path string) string {
	parts := strings.Split(path, "/")
	for n := range parts {
		parts[n] = url.PathEscape(parts[n])
	}
	return strings.Join(parts, "/")
}
func syncPrimary(ctx context.Context, f fetcher, p wowapi.Project) error {
	lock, e := sourcelock.ReadSources(p.SourcesPath())
	if e != nil {
		return e
	}
	var jobs []sourceJob
	for _, group := range keys(lock) {
		source := lock[group]
		for _, r := range source.Files {
			jobs = append(jobs, sourceJob{group, source.Repository, source.Commit, strings.Trim(source.Path+"/"+r.Name, "/"), r.Name, r.SHA, r.Size})
		}
	}
	return parallel(ctx, jobs, func(ctx context.Context, j sourceJob) error {
		path, e := p.InputPath(j.group, j.name)
		if e != nil {
			return e
		}
		if _, e := safeJoin(filepath.Dir(path), filepath.Base(path)); e != nil {
			return e
		}
		if exists(path) {
			b, e := sourcelock.ReadRegular(path, maxFile)
			if e != nil {
				return e
			}
			if len(b) != j.size || sourcelock.BlobSHA(b) != j.blob {
				return fmt.Errorf("modified pinned source: %s", path)
			}
			return nil
		}
		if j.group == "blizzard" || j.group == "blizzard_version" {
			return fmt.Errorf("missing verified native input: %s", path)
		}
		if !sourcelock.IsGitSHA(j.commit) {
			return fmt.Errorf("community source requires immutable revision")
		}
		b, e := f.get(ctx, "https://raw.githubusercontent.com/"+j.repo+"/"+j.commit+"/"+escapePath(j.upstream), maxFile)
		if e != nil {
			return e
		}
		if len(b) != j.size || sourcelock.BlobSHA(b) != j.blob {
			return fmt.Errorf("community blob identity mismatch: %s", j.name)
		}
		return atomicWrite(path, b, 0644)
	})
}

// Bootstrap selects verified native inputs and populates pinned community inputs.
func Bootstrap(ctx context.Context, root, snapshot string) (wowapi.Project, error) {
	return bootstrap(ctx, newFetcher(), root, snapshot)
}
func bootstrap(ctx context.Context, f fetcher, root, snapshot string) (wowapi.Project, error) {
	p, e := projectSources(root, snapshot)
	if e != nil {
		return p, e
	}
	return p, syncPrimary(ctx, f, p)
}

// Cache files remain optional, integrity-checked projections rather than authorities.
func seedCaches(origin, target string) error {
	p := wowapi.NewProject(origin)
	lock, e := sourcelock.ReadSources(p.SourcesPath())
	if e != nil {
		return e
	}
	for _, group := range keys(lock) {
		if group == "blizzard" || group == "blizzard_version" {
			continue
		}
		for _, row := range lock[group].Files {
			name := row.Name
			src, e := p.InputPath(group, name)
			if e != nil {
				return e
			}
			if !exists(src) {
				continue
			}
			b, e := sourcelock.ReadRegular(src, maxFile)
			if e != nil {
				return e
			}
			if !row.Matches(b) {
				return fmt.Errorf("modified pinned community cache: %s", name)
			}
			dst, e := wowapi.NewProject(target).InputPath(group, name)
			if e != nil {
				return e
			}
			if e = atomicWrite(dst, b, 0644); e != nil {
				return e
			}
		}
	}
	return nil
}
