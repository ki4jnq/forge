package k8

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"io/ioutil"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/ki4jnq/forge/deploy/engine"
)

var (
	ErrNonUniqueDeploymentName = errors.New("The deployment name matched more than one deployment in Kubernetes.")
	ErrUnmatchedDeploymentName = errors.New("The deployment could not be found on this Kubernetes cluster.")
	ErrNoMatchingContainer     = errors.New("A container could not be found that matched the name.")
	ErrNoImage                 = errors.New("No matching container image could be found.")
)

type K8 struct {
	*k8DeployWatcher
	*k8ClientProvider

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
		err = client.ExtensionsV1beta1().Deployments(k8Namespace).Rollback(rollback)
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

	deployment, err := ks.getCurrentDeployment(client)
	if err != nil {
		return err
	}

	if err := ks.updateDeploymentObject(deployment, tag); err != nil {
		return err
	}

	if err := ks.updateK8Deployment(client, deployment); err != nil {
		return err
	}
	ks.needsRollback = true // Do we need this if `watchIt` works properly?

	if err := ks.watchIt(client, ks.mustLookup("name"), tag, *deployment.Spec.Replicas, deployment.Status.ObservedGeneration); err != nil {
		return err
	}

	return nil
}

// getCurrentDeployment retrieves the deployment object whose "app" label
// matches the "name" from Forge's config.
func (ks *K8) getCurrentDeployment(client *kubernetes.Clientset) (*v1beta1.Deployment, error) {
	deployments, err := client.ExtensionsV1beta1().
		Deployments(k8Namespace).
		List(v1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%v", ks.mustLookup("name")),
		})
	if err != nil {
		return nil, err
	} else if len(deployments.Items) > 1 {
		return nil, ErrNonUniqueDeploymentName
	} else if len(deployments.Items) < 1 {
		return nil, ErrUnmatchedDeploymentName
	}
	return &deployments.Items[0], nil
}

// updateDeploymentObject updates the deployment object's container.image to the one
// that is going to be deployed based on the VERSION file.
func (ks *K8) updateDeploymentObject(deployment *v1beta1.Deployment, tag string) error {
	containers := make([]*v1.Container, 0, len(deployment.Spec.Template.Spec.Containers))

	imageName := ks.mustLookup("image")
	for i, _ := range deployment.Spec.Template.Spec.Containers {
		c := &deployment.Spec.Template.Spec.Containers[i]
		parts := strings.Split(c.Image, ":")
		if parts[0] == imageName {
			containers = append(containers, c)
		}
	}
	if len(containers) <= 0 {
		return ErrNoMatchingContainer
	}

	if imageName == "" {
		return ErrNoImage
	}

	// Also make sure to update the deployments metadata to match the new tag.
	deployment.Labels["version"] = tag
	deployment.Spec.Template.Labels["version"] = tag

	for _, c := range containers {
		c.Image = fmt.Sprintf("%v:%v", imageName, tag)
	}

	return nil
}

func (ks *K8) updateK8Deployment(client *kubernetes.Clientset, deployment *v1beta1.Deployment) error {
	_, err := client.ExtensionsV1beta1().Deployments(k8Namespace).Update(deployment)
	return err
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
