package apistructs

// IdentityInfo represents operator identity info.
// Fields will not be json marshal/unmarshal.
type IdentityInfo struct {
	// UserID is user id. It must be provided in some cases.
	// Cannot be null if InternalClient is null.
	// +optional
	UserID string `json:"userID"`

	// InternalClient records the internal client, such as: bundle.
	// Cannot be null if UserID is null.
	// +optional
	InternalClient string `json:"-"`
}

func (info *IdentityInfo) IsInternalClient() bool {
	return info.InternalClient != ""
}

func (info *IdentityInfo) Empty() bool {
	return info.UserID == "" && info.InternalClient == ""
}
