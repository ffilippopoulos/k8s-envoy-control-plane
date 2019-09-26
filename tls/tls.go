package tls

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var k8sClient kubernetes.Interface

func Init(client kubernetes.Interface) {
	k8sClient = client
	log.Debug("tls k8s client initialized")
}

func GetTLS(namespace, secretName string) (Certificate, error) {
	if secretName == "" {
		log.Warn("No secret name provided")
		return Certificate{}, errors.New("secret name empty")
	}

	secret, err := k8sClient.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return Certificate{}, errors.New(fmt.Sprintf("failed to fetch secret %s/%s: %v", namespace, secretName, err))
	}

	cert, key, err := getCertificateBlocks(secret)
	if err != nil {
		return Certificate{}, errors.New(fmt.Sprintf("failed to get certificate blocks for secret: %s/%s: %v", namespace, secretName, err))
	}

	return Certificate{
		Cert: cert,
		Key:  key,
	}, nil
}

func getCertificateBlocks(secret *corev1.Secret) (string, string, error) {

	tlsCrtData, tlsCrtExists := secret.Data["tls.crt"]
	if !tlsCrtExists {
		return "", "", errors.New("missing entry: tls.crt")
	}

	tlsKeyData, tlsKeyExists := secret.Data["tls.key"]
	if !tlsKeyExists {
		return "", "", errors.New("missing entry: tls.key")
	}

	cert := string(tlsCrtData)
	if cert == "" {
		return "", "", errors.New("tls.crt is empty")
	}

	key := string(tlsKeyData)
	if key == "" {
		return "", "", errors.New("tls.key is empty")
	}

	return cert, key, nil
}

func GetCA(namespace, secretName string) (string, error) {
	if secretName == "" {
		log.Warn("No secret name provided")
		return "", errors.New("secret name empty")
	}

	secret, err := k8sClient.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to fetch secret %s/%s: %v", namespace, secretName, err))
	}

	ca, err := getCAFromSecret(secret)
	if err != nil {
		return ca, errors.New(fmt.Sprintf("failed to get ca.crt for secret: %s/%s: %v", namespace, secretName, err))
	}

	return ca, nil
}

func getCAFromSecret(secret *corev1.Secret) (string, error) {
	caData, caExists := secret.Data["ca.crt"]
	if !caExists {
		return "", errors.New("missing entry: ca.crt")
	}

	ca := string(caData)
	if ca == "" {
		return "", errors.New("ca.crt is empty")
	}

	return ca, nil
}
