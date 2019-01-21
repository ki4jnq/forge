package k8

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"io/ioutil"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"

	"github.com/ki4jnq/forge/deploy/engine"
)

var (
	ErrNonUniqueName       = errors.New("The resource name matched more than one deployment in Kubernetes.")
	ErrUnmatchedName       = errors.New("The resource could not be found on this Kubernetes cluster.")
	ErrNoMatchingContainer = errors.New("A container could not be found that matched the name.")
	ErrNoImage             = errors.New("No matching container image could be found.")
)

type updater interface {
	update(cl *kubernetes.Clientset, name, image, tag string) error
}

type K8 struct {
	*k8DeployWatcher
	*k8ClientProvider

	// updater manages updating specific objects in Kubernetes, e.g.
	// Deployments.
	updater updater

	// needsRollback tracks whether any work must be done to rollback
	// the deployment or not.
	needsRollback bool
}

func NewK8Shipper(opts map[string]interface{}) *K8 {
	return &K8{
		k8DeployWatcher: newK8DeployWatcher(),
		k8ClientProvider: &k8ClientProvider{
			Opts: opts,
		},
	}
}

func (ks *K8) ShipIt(ctx context.Context) chan error {
	ch := make(chan error)
	// Run the actual deploy work in a seperate goroutine and send its errors
	// to the `ch` channel.
	go func() {
		defer close(ch)
		defer ks.savePanics(ch)

		if err := ks.runDeploy(ctx, ch); err != nil {
			ch <- err
		}
	}()
	return ch
}

func (ks *K8) Rollback(ctx context.Context) chan error {
	ch := make(chan error)
	if !ks.needsRollback {
		close(ch)
		return ch
	}

	go func() {
		defer close(ch)
		defer ks.savePanics(ch)

		client, err := ks.getK8Client()
		if err != nil {
			ch <- err
			return
		}

		rollback := &v1beta1.DeploymentRollback{Name: ks.mustLookup("name")}
		err = client.ExtensionsV1beta1().
			Deployments(k8Namespace).
			Rollback(rollback)
		if err != nil {
			ch <- err
		}
	}()
	return ch
}

// runDeploy coordinates all of the actual work performed during the deploy.
func (ks *K8) runDeploy(ctx context.Context, ch chan error) error {
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
	//ks.needsRollback = true // Do we need this if `watchIt` works properly?

	//err = ks.watchIt(
	//	client,
	//	ks.mustLookup("name"),
	//	tag,
	//	*deployment.Spec.Replicas,
	//	deployment.Status.ObservedGeneration,
	//)
	//if err != nil {
	//	return err
	//}

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
