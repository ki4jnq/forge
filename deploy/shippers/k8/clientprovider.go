package k8

import (
	"errors"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	ErrMissingConfig = errors.New("Missing some expected configuration values.")
	ErrConfigInvalid = errors.New("A configuration option is invalid.")
)

type k8ClientProvider struct {
	client *kubernetes.Clientset
}

// getK8Client returns a Kubernetes Clientset configured to talk to a
// particular K8 cluster.
func (kcp *k8ClientProvider) getK8Client(opts Options) (*kubernetes.Clientset, error) {
	if kcp.client != nil {
		return kcp.client, nil
	}

	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", configGetter(opts))
	if err != nil {
		return nil, err
	}
	kcp.client, err = kubernetes.NewForConfig(config)
	return kcp.client, err
}

// configGetter is a function passed to the K8 client lib and returns a K8
// config setup as specified in the Forgefile.
func configGetter(opts Options) func() (*api.Config, error) {
	return func() (*api.Config, error) {
		kubeConfig := api.NewConfig()
		context := api.NewContext()
		cluster := api.NewCluster()
		auth := api.NewAuthInfo()

		attrs := []string{"server"}
		basicAuthAttrs := []string{"username", "password"}
		tokenAuthAttrs := []string{"token"}
		certAuthAttrs := []string{"apiCertFile", "apiKeyFile"}
		certAuthDataAttrs := []string{"apiCert", "apiKey"}

		err := opts.readConfigInto(attrs, &cluster.Server)
		if err != nil {
			return nil, err
		}

		// This is not being checked for an error (we should still be OK, though,
		// because Kubernetes will also be checking the config.
		opts.readConfigInto([]string{"caFile"}, &cluster.CertificateAuthority)
		opts.readConfigInto([]string{"ca"}, &cluster.CertificateAuthorityData)

		readers := [](func() error){
			func() error { return opts.readConfigInto(basicAuthAttrs, &auth.Username, &auth.Password) },
			func() error { return opts.readConfigInto(certAuthAttrs, &auth.ClientCertificate, &auth.ClientKey) },
			func() error { return opts.readConfigInto(tokenAuthAttrs, &auth.Token) },
			func() error {
				return opts.readConfigInto(certAuthDataAttrs, &auth.ClientCertificateData, &auth.ClientKeyData)
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
}
