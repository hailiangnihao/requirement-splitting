package domain

type ProjectStatus string

const (
	ProjectStatusPlanning ProjectStatus = "planning"
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusPaused   ProjectStatus = "paused"
	ProjectStatusAccepted ProjectStatus = "accepted"
	ProjectStatusLaunched ProjectStatus = "launched"
	ProjectStatusArchived ProjectStatus = "archived"
)

type HealthLevel string

const (
	HealthLevelHealthy    HealthLevel = "healthy"
	HealthLevelAttention  HealthLevel = "attention"
	HealthLevelRisk       HealthLevel = "risk"
	HealthLevelSevereRisk HealthLevel = "severe_risk"
)

type AIDraftStatus string

const (
	AIDraftStatusDraft      AIDraftStatus = "draft"
	AIDraftStatusValidated  AIDraftStatus = "validated"
	AIDraftStatusInvalid    AIDraftStatus = "invalid"
	AIDraftStatusPublished  AIDraftStatus = "published"
	AIDraftStatusSuperseded AIDraftStatus = "superseded"
)

type TaskStatus string

const (
	TaskStatusPendingDev   TaskStatus = "pending_dev"
	TaskStatusDeveloping   TaskStatus = "developing"
	TaskStatusPendingTest  TaskStatus = "pending_test"
	TaskStatusTesting      TaskStatus = "testing"
	TaskStatusPendingCheck TaskStatus = "pending_acceptance"
	TaskStatusAccepted     TaskStatus = "accepted"
	TaskStatusLaunched     TaskStatus = "launched"
)

type TestCaseConfirmationStatus string

const (
	TestCaseConfirmationPending   TestCaseConfirmationStatus = "pending_confirmation"
	TestCaseConfirmationConfirmed TestCaseConfirmationStatus = "confirmed"
	TestCaseConfirmationRejected  TestCaseConfirmationStatus = "rejected"
)

type TestRunReviewStatus string

const (
	TestRunReviewPending TestRunReviewStatus = "pending_review"
	TestRunReviewPassed  TestRunReviewStatus = "passed"
	TestRunReviewFailed  TestRunReviewStatus = "failed"
	TestRunReviewRetest  TestRunReviewStatus = "needs_retest"
	TestRunReviewIgnored TestRunReviewStatus = "ignored"
)

type DefectStatus string

const (
	DefectStatusPendingConfirm    DefectStatus = "pending_confirm"
	DefectStatusPendingFix        DefectStatus = "pending_fix"
	DefectStatusFixing            DefectStatus = "fixing"
	DefectStatusPendingRegression DefectStatus = "pending_regression"
	DefectStatusRegressionPassed  DefectStatus = "regression_passed"
	DefectStatusClosed            DefectStatus = "closed"
	DefectStatusRejected          DefectStatus = "rejected"
)

type ChangeRequestStatus string

const (
	ChangeRequestStatusSubmitted ChangeRequestStatus = "submitted"
	ChangeRequestStatusAnalyzed  ChangeRequestStatus = "analyzed"
	ChangeRequestStatusAccepted  ChangeRequestStatus = "accepted"
	ChangeRequestStatusApplied   ChangeRequestStatus = "applied"
	ChangeRequestStatusRejected  ChangeRequestStatus = "rejected"
)

type AcceptanceStatus string

const (
	AcceptanceStatusPending AcceptanceStatus = "pending"
	AcceptanceStatusPassed  AcceptanceStatus = "passed"
	AcceptanceStatusFailed  AcceptanceStatus = "failed"
	AcceptanceStatusBlocked AcceptanceStatus = "blocked"
)
