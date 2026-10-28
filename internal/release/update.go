package release

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi"
)

// writeCandidateArtifacts only writes inside the private workspace owned by Update.
func writeCandidateArtifacts(workspace string, files map[string][]byte) error {
	for _, name := range keys(files) {
		path, err := safeJoin(workspace, name)
		if err != nil {
			return err
		}
		if err = atomicWrite(path, files[name], 0644); err != nil {
			return err
		}
	}
	return nil
}
func compareArtifacts(root string, files map[string][]byte) error {
	for _, name := range keys(files) {
		path, err := safeJoin(root, name)
		if err != nil {
			return err
		}
		data, err := sourcelock.ReadRegular(path, maxPayload)
		if err != nil {
			return err
		}
		if !bytes.Equal(data, files[name]) {
			return fmt.Errorf("stale generated output: %s", name)
		}
	}
	return nil
}

// generateCandidate returns the checked workspace and the names of every output
// it derived there.
func generateCandidate(ctx context.Context, root, runDir string, selected consumer.Version) (string, []string, error) {
	lock, e := sourcelock.ReadSnapshot(wowapi.NewProject(root).ClientSourcesPath())
	if e != nil {
		return "", nil, e
	}
	target := filepath.Join(runDir, "workspace")
	if e = copyMaintained(root, target); e != nil {
		return "", nil, e
	}
	if e = seedCaches(root, target); e != nil {
		return "", nil, e
	}
	f := newFetcher()
	selection := lockedSelection(lock)
	if selected.Full != lock.Version {
		selection = offeredSelection(selected, f, lock.Region)
	}
	snapshot, e := resolveSnapshot(ctx, root, selection, extractor(root, ""))
	if e != nil {
		return "", nil, e
	}
	if selection.lock == nil {
		if e = stageNative(target, snapshot); e != nil {
			return "", nil, e
		}
	}
	p, e := bootstrap(ctx, f, target, snapshot)
	if e != nil {
		return "", nil, e
	}

	artifacts, e := p.Generate(ctx)
	if e != nil {
		return "", nil, e
	}
	if e = writeCandidateArtifacts(target, artifacts.Files); e != nil {
		return "", nil, e
	}
	metadata, e := prepareMetadata(target)
	if e != nil {
		return "", nil, e
	}
	if e = writeCandidateArtifacts(target, metadata); e != nil {
		return "", nil, e
	}
	example := consumer.Selection{Sources: []string{filepath.Join(target, "examples/WowTealExample/src")}}
	checks := []struct {
		name string
		run  func(io.Writer) error
	}{
		{"verify", func(io.Writer) error { _, e := consumer.Verify(target); return e }},
		{"check-types", func(w io.Writer) error { return consumer.CheckTypes(ctx, target, w, w) }},
		{"check", func(w io.Writer) error { return consumer.Check(ctx, target, example, w, w) }},
	}
	for _, check := range checks {
		log := filepath.Join(runDir, "logs/consumer-"+check.name+".log")
		f, e := os.Create(log)
		if e != nil {
			if er := os.MkdirAll(filepath.Dir(log), 0755); er != nil {
				return "", nil, er
			}
			f, e = os.Create(log)
		}
		if e != nil {
			return "", nil, e
		}
		e = check.run(f)
		ce := f.Close()
		if e != nil {
			return "", nil, fmt.Errorf("consumer checks failed; see %s: %w", log, e)
		}
		if ce != nil {
			return "", nil, ce
		}
	}
	again, e := p.Generate(ctx)
	if e != nil {
		return "", nil, e
	}
	if e = compareArtifacts(target, again.Files); e != nil {
		return "", nil, e
	}
	metadata, e = prepareMetadata(target)
	if e != nil {
		return "", nil, e
	}
	if e = compareArtifacts(target, metadata); e != nil {
		return "", nil, e
	}
	outputs := []string{wowapi.SourcesFile, wowapi.ClientSourcesFile}
	outputs = append(outputs, keys(artifacts.Files)...)
	outputs = append(outputs, keys(metadata)...)
	return target, outputs, nil
}

// Update acquires and validates a private candidate before changing active outputs.
func Update(ctx context.Context, root string) (UpdateResult, error) {
	root, e := filepath.Abs(root)
	if e != nil {
		return UpdateResult{}, e
	}
	unlock, e := updateLock(root)
	if e != nil {
		return UpdateResult{}, e
	}
	defer unlock()
	if exists(filepath.Join(root, "build/wow-update/pending.json")) {
		return UpdateResult{}, fmt.Errorf("incomplete publication journal retained; inspect or run wow-types recover")
	}
	selected, e := consumer.ReadVersion(root)
	if e != nil {
		return UpdateResult{}, e
	}
	expected, e := captureInputs(root)
	if e != nil {
		return UpdateResult{}, e
	}
	runDir, e := os.MkdirTemp(filepath.Join(root, "build/wow-update"), selected.Full+"-go-")
	if e != nil {
		return UpdateResult{}, e
	}
	workspace, outputs, e := generateCandidate(ctx, root, runDir, selected)
	if e != nil {
		_ = writeObject(filepath.Join(runDir, "failure.json"), updateFailure{Status: "failed", WoWVersion: selected.Full, Error: e.Error(), Candidate: runDir, ActiveOutputs: "not published"})
		return UpdateResult{}, fmt.Errorf("catalogue update stopped; candidate retained at %s: %w", runDir, e)
	}
	changed, e := publish(ctx, root, workspace, runDir, outputs, expected, os.Rename)
	if e != nil {
		_ = writeObject(filepath.Join(runDir, "failure.json"), updateFailure{Status: "failed", Error: e.Error(), Candidate: runDir})
		return UpdateResult{}, e
	}
	status := "unchanged"
	if len(changed) > 0 {
		status = "updated"
	}
	result := UpdateResult{Status: status, WoWVersion: selected.Full, ChangedFiles: changed, Candidate: workspace, Checks: "type, secret, package, source and determinism checks passed"}
	if e = writeObject(filepath.Join(runDir, "result.json"), result); e != nil {
		return UpdateResult{}, e
	}
	return result, nil
}

// CheckGenerator compares regenerated artifacts with the active cohort without publishing.
func CheckGenerator(ctx context.Context, root, snapshot string) (wowapi.CatalogueStatistics, error) {
	p, err := Bootstrap(ctx, root, snapshot)
	if err != nil {
		return wowapi.CatalogueStatistics{}, err
	}
	artifacts, err := p.Generate(ctx)
	if err != nil {
		return wowapi.CatalogueStatistics{}, err
	}
	if err = compareArtifacts(root, artifacts.Files); err != nil {
		return wowapi.CatalogueStatistics{}, err
	}
	metadata, err := prepareMetadata(root)
	if err != nil {
		return wowapi.CatalogueStatistics{}, err
	}
	if err = compareArtifacts(root, metadata); err != nil {
		return wowapi.CatalogueStatistics{}, err
	}
	var diagnostics bytes.Buffer
	if err = consumer.CheckTypes(ctx, root, &diagnostics, &diagnostics); err != nil {
		return wowapi.CatalogueStatistics{}, fmt.Errorf("generated declaration check: %w\n%s", err, strings.TrimSpace(diagnostics.String()))
	}
	return artifacts.Statistics, nil
}
