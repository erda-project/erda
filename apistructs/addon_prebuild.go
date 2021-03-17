package apistructs

// AddonPrebuildReq addon prebuild request body
type AddonPrebuildReq struct {
	Name       string            `json:"name"`
	Plan       string            `json:"plan"`
	Type       string            `json:"type"`
	InstanceID string            `json:"instanceId"`
	Config     map[string]string `json:"config"`
	Options    map[string]string `json:"options"`
	Actions    map[string]string `json:"actions"`
}

// AddonPrebuildOverlayReq addon prebuild overlay request body，rds覆盖mysql
type AddonPrebuildOverlayReq struct {
	To   string `json:"to"`
	Type string `json:"type"`
}

// SaveAddonPrebuildReq addon prebuild overlay request body，rds覆盖mysql
type SaveAddonPrebuildReq struct {
	Addons        []AddonPrebuildReq        `json:"addons"`
	AddonsOverlay []AddonPrebuildOverlayReq `json:"addons_overlay"`
}
