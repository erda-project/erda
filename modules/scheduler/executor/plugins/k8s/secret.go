package k8s

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
)

// NewImageSecret 新建 image pull secret
// 1, 创建该 namespace 下的 imagePullSecret
// 2, 将这个 secret 加到该 namespace 的 serviceaccount 中去
func (k *Kubernetes) NewImageSecret(namespace string) error {
	// 集群初始化的时候会在 default namespace 下创建一个拉镜像的 secret
	s, err := k.secret.Get("default", AliyunRegistry)
	if err != nil {
		return err
	}

	mysecret := &apiv1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: namespace,
		},
		Data:       s.Data,
		StringData: s.StringData,
		Type:       s.Type,
	}

	if err := k.secret.Create(mysecret); err != nil {
		return err
	}

	return k.updateDefaultServiceAccountForImageSecret(namespace, s.Name)
}

// NewImageSecret 新建 image pull secret
// 1, 创建该 namespace 下的 imagePullSecret
// 2, 将需要认证的image的secret添加到该namespace的secret中
// 3, 将这个 secret 加到该 namespace 的 serviceaccount 中去
func (k *Kubernetes) NewRuntimeImageSecret(namespace string, sg *apistructs.ServiceGroup) error {
	// 集群初始化的时候会在 default namespace 下创建一个拉镜像的 secret
	s, err := k.secret.Get("default", AliyunRegistry)
	if err != nil {
		return err
	}

	var dockerConfigJson apistructs.RegistryAuthJson
	if err := json.Unmarshal(s.Data[".dockerconfigjson"], &dockerConfigJson); err != nil {
		return err
	}

	//将有设置用户名密码的runtime的secret追加到secret中
	for _, service := range sg.Services {
		if service.ImageUsername != "" {
			u := strings.Split(service.Image, "/")[0]
			authString := base64.StdEncoding.EncodeToString([]byte(service.ImageUsername + ":" + service.ImagePassword))
			dockerConfigJson.Auths[u] = apistructs.RegistryUserInfo{Auth: authString}
		}
	}

	var sData []byte
	if sData, err = json.Marshal(dockerConfigJson); err != nil {
		return err
	}

	mysecret := &apiv1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: namespace,
		},
		Data: map[string][]byte{".dockerconfigjson": sData},
		Type: s.Type,
	}

	if err := k.secret.Create(mysecret); err != nil {
		return err
	}

	return k.updateDefaultServiceAccountForImageSecret(namespace, s.Name)
}

// CopyDiceSecrets 将 orignns namespace 下的 secret 复制到 dstns 下
func (k *Kubernetes) CopyDiceSecrets(originns, dstns string) ([]apiv1.Secret, error) {
	secrets, err := k.secret.List(originns)
	if err != nil {
		return nil, err
	}
	result := []apiv1.Secret{}
	for _, secret := range secrets.Items {
		// ignore default token
		if !strutil.HasPrefixes(secret.Name, "dice-") {
			continue
		}
		dstsecret := &apiv1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      secret.Name,
				Namespace: dstns,
			},
			Data:       secret.Data,
			StringData: secret.StringData,
			Type:       secret.Type,
		}
		if err := k.secret.CreateIfNotExist(dstsecret); err != nil {
			return nil, err
		}
		result = append(result, secret)
	}
	return result, nil
}

// SecretVolume
func (k *Kubernetes) SecretVolume(secret *apiv1.Secret) (apiv1.Volume, apiv1.VolumeMount) {
	return apiv1.Volume{
			Name: secret.Name,
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName: secret.Name,
				},
			},
		},
		apiv1.VolumeMount{
			Name:      secret.Name,
			MountPath: fmt.Sprintf("/%s", secret.Name),
			ReadOnly:  true,
		}
}
