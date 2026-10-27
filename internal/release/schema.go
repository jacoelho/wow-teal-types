package release

import "wow-teal-types/internal/sourcelock"

type extractorReceipt struct {
	sourcelock.ClientIdentity
	Installation string                  `json:"installation,omitempty"`
	Files        []sourcelock.ClientFile `json:"files"`
	Manifests    []struct {
		FileDataID uint32 `json:"file_data_id"`
		Path       string `json:"path"`
	} `json:"manifests"`
	ArchiveIndexVerification struct {
		ArchiveCount   int   `json:"archive_count"`
		ArchiveBytes   int64 `json:"archive_bytes"`
		ArchiveEntries int64 `json:"archive_entries"`
		ArchiveBlocks  int64 `json:"archive_blocks"`
	} `json:"archive_index_verification"`
	ArchiveIndexSizeHints []struct {
		Archive      string `json:"archive"`
		DeclaredSize int64  `json:"declared_size"`
		ActualSize   int64  `json:"actual_size"`
	} `json:"archive_index_size_hints"`
}
type projectionFile struct {
	ClientPath *string `json:"client_path"`
	Group      string  `json:"group"`
	Name       string  `json:"name"`
	Source     string  `json:"source"`
}
type projectionReceipt struct {
	Files          []projectionFile `json:"files"`
	SnapshotSHA256 string           `json:"snapshot_sha256"`
}
type PublicationResult struct {
	Files  int    `json:"files,omitempty"`
	Status string `json:"status"`
}
type publicationPointer struct {
	Journal string `json:"journal"`
	SHA256  string `json:"sha256"`
}
type publicationOutcome struct {
	Error         string `json:"error,omitempty"`
	RollbackError string `json:"rollback_error,omitempty"`
	Status        string `json:"status"`
}
type UpdateResult struct {
	Candidate    string   `json:"candidate"`
	ChangedFiles []string `json:"changed_files"`
	Checks       string   `json:"checks"`
	Status       string   `json:"status"`
	WoWVersion   string   `json:"wow_version"`
}
type updateFailure struct {
	ActiveOutputs string `json:"active_outputs,omitempty"`
	Candidate     string `json:"candidate"`
	Error         string `json:"error"`
	Status        string `json:"status"`
	WoWVersion    string `json:"wow_version,omitempty"`
}
type StageResult struct {
	Blockers           []string `json:"blockers"`
	NativeBuild        string   `json:"nativeBuild"`
	ReviewRequired     bool     `json:"reviewRequired"`
	Status             string   `json:"status"`
	ValidationRequired []string `json:"validationRequired"`
	Workspace          string   `json:"workspace"`
}
type stageFailure struct {
	Status    string `json:"status"`
	Workspace string `json:"workspace"`
}
