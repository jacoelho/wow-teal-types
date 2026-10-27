package wowapi

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestTSMCrosscheckMatchesReviewedCohort(t *testing.T) {
	root := catalogueTestRoot(t)
	p := NewProject(root)
	if _, e := os.Stat(p.InputPath("blizzard", "BuildDocumentation.lua")); e != nil {
		t.Skip("native inputs absent; run bootstrap first")
	}
	var c *catalogue
	var constructionError error
	func() {
		defer captureError(&constructionError)
		c = newCatalogue(p)
		c.loadQuestie()
		c.loadNative()
		c.loadFrameXML()
		c.applyOverrides()
	}()
	if constructionError != nil {
		t.Fatal(constructionError)
	}
	projection, e := c.finalize(t.Context())
	if e != nil {
		t.Fatal(e)
	}
	outputs := projection.render()
	actual, e := p.crosscheck(t.Context(), c, c.manifest(), projection, outputs)
	if actual != nil {
		data, marshalErr := jsonPretty(actual)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		debug := filepath.Join(root, "build/go-evidence-parity")
		if mkdirErr := os.MkdirAll(debug, 0755); mkdirErr != nil {
			t.Fatal(mkdirErr)
		}
		if writeErr := os.WriteFile(filepath.Join(debug, "wow_tsm_crosscheck.json"), data, 0644); writeErr != nil {
			t.Fatal(writeErr)
		}
		if e == nil {
			want, readErr := os.ReadFile(filepath.Join(root, "evidence/generated/wow_tsm_crosscheck.json"))
			if readErr != nil {
				t.Fatal(readErr)
			}
			if !bytes.Equal(data, want) {
				t.Errorf("TSM report differs from reviewed cohort; derived report saved in %s", debug)
			}
		}
	}
	if e != nil {
		t.Fatal(e)
	}
	// Finalization owns emitted signatures, including secondary-native evidence.
	// Changing discarded builder projection state must not change the audit.
	methods := asMap(asMap(c.defs["DurationTextBindingObject"])["methods"])
	originalMethod := asMap(methods["GetDuration"])
	if originalMethod == nil {
		t.Fatal("recorded cohort lacks DurationTextBindingObject.GetDuration")
	}
	changedMethod := mapCopy(originalMethod)
	changedMethod["resolved_signature"] = "function(self): string"
	methods["GetDuration"] = changedMethod
	rechecked, recheckErr := p.crosscheck(t.Context(), c, c.manifest(), projection, outputs)
	methods["GetDuration"] = originalMethod
	if recheckErr != nil {
		t.Fatalf("audit recomputed finalized signatures: %v", recheckErr)
	}
	originalReport, err := jsonPretty(actual)
	if err != nil {
		t.Fatal(err)
	}
	recheckedReport, err := jsonPretty(rechecked)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(recheckedReport, originalReport) {
		t.Fatal("discarded builder signature changed finalized audit report")
	}
	// Permission review is independent of the emitted type graph. Change a real
	// native contract field whose policy is retained outside Teal signatures.
	originalGraph := actual["declaration_graph_sha256"]
	changedSymbol := ""
	for _, doc := range c.nativeDocs {
		functions := asList(doc.Metadata["Functions"])
		if len(functions) == 0 {
			continue
		}
		function := asMap(functions[0])
		policy := "NotAllowed"
		if function["SecretArguments"] == policy {
			policy = "AllowedWhenUntainted"
		}
		function["SecretArguments"] = policy
		owner := asString(doc.Metadata["Namespace"])
		if doc.Metadata["Type"] == "ScriptObject" {
			owner = scriptName(asString(doc.Metadata["Name"]))
		}
		changedSymbol = asString(function["Name"])
		if owner != "" {
			changedSymbol = owner + "." + changedSymbol
		}
		break
	}
	if changedSymbol == "" {
		t.Fatal("recorded cohort has no native function")
	}
	changed, rejection := p.crosscheck(t.Context(), c, c.manifest(), projection, outputs)
	if rejection == nil || changed == nil {
		t.Fatal("changed native permission was approved or report was lost")
	}
	if changed["declaration_graph_sha256"] != originalGraph {
		t.Fatal("permission-only change unexpectedly altered Teal graph")
	}
	decisions := asList(asMap(changed["reconciliation"])["decisions"])
	found := false
	for _, v := range decisions {
		row := asMap(v)
		if row["symbol"] == changedSymbol {
			found = true
			if row["state"] != "changed-contract" {
				t.Fatalf("changed permission state = %v", row["state"])
			}
		}
	}
	if !found {
		t.Fatal("changed native contract disappeared from reconciliation")
	}
}
