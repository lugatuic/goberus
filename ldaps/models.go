package ldaps

// MemberInfo is a minimal struct representing attributes returned by GetMemberInfo.
type MemberInfo struct {
	DN              string   `json:"distinguishedName,omitempty"`
	CN              string   `json:"cn,omitempty"`
	DisplayName     string   `json:"displayName,omitempty"`
	Mail            string   `json:"mail,omitempty"`
	SAMAccountName  string   `json:"sAMAccountName,omitempty"`
	MemberOf        []string `json:"memberOf,omitempty"`
	Description     string   `json:"description,omitempty"`
	BadPasswordTime string   `json:"badPasswordTime,omitempty"`
}

// UserInfo represents the minimal user registration payload used by AddUser.
type UserInfo struct {
	Username           string `json:"username"`
	Password           string `json:"password"`
	GivenName          string `json:"givenName,omitempty"`
	Surname            string `json:"surname,omitempty"`
	DisplayName        string `json:"displayName,omitempty"`
	Mail               string `json:"mail,omitempty"`
	Phone              string `json:"phone,omitempty"`
	Major              string `json:"major,omitempty"`
	College            string `json:"college,omitempty"`
	Description        string `json:"description,omitempty"`
	OrganizationalUnit string `json:"ou,omitempty"`
}
