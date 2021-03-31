package labelmanager

func (s *LabelManagerImpl) List() map[string]bool {
	return map[string]bool{
		"locked":            false,
		"platform":          false,
		"pack-job":          false,
		"bigdata-job":       false,
		"job":               false,
		"stateful-service":  false,
		"stateless-service": false,
		"workspace-dev":     false,
		"workspace-test":    false,
		"workspace-staging": false,
		"workspace-prod":    false,
		"org-":              true,
		"location-":         true,
	}
}
