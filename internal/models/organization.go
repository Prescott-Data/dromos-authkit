package models

// OrgRole represents organization-level role types.
type OrgRole string

// Organization role constants matching Zitadel role keys.
const (
	// OrgRoleOwner has full control over the organization.
	OrgRoleOwner OrgRole = "orgowner"

	// OrgRoleAdmin can manage members and most organization settings.
	OrgRoleAdmin OrgRole = "orgadmin"

	// OrgRoleMember has standard access to organization resources.
	OrgRoleMember OrgRole = "orgmember"

	// OrgRoleViewer has read-only access to organization resources.
	OrgRoleViewer OrgRole = "orgviewer"
)
