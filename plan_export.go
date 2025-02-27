package tfe

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ PlanExports = (*planExports)(nil)

// PlanExports describes all the plan export related methods that the Terraform
// Enterprise API supports.
//
// TFE API docs: https://www.terraform.io/docs/cloud/api/plan-exports.html
type PlanExports interface {
	// Export a plan by its ID with the given options.
	Create(ctx context.Context, options PlanExportCreateOptions) (*PlanExport, error)

	// Read a plan export by its ID.
	Read(ctx context.Context, planExportID string) (*PlanExport, error)

	// Delete a plan export by its ID.
	Delete(ctx context.Context, planExportID string) error

	// Download the data of an plan export.
	Download(ctx context.Context, planExportID string) ([]byte, error)
}

// planExports implements PlanExports.
type planExports struct {
	client *Client
}

// PlanExportDataType represents the type of data exported from a plan.
type PlanExportDataType string

// List all available plan export data types.
const (
	PlanExportSentinelMockBundleV0 PlanExportDataType = "sentinel-mock-bundle-v0"
)

// PlanExportStatus represents a plan export state.
type PlanExportStatus string

// List all available plan export statuses.
const (
	PlanExportCanceled PlanExportStatus = "canceled"
	PlanExportErrored  PlanExportStatus = "errored"
	PlanExportExpired  PlanExportStatus = "expired"
	PlanExportFinished PlanExportStatus = "finished"
	PlanExportPending  PlanExportStatus = "pending"
	PlanExportQueued   PlanExportStatus = "queued"
)

// PlanExportStatusTimestamps holds the timestamps for plan export statuses.
type PlanExportStatusTimestamps struct {
	CanceledAt time.Time `jsonapi:"attr,canceled-at,rfc3339"`
	ErroredAt  time.Time `jsonapi:"attr,errored-at,rfc3339"`
	ExpiredAt  time.Time `jsonapi:"attr,expired-at,rfc3339"`
	FinishedAt time.Time `jsonapi:"attr,finished-at,rfc3339"`
	QueuedAt   time.Time `jsonapi:"attr,queued-at,rfc3339"`
}

// PlanExport represents an export of Terraform Enterprise plan data.
type PlanExport struct {
	ID               string                      `jsonapi:"primary,plan-exports"`
	DataType         PlanExportDataType          `jsonapi:"attr,data-type"`
	Status           PlanExportStatus            `jsonapi:"attr,status"`
	StatusTimestamps *PlanExportStatusTimestamps `jsonapi:"attr,status-timestamps"`
}

// PlanExportCreateOptions represents the options for exporting data from a plan.
type PlanExportCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,plan-exports"`

	// Required: The plan to export.
	Plan *Plan `jsonapi:"relation,plan"`

	// Required: The name of the policy set.
	DataType *PlanExportDataType `jsonapi:"attr,data-type"`
}

// Create a plan export
func (s *planExports) Create(ctx context.Context, options PlanExportCreateOptions) (*PlanExport, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}

	req, err := s.client.newRequest("POST", "plan-exports", &options)
	if err != nil {
		return nil, err
	}

	pe := &PlanExport{}
	err = s.client.do(ctx, req, pe)
	if err != nil {
		return nil, err
	}

	return pe, err
}

// Read a plan export by its ID.
func (s *planExports) Read(ctx context.Context, planExportID string) (*PlanExport, error) {
	if !validStringID(&planExportID) {
		return nil, ErrInvalidPlanExportID
	}

	u := fmt.Sprintf("plan-exports/%s", url.QueryEscape(planExportID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	pe := &PlanExport{}
	err = s.client.do(ctx, req, pe)
	if err != nil {
		return nil, err
	}

	return pe, nil
}

// Delete a plan export by ID.
func (s *planExports) Delete(ctx context.Context, planExportID string) error {
	if !validStringID(&planExportID) {
		return ErrInvalidPlanExportID
	}

	u := fmt.Sprintf("plan-exports/%s", url.QueryEscape(planExportID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// Download a plan export's data. Data is exported in a .tar.gz format.
func (s *planExports) Download(ctx context.Context, planExportID string) ([]byte, error) {
	if !validStringID(&planExportID) {
		return nil, ErrInvalidPlanExportID
	}

	u := fmt.Sprintf("plan-exports/%s/download", url.QueryEscape(planExportID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = s.client.do(ctx, req, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (o PlanExportCreateOptions) valid() error {
	if o.Plan == nil {
		return ErrRequiredPlan
	}
	if o.DataType == nil {
		return ErrRequiredDataType
	}
	return nil
}
