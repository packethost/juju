package packet

import (
	"github.com/juju/errors"

	"github.com/juju/juju/cloud"
	"github.com/juju/juju/environs"
)

type environProviderCredentials struct{}

const (
	credAPIToken  = "packet-api-token"
	credProjectID = "packet-project-id"
)

// CredentialSchemas is part of the environs.ProviderCredentials interface.
func (environProviderCredentials) CredentialSchemas() map[cloud.AuthType]cloud.CredentialSchema {
	return map[cloud.AuthType]cloud.CredentialSchema{
		cloud.ApiTokenAuthType: {
			{
				credAPIToken, cloud.CredentialAttr{
					Description: "Packet API Token",
					Hidden:      true,
				},
			}, {
				credProjectID, cloud.CredentialAttr{Description: "UUID of your project in Packet"},
			},
		},
	}
}

// DetectCredentials is part of the environs.ProviderCredentials interface.
func (environProviderCredentials) DetectCredentials() (*cloud.CloudCredential, error) {
	return nil, errors.NotFoundf("credentials")
}

// FinalizeCredential is part of the environs.ProviderCredentials interface.
func (environProviderCredentials) FinalizeCredential(_ environs.FinalizeCredentialContext, args environs.FinalizeCredentialParams) (*cloud.Credential, error) {
	return &args.Credential, nil
}
