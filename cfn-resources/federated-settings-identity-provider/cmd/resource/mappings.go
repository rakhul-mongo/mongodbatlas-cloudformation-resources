package resource

import (
	admin20250312010 "go.mongodb.org/atlas-sdk/v20250312010/admin"

	"github.com/mongodb/mongodbatlas-cloudformation-resources/util"
)

const (
	ProtocolSAML = "SAML"
	ProtocolOIDC = "OIDC"
)

// GetFederatedSettingsIdentityProviderModel converts an Atlas API response to a CFN Model.
// It preserves identifier fields from currentModel when provided.
func GetFederatedSettingsIdentityProviderModel(api *admin20250312010.FederationIdentityProvider, currentModel *Model) *Model {
	var model *Model
	if currentModel != nil {
		model = currentModel
	} else {
		model = &Model{}
	}

	if api == nil {
		return model
	}

	// Computed identifiers (Terraform exposes both okta_idp_id and idp_id)
	oktaID := api.GetOktaIdpId()
	model.OktaIdpId = &oktaID
	idpID := api.GetId()
	model.IdpId = &idpID

	// Common fields (Terraform always sets these)
	displayName := api.GetDisplayName()
	model.Name = &displayName
	issuerURI := api.GetIssuerUri()
	model.IssuerUri = &issuerURI
	protocol := api.GetProtocol()
	model.Protocol = &protocol

	// These are set in Terraform outside protocol-specific branches
	description := api.GetDescription()
	model.Description = &description
	authorizationType := api.GetAuthorizationType()
	model.AuthorizationType = &authorizationType
	idpType := api.GetIdpType()
	model.IdpType = &idpType

	// Protocol-specific fields: Terraform sets only the fields relevant to the protocol.
	if protocol == ProtocolSAML {
		requestBinding := api.GetRequestBinding()
		model.RequestBinding = &requestBinding
		responseSignatureAlgorithm := api.GetResponseSignatureAlgorithm()
		model.ResponseSignatureAlgorithm = &responseSignatureAlgorithm
		model.SsoDebugEnabled = api.SsoDebugEnabled
		ssoURL := api.GetSsoUrl()
		model.SsoUrl = &ssoURL
		status := api.GetStatus()
		model.Status = &status

		// Also set associated_domains for both protocols (Terraform does this outside branch)
		model.AssociatedDomains = api.GetAssociatedDomains()
	} else if protocol == ProtocolOIDC {
		audience := api.GetAudience()
		model.Audience = &audience
		clientID := api.GetClientId()
		model.ClientId = &clientID
		groupsClaim := api.GetGroupsClaim()
		model.GroupsClaim = &groupsClaim
		model.RequestedScopes = api.GetRequestedScopes()
		userClaim := api.GetUserClaim()
		model.UserClaim = &userClaim

		// Also set associated_domains for both protocols (Terraform does this outside branch)
		model.AssociatedDomains = api.GetAssociatedDomains()
	}

	return model
}

func expandOIDCCreateRequest(model *Model) *admin20250312010.FederationOidcIdentityProviderUpdate {
	// Terraform always passes pointers (even for empty strings) for most OIDC fields.
	// We mirror that behavior for parity.
	var associatedDomains []string
	if model.AssociatedDomains != nil {
		associatedDomains = model.AssociatedDomains
	} else {
		associatedDomains = []string{}
	}
	var requestedScopes []string
	if model.RequestedScopes != nil {
		requestedScopes = model.RequestedScopes
	} else {
		requestedScopes = []string{}
	}

	return &admin20250312010.FederationOidcIdentityProviderUpdate{
		Audience:          util.Pointer(util.SafeString(model.Audience)),
		AssociatedDomains: &associatedDomains,
		AuthorizationType: util.Pointer(util.SafeString(model.AuthorizationType)),
		ClientId:          util.Pointer(util.SafeString(model.ClientId)),
		Description:       util.Pointer(util.SafeString(model.Description)),
		DisplayName:       util.Pointer(util.SafeString(model.Name)),
		GroupsClaim:       util.Pointer(util.SafeString(model.GroupsClaim)),
		IdpType:           util.Pointer(util.SafeString(model.IdpType)),
		IssuerUri:         util.Pointer(util.SafeString(model.IssuerUri)),
		Protocol:          util.Pointer(util.SafeString(model.Protocol)),
		RequestedScopes:   &requestedScopes,
		UserClaim:         util.Pointer(util.SafeString(model.UserClaim)),
	}
}

func normalizeIdForBackCompat(model *Model) {
	// Terraform historically used okta_idp_id for the ID encoding; CFN primary identifier uses IdpId.
	// To support resources created/imported with OktaIdpId, treat it as IdpId when IdpId isn't set.
	if model != nil && model.IdpId == nil && model.OktaIdpId != nil {
		model.IdpId = model.OktaIdpId
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
