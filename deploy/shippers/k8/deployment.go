package k8

import (
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type deployment struct{}

func (d *deployment) update(client *kubernetes.Clientset, name, image, tag string) error {
	deployment, err := d.getCurrentDeployment(client, name)
	if err != nil {
		return err
	}

	if err := d.updateDeploymentObject(deployment, image, tag); err != nil {
		return err
	}

	if err := d.updateK8Deployment(client, deployment); err != nil {
		return err
	}

	return nil
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

// updateDeploymentObject updates the deployment object's container.image to the one
// that is going to be deployed based on the VERSION file.
func (d *deployment) updateDeploymentObject(
	deployment *v1beta1.Deployment,
	image string,
	tag string,
) error {
	containers := make([]*v1.Container, 0, len(deployment.Spec.Template.Spec.Containers))

	for _, c := range deployment.Spec.Template.Spec.Containers {
		parts := strings.Split(c.Image, ":")
		if parts[0] == image {
			containers = append(containers, &c)
		}
	}
	if len(containers) <= 0 {
		return ErrNoMatchingContainer
	}

	if image == "" {
		return ErrNoImage
	}

	// Also make sure to update the deployments metadata to match the new tag.
	deployment.Labels["version"] = tag
	deployment.Spec.Template.Labels["version"] = tag

	for _, c := range containers {
		c.Image = fmt.Sprintf("%v:%v", image, tag)
	}

	return nil
}

func (d *deployment) updateK8Deployment(
	client *kubernetes.Clientset,
	deployment *v1beta1.Deployment,
) error {
	_, err := client.ExtensionsV1beta1().Deployments(k8Namespace).Update(deployment)
	return err
}
