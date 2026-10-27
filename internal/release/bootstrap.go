package release

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"
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
func parallel[T any](ctx context.Context, items []T, fn func(context.Context, T) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	work := make(chan T)
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
func matches(b []byte, row M) bool {
	if len(b) > maxPayload {
		return false
	}
	if n, ok := row["size"]; ok && len(b) != i(n) {
		return false
	}
	if sha(b) != s(row["sha256"]) {
		return false
	}
	for _, k := range []string{"git_blob_sha1", "sha"} {
		if v, ok := row[k]; ok && blob(b) != s(v) {
			return false
		}
	}
	return true
}

type sourceJob struct {
	group, repo, commit, upstream, name string
	row                                 M
}

func secondaryJobs(root string) ([]sourceJob, error) {
	var jobs []sourceJob
	tsm, e := readObject(filepath.Join(root, "scripts/wow_api/tsm_sources.json"))
	if e != nil {
		return nil, e
	}
	for _, v := range arr(tsm["files"]) {
		r := obj(v)
		up := s(r["path"])
		if up == "" {
			prefix := "https://raw.githubusercontent.com/" + s(tsm["repository"]) + "/" + s(tsm["commit"]) + "/"
			if !strings.HasPrefix(s(r["url"]), prefix) {
				return nil, fmt.Errorf("TSM source URL is not pinned")
			}
			up = strings.TrimPrefix(s(r["url"]), prefix)
		}
		jobs = append(jobs, sourceJob{"tsm", s(tsm["repository"]), s(tsm["commit"]), up, s(r["name"]), r})
	}
	b := obj(tsm["blizzard"])
	jobs = append(jobs, sourceJob{"tsm", s(b["repository"]), s(b["commit"]), "version.txt", "blizzard-version.txt", M{"sha256": b["version_sha256"]}})
	native, e := readObject(filepath.Join(root, "scripts/wow_api/native_crosscheck_sources.json"))
	if e != nil {
		return nil, e
	}
	for _, v := range arr(native["files"]) {
		r := obj(v)
		jobs = append(jobs, sourceJob{"native-crosscheck", s(r["repository"]), s(r["commit"]), s(r["upstream_path"]), s(r["name"]), r})
	}
	backup, e := readObject(filepath.Join(root, "scripts/wow_api/native_backup_sources.json"))
	if e != nil {
		return nil, e
	}
	for _, v := range arr(backup["files"]) {
		r := obj(v)
		jobs = append(jobs, sourceJob{"native-backup", s(backup["repository"]), s(backup["commit"]), s(backup["path"]) + "/" + s(r["name"]), s(r["name"]), r})
	}
	return jobs, nil
}
func syncSecondary(ctx context.Context, root string, online bool) error {
	jobs, e := secondaryJobs(root)
	if e != nil {
		return e
	}
	total := 0
	for _, j := range jobs {
		size := maxPayload
		if n, ok := j.row["size"]; ok {
			size = i(n)
		}
		total += size
	}
	if total > maxTotal {
		return fmt.Errorf("locked secondary inputs exceed aggregate limit")
	}
	return parallel(ctx, jobs, func(ctx context.Context, j sourceJob) error {
		if !hex40.MatchString(j.commit) {
			return fmt.Errorf("acquisition requires immutable commit")
		}
		if _, e := safeJoin("/", j.upstream); e != nil {
			return e
		}
		path, e := safeJoin(root, "references/wow-api/"+j.group+"/"+j.name)
		if e != nil {
			return e
		}
		var local []byte
		if exists(path) {
			local, e = readFile(path, maxPayload)
			if e != nil {
				return e
			}
			if !matches(local, j.row) {
				return fmt.Errorf("modified pinned input: %s", path)
			}
		}
		if online || local == nil {
			address := "https://raw.githubusercontent.com/" + j.repo + "/" + j.commit + "/" + escapePath(j.upstream)
			b, e := fetch(ctx, address, maxPayload)
			if e != nil {
				return e
			}
			if strings.HasSuffix(j.name, ".bin.zst") && bytes.HasPrefix(b, []byte("version https://git-lfs.github.com/")) {
				parts := strings.Split(string(b), "\n")
				if len(parts) != 4 || parts[0] != "version https://git-lfs.github.com/spec/v1" || !strings.HasPrefix(parts[1], "oid sha256:") || !strings.HasPrefix(parts[2], "size ") {
					return fmt.Errorf("invalid LFS pointer: %s", j.name)
				}
				var size int
				fmt.Sscanf(parts[2], "size %d", &size)
				if size < 1 || size > maxPayload {
					return fmt.Errorf("LFS payload exceeds limit")
				}
				b, e = fetch(ctx, "https://media.githubusercontent.com/media/"+j.repo+"/"+j.commit+"/"+escapePath(j.upstream), maxPayload)
				if e != nil {
					return e
				}
				if len(b) != size || sha(b) != strings.TrimPrefix(parts[1], "oid sha256:") {
					return fmt.Errorf("Git LFS identity mismatch: %s", j.name)
				}
			}
			if !matches(b, j.row) {
				return fmt.Errorf("locked input identity mismatch: %s", address)
			}
			if local != nil && !bytes.Equal(local, b) {
				return fmt.Errorf("online pinned input differs: %s", path)
			}
			if local == nil {
				return atomicWrite(path, b, 0644)
			}
		}
		return nil
	})
}
func escapePath(path string) string {
	parts := strings.Split(path, "/")
	for n := range parts {
		parts[n] = url.PathEscape(parts[n])
	}
	return strings.Join(parts, "/")
}
func syncPrimary(ctx context.Context, p wowapi.Project) error {
	lock, e := p.Lock("sources.json")
	if e != nil {
		return e
	}
	var jobs []sourceJob
	for _, group := range keys(lock) {
		source := obj(lock[group])
		for _, v := range arr(source["files"]) {
			r := obj(v)
			jobs = append(jobs, sourceJob{group, s(source["repository"]), s(source["commit"]), strings.Trim(s(source["path"])+"/"+s(r["name"]), "/"), s(r["name"]), r})
		}
	}
	return parallel(ctx, jobs, func(ctx context.Context, j sourceJob) error {
		path := p.InputPath(j.group, j.name)
		if _, e := safeJoin(filepath.Dir(path), filepath.Base(path)); e != nil {
			return e
		}
		if exists(path) {
			b, e := readFile(path, maxFile)
			if e != nil {
				return e
			}
			if len(b) != i(j.row["size"]) || blob(b) != s(j.row["sha"]) {
				return fmt.Errorf("modified pinned source: %s", path)
			}
			return nil
		}
		if j.group == "blizzard" || j.group == "blizzard_support" || j.group == "blizzard_version" {
			return fmt.Errorf("missing verified native input: %s", path)
		}
		if !hex40.MatchString(j.commit) {
			return fmt.Errorf("community source requires immutable revision")
		}
		b, e := fetch(ctx, "https://raw.githubusercontent.com/"+j.repo+"/"+j.commit+"/"+escapePath(j.upstream), maxFile)
		if e != nil {
			return e
		}
		if len(b) != i(j.row["size"]) || blob(b) != s(j.row["sha"]) {
			return fmt.Errorf("community blob identity mismatch: %s", j.name)
		}
		return atomicWrite(path, b, 0644)
	})
}

// Bootstrap populates ordinary caches only from a verified native projection and pinned community inputs.
func Bootstrap(ctx context.Context, root, snapshot string) (wowapi.Project, error) {
	p, e := Projection(root, snapshot, filepath.Join(root, "scripts/wow_api/client_sources.json"))
	if e != nil {
		return p, e
	}
	if e = syncPrimary(ctx, p); e != nil {
		return p, e
	}
	ordinary := wowapi.NewProject(root)
	lock, e := p.Lock("sources.json")
	if e != nil {
		return p, e
	}
	pending := map[string][]byte{}
	for _, group := range []string{"blizzard", "blizzard_support", "blizzard_version"} {
		for _, value := range arr(obj(lock[group])["files"]) {
			row := obj(value)
			name := s(row["name"])
			b, e := readFile(p.InputPath(group, name), maxFile)
			if e != nil {
				return p, e
			}
			target := ordinary.InputPath(group, name)
			if exists(target) {
				old, e := readFile(target, maxFile)
				if e != nil {
					return p, e
				}
				if !bytes.Equal(old, b) {
					return p, fmt.Errorf("modified pinned native cache: %s", target)
				}
			} else {
				pending[target] = b
			}
		}
	}
	for _, target := range keys(pending) {
		if e = atomicWrite(target, pending[target], 0644); e != nil {
			return p, e
		}
	}
	if e = syncSecondary(ctx, root, false); e != nil {
		return p, e
	}
	return p, nil
}

// Cache files remain optional, integrity-checked projections rather than authorities.
func seedCaches(origin, target string) error {
	p := wowapi.NewProject(origin)
	lock, e := p.Lock("sources.json")
	if e != nil {
		return e
	}
	for _, group := range keys(lock) {
		if group == "blizzard" || group == "blizzard_support" || group == "blizzard_version" {
			continue
		}
		for _, value := range arr(obj(lock[group])["files"]) {
			row := obj(value)
			name := s(row["name"])
			src := p.InputPath(group, name)
			if !exists(src) {
				continue
			}
			b, e := readFile(src, maxFile)
			if e != nil {
				return e
			}
			if len(b) != i(row["size"]) || blob(b) != s(row["sha"]) {
				return fmt.Errorf("modified pinned community cache: %s", name)
			}
			dst := wowapi.NewProject(target).InputPath(group, name)
			if e = atomicWrite(dst, b, 0644); e != nil {
				return e
			}
		}
	}
	jobs, e := secondaryJobs(target)
	if e != nil {
		return e
	}
	for _, j := range jobs {
		relative := "references/wow-api/" + j.group + "/" + j.name
		src, e := safeJoin(origin, relative)
		if e != nil {
			return e
		}
		if !exists(src) {
			continue
		}
		b, e := readFile(src, maxPayload)
		if e != nil {
			return e
		}
		if !matches(b, j.row) {
			return fmt.Errorf("modified pinned cross-check cache: %s", j.name)
		}
		dst, e := safeJoin(target, relative)
		if e != nil {
			return e
		}
		if e = atomicWrite(dst, b, 0644); e != nil {
			return e
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
