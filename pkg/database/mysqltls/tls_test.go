// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mysqltls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		assert.NoError(t, Open("", "", "", ""))
	})

	t.Run("built-in mode remains compatible", func(t *testing.T) {
		assert.NoError(t, Open("true", "", "", ""))
	})

	t.Run("custom profile requires CA", func(t *testing.T) {
		assert.Error(t, Open("mysql", "", "", ""))
	})

	t.Run("certificate path requires profile", func(t *testing.T) {
		assert.Error(t, Open("", "/tmp/ca.pem", "", ""))
	})

	t.Run("built-in mode rejects ignored CA", func(t *testing.T) {
		assert.Error(t, Open("true", "/tmp/ca.pem", "", ""))
	})

	t.Run("invalid CA fails", func(t *testing.T) {
		caPath := filepath.Join(t.TempDir(), "ca.pem")
		require.NoError(t, os.WriteFile(caPath, []byte("not a PEM certificate"), 0600))
		assert.Error(t, Open("rds-invalid", caPath, "", ""))
	})

	t.Run("unreadable CA fails", func(t *testing.T) {
		caPath := filepath.Join(t.TempDir(), "missing-ca.pem")
		assert.Error(t, Open("rds-missing-file", caPath, "", ""))
	})
}

func TestOpenRegistersHostnameVerifyingProfile(t *testing.T) {
	const tlsName = "rds-identity-test"
	caPath := filepath.Join(t.TempDir(), "ca.pem")
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: testCertificate(t)})
	require.NoError(t, os.WriteFile(caPath, cert, 0600))
	require.NoError(t, Open(tlsName, caPath, "", ""))
	defer mysql.DeregisterTLSConfig(tlsName)

	cfg, err := mysql.ParseDSN("erda:secret@tcp(example.mysql.rds.aliyuncs.com:3306)/erda?tls=" + tlsName)
	require.NoError(t, err)
	require.NotNil(t, cfg.TLS)
	assert.NotNil(t, cfg.TLS.RootCAs)
	assert.False(t, cfg.TLS.InsecureSkipVerify)
	assert.Equal(t, "example.mysql.rds.aliyuncs.com", cfg.TLS.ServerName)
}

func testCertificate(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	return der
}
