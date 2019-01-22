package k8

import (
	"fmt"

	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type deployment struct {
	// Set to true if the deployment has reached the point that a rollback
	// should be triggered.
	needsRollback bool
}

func (d *deployment) update(client *kubernetes.Clientset, name, image, tag string) error {
	deployment, err := d.getCurrentDeployment(client, name)
	if err != nil {
		return err
	}

	d.updateDeploymentObject(deployment, image, tag)

	if err := d.updateK8Deployment(client, deployment); err != nil {
		return err
	}
	d.needsRollback = true

	watcher := newK8DeployWatcher()
	err = watcher.watchIt(
		client,
		name,
		tag,
		*deployment.Spec.Replicas,
		deployment.Status.ObservedGeneration,
	)
	if err != nil {
		return err
	}

	return nil
}

func (d *deployment) rollback(client *kubernetes.Clientset, name string) error {
	if !d.needsRollback {
		return nil
	}

	rollback := &v1beta1.DeploymentRollback{Name: name}
	return client.ExtensionsV1beta1().
		Deployments(k8Namespace).
		Rollback(rollback)
}

// getCurrentDeployment retrieves the deployment object whose "app" label
// matches the "name" from Forge's config.
func (d *deployment) getCurrentDeployment(
	client *kubernetes.Clientset,
	name string,
) (
	*v1beta1.Deployment,
	error,
) {
	deployments, err := client.ExtensionsV1beta1().
		Deployments(k8Namespace).
		List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%v", name),
		})

	if err != nil {
		return nil, err
	} else if len(deployments.Items) > 1 {
		return nil, ErrNonUniqueName
	} else if len(deployments.Items) < 1 {
		return nil, ErrUnmatchedName
	}

	return &deployments.Items[0], nil
}

// updateDeploymentObject updates the deployment object's container.image to
// the new `tag`.
func (d *deployment) updateDeploymentObject(
	deployment *v1beta1.Deployment,
	image string,
	tag string,
) {
	containers := updateContainerImages(
		image,
		tag,
		deployment.Spec.Template.Spec.Containers,
	)

	if deployment.Spec.Template.Labels == nil {
		deployment.Spec.Template.Labels = make(map[string]string)
	}

	// Also make sure to update the deployment's metadata to match the new tag.
	deployment.Labels["version"] = tag
	deployment.Spec.Template.Labels["version"] = tag
	deployment.Spec.Template.Spec.Containers = containers
}

func (d *deployment) updateK8Deployment(
	client *kubernetes.Clientset,
	deployment *v1beta1.Deployment,
) error {
	_, err := client.ExtensionsV1beta1().Deployments(k8Namespace).Update(deployment)
	return err
}
