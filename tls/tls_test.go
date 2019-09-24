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

func TestEmptyTLSSecretName(t *testing.T) {
	client := fake.NewSimpleClientset()
	Init(client)

	_, err := GetTLS("namespace", "")

	expectedError := errors.New("secret name empty")

	if assert.Error(t, err) {
		assert.Equal(t, expectedError, err)
	}
}

func TestNoTLSSecretFound(t *testing.T) {
	client := fake.NewSimpleClientset()
	Init(client)

	_, err := GetTLS("namespace", "secret")

	expectedError := errors.New(fmt.Sprintf("failed to fetch secret %s/%s: secrets \"%s\" not found", "namespace", "secret", "secret"))

	if assert.Error(t, err) {
		assert.Equal(t, expectedError, err)
	}
}

func TestTLSSecretEmpty(t *testing.T) {
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

func TestTLSSecretKeyMissing(t *testing.T) {
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

func TestTLSSecretCertEmpty(t *testing.T) {
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

func TestTLSSecretKeyEmpty(t *testing.T) {
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

func TestEmptyCASecretName(t *testing.T) {
	client := fake.NewSimpleClientset()
	Init(client)

	_, err := GetCA("namespace", "")

	expectedError := errors.New("secret name empty")

	if assert.Error(t, err) {
		assert.Equal(t, expectedError, err)
	}
}

func TestNoCASecretFound(t *testing.T) {
	client := fake.NewSimpleClientset()
	Init(client)

	_, err := GetCA("namespace", "secret")

	expectedError := errors.New(fmt.Sprintf("failed to fetch secret %s/%s: secrets \"%s\" not found", "namespace", "secret", "secret"))

	if assert.Error(t, err) {
		assert.Equal(t, expectedError, err)
	}
}

func TestCAKeyMissing(t *testing.T) {
	client := fake.NewSimpleClientset()
	data := make(map[string][]byte)
	data["blah.crt"] = []byte("AAAAAAAAAAAAAAA=")
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

	_, err = GetCA("namespace", "secret")

	errCrtMissing := errors.New("missing entry: ca.crt")
	expectedError := errors.New(fmt.Sprintf("failed to get ca.crt for secret: %s/%s: %v", "namespace", "secret", errCrtMissing))

	if assert.Error(t, err) {
		assert.Equal(t, expectedError, err)
	}
}

func TestCASecretEmpty(t *testing.T) {
	client := fake.NewSimpleClientset()
	data := make(map[string][]byte)
	data["ca.crt"] = []byte{}
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

	_, err = GetCA("namespace", "secret")

	errCrtMissing := errors.New("ca.crt is empty")
	expectedError := errors.New(fmt.Sprintf("failed to get ca.crt for secret: %s/%s: %v", "namespace", "secret", errCrtMissing))

	if assert.Error(t, err) {
		assert.Equal(t, expectedError, err)
	}
}

func TestGetCA(t *testing.T) {
	client := fake.NewSimpleClientset()
	data := make(map[string][]byte)
	data["ca.crt"] = []byte("AAAAAAAAAAAAAAA=")
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

	ca, err := GetCA("namespace", "secret")
	if err != nil {
		t.Fatalf("error getting tls %v", err)
	}

	expectedCA := "AAAAAAAAAAAAAAA="
	assert.Equal(t, expectedCA, ca)
}
