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

package license

import (
	"github.com/bestchains/bestchains-contracts/contracts/nonce"
	"github.com/bestchains/bestchains-contracts/library/context"
)

type Repository struct {
	Owner string

	ID  string
	URL string
}

type Software struct {
	Owner string

	ID      string
	Version string
	RepoID  string
}

type LicenseStatus string

const (
	Applying LicenseStatus = "Applying"
	Issued   LicenseStatus = "Issued"
	Rejected LicenseStatus = "Rejected"
	Consumed LicenseStatus = "Consumed"
)

type License struct {
	Issuer string
	Owner  string

	ID          string
	SoftwareID  string
	ExpiredDate uint64
	Status      string

	// TO BE DECIDE
	// EncActivationCode is the encrypted activation code
	// - encrypted by the owner's public credential
	EncActivationCode string
}

type ILicense interface {
	nonce.INonce
	Initialize(ctx context.ContextInterface) error
	// Repository
	// Any
	CreateRepo(ctx context.ContextInterface, msg context.Message, url string) (string, error)
	// Only Repo Owner
	UpdateRepo(ctx context.ContextInterface, msg context.Message, repoID string, newUrl string) error
	// Any
	GetRepos(ctx context.ContextInterface) ([]Repository, error)

	// Software
	// Any
	PublishSoftware(ctx context.ContextInterface, msg context.Message, repoID string, version string) error
	// Only repo owner
	UpgradeSoftware(ctx context.ContextInterface, msg context.Message, softwareID string, version string) error

	// License
	// Any
	GetLicense(ctx context.ContextInterface, licenseID string) (License, error)
	GetLicenses(ctx context.ContextInterface) ([]License, error)

	// Any
	ApplyLicense(ctx context.ContextInterface, msg context.Message, softwareID string) (string, error)

	// License Owner(Software Owner)
	IssueLicense(ctx context.ContextInterface, licenseID string, encActivationCode string) error
	RejectLicense(ctx context.ContextInterface, licenseID string) error

	// License Owner
	// Consume license by license owner
	ConsumeLicense(ctx context.ContextInterface, msg context.Message, licenseID string) error
}
