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

package market

import (
	"encoding/hex"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/bestchains/bestchains-contracts/contracts/nonce"
	"github.com/bestchains/bestchains-contracts/library"
	"github.com/bestchains/bestchains-contracts/library/context"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"golang.org/x/crypto/sha3"
)

var (
	RepoKeyPrefix = "repo"
	// sofware~repoId~swid
	ComponentKeyPrefix = "component"
	LicenseKeyPrefix   = "license"
)

// MarketContract implements IMarket interface
type MarketContract struct {
	contractapi.Contract
	nonce.INonce
}

var _ IMarket = new(MarketContract)

// NewMarketContract creates a new instance of MarketContract with the specified nonce.
func NewMarketContract(nonceContract nonce.INonce) *MarketContract {
	// Create a new MarketContract instance.
	marketContract := new(MarketContract)

	// Set the name of the MarketContract instance.
	marketContract.Name = "org.bestchains.com.MarketContract"

	// Set the nonce of the MarketContract instance.
	marketContract.INonce = nonceContract

	// Set the transaction context handler of the MarketContract instance.
	marketContract.TransactionContextHandler = new(context.Context)

	// Set the before transaction handler of the MarketContract instance.
	marketContract.BeforeTransaction = context.BeforeTransaction

	// Return the newly created MarketContract instance.
	return marketContract
}

func (lc *MarketContract) Initialize(ctx context.ContextInterface) error {
	return errors.New("Initialize: not implemented")
}

// CreateRepo creates a new Repository object and stores it on the ledger.
//
// ctx: is the context interface for calling chaincode functions, msg is the
// message: being processed, and url is the URL for the new repository.
// The function returns the ID of the new repository and an error, if any.
func (lc *MarketContract) CreateRepo(ctx context.ContextInterface, msg context.Message, url string,
) (string, error) {
	// Increment nonce
	curr, err := lc.INonce.Increment(ctx, ctx.MsgSender().String())
	if err != nil {
		return "", errors.Wrap(err, "MarketContract: failed to increment nonce")
	}

	// Calculate repository ID
	// nonce here is greater than 0
	id, err := calculateRepoID(ctx, ctx.MsgSender(), curr-1, url)
	if err != nil {
		return "", errors.Wrap(err, "MarketContract: failed to calculate repo id")
	}

	// Create new Repository object
	repo := &Repository{
		ID:    id,
		Owner: ctx.MsgSender().String(),
		URL:   url,
	}
	val, _ := json.Marshal(repo)

	// Save value to ledger
	repoKey, err := ctx.GetStub().CreateCompositeKey(RepoKeyPrefix, []string{id})
	if err != nil {
		return "", errors.Wrap(err, "MarketContract: invalid composite RepoKey")
	}
	err = ctx.GetStub().PutState(repoKey, []byte(val))
	if err != nil {
		return "", errors.Wrap(err, "MarketContract: failed to put Repository")
	}

	// Emit event
	err = ctx.EmitEvent("CreateRepository", &EventCreateRepository{
		Owner: repo.Owner,
		ID:    id,
		URL:   repo.URL,
	})
	if err != nil {
		return "", errors.Wrap(err, "MarketContract: failed to emit event CreateRepository")
	}

	return id, nil
}

// UpdateRepo updates the repository URL for the given repoID.
//
// ctx: the context interface.
// msg: the message context.
// repoID: the ID of the repository to be updated.
// newUrl: the new URL to replace the current one.
// error: returns an error if the function fails to update the repository.
func (lc *MarketContract) UpdateRepo(
	ctx context.ContextInterface, msg context.Message,
	repoID string, newUrl string,
) error {
	// TODO: implement the code for updating the repository URL.
	return errors.New("UpdateRepo: not implemented")
}

// GetRepos gets all the repositories from the state by iterating over all the
// composite keys with a prefix of "repo". It returns a slice of Repositories or
// an error if any occurred.
//
// ctx: the context of the function with a GetStub() method that returns an
// interface that can read/write to the state database.
// ([]Repository, error): a slice of Repositories and/or an error if any occurred.
func (lc *MarketContract) GetRepos(ctx context.ContextInterface) ([]Repository, error) {
	// Get all the repos from state by iterating over all composite keys with a prefix of "repo"
	itr, err := ctx.GetStub().GetStateByPartialCompositeKey(RepoKeyPrefix, []string{})
	if err != nil {
		return nil, errors.Wrap(err, "MarketContract: failed to get repos")
	}
	defer itr.Close()

	// Create an empty slice of repositories to store the results
	repos := make([]Repository, 0)

	// Iterate over all composite keys and unmarshal the values into Repository structs
	for itr.HasNext() {
		kv, err := itr.Next()
		if err != nil {
			return nil, errors.Wrap(err, "MarketContract: failed to get next iteration key")
		}
		repo := new(Repository)
		err = json.Unmarshal(kv.Value, &repo)
		if err != nil {
			return nil, errors.Wrap(err, "MarketContract: failed to unmarshal repo")
		}
		repos = append(repos, *repo)
	}

	return repos, nil
}

// PublishComponent publishes a component with a specific version to a repository.
// Only when repo owner verified this swid, the component can be really published.
//
// Parameters:
//   - ctx: the context interface.
//   - msg: the context message.
//   - repoID: the ID of the repository to which the component belongs.
//   - swUUID: the ID of the component to be published.
//   - version: the version of the component to be published.
//
// Returns:
//   - string: an empty string.
//   - error: an error if there is an issue with creating a composite key.
func (lc *MarketContract) PublishComponent(ctx context.ContextInterface, msg context.Message, repoID string, swUUID string, version string) (string, error) {
	var err error

	if swUUID == "" || version == "" {
		return "", errors.New("PublishComponent: invalid input")
	}

	// Increment the nonce.
	_, err = lc.INonce.Increment(ctx, ctx.MsgSender().String())
	if err != nil {
		return "", errors.Wrap(err, "MarketContract: failed to increment nonce")
	}

	// Create composite key.
	swKey, err := ctx.GetStub().CreateCompositeKey(ComponentKeyPrefix, []string{repoID, swUUID})
	if err != nil {
		return "", errors.Wrap(err, "MarketContract: invalid composite ComponentKey")
	}

	// Get the component.
	swBytes, err := ctx.GetStub().GetState(swKey)
	if err != nil {
		return "", errors.Wrap(err, "MarketContract: failed to get Component")
	}

	sw := new(Component)
	if swBytes != nil {
		// Unmarshal the component.
		err = json.Unmarshal(swBytes, sw)
		if err != nil {
			return "", errors.Wrap(err, "MarketContract: failed to unmarshal Component")
		}

		// Check if the sender is the owner of the component.
		if sw.Owner == ctx.MsgSender().String() {
			return "", errors.New("PublishComponent: only component owner can publish a new version")
		}

		// Check if the version already exists.
		for _, versioned := range sw.Versions {
			if versioned.Number == version {
				return "", errors.New("PublishComponent: version conflict")
			}
		}
	} else {
		// Create a new component.
		sw.UUID = swUUID
		sw.Owner = ctx.MsgSender().String()
		sw.RepoID = repoID
		sw.Versions = append(sw.Versions, Version{Number: version})
	}

	// Marshal the component and put it in the state.
	val, _ := json.Marshal(sw)
	err = ctx.GetStub().PutState(swKey, []byte(val))
	if err != nil {
		return "", errors.Wrap(err, "MarketContract: failed to put Component")
	}

	// Emit the PublishComponent event.
	err = ctx.EmitEvent("PublishComponent", &EventPublishComponent{
		UUID:    sw.UUID,
		Owner:   sw.Owner,
		Version: version,
		RepoID:  repoID,
	})
	if err != nil {
		return "", errors.Wrap(err, "MarketContract: failed to emit event PublishComponent")
	}

	return "", nil
}

// EndorseComponent endorses a specific version of a component in a repository.
// It checks if the caller is the owner of the repository and if the version hasn't been endorsed yet.
// If the checks pass, it sets the SignedByRepoOwner flag to true for the selected version and emits an EndorseComponent event.
func (lc *MarketContract) EndorseComponent(ctx context.ContextInterface, msg context.Message, repoID string, swUUID string, version string) error {
	var err error

	// Increment nonce
	_, err = lc.INonce.Increment(ctx, ctx.MsgSender().String())
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to increment nonce")
	}

	// Check if repository exists
	repoKey, _ := ctx.GetStub().CreateCompositeKey(RepoKeyPrefix, []string{repoID})
	repoBytes, _ := ctx.GetStub().GetState(repoKey)
	if repoBytes == nil {
		return errors.New("EndorseComponent: repo not found")
	}
	repo := new(Repository)
	err = json.Unmarshal(repoBytes, repo)
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to unmarshal repo")
	}

	// Check if caller is the owner of the repository
	if repo.Owner != ctx.MsgSender().String() {
		return errors.New("EndorseComponent: only repo owner can endorse a component")
	}

	// Check if component exists
	swKey, _ := ctx.GetStub().CreateCompositeKey(ComponentKeyPrefix, []string{repoID, swUUID})
	swBytes, _ := ctx.GetStub().GetState(swKey)
	if swBytes == nil {
		return errors.New("EndorseComponent: component not found")
	}
	sw := new(Component)
	err = json.Unmarshal(swBytes, sw)
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to unmarshal Component")
	}

	// Find the selected version and set the SignedByRepoOwner flag to true
	for _, versioned := range sw.Versions {
		if versioned.Number == version {
			if versioned.SignedByRepoOwner {
				return errors.New("EndorseComponent: this version already endorsed")
			}
			versioned.SignedByRepoOwner = true
			val, _ := json.Marshal(sw)
			err = ctx.GetStub().PutState(swKey, []byte(val))
			if err != nil {
				return errors.Wrap(err, "MarketContract: failed to put Component")
			}

			// Emit EndorseComponent event
			ctx.EmitEvent("EndorseComponent", &EventEndorseComponent{
				UUID: sw.UUID,
			})

			return nil
		}
	}

	return errors.New("EndorseComponent: version not found")
}

// GetComponents retrieves all components from a repository using the provided context and repository ID.
func (lc *MarketContract) GetComponents(ctx context.ContextInterface, repoID string) ([]Component, error) {
	// List all components within a repo.
	itr, err := ctx.GetStub().GetStateByPartialCompositeKey(ComponentKeyPrefix, []string{repoID})
	if err != nil {
		return nil, errors.Wrap(err, "MarketContract: failed to get components")
	}
	defer itr.Close()

	var sws []Component
	for itr.HasNext() {
		res, err := itr.Next()
		if err != nil {
			return nil, errors.Wrap(err, "MarketContract: failed to get next component")
		}
		sw := new(Component)
		err = json.Unmarshal(res.Value, sw)
		if err != nil {
			return nil, errors.Wrap(err, "MarketContract: failed to unmarshal Component")
		}
		sws = append(sws, *sw)
	}

	return sws, nil
}

// ApplyLicense applies a license to a component for a specific repository.
// It returns the license ID if successful, and an error otherwise.
func (lc *MarketContract) ApplyLicense(ctx context.ContextInterface, msg context.Message, repoID string, componentID string) (string, error) {
	// increment sender's nonce
	curr, err := lc.INonce.Increment(ctx, ctx.MsgSender().String())
	if err != nil {
		return "", errors.Wrap(err, "MarketContract: failed to increment nonce")
	}

	// calculate license id
	licenseID, err := calculateLicenseID(ctx, ctx.MsgSender(), curr-1, repoID, componentID)
	if err != nil {
		return "", errors.Wrap(err, "MarketContract: failed to calculate license id")
	}

	// init license key
	licenseKey, err := ctx.GetStub().CreateCompositeKey(LicenseKeyPrefix, []string{licenseID})
	if err != nil {
		return "", errors.Wrap(err, "MarketContract: failed to create license key")
	}

	// init license
	license := &License{
		ID:          licenseID,
		RepoID:      repoID,
		ComponentID: componentID,
		IssueBy:     ctx.MsgSender().String(),
		IssueTo:     ctx.MsgSender().String(),
		Status:      Applying,
	}

	// store license with status Applying
	val, _ := json.Marshal(license)
	err = ctx.GetStub().PutState(licenseKey, []byte(val))
	if err != nil {
		return "", errors.Wrap(err, "MarketContract: failed to put License")
	}

	return licenseID, nil
}

// GetLicenses returns a list of all licenses stored on the ledger.
// It takes a context interface as an argument and returns a slice of Licenses
// and an error, if any.
func (lc *MarketContract) GetLicenses(ctx context.ContextInterface) ([]License, error) {
	// list all licenses
	itr, err := ctx.GetStub().GetStateByPartialCompositeKey(LicenseKeyPrefix, []string{})
	if err != nil {
		return nil, errors.Wrap(err, "MarketContract: failed to get licenses")
	}
	defer itr.Close()

	var lics []License
	for itr.HasNext() {
		res, err := itr.Next()
		if err != nil {
			return nil, errors.Wrap(err, "MarketContract: failed to get next license")
		}

		// unmarshal the license JSON into a License object
		lic := new(License)
		err = json.Unmarshal(res.Value, lic)
		if err != nil {
			return nil, errors.Wrap(err, "MarketContract: failed to unmarshal License")
		}

		// append the license to the slice
		lics = append(lics, *lic)
	}

	return lics, nil
}

// IssueLicense issues a license for a component
func (lc *MarketContract) IssueLicense(ctx context.ContextInterface, msg context.Message, licenseID string, encActivationCode string) error {
	// increment sender's nonce
	_, err := lc.INonce.Increment(ctx, ctx.MsgSender().String())
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to increment nonce")
	}

	// get license from licenseID
	licenseKey, err := ctx.GetStub().CreateCompositeKey(LicenseKeyPrefix, []string{licenseID})
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to create license key")
	}
	licenseBytes, err := ctx.GetStub().GetState(licenseKey)
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to get license")
	}
	if licenseBytes == nil {
		return errors.New("MarketContract: license not found")
	}
	lic := new(License)
	err = json.Unmarshal(licenseBytes, lic)
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to unmarshal License")
	}

	// get the relevent component
	componentKey, _ := ctx.GetStub().CreateCompositeKey(ComponentKeyPrefix, []string{lic.RepoID, lic.ComponentID})
	componentBytes, err := ctx.GetStub().GetState(componentKey)
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to get component")
	}
	if componentBytes == nil {
		return errors.New("MarketContract: component not found")
	}
	sw := new(Component)
	err = json.Unmarshal(componentBytes, sw)
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to unmarshal Component")
	}

	// check whether message sender is the component owner
	if ctx.MsgSender().String() != sw.Owner {
		return errors.New("MarketContract: message sender is not the component owner")
	}

	// set license status to `Issued`,then store it to database
	preStatus := lic.Status
	lic.Status = Issued
	lic.IssueBy = ctx.MsgSender().String()

	val, _ := json.Marshal(lic)
	err = ctx.GetStub().PutState(licenseKey, []byte(val))
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to put License")
	}

	// emit event EventChangeLicenseStatus
	ctx.EmitEvent("ChangeLicenseStatus", &EventChangeLicenseStatus{
		LicenseID: lic.ID,
		PreStatus: preStatus,
		NewStatus: Issued,
	})

	return nil
}
func (lc *MarketContract) RejectLicense(ctx context.ContextInterface, msg context.Message, licenseID string) error {
	// increment sender's nonce
	_, err := lc.INonce.Increment(ctx, ctx.MsgSender().String())
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to increment nonce")
	}

	// get license from licenseID
	licenseKey, err := ctx.GetStub().CreateCompositeKey(LicenseKeyPrefix, []string{licenseID})
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to create license key")
	}
	licenseBytes, err := ctx.GetStub().GetState(licenseKey)
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to get license")
	}
	if licenseBytes == nil {
		return errors.New("MarketContract: license not found")
	}
	lic := new(License)
	err = json.Unmarshal(licenseBytes, lic)
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to unmarshal License")
	}

	// get the relevent component
	componentKey, _ := ctx.GetStub().CreateCompositeKey(ComponentKeyPrefix, []string{lic.RepoID, lic.ComponentID})
	componentBytes, err := ctx.GetStub().GetState(componentKey)
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to get component")
	}
	if componentBytes == nil {
		return errors.New("MarketContract: component not found")
	}
	sw := new(Component)
	err = json.Unmarshal(componentBytes, sw)
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to unmarshal Component")
	}
	// check whether message sender is the component owner
	if ctx.MsgSender().String() != sw.Owner {
		return errors.New("MarketContract: message sender is not the component owner")
	}

	// set license status to `Issued`,then store it to database
	preStatus := lic.Status
	lic.Status = Rejected
	lic.IssueBy = ctx.MsgSender().String()

	val, _ := json.Marshal(lic)
	err = ctx.GetStub().PutState(licenseKey, []byte(val))
	if err != nil {
		return errors.Wrap(err, "MarketContract: failed to put License")
	}

	// emit event EventChangeLicenseStatus
	ctx.EmitEvent("ChangeLicenseStatus", &EventChangeLicenseStatus{
		LicenseID: lic.ID,
		PreStatus: preStatus,
		NewStatus: Rejected,
	})

	return nil
}

func (lc *MarketContract) ConsumeLicense(ctx context.ContextInterface, msg context.Message, licenseID string) error {
	panic("not implemented")
}

// calculateRepoID calculates the id of a repository given its owner, nonce, and URL.
func calculateRepoID(ctx context.ContextInterface, owner library.Address, nonce uint64, repoUrl string) (string, error) {
	// Concatenate the owner address and nonce.
	data := append(owner.Bytes(), library.NewCounter(nonce).Bytes()...)
	// Append the repository URL to the data.
	data = append(data, []byte(repoUrl)...)
	// Hash the resulting data using SHA3-256.
	hashedPubKey := sha3.Sum256(data)
	// Return the last 20 bytes of the hash as a hexadecimal string.
	return hex.EncodeToString(hashedPubKey[12:]), nil
}

// calculateLicenseID calculates the license ID for a given applicant, nonce, repoID, and componentID.
// It returns the license ID as a string and an error (if any).
func calculateLicenseID(ctx context.ContextInterface, applicant library.Address, nonce uint64, repoID string, componentID string) (string, error) {
	// Combine the applicant bytes, nonce bytes, repoID bytes, and componentID bytes to create the data to be hashed.
	data := append(applicant.Bytes(), library.NewCounter(nonce).Bytes()...)
	data = append(data, []byte(repoID)...)
	data = append(data, []byte(componentID)...)

	// Hash the data using sha3.
	hashedPubKey := sha3.Sum256(data)

	// Return the license ID as a hex-encoded string (starting from the 12th byte of the hash) and nil error.
	return hex.EncodeToString(hashedPubKey[12:]), nil
}
