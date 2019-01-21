package k8

import (
	"fmt"
	"strings"

	"k8s.io/api/batch/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type cronjob struct{}

func (cj *cronjob) update(client *kubernetes.Clientset, name, image, tag string) error {
	job, err := cj.getCurrentJob(client, name)
	if err != nil {
		return err
	}

	if err := cj.updateObject(job, image, tag); err != nil {
		return err
	}

	if err := cj.updateCronJobObject(client, job); err != nil {
		return err
	}

	return nil
}

// getCurrentJob retrieves the Cron Job object whose "app" label
// matches the "name" from Forge's config.
func (cj *cronjob) getCurrentJob(
	client *kubernetes.Clientset,
	name string,
) (
	*v1beta1.CronJob,
	error,
) {
	jobs, err := client.BatchV1beta1().
		CronJobs(k8Namespace).
		List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%v", name),
		})

	if err != nil {
		return nil, err
	} else if len(jobs.Items) > 1 {
		return nil, ErrNonUniqueName
	} else if len(jobs.Items) < 1 {
		return nil, ErrUnmatchedName
	}

	return &jobs.Items[0], nil
}

// updateObject updates the cron job object's container.image to the use the
// new `tag`.
func (cj *cronjob) updateObject(
	jobObj *v1beta1.CronJob,
	image string,
	tag string,
) error {
	containerList := jobObj.Spec.JobTemplate.Spec.Template.Spec.Containers
	containers := make([]*v1.Container, 0, len(containerList))

	for _, c := range containerList {
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

	if jobObj.Spec.JobTemplate.Labels == nil {
		jobObj.Spec.JobTemplate.Labels = make(map[string]string)
	}

	// Also make sure to update the cronJob's metadata to match the new tag.
	jobObj.Labels["version"] = tag
	jobObj.Spec.JobTemplate.Labels["version"] = tag

	for _, c := range containers {
		c.Image = fmt.Sprintf("%v:%v", image, tag)
	}

	return nil
}

func (cj *cronjob) updateCronJobObject(
	client *kubernetes.Clientset,
	jobObj *v1beta1.CronJob,
) error {
	_, err := client.BatchV1beta1().CronJobs(k8Namespace).Update(jobObj)
	return err
}
