package k8

import (
	"k8s.io/api/core/v1"
	"strings"
)

// updateContainerImages iterates the list of containers and updates the
// image:tag attribute of any containers with the matching image.
func updateContainerImages(image, tag string, containers []v1.Container) []v1.Container {
	var newContainers []v1.Container

	for _, c := range containers {
		parts := strings.Split(c.Image, ":")
		if parts[0] == image {
			c.Image = image + ":" + tag
		}

		newContainers = append(newContainers, c)
	}

	return newContainers
}
