/*
Copyright 2023 The Bestchains Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Copyright 2023 The Bestchains Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, component
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package market

import (
	"github.com/bestchains/bestchains-contracts/contracts/nonce"
	"github.com/bestchains/bestchains-contracts/library/context"
)

// EventCreateRepository represents the event of creating a new repository.
type EventCreateRepository struct {
	ID    string `json:"id,omitempty"`
	Owner string `json:"owner,omitempty"`
	URL   string `json:"url,omitempty"`
}

// Repository represents a component repository.
type Repository struct {
	ID    string `json:"id,omitempty"`
	Owner string `json:"owner,omitempty"`
	URL   string `json:"url,omitempty"`
}

// EventPublishComponent represents the event of publishing a new component.
type EventPublishComponent struct {
	UUID    string `json:"uuid,omitempty"`
	Version string `json:"version,omitempty"`
	Owner   string `json:"owner,omitempty"`
	RepoID  string `json:"repoId,omitempty"`
}

// EventEndorseComponent represents the event of endorsing a component.
type EventEndorseComponent struct {
	RepoID  string `json:"repoID,omitempty"`
	UUID    string `json:"uuid,omitempty"`
	Version string `json:"version,omitempty"`
}

// Component represents a software component.
type Component struct {
	UUID     string    `json:"uuid,omitempty"`
	Owner    string    `json:"owner,omitempty"`
	RepoID   string    `json:"repoId,omitempty"`
	Versions []Version `json:"versions,omitempty"`
}

// Version represents a version of a software component.
type Version struct {
	Number string `json:"number,omitempty"`

	// SignedByRepoOwner indicates whether the version has been signed by the repository owner.
	SignedByRepoOwner bool `json:"signedByRepoOwner,omitempty"`
}

// EventApplyLicense represents the event of applying for a license.
type EventApplyLicense struct {
	LicenseID   string `json:"licenseId,omitempty"`
	RepoID      string `json:"repoId,omitempty"`
	ComponentID string `json:"componentId,omitempty"`
}

// EventChangeLicenseStatus represents the event of changing the status of a license.
type EventChangeLicenseStatus struct {
	LicenseID string        `json:"licenseId,omitempty"`
	PreStatus LicenseStatus `json:"preStatus,omitempty"`
	NewStatus LicenseStatus `json:"newStatus,omitempty"`
}

// LicenseStatus represents the status of a license.
type LicenseStatus string

const (
	// Applying is the status of a license when it is being applied for.
	Applying LicenseStatus = "Applying"
	// Issued is the status of a license when it has been issued.
	Issued LicenseStatus = "Issued"
	// Rejected is the status of a license when it has been rejected.
	Rejected LicenseStatus = "Rejected"
	// Consumed is the status of a license when it has been consumed.
	Consumed LicenseStatus = "Consumed"
)

// License represents a component license
type License struct {
	// ID is the unique identifier for the license
	ID string `json:"id,omitempty"`

	// RepoID is the ID of the repository that the license applies to
	RepoID string `json:"repoID,omitempty"`

	// ComponentID is the ID of the component that the license applies to
	ComponentID string `json:"componentId,omitempty"`

	// IssueBy is the entity that issued the license
	IssueBy string `json:"issueBy,omitempty"`

	// IssueTo is the entity that the license was issued to
	IssueTo string `json:"owner,omitempty"`

	// Status is the current status of the license
	Status LicenseStatus `json:"status,omitempty"`

	// EncActivationCode is the encrypted activation code for the license
	// It is encrypted by the owner's public credential
	EncActivationCode string `json:"encActivationCode,omitempty"`
}

// IMarket is the interface for the market service
type IMarket interface {
	// INonce is an interface for generating nonces
	nonce.INonce

	// Initialize initializes the market service
	Initialize(ctx context.ContextInterface) error

	// Repository

	// CreateRepo creates a new repository
	CreateRepo(ctx context.ContextInterface, msg context.Message, url string) (string, error)

	// UpdateRepo updates an existing repository
	// Only the owner of the repository can update it
	UpdateRepo(ctx context.ContextInterface, msg context.Message, repoID string, newUrl string) error

	// GetRepos returns a list of all repositories
	GetRepos(ctx context.ContextInterface) ([]Repository, error)

	// Component

	// PublishComponent publishes a new software component to a repository
	PublishComponent(ctx context.ContextInterface, msg context.Message, repoID string, swUUID string, version string) (string, error)

	// EndorseComponent endorses a software component in a repository
	// Only the owner of the repository can endorse a component
	EndorseComponent(ctx context.ContextInterface, msg context.Message, repoID string, swUUID string, version string) error

	// GetComponents returns a list of all components in a repository
	GetComponents(ctx context.ContextInterface, repoID string) ([]Component, error)

	// License

	// GetLicenses returns a list of all licenses
	GetLicenses(ctx context.ContextInterface) ([]License, error)

	// ApplyLicense applies a license to a component in a repository
	// Any entity can apply a license
	ApplyLicense(ctx context.ContextInterface, msg context.Message, repoID string, componentID string) (string, error)

	// IssueLicense issues a license
	// Only the owner of the component can issue it
	IssueLicense(ctx context.ContextInterface, msg context.Message, licenseID string, encActivationCode string) error

	// RejectLicense rejects a license
	// Only the owner of the component can reject it
	RejectLicense(ctx context.ContextInterface, msg context.Message, licenseID string) error

	// ConsumeLicense consumes a license
	// Only the owner of the license can consume it
	ConsumeLicense(ctx context.ContextInterface, msg context.Message, licenseID string) error
}
