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

func (cj *cronjob) updateObject(jobObj *v1beta1.CronJob, image, tag string) error {
	var containers []v1.Container

	for _, c := range jobObj.Spec.JobTemplate.Spec.Template.Spec.Containers {
		parts := strings.Split(c.Image, ":")
		if parts[0] == image {
			containers = append(containers, c)
		}
	}

	if len(containers) <= 0 {
		return ErrNoMatchingContainer
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
