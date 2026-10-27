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

type sourceDigest struct {
	size         *int
	sha256, blob string
}

func matches(b []byte, digest sourceDigest) bool {
	return len(b) <= maxPayload && (digest.size == nil || len(b) == *digest.size) && (digest.sha256 == "" || sha(b) == digest.sha256) && (digest.blob == "" || blob(b) == digest.blob)
}

type sourceJob struct {
	group, repo, commit, upstream, name string
	digest                              sourceDigest
}

func lockedDigest(row sourcelock.LockedFile) sourceDigest {
	git := row.SHA
	if row.GitBlobSHA1 != "" {
		git = row.GitBlobSHA1
	}
	return sourceDigest{&row.Size, row.SHA256, git}
}
func secondaryJobs(root string) ([]sourceJob, error) {
	var jobs []sourceJob
	tsm, err := sourcelock.ReadTSM(filepath.Join(root, "scripts/wow_api/tsm_sources.json"))
	if err != nil {
		return nil, err
	}
	for _, row := range tsm.Files {
		upstream := row.Path
		if upstream == "" {
			prefix := "https://raw.githubusercontent.com/" + tsm.Repository + "/" + tsm.Commit + "/"
			if !strings.HasPrefix(row.URL, prefix) {
				return nil, fmt.Errorf("TSM source URL is not pinned")
			}
			upstream = strings.TrimPrefix(row.URL, prefix)
		}
		jobs = append(jobs, sourceJob{"tsm", tsm.Repository, tsm.Commit, upstream, row.Name, lockedDigest(row)})
	}
	b := tsm.Blizzard
	jobs = append(jobs, sourceJob{"tsm", b.Repository, b.Commit, "version.txt", "blizzard-version.txt", sourceDigest{sha256: b.VersionSHA256}})
	crosscheck, err := sourcelock.ReadCrosscheck(filepath.Join(root, "scripts/wow_api/native_crosscheck_sources.json"))
	if err != nil {
		return nil, err
	}
	for _, row := range crosscheck.Files {
		jobs = append(jobs, sourceJob{"native-crosscheck", row.Repository, row.Commit, row.UpstreamPath, row.Name, lockedDigest(row)})
	}
	backup, err := sourcelock.ReadBackup(filepath.Join(root, "scripts/wow_api/native_backup_sources.json"))
	if err != nil {
		return nil, err
	}
	for _, row := range backup.Files {
		jobs = append(jobs, sourceJob{"native-backup", backup.Repository, backup.Commit, backup.Path + "/" + row.Name, row.Name, lockedDigest(row)})
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
		if j.digest.size != nil {
			size = *j.digest.size
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
			if !matches(local, j.digest) {
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
			if !matches(b, j.digest) {
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
	lock, e := sourcelock.ReadSources(filepath.Join(p.Root, "scripts/wow_api/sources.json"))
	if e != nil {
		return e
	}
	var jobs []sourceJob
	for _, group := range keys(lock) {
		source := lock[group]
		for _, r := range source.Files {
			jobs = append(jobs, sourceJob{group, source.Repository, source.Commit, strings.Trim(source.Path+"/"+r.Name, "/"), r.Name, sourceDigest{size: &r.Size, blob: r.SHA}})
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
			if len(b) != *j.digest.size || blob(b) != j.digest.blob {
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
		if len(b) != *j.digest.size || blob(b) != j.digest.blob {
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
	lock, e := sourcelock.ReadSources(filepath.Join(p.Root, "scripts/wow_api/sources.json"))
	if e != nil {
		return p, e
	}
	pending := map[string][]byte{}
	for _, group := range []string{"blizzard", "blizzard_support", "blizzard_version"} {
		for _, row := range lock[group].Files {
			name := row.Name
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
	lock, e := sourcelock.ReadSources(filepath.Join(p.Root, "scripts/wow_api/sources.json"))
	if e != nil {
		return e
	}
	for _, group := range keys(lock) {
		if group == "blizzard" || group == "blizzard_support" || group == "blizzard_version" {
			continue
		}
		for _, row := range lock[group].Files {
			name := row.Name
			src := p.InputPath(group, name)
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
		if !matches(b, j.digest) {
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
