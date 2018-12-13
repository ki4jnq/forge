package k8

import (
	"errors"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	// These are arbitrary and serve only to identify items within the K8
	// configuration itself.
	currentK8Context = "theonlyonewecareabout"
	currentK8User    = "justme"
	currentK8Cluster = "overthere"
	// Unlike the previous consts, k8Namespace is non-arbitrary.
	k8Namespace = "default"
)

var (
	ErrMissingConfig = errors.New("Missing some expected configuration values.")
	ErrConfigInvalid = errors.New("A configuration option is invalid.")
)

type k8ClientProvider struct {
	client *kubernetes.Clientset
	Opts   map[string]interface{}
}

// getK8Client returns a Kubernetes Clientset configured to talk to a
// particular K8 cluster.
func (kcp *k8ClientProvider) getK8Client() (*kubernetes.Clientset, error) {
	if kcp.client != nil {
		return kcp.client, nil
	}

	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", kcp.configGetter)
	if err != nil {
		return nil, err
	}
	kcp.client, err = kubernetes.NewForConfig(config)
	return kcp.client, err
}

// configGetter is a function passed to the K8 client lib and returns a K8
// config setup as specified in the Forgefile.
func (kcp *k8ClientProvider) configGetter() (*api.Config, error) {
	kubeConfig := api.NewConfig()
	context := api.NewContext()
	cluster := api.NewCluster()
	auth := api.NewAuthInfo()

	attrs := []string{"server"}
	basicAuthAttrs := []string{"username", "password"}
	tokenAuthAttrs := []string{"token"}
	certAuthAttrs := []string{"apiCertFile", "apiKeyFile"}
	certAuthDataAttrs := []string{"apiCert", "apiKey"}

	err := kcp.readConfigInto(attrs, &cluster.Server)
	if err != nil {
		return nil, err
	}

	// This is not being checked for an error (we should still be OK, though,
	// because Kubernetes will also be checking the config.
	kcp.readConfigInto([]string{"caFile"}, &cluster.CertificateAuthority)
	kcp.readConfigInto([]string{"ca"}, &cluster.CertificateAuthorityData)

	readers := [](func() error){
		func() error { return kcp.readConfigInto(basicAuthAttrs, &auth.Username, &auth.Password) },
		func() error { return kcp.readConfigInto(certAuthAttrs, &auth.ClientCertificate, &auth.ClientKey) },
		func() error { return kcp.readConfigInto(tokenAuthAttrs, &auth.Token) },
		func() error {
			return kcp.readConfigInto(certAuthDataAttrs, &auth.ClientCertificateData, &auth.ClientKeyData)
		},
	}

loop:
	for idx, reader := range readers {
		switch err := reader(); err {
		case nil: // Success
			break loop
		case ErrMissingConfig: // Non-fatal error
			if idx >= len(readers)-1 {
				return nil, err
			}
		default: // Fatal error
			return nil, err
		}
	}

	context.Namespace = "default"
	context.AuthInfo = currentK8User
	context.Cluster = currentK8Cluster

	kubeConfig.AuthInfos = map[string]*api.AuthInfo{currentK8User: auth}
	kubeConfig.Clusters = map[string]*api.Cluster{currentK8Cluster: cluster}
	kubeConfig.Contexts = map[string]*api.Context{currentK8Context: context}
	kubeConfig.APIVersion = "v1"
	kubeConfig.CurrentContext = currentK8Context

	return kubeConfig, nil
}

// readConfigInto reads options from the Forge config into variables passed
// in the `targets` varargs. Care should be taken to ensure that len(optNames)
// is <= len(targets).
func (kcp *k8ClientProvider) readConfigInto(optNames []string, targets ...interface{}) error {
	for idx, name := range optNames {
		if _, ok := kcp.Opts[name]; !ok {
			return ErrMissingConfig
		}

		value, ok := kcp.Opts[name].(string)
		if !ok {
			return ErrConfigInvalid
		}

		switch target := targets[idx].(type) {
		case *string:
			*target = value
			continue
		case *[]byte:
			*target = []byte(value)
			continue
		default:
			fmt.Println("failed to match %v\n", targets[idx])
		}
		return ErrConfigInvalid
	}
	return nil
}
