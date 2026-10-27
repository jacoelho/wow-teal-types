package release

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi"
)

func readBounded(r io.Reader, limit int64) ([]byte, error) {
	b, e := io.ReadAll(io.LimitReader(r, limit+1))
	if e != nil {
		return nil, e
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("acquisition response exceeds %d bytes", limit)
	}
	return b, nil
}
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
func syncPrimary(ctx context.Context, p wowapi.Project) error {
	lock, e := sourcelock.ReadSources(filepath.Join(p.Root, "internal/wowapi/inputs/sources.json"))
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
			b, e := readFile(path, maxFile)
			if e != nil {
				return e
			}
			if len(b) != j.size || blob(b) != j.blob {
				return fmt.Errorf("modified pinned source: %s", path)
			}
			return nil
		}
		if j.group == "blizzard" || j.group == "blizzard_version" {
			return fmt.Errorf("missing verified native input: %s", path)
		}
		if !hex40.MatchString(j.commit) {
			return fmt.Errorf("community source requires immutable revision")
		}
		b, e := fetch(ctx, "https://raw.githubusercontent.com/"+j.repo+"/"+j.commit+"/"+escapePath(j.upstream), maxFile)
		if e != nil {
			return e
		}
		if len(b) != j.size || blob(b) != j.blob {
			return fmt.Errorf("community blob identity mismatch: %s", j.name)
		}
		return atomicWrite(path, b, 0644)
	})
}

// Bootstrap selects verified native inputs and populates pinned community inputs.
func Bootstrap(ctx context.Context, root, snapshot string) (wowapi.Project, error) {
	p, e := Projection(root, snapshot, filepath.Join(root, "internal/wowapi/inputs/client_sources.json"))
	if e != nil {
		return p, e
	}
	if e = syncPrimary(ctx, p); e != nil {
		return p, e
	}
	return p, nil
}

// Cache files remain optional, integrity-checked projections rather than authorities.
func seedCaches(origin, target string) error {
	p := wowapi.NewProject(origin)
	lock, e := sourcelock.ReadSources(filepath.Join(p.Root, "internal/wowapi/inputs/sources.json"))
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
			b, e := readFile(src, maxFile)
			if e != nil {
				return e
			}
			if len(b) != row.Size || blob(b) != row.SHA {
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

func acquireHTTPClient() *http.Client {
	return &http.Client{Timeout: 45 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 || req.URL.Scheme != "https" {
			return fmt.Errorf("unsupported acquisition redirect")
		}
		return nil
	}}
}
