package oauth2

import (
	"testing"

	"github.com/erda-project/erda/apistructs"
)

func TestValidateOAuth2TokenPayload(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
		payload  *apistructs.OAuth2TokenPayload
		wantErr  bool
	}{
		{
			name:     "openapi allow all",
			clientID: "openapi",
			payload: &apistructs.OAuth2TokenPayload{
				AllowAccessAllAPIs: true,
				Metadata:           map[string]string{"User-ID": "1"},
			},
		},
		{
			name:     "pipeline allow all rejected",
			clientID: "pipeline",
			payload:  &apistructs.OAuth2TokenPayload{AllowAccessAllAPIs: true},
			wantErr:  true,
		},
		{
			name:     "pipeline metadata",
			clientID: "pipeline",
			payload: &apistructs.OAuth2TokenPayload{Metadata: map[string]string{
				"User-ID":                "1",
				"Org-ID":                 "2",
				"Internal-Client":        "pipeline",
				"Internal-Action-Client": "pipeline",
				"pipelineID":             "3",
				"taskID":                 "4",
			}},
		},
		{
			name:     "pipeline unknown metadata",
			clientID: "pipeline",
			payload:  &apistructs.OAuth2TokenPayload{Metadata: map[string]string{"Admin": "true"}},
			wantErr:  true,
		},
		{
			name:     "fdp lowercase metadata",
			clientID: "fdp",
			payload: &apistructs.OAuth2TokenPayload{Metadata: map[string]string{
				"uuid":            "u",
				"user-id":         "1",
				"org-id":          "2",
				"internal-client": "fdp",
			}},
		},
		{
			name:     "action runner internal client rejected",
			clientID: "action-runner",
			payload:  &apistructs.OAuth2TokenPayload{Metadata: map[string]string{"Internal-Client": "x"}},
			wantErr:  true,
		},
		{
			name:     "unknown client common metadata",
			clientID: "custom",
			payload:  &apistructs.OAuth2TokenPayload{Metadata: map[string]string{"User-ID": "1", "Org-ID": "2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateOAuth2TokenPayload(tt.clientID, tt.payload); (err != nil) != tt.wantErr {
				t.Errorf("validateOAuth2TokenPayload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
