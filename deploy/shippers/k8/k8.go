package k8

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"io/ioutil"

	"k8s.io/client-go/kubernetes"

	"github.com/ki4jnq/forge/deploy/engine"
)

var (
	ErrNonUniqueName = errors.New("The resource name matched more than one deployment in Kubernetes.")
	ErrUnmatchedName = errors.New("The resource could not be found on this Kubernetes cluster.")
)

type updater interface {
	update(cl *kubernetes.Clientset, name, image, tag string) error
	rollback(cl *kubernetes.Clientset, name string) error
}

type K8 struct {
	*k8ClientProvider

	// updater manages updating specific objects in Kubernetes, e.g.
	// Deployments.
	updater updater
}

func newK8Shipper(opts map[string]interface{}) *K8 {
	return &K8{
		k8ClientProvider: &k8ClientProvider{
			Opts: opts,
		},
	}
}

func NewCronShipper(opts map[string]interface{}) *K8 {
	shipper := newK8Shipper(opts)
	shipper.updater = &cronjob{}
	return shipper
}

func NewDeploymentShipper(opts map[string]interface{}) *K8 {
	shipper := newK8Shipper(opts)
	shipper.updater = &deployment{}
	return shipper
}

func (ks *K8) ShipIt(ctx context.Context) chan error {
	ch := make(chan error)
	// Run the actual deploy work in a seperate goroutine and send its errors
	// to the `ch` channel.
	go func() {
		defer close(ch)
		defer ks.savePanics(ch)

		if err := ks.runDeploy(ctx); err != nil {
			ch <- err
		}
	}()
	return ch
}

func (ks *K8) Rollback(ctx context.Context) chan error {
	ch := make(chan error)

	go func() {
		defer close(ch)
		defer ks.savePanics(ch)

		client, err := ks.getK8Client()
		if err != nil {
			ch <- err
			return
		}

		ks.updater.rollback(client, ks.mustLookup("name"))
		if err != nil {
			ch <- err
		}
	}()
	return ch
}

// runDeploy coordinates all of the actual work performed during the deploy.
func (ks *K8) runDeploy(ctx context.Context) error {
	tag, err := ks.readTag(ctx)
	if err != nil {
		return err
	}

	client, err := ks.getK8Client()
	if err != nil {
		return err
	}

	err = ks.updater.update(
		client,
		ks.mustLookup("name"),
		ks.mustLookup("image"),
		tag,
	)
	if err != nil {
		return err
	}

	return nil
}

func (ks *K8) readTag(ctx context.Context) (string, error) {
	version := engine.OptionsFromContext(ctx).Version
	if version != "" {
		return version, nil
	}

	buffer, err := ioutil.ReadFile("VERSION")
	return strings.Trim(string(buffer), " \n"), err
}

// TODO: The configuration should be verified at an early step, as opposed to
// paniking if things aren't exactly what we expect.
func (ks *K8) mustLookup(key string) string {
	name, ok := ks.Opts[key].(string)
	if !ok {
		panic(ConfigErr{key})
	}
	return name
}

func (ks *K8) savePanics(ch chan error) {
	if obj := recover(); obj != nil {
		switch err := obj.(type) {
		case error:
			ch <- err.(error)
		case string:
			ch <- errors.New(err)
		default:
			ch <- errors.New(fmt.Sprintf(
				"Encountered an unknown error. String representation is: %v", err,
			))
		}
	}
}
