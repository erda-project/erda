package constraints

// PodLabelsForAffinity used to set the parameters required for podantiaffinity,
type PodLabelsForAffinity struct {
	PodLabels map[string]string
	// Required use 'required' or 'preferred'
	Required bool
}

// Constraints each executor's constraints should implement it
type Constraints interface {
	IsConstraints()
}

type HostnameUtil interface {
	IPToHostname(ip string) string
}
