// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package config

type Config struct {
	Debug            bool   `env:"DEBUG" default:"false" desc:"enable debug logging"`
	RepoURL          string `env:"HELM_REPO_URL" desc:"helm repo url"`
	Reinstall        bool   `env:"REINSTALL" desc:"reinstall erda comp"`
	Version          string `env:"ERDA_CHART_VERSION" desc:"erda chart version"`
	InstallMode      string `env:"INSTALL_MODE" desc:"install mode, remote or local"`
	RepoMode         string `env:"REPO_MODE" desc:"get chart mode, download or use build in charts"`
	TargetCluster    string `env:"TARGET_CLUSTER" desc:"special when CREDENTIAL_FROM=CLUSTER_MANAGER"`
	ErdaBaseValues   string `env:"ERDA_BASE_VALUES" desc:"provide erda base values"`
	ErdaAddonsValues string `env:"ERDA_ADDONS_VALUES" desc:"provide erda addons values"`
	ErdaValues       string `env:"ERDA_VALUES" desc:"provide erda values"`
	// HELM_NAMESPACE: helm deploy namespace
	// HELM_REPO_URL: helm repo address
	// HELM_REPOSITORY_CONFIG: helm repository store path
	// HELM_REPOSITORY_CACHE: helm charts cache path
}
