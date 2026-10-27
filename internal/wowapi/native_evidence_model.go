package wowapi

import (
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi/literal"
)

type NativeEvidenceSources struct {
	sourcelock.Crosscheck
	Backup                sourcelock.Backup     `json:"backup"`
	NativeObservations    ObservationLock       `json:"native_observations"`
	NativeBehaviorReviews BehaviorReviewSources `json:"native_behavior_reviews"`
}
type backupContext struct {
	Documents            []string `json:"documents"`
	DefinitionComparison string   `json:"definition_comparison"`
	Scope                string   `json:"scope"`
}
type nativeCoverage struct {
	Corroborated                []string `json:"corroborated"`
	NativeObservationScope      string   `json:"native_observation_scope"`
	NotIndependentlyEstablished []string `json:"not_independently_established"`
}
type NativeEvidence struct {
	Sources           NativeEvidenceSources            `json:"sources"`
	BackupContext     backupContext                    `json:"backup_context"`
	Methods           map[string]*NativeMethodEvidence `json:"methods,omitempty"`
	ReceiverHierarchy map[string]receiverHierarchy     `json:"receiver_hierarchy"`
	Statistics        map[string]int                   `json:"statistics"`
	Coverage          nativeCoverage                   `json:"coverage"`
}
type primaryObservation struct {
	Source                string   `json:"source"`
	Line                  int      `json:"line"`
	Call                  string   `json:"call"`
	ObservedArgumentCount int      `json:"observed_argument_count"`
	ArgumentRoles         []string `json:"argument_roles"`
	Scope                 string   `json:"scope"`
}
type backupMethodEvidence struct {
	Source             string           `json:"source"`
	MetadataComparison string           `json:"metadata_comparison"`
	Security           securityEvidence `json:"security"`
	Scope              string           `json:"scope"`
}
type NativeMethodEvidence struct {
	Symbol                                  string               `json:"symbol"`
	Family                                  string               `json:"family"`
	Coverage                                string               `json:"coverage"`
	NativeSource                            string               `json:"native_source"`
	NativeMetadata                          literal.Object       `json:"native_metadata"`
	SecondarySource                         string               `json:"secondary_source"`
	SecondaryReceiver                       string               `json:"secondary_receiver"`
	SecondaryMethod                         SecondaryMethod      `json:"secondary_method"`
	SignatureComparison                     secondaryComparison  `json:"signature_comparison"`
	ReceiverHierarchy                       receiverHierarchy    `json:"receiver_hierarchy"`
	NativeSecurityNotRepresentedBySecondary securityEvidence     `json:"native_security_not_represented_by_secondary"`
	BackupMetadata                          backupMethodEvidence `json:"backup_metadata"`
	EmittedSignatures                       []string             `json:"emitted_signatures"`
	PrimaryObservation                      *primaryObservation  `json:"primary_observation"`
	NativeInvocationObservation             *CaptureInvocation   `json:"native_invocation_observation"`
	ReviewedBehaviorObservation             *BehaviorReview      `json:"reviewed_behavior_observation,omitempty"`
	Limits                                  []string             `json:"limits"`
}
