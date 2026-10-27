package release

import "wow-teal-types/internal/sourcelock"

type extractorReceipt struct {
	sourcelock.ClientIdentity
	Files []sourcelock.ClientFile `json:"files"`
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
	NativeBuild string `json:"nativeBuild"`
	Status      string `json:"status"`
	Workspace   string `json:"workspace"`
}
type stageFailure struct {
	Status    string `json:"status"`
	Workspace string `json:"workspace"`
}
