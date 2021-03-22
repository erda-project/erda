package pipelineyml

type SecretNotFoundSecret struct {
	data    []byte
	secrets map[string]string
}

func NewSecretNotFoundSecret(data []byte, secrets map[string]string) *SecretNotFoundSecret {
	return &SecretNotFoundSecret{
		data:    data,
		secrets: secrets,
	}
}

func (v *SecretNotFoundSecret) Visit(s *Spec) {
	if _, err := RenderSecrets(v.data, v.secrets); err != nil {
		s.appendError(err)
		return
	}
}
