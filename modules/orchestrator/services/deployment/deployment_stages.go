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

package deployment

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
)

func (d *Deployment) DeployStageAddons(deploymentID uint64) (*apistructs.DeploymentCreateResponseDTO, error) {
	fsm := NewFSMContext(deploymentID, d.db, d.evMgr, d.bdl, d.addon, d.migration, d.encrypt, d.resource)
	if err := fsm.Load(); err != nil {
		return nil, errors.Wrapf(err, "failed to load fsm, deployment: %d, (%v)", deploymentID, err)
	}
	var err error
	switch fsm.Deployment.Status {
	case apistructs.DeploymentStatusWaitApprove:

	case apistructs.DeploymentStatusWaiting:
		err = fsm.continueWaiting()
		if err != nil {
			break
		}
		fallthrough
	case apistructs.DeploymentStatusDeploying:
		switch fsm.Deployment.Phase {
		case apistructs.DeploymentPhaseInit:
			err = fsm.continuePhasePreAddon()
			if err != nil {
				break
			}
			fallthrough
		case apistructs.DeploymentPhaseAddon:
			err = fsm.continuePhaseAddon()
			if err != nil {
				break
			}
		}
	default:
		return nil, errors.Errorf("DeployStageAddons: deployment status != WAITING or DEPLOYING or WAITAPPROVE")
	}
	return &apistructs.DeploymentCreateResponseDTO{
		DeploymentID:  fsm.deploymentID,
		ApplicationID: fsm.Runtime.ApplicationID,
		RuntimeID:     fsm.Runtime.ID,
	}, err
}

func (d *Deployment) DeployStageServices(deploymentID uint64) (*apistructs.DeploymentCreateResponseDTO, error) {
	fsm := NewFSMContext(deploymentID, d.db, d.evMgr, d.bdl, d.addon, d.migration, d.encrypt, d.resource)
	if err := fsm.Load(); err != nil {
		return nil, errors.Wrapf(err, "failed to load fsm, deployment: %d, (%v)", deploymentID, err)
	}
	var err error
	switch fsm.Deployment.Status {
	case apistructs.DeploymentStatusDeploying:
		switch fsm.Deployment.Phase {
		case apistructs.DeploymentPhaseScript:
			err = fsm.continuePhasePreService()
			if err != nil {
				break
			}
			fallthrough
		case apistructs.DeploymentPhaseService:
			err = fsm.continuePhaseService()
			if err != nil {
				break
			}
		}
	default:
		return nil, errors.Errorf("DeployStageServices: deployment status != DEPLOYING")
	}
	return &apistructs.DeploymentCreateResponseDTO{
		DeploymentID:  deploymentID,
		ApplicationID: fsm.Runtime.ApplicationID,
		RuntimeID:     fsm.Runtime.ID,
	}, err
}

func (d *Deployment) DeployStageDomains(deploymentID uint64) (*apistructs.DeploymentCreateResponseDTO, error) {
	fsm := NewFSMContext(deploymentID, d.db, d.evMgr, d.bdl, d.addon, d.migration, d.encrypt, d.resource)
	if err := fsm.Load(); err != nil {
		return nil, errors.Wrapf(err, "failed to load fsm, deployment: %d, (%v)", deploymentID, err)
	}
	var err error
	switch fsm.Deployment.Status {
	case apistructs.DeploymentStatusDeploying:
		switch fsm.Deployment.Phase {
		case apistructs.DeploymentPhaseRegister:
			err = fsm.continuePhaseRegister()
			if err != nil {
				break
			}
			fallthrough
		case apistructs.DeploymentPhaseCompleted:
			err = fsm.continuePhaseCompleted()
			if err != nil {
				break
			}
		}
	default:
		return nil, errors.Errorf("DeployStageDomains: deployment status != DEPLOYING")
	}
	return &apistructs.DeploymentCreateResponseDTO{
		DeploymentID:  fsm.deploymentID,
		ApplicationID: fsm.Runtime.ApplicationID,
		RuntimeID:     fsm.Runtime.ID,
	}, err
}
