package release

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/sourcelock"
)

type fileIdentity struct {
	SHA256 string `json:"sha256"`
	Mode   uint32 `json:"mode"`
}

func (id *fileIdentity) UnmarshalJSON(data []byte) error {
	value, err := decodeJSON[struct {
		SHA256 string  `json:"sha256"`
		Mode   *uint32 `json:"mode"`
	}](data)
	if err != nil {
		return err
	}
	if value.Mode == nil || *value.Mode > 0777 || !sourcelock.IsSHA256(value.SHA256) {
		return fmt.Errorf("invalid publication file identity")
	}
	*id = fileIdentity{SHA256: value.SHA256, Mode: *value.Mode}
	return nil
}

type inputState map[string]fileIdentity
type publicationRow struct {
	Path   string        `json:"path"`
	Before *fileIdentity `json:"before"`
	After  *fileIdentity `json:"after"`
	Backup string        `json:"backup"`
}
type journal struct {
	Schema int              `json:"schema"`
	Root   string           `json:"root"`
	Files  []publicationRow `json:"files"`
}

var ownedOutputs = []string{"types/wow/wow.d.tl", "types/wow/wow_env.d.tl", "types/wow/wow_secure_env.d.tl", "types/wow/call_policy.lua", "internal/wowapi/inputs/sources.json", "internal/wowapi/inputs/client_sources.json", "README.md", "examples/minimal/WowTealExample.toc", "package.json"}

func sameIdentity(a, b *fileIdentity) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}
func validFileIdentity(id *fileIdentity) bool {
	return id != nil && sourcelock.IsSHA256(id.SHA256) && id.Mode <= 0777
}
func identityOf(path string) (*fileIdentity, error) {
	if !exists(path) {
		return nil, nil
	}
	b, e := readFile(path, maxPayload)
	if e != nil {
		return nil, e
	}
	st, e := os.Stat(path)
	if e != nil {
		return nil, e
	}
	return &fileIdentity{sha(b), uint32(st.Mode().Perm())}, nil
}
func captureInputs(root string) (inputState, error) {
	state := inputState{}
	total := int64(0)
	add := func(path string) error {
		r, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}
		name := filepath.ToSlash(r)
		if _, ok := state[name]; ok {
			return nil
		}
		id, e := identityOf(path)
		if e != nil {
			return e
		}
		if id == nil {
			return nil
		}
		st, e := os.Stat(path)
		if e != nil {
			return e
		}
		total += st.Size()
		if total > maxTotal {
			return fmt.Errorf("maintained update inputs exceed aggregate limit")
		}
		state[name] = *id
		return nil
	}
	for _, name := range maintainedDirs {
		dir := filepath.Join(root, filepath.FromSlash(name))
		if !exists(dir) {
			continue
		}
		e := filepath.WalkDir(dir, func(path string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if path != dir && d.IsDir() && (d.Name() == "target" || d.Name() == ".git" || d.Name() == "__pycache__") {
				return filepath.SkipDir
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("maintained update input contains symlink: %s", path)
			}
			if d.IsDir() {
				return nil
			}
			return add(path)
		})
		if e != nil {
			return nil, e
		}
	}
	for _, name := range maintainedFiles {
		p, e := safeJoin(root, name)
		if e != nil {
			return nil, e
		}
		if exists(p) {
			if e = add(p); e != nil {
				return nil, e
			}
		}
	}
	return state, nil
}
func requireUnchanged(root string, expected inputState) error {
	actual, e := captureInputs(root)
	if e != nil {
		return e
	}
	if !maps.Equal(actual, expected) {
		var changed []string
		names := map[string]bool{}
		for n := range actual {
			names[n] = true
		}
		for n := range expected {
			names[n] = true
		}
		for _, n := range keys(names) {
			if actual[n] != expected[n] {
				changed = append(changed, n)
			}
		}
		return fmt.Errorf("active inputs changed during generation: %s", strings.Join(changed[:min(10, len(changed))], ", "))
	}
	return nil
}
func updateLock(root string) (func(), error) {
	dir := filepath.Join(root, "build/wow-update")
	if e := os.MkdirAll(dir, 0755); e != nil {
		return nil, e
	}
	path, e := safeJoin(root, "build/wow-update/update.lock")
	if e != nil {
		return nil, e
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if e != nil {
		return nil, e
	}
	if e = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); e != nil {
		f.Close()
		return nil, fmt.Errorf("another catalogue update owns this checkout: %w", e)
	}
	return func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN); _ = f.Close() }, nil
}
func outputNames(workspace string) ([]string, error) {
	data, e := readFile(filepath.Join(workspace, "package.json"), maxFile)
	if e != nil {
		return nil, e
	}
	m, e := consumer.ParsePackageManifest(data)
	if e != nil {
		return nil, e
	}
	v := m.Version
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(v) {
		return nil, fmt.Errorf("invalid checked package version")
	}
	names := append([]string{}, ownedOutputs...)
	names = append(names, "rockspecs/wow-tl-type-"+v+"-1.rockspec.in")
	sortStrings(names)
	for k, n := range names {
		if n == "package.json" {
			names = append(append(names[:k], names[k+1:]...), n)
			break
		}
	}
	return names, nil
}
func rollback(root, transaction string, rows []publicationRow) error {
	var failures []error
	for k := len(rows) - 1; k >= 0; k-- {
		row := rows[k]
		path, e := safeJoin(root, row.Path)
		if e != nil {
			failures = append(failures, e)
			continue
		}
		current, e := identityOf(path)
		if e != nil {
			failures = append(failures, e)
			continue
		}
		if sameIdentity(current, row.Before) {
			continue
		}
		if !sameIdentity(current, row.After) {
			failures = append(failures, fmt.Errorf("preserved external edit: %s", row.Path))
			continue
		}
		if row.Before == nil {
			e = os.Remove(path)
		} else {
			backup := filepath.Join(transaction, "backup", row.Backup)
			saved, err := identityOf(backup)
			if err != nil {
				e = err
			} else if saved == nil || saved.SHA256 != row.Before.SHA256 {
				e = fmt.Errorf("publication backup identity mismatch: %s", row.Path)
			} else {
				b, err := readFile(backup, maxPayload)
				if err != nil {
					e = err
				} else {
					e = atomicWrite(path, b, fs.FileMode(row.Before.Mode))
				}
			}
		}
		if e != nil {
			failures = append(failures, fmt.Errorf("%s: %w", row.Path, e))
		}
	}
	return errors.Join(failures...)
}

// publish offers per-file atomic replacement, guarded publication and rollback.
// It does not promise a filesystem-wide atomic commit or power-loss durability.
func publish(ctx context.Context, root, workspace, runDir string, expected inputState, replace func(string, string) error) (changed []string, err error) {
	if e := requireUnchanged(root, expected); e != nil {
		return nil, e
	}
	names, e := outputNames(workspace)
	if e != nil {
		return nil, e
	}
	rows := []publicationRow{}
	transaction := filepath.Join(runDir, "publication")
	for _, name := range names {
		src, e := safeJoin(workspace, name)
		if e != nil {
			return nil, e
		}
		dst, e := safeJoin(root, name)
		if e != nil {
			return nil, e
		}
		after, e := identityOf(src)
		if e != nil {
			return nil, e
		}
		if after == nil {
			return nil, fmt.Errorf("checked candidate missing output: %s", name)
		}
		before, e := identityOf(dst)
		if e != nil {
			return nil, e
		}
		if before != nil {
			after.Mode = before.Mode
		}
		if sameIdentity(before, after) {
			continue
		}
		rows = append(rows, publicationRow{name, before, after, strconv.Itoa(len(rows))})
	}
	if len(rows) == 0 {
		return []string{}, nil
	}
	if e = os.MkdirAll(filepath.Join(transaction, "backup"), 0755); e != nil {
		return nil, e
	}
	if e = os.MkdirAll(filepath.Join(transaction, "staged"), 0755); e != nil {
		return nil, e
	}
	for _, r := range rows {
		if r.Before != nil {
			b, e := readFile(filepath.Join(root, filepath.FromSlash(r.Path)), maxPayload)
			if e != nil {
				return nil, e
			}
			if e = atomicWrite(filepath.Join(transaction, "backup", r.Backup), b, fs.FileMode(r.Before.Mode)); e != nil {
				return nil, e
			}
		}
		b, e := readFile(filepath.Join(workspace, filepath.FromSlash(r.Path)), maxPayload)
		if e != nil {
			return nil, e
		}
		if e = atomicWrite(filepath.Join(transaction, "staged", r.Backup), b, fs.FileMode(r.After.Mode)); e != nil {
			return nil, e
		}
	}
	body, e := jsonBytes(journal{1, root, rows})
	if e != nil {
		return nil, e
	}
	jp := filepath.Join(transaction, "journal.json")
	if e = atomicWrite(jp, body, 0644); e != nil {
		return nil, e
	}
	marker := filepath.Join(root, "build/wow-update/pending.json")
	if exists(marker) {
		return nil, fmt.Errorf("another publication journal exists")
	}
	if e = writeObject(marker, publicationPointer{Journal: jp, SHA256: sha(body)}); e != nil {
		return nil, e
	}
	applied := []publicationRow{}
	defer func() {
		if err != nil {
			failure := rollback(root, transaction, applied)
			_ = writeObject(filepath.Join(transaction, "result.json"), publicationOutcome{Status: "rolled-back", Error: err.Error(), RollbackError: fmt.Sprint(failure)})
			if failure == nil {
				_ = os.Remove(marker)
			} else {
				err = fmt.Errorf("publication rollback requires recovery: %w; initial failure: %v", failure, err)
			}
		}
	}()
	if e = requireUnchanged(root, expected); e != nil {
		return nil, e
	}
	for _, r := range rows {
		if e = ctx.Err(); e != nil {
			return nil, e
		}
		dst, e := safeJoin(root, r.Path)
		if e != nil {
			return nil, e
		}
		current, e := identityOf(dst)
		if e != nil {
			return nil, e
		}
		if !sameIdentity(current, r.Before) {
			return nil, fmt.Errorf("publication destination changed: %s", r.Path)
		}
		if e = os.MkdirAll(filepath.Dir(dst), 0755); e != nil {
			return nil, e
		}
		applied = append(applied, r)
		if e = replace(filepath.Join(transaction, "staged", r.Backup), dst); e != nil {
			return nil, e
		}
		changed = append(changed, r.Path)
	}
	after := inputState{}
	for k, v := range expected {
		after[k] = v
	}
	for _, r := range rows {
		after[r.Path] = *r.After
	}
	if e = requireUnchanged(root, after); e != nil {
		return nil, e
	}
	if e = writeObject(filepath.Join(transaction, "result.json"), publicationOutcome{Status: "complete"}); e != nil {
		return nil, e
	}
	if e = os.Remove(marker); e != nil {
		return nil, e
	}
	return changed, nil
}
func Recover(root string) (PublicationResult, error) {
	unlock, e := updateLock(root)
	if e != nil {
		return PublicationResult{}, e
	}
	defer unlock()
	marker := filepath.Join(root, "build/wow-update/pending.json")
	if !exists(marker) {
		return PublicationResult{Status: "no-pending-publication"}, nil
	}
	pointer, e := readJSON[publicationPointer](marker)
	if e != nil {
		return PublicationResult{}, e
	}
	path := pointer.Journal
	relative, e := filepath.Rel(root, path)
	if e != nil || !strings.HasPrefix(filepath.ToSlash(relative), "build/wow-update/") {
		return PublicationResult{}, fmt.Errorf("publication journal outside owned directory")
	}
	path, e = safeJoin(root, filepath.ToSlash(relative))
	if e != nil {
		return PublicationResult{}, e
	}
	body, e := readFile(path, maxFile)
	if e != nil {
		return PublicationResult{}, e
	}
	if sha(body) != pointer.SHA256 {
		return PublicationResult{}, fmt.Errorf("publication journal identity mismatch")
	}
	j, e := decodeJSON[journal](body)
	if e != nil {
		return PublicationResult{}, e
	}
	if j.Schema != 1 || j.Root != root || len(j.Files) < 1 || len(j.Files) > len(ownedOutputs)+1 {
		return PublicationResult{}, fmt.Errorf("invalid publication journal inventory")
	}
	allowed := map[string]bool{}
	for _, n := range ownedOutputs {
		allowed[n] = true
	}
	seen := map[string]bool{}
	recipe := regexp.MustCompile(`^rockspecs/wow-tl-type-\d+\.\d+\.\d+-1\.rockspec\.in$`)
	for k, row := range j.Files {
		if seen[row.Path] || row.Backup != strconv.Itoa(k) || (!allowed[row.Path] && !recipe.MatchString(row.Path)) || row.After == nil {
			return PublicationResult{}, fmt.Errorf("unowned or duplicate publication journal path")
		}
		if !validFileIdentity(row.After) || (row.Before != nil && !validFileIdentity(row.Before)) {
			return PublicationResult{}, fmt.Errorf("invalid publication file identity")
		}
		seen[row.Path] = true
		dst, e := safeJoin(root, row.Path)
		if e != nil {
			return PublicationResult{}, e
		}
		current, e := identityOf(dst)
		if e != nil {
			return PublicationResult{}, e
		}
		if !sameIdentity(current, row.Before) && !sameIdentity(current, row.After) {
			return PublicationResult{}, fmt.Errorf("recovery preserved external edit: %s", row.Path)
		}
	}
	if e = rollback(root, filepath.Dir(path), j.Files); e != nil {
		return PublicationResult{}, e
	}
	if e = os.Remove(marker); e != nil {
		return PublicationResult{}, e
	}
	return PublicationResult{Status: "restored-prior-publication", Files: len(j.Files)}, nil
}
