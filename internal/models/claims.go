package models

// Claims represents the validated JWT claims from Zitadel.
type Claims struct {
	Sub string `json:"sub"`
	Email string `json:"email"`
	OrgID string `json:"urn:zitadel:iam:org:id"`
	OrgDomain string `json:"urn:zitadel:iam:user:resourceowner:primary_domain"`
	Roles map[string]interface{} `json:"urn:zitadel:iam:org:project:roles"`
}
