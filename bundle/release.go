package bundle

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

// GetDiceYAML 拉取 dice.yml
func (b *Bundle) GetDiceYAML(releaseID string, workspace ...string) (*diceyml.DiceYaml, error) {
	host, err := b.urls.DiceHub()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var buf bytes.Buffer
	r, err := hc.Get(host).Path(fmt.Sprintf("/api/releases/%s/actions/get-dice", releaseID)).
		Header("Accept", "application/x-yaml").
		Header("Internal-Client", "true").
		Do().Body(&buf)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		err = errors.Errorf("failed to fetch diceYAML, releaseID: %s, status-code: %d, raw body: %s",
			releaseID, r.StatusCode(), buf.String())
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	// do parse
	dice, err := diceyml.New(buf.Bytes(), true)
	if err != nil {
		return nil, apierrors.ErrInvoke.InvalidState(err.Error())
	}
	if len(workspace) > 0 {
		if err := dice.MergeEnv(workspace[0]); err != nil {
			return nil, apierrors.ErrInvoke.InvalidState(err.Error())
		}
	}
	return dice, nil
}

// GetRelease 获取release信息
func (b *Bundle) GetRelease(releaseID string) (*apistructs.ReleaseGetResponseData, error) {
	host, err := b.urls.DiceHub()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var releaseResp apistructs.ReleaseGetResponse
	r, err := hc.Get(host).Path(fmt.Sprintf("/api/releases/%s", releaseID)).
		Header("Accept", "application/json").
		Header("Internal-Client", "true").
		Do().JSON(&releaseResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if !r.IsOK() {
		err = errors.Errorf("failed to fetch release info, releaseID: %s, status-code: %d",
			releaseID, r.StatusCode())
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	return &(releaseResp.Data), nil
}

func (b *Bundle) ListReleases(req apistructs.ReleaseListRequest) (*apistructs.ReleaseListResponseData, error) {
	host, err := b.urls.DiceHub()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var releasesResp apistructs.ReleaseListResponse
	resp, err := hc.Get(host).Path("/api/releases").
		Header("Internal-Client", "true").
		Params(req.ConvertToQueryParams()).
		Do().JSON(&releasesResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !releasesResp.Success {
		return nil, toAPIError(resp.StatusCode(), releasesResp.Error)
	}
	return &releasesResp.Data, nil
}

// UpdateReference 更新 release 引用
func (b *Bundle) UpdateReference(releaseID string, increase ...bool) error {
	host, err := b.urls.DiceHub()
	if err != nil {
		return err
	}
	hc := b.hc

	doIncrease := true // default is increase
	if len(increase) > 0 && !increase[0] {
		doIncrease = false
	}
	req := apistructs.ReleaseReferenceUpdateRequest{
		ReleaseID: releaseID,
		Increase:  doIncrease,
	}
	var resp httpserver.Resp
	r, err := hc.Put(host).Path(fmt.Sprintf("/api/releases/%s/reference/actions/change", releaseID)).
		Header("Internal-Client", "true").
		JSONBody(&req).Do().JSON(&resp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() || !resp.Success {
		return toAPIError(r.StatusCode(), resp.Err)
	}
	return nil
}

// IncreaseReference 增加 release 引用
func (b *Bundle) IncreaseReference(releaseID string) error {
	return b.UpdateReference(releaseID, true)
}

// DecreaseReference 减小 release 引用
func (b *Bundle) DecreaseReference(releaseID string) error {
	return b.UpdateReference(releaseID, false)
}
