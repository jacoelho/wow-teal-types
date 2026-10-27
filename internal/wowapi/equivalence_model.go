package wowapi

import "wow-teal-types/internal/sourcelock"

type EquivalenceReceipt struct {
	BaselineSourceLocksSHA256 string             `json:"baseline_source_locks_sha256"`
	BaselineSources           sourcelock.Sources `json:"baseline_sources"`
	BaselineVersion           string             `json:"baseline_version"`
	LedgerSHA256              string             `json:"ledger_sha256"`
	Policy                    string             `json:"policy"`
	PolicySHA256              map[string]string  `json:"policy_sha256"`
	SchemaVersion             int                `json:"schema_version"`
	TargetSourcesSHA256       string             `json:"target_sources_sha256"`
	TargetVersion             string             `json:"target_version"`
}
type EquivalenceProof struct {
	BaselineVersion    string `json:"baseline_version"`
	CheckedNativeFiles int    `json:"checked_native_files"`
	LedgerSHA256       string `json:"ledger_sha256"`
	Policy             string `json:"policy"`
	ReceiptSHA256      string `json:"receipt_sha256"`
	TargetVersion      string `json:"target_version"`
}
