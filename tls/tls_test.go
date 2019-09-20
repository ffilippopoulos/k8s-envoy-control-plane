package tls

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEmptySecretName(t *testing.T) {
	client := fake.NewSimpleClientset()
	Init(client)

	_, err := GetTLS("namespace", "")

	expectedError := errors.New("secret name empty")

	if assert.Error(t, err) {
		assert.Equal(t, expectedError, err)
	}
}

func TestNoSecretFound(t *testing.T) {
	client := fake.NewSimpleClientset()
	Init(client)

	_, err := GetTLS("namespace", "secret")

	expectedError := errors.New(fmt.Sprintf("failed to fetch secret %s/%s: secrets \"%s\" not found", "namespace", "secret", "secret"))

	if assert.Error(t, err) {
		assert.Equal(t, expectedError, err)
	}
}

func TestSecretEmpty(t *testing.T) {
	client := fake.NewSimpleClientset()
	s := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret",
			Namespace: "namespace",
		},
	}
	_, err := client.CoreV1().Secrets("namespace").Create(&s)
	if err != nil {
		t.Fatalf("error creating secret %v: %v", s, err)
	}
	Init(client)

	_, err = GetTLS("namespace", "secret")

	errCrtMissing := errors.New("missing entry: tls.crt")
	expectedError := errors.New(fmt.Sprintf("failed to get certificate blocks for secret: %s/%s: %v", "namespace", "secret", errCrtMissing))

	if assert.Error(t, err) {
		assert.Equal(t, expectedError, err)
	}
}

func TestSecretKeyMissing(t *testing.T) {
	client := fake.NewSimpleClientset()
	data := make(map[string][]byte)
	data["tls.crt"] = []byte("AAAAAAAAAAAAAAA=")
	s := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret",
			Namespace: "namespace",
		},
		Data: data,
	}
	_, err := client.CoreV1().Secrets("namespace").Create(&s)
	if err != nil {
		t.Fatalf("error creating secret %v: %v", s, err)
	}
	Init(client)

	_, err = GetTLS("namespace", "secret")

	errCrtMissing := errors.New("missing entry: tls.key")
	expectedError := errors.New(fmt.Sprintf("failed to get certificate blocks for secret: %s/%s: %v", "namespace", "secret", errCrtMissing))

	if assert.Error(t, err) {
		assert.Equal(t, expectedError, err)
	}
}

func TestSecretCertEmpty(t *testing.T) {
	client := fake.NewSimpleClientset()
	data := make(map[string][]byte)
	data["tls.crt"] = []byte{}
	data["tls.key"] = []byte{}
	s := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret",
			Namespace: "namespace",
		},
		Data: data,
	}
	_, err := client.CoreV1().Secrets("namespace").Create(&s)
	if err != nil {
		t.Fatalf("error creating secret %v: %v", s, err)
	}
	Init(client)

	_, err = GetTLS("namespace", "secret")

	errCrtMissing := errors.New("tls.crt is empty")
	expectedError := errors.New(fmt.Sprintf("failed to get certificate blocks for secret: %s/%s: %v", "namespace", "secret", errCrtMissing))

	if assert.Error(t, err) {
		assert.Equal(t, expectedError, err)
	}
}

func TestSecretKeyEmpty(t *testing.T) {
	client := fake.NewSimpleClientset()
	data := make(map[string][]byte)
	data["tls.crt"] = []byte("AAAAAAAAAAAAAAA=")
	data["tls.key"] = []byte{}
	s := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret",
			Namespace: "namespace",
		},
		Data: data,
	}
	_, err := client.CoreV1().Secrets("namespace").Create(&s)
	if err != nil {
		t.Fatalf("error creating secret %v: %v", s, err)
	}
	Init(client)

	_, err = GetTLS("namespace", "secret")

	errCrtMissing := errors.New("tls.key is empty")
	expectedError := errors.New(fmt.Sprintf("failed to get certificate blocks for secret: %s/%s: %v", "namespace", "secret", errCrtMissing))

	if assert.Error(t, err) {
		assert.Equal(t, expectedError, err)
	}
}

func TestGetTLS(t *testing.T) {
	client := fake.NewSimpleClientset()
	data := make(map[string][]byte)
	data["tls.crt"] = []byte("AAAAAAAAAAAAAAA=")
	data["tls.key"] = []byte("AAAAAAAAAAAAAAA=")
	s := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret",
			Namespace: "namespace",
		},
		Data: data,
	}
	_, err := client.CoreV1().Secrets("namespace").Create(&s)
	if err != nil {
		t.Fatalf("error creating secret %v: %v", s, err)
	}
	Init(client)

	cert, err := GetTLS("namespace", "secret")
	if err != nil {
		t.Fatalf("error getting tls %v", err)
	}

	expectedCert := Certificate{
		Cert: "AAAAAAAAAAAAAAA=",
		Key:  "AAAAAAAAAAAAAAA=",
	}
	assert.Equal(t, expectedCert, cert)
}
