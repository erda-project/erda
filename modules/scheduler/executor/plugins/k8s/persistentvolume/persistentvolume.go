// Package persistentvolume manipulates the k8s api of persistentvolume object
package persistentvolume

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

// PersistentVolume is the object to manipulate k8s api of persistentVolumeClaim
type PersistentVolume struct {
	addr   string
	client *httpclient.HTTPClient
}

// Option configures a PersistentVolume
type Option func(*PersistentVolume)

// New news a PersistentVolumeClaim
func New(options ...Option) *PersistentVolume {
	pv := &PersistentVolume{}

	for _, op := range options {
		op(pv)
	}

	return pv
}

// WithCompleteParams provides an Option
func WithCompleteParams(addr string, client *httpclient.HTTPClient) Option {
	return func(pv *PersistentVolume) {
		pv.addr = addr
		pv.client = client
	}
}

// Create creates a k8s persistentVolume
func (p *PersistentVolume) Create(pv *apiv1.PersistentVolume) error {
	var b bytes.Buffer

	resp, err := p.client.Post(p.addr).
		Path("/api/v1/persistentvolumes").
		JSONBody(pv).
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to create pv, name: %s, (%v)", pv.Name, err)
	}
	if !resp.IsOK() {
		return errors.Errorf("failed to create pv, name: %s, statuscode: %v, body: %v",
			pv.Name, resp.StatusCode(), b.String())
	}
	return nil
}

// Delete deletes a k8s persistentVolume
func (p *PersistentVolume) Delete(name string) error {
	var b bytes.Buffer
	path := strutil.Concat("/api/v1/persistentvolumes/", name)

	resp, err := p.client.Delete(p.addr).
		Path(path).
		Do().
		Body(&b)

	if err != nil {
		return errors.Errorf("failed to delete pv, name: %s, (%v)", name, err)
	}
	if !resp.IsOK() {
		if resp.IsNotfound() {
			return k8serror.ErrNotFound
		}

		return errors.Errorf("failed to delete pv, name: %s, statuscode: %v, body: %v",
			name, resp.StatusCode(), b.String())
	}
	return nil
}

// List lists a k8s persistentVolumes
func (p *PersistentVolume) List(name string) (apiv1.PersistentVolumeList, error) {
	var b bytes.Buffer
	var list apiv1.PersistentVolumeList
	path := "/api/v1/persistentvolumes/"

	resp, err := p.client.Get(p.addr).
		Path(path).
		Do().
		Body(&b)

	if err != nil {
		return list, errors.Errorf("failed to list related pv, name: %s, (%v)", name, err)
	}
	if !resp.IsOK() {
		if resp.IsNotfound() {
			return list, k8serror.ErrNotFound
		}

		return list, errors.Errorf("failed to list related pv, name: %s, statuscode: %v, body: %v",
			name, resp.StatusCode(), b.String())
	}
	if err := json.Unmarshal(b.Bytes(), &list); err != nil {
		return list, err
	}
	return list, nil
}
