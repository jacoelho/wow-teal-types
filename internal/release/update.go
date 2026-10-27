package release

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi"
)

func ensureSnapshot(ctx context.Context, root string, lock sourcelock.Snapshot) (string, error) {
	snapshot := filepath.Join(root, "build/wow-client", fmt.Sprint(lock.Build))
	if exists(snapshot) {
		_, e := VerifySnapshot(snapshot, &lock)
		return snapshot, e
	}
	lockData, e := jsonBytes(lock)
	if e != nil {
		return "", e
	}
	snapshot = filepath.Join(root, "build/wow-update/snapshots", sha(lockData))
	if exists(snapshot) {
		_, e = VerifySnapshot(snapshot, &lock)
		return snapshot, e
	}
	_, e = Stage(ctx, root, snapshot, "", lock.Identity(), &lock)
	return snapshot, e
}

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
		data, err := readFile(path, maxPayload)
		if err != nil {
			return err
		}
		if !bytes.Equal(data, files[name]) {
			return fmt.Errorf("stale generated output: %s", name)
		}
	}
	return nil
}
func buildReader(ctx context.Context, root, workspace, runDir string) (string, error) {
	target := filepath.Join(root, "build/wow-update/reader")
	if e := command(ctx, workspace, filepath.Join(runDir, "logs/reader.log"), 10*time.Minute, "cargo", "build", "--locked", "--manifest-path", "scripts/wow_api/tsm_resolved_export/Cargo.toml", "--target-dir", target); e != nil {
		return "", e
	}
	binary := filepath.Join(target, "debug/tsm-resolved-export")
	if !exists(binary) {
		return "", fmt.Errorf("pinned reader build did not produce executable")
	}
	return binary, nil
}
func auditBeforeConversion(p wowapi.Project) error {
	report, e := p.AuditMetadata(nil)
	if e != nil {
		return e
	}
	b, e := wowapi.MetadataJSON(report)
	if e != nil {
		return e
	}
	if e = atomicWrite(filepath.Join(p.Root, "evidence/generated/wow_native_metadata.json"), b, 0644); e != nil {
		return e
	}
	return wowapi.RequireKnownMetadata(report)
}
func generateCandidate(ctx context.Context, root, runDir string, selected consumer.Version) (string, error) {
	lock, e := sourcelock.ReadSnapshot(filepath.Join(root, "scripts/wow_api/client_sources.json"))
	if e != nil {
		return "", e
	}
	baseline := filepath.Join(runDir, "baseline/workspace")
	if e = copyMaintained(root, baseline); e != nil {
		return "", e
	}
	if e = atomicWrite(filepath.Join(baseline, ".wow-version"), []byte(lock.Version+"\n"), 0644); e != nil {
		return "", e
	}
	if e = seedCaches(root, baseline); e != nil {
		return "", e
	}
	snapshot, e := ensureSnapshot(ctx, root, lock)
	if e != nil {
		return "", e
	}
	if _, e = Bootstrap(ctx, baseline, snapshot); e != nil {
		return "", e
	}
	reader, e := buildReader(ctx, root, baseline, runDir)
	if e != nil {
		return "", e
	}
	target := baseline
	if selected.Full != lock.Version {
		targetSnapshot := filepath.Join(root, "build/wow-client", fmt.Sprint(selected.Build))
		if exists(targetSnapshot) {
			record, e := VerifySnapshot(targetSnapshot, nil)
			if e != nil {
				return "", e
			}
			if e = selected.Require(record.Version, record.Build); e != nil {
				return "", e
			}
		} else {
			targetSnapshot = filepath.Join(root, "build/wow-update/snapshots", selected.Full)
			if exists(targetSnapshot) {
				record, e := VerifySnapshot(targetSnapshot, nil)
				if e != nil {
					return "", e
				}
				if e = selected.Require(record.Version, record.Build); e != nil {
					return "", e
				}
			} else {
				identity, e := ResolveBuild(ctx, lock.Region, selected.Build)
				if e != nil {
					return "", e
				}
				if e = selected.Require(identity.Version, identity.Build); e != nil {
					return "", e
				}
				if _, e = Stage(ctx, root, targetSnapshot, "", identity, nil); e != nil {
					return "", e
				}
			}
		}
		target = filepath.Join(runDir, "candidate/workspace")
		if e = copyMaintained(baseline, target); e != nil {
			return "", e
		}
		if e = seedCaches(baseline, target); e != nil {
			return "", e
		}
		versionBytes, e := readFile(filepath.Join(root, ".wow-version"), 64)
		if e != nil {
			return "", e
		}
		if e = atomicWrite(filepath.Join(target, ".wow-version"), versionBytes, 0644); e != nil {
			return "", e
		}
		if _, e = stageNative(target, targetSnapshot); e != nil {
			return "", e
		}
		snapshot = targetSnapshot

	} else {
		versionBytes, e := readFile(filepath.Join(root, ".wow-version"), 64)
		if e != nil {
			return "", e
		}
		if e = atomicWrite(filepath.Join(target, ".wow-version"), versionBytes, 0644); e != nil {
			return "", e
		}
	}
	p, e := Bootstrap(ctx, target, snapshot)
	if e != nil {
		return "", e
	}
	p = p.WithReader(reader)
	if e = prepareContractReview(p, baseline, selected.Full != lock.Version); e != nil {
		return "", e
	}

	artifacts, e := p.Generate(ctx)
	if e != nil {
		if len(artifacts.Files) > 0 {
			_ = writeCandidateArtifacts(target, artifacts.Files)
		}
		return "", e
	}
	if e = writeCandidateArtifacts(target, artifacts.Files); e != nil {
		return "", e
	}
	manifest, e := wowapi.ReadManifest(filepath.Join(target, "evidence/generated/wow_manifest.json"))
	if e != nil {
		return "", e
	}
	coverage, e := p.Coverage(ctx, snapshot, manifest)
	if e != nil {
		return "", e
	}
	if e = writeCandidateArtifacts(target, coverage); e != nil {
		return "", e
	}
	metadata, e := prepareMetadata(target)
	if e != nil {
		return "", e
	}
	if e = writeCandidateArtifacts(target, metadata); e != nil {
		return "", e
	}
	for _, args := range [][]string{{"verify"}, {"check-types"}, {"check", filepath.Join(target, "examples/minimal")}} {
		log := filepath.Join(runDir, "logs/consumer-"+args[0]+".log")
		f, e := os.Create(log)
		if e != nil {
			if er := os.MkdirAll(filepath.Dir(log), 0755); er != nil {
				return "", er
			}
			f, e = os.Create(log)
		}
		if e != nil {
			return "", e
		}
		e = consumer.Run(ctx, target, args, f, f)
		ce := f.Close()
		if e != nil {
			return "", fmt.Errorf("consumer checks failed; see %s: %w", log, e)
		}
		if ce != nil {
			return "", ce
		}
	}
	again, e := p.Generate(ctx)
	if e != nil {
		return "", e
	}
	if e = compareArtifacts(target, again.Files); e != nil {
		return "", e
	}
	metadata, e = prepareMetadata(target)
	if e != nil {
		return "", e
	}
	if e = compareArtifacts(target, metadata); e != nil {
		return "", e
	}
	return target, nil
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
	workspace, e := generateCandidate(ctx, root, runDir, selected)
	if e != nil {
		_ = writeObject(filepath.Join(runDir, "failure.json"), updateFailure{Status: "failed", WoWVersion: selected.Full, Error: e.Error(), Candidate: runDir, ActiveOutputs: "not published"})
		return UpdateResult{}, fmt.Errorf("catalogue update stopped; candidate retained at %s: %w", runDir, e)
	}
	changed, e := publish(ctx, root, workspace, runDir, expected, os.Rename)
	if e != nil {
		_ = writeObject(filepath.Join(runDir, "failure.json"), updateFailure{Status: "failed", Error: e.Error(), Candidate: runDir})
		return UpdateResult{}, e
	}
	status := "unchanged"
	if len(changed) > 0 {
		status = "updated"
	}
	result := UpdateResult{Status: status, WoWVersion: selected.Full, ChangedFiles: changed, Candidate: workspace, Checks: "type, secret, metadata, source and determinism checks passed"}
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
	reader, err := buildReader(ctx, root, root, filepath.Join(root, "build/wow-generate"))
	if err != nil {
		return wowapi.CatalogueStatistics{}, err
	}
	p = p.WithReader(reader)
	artifacts, err := p.Generate(ctx)
	if err != nil {
		return wowapi.CatalogueStatistics{}, err
	}
	if err = validateDeclarations(ctx, root, artifacts.Files); err != nil {
		return wowapi.CatalogueStatistics{}, err
	}
	if err = compareArtifacts(root, artifacts.Files); err != nil {
		return wowapi.CatalogueStatistics{}, err
	}
	manifest, err := wowapi.ReadManifest(filepath.Join(root, "evidence/generated/wow_manifest.json"))
	if err != nil {
		return wowapi.CatalogueStatistics{}, err
	}
	coverage, err := p.Coverage(ctx, snapshot, manifest)
	if err != nil {
		return wowapi.CatalogueStatistics{}, err
	}
	if err = compareArtifacts(root, coverage); err != nil {
		return wowapi.CatalogueStatistics{}, err
	}
	metadata, err := prepareMetadata(root)
	if err != nil {
		return wowapi.CatalogueStatistics{}, err
	}
	if err = compareArtifacts(root, metadata); err != nil {
		return wowapi.CatalogueStatistics{}, err
	}
	return artifacts.Statistics, nil
}

// Every rejected candidate retains its own full metadata before source/contract
// equivalence can stop review. Copied reports from the baseline are insufficient.
func prepareContractReview(p wowapi.Project, baseline string, transition bool) error {
	if e := auditBeforeConversion(p); e != nil {
		return e
	}
	if transition {
		receipt, e := wowapi.CreateEquivalence(baseline, p.Root)
		if e != nil {
			return e
		}
		return writeObject(filepath.Join(p.Root, equivalenceReceipt), receipt)
	}
	return nil
}
