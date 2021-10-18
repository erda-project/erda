package chtype

import "github.com/erda-project/erda/pkg/common/errors"

type AliShortMessage struct {
	AccessKeyId     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
}

func (asm *AliShortMessage) Validate() error {
	if asm.SignName == "" {
		return errors.NewMissingParameterError("signName")
	}
	if asm.TemplateCode == "" {
		return errors.NewMissingParameterError("templateCode")
	}
	if asm.AccessKeyId == "" {
		return errors.NewMissingParameterError("accessKeyId")
	}
	if asm.AccessKeySecret == "" {
		return errors.NewMissingParameterError("accessKeySecret")
	}
	return nil
}
