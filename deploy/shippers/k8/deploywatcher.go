package k8

import (
	"errors"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

type deployStatus int

const (
	statRunning deployStatus = iota
	statFailed
	statDone
)

var acceptableWaitingReasons = [...]string{
	"ContainerCreating",
}

var ErrPodsFailedToStart = errors.New("The new Kubernetes pods failed to start.")

type podSet map[string]bool

type k8DeployWatcher struct {
	running podSet // Pod IDs that have started up but have not completed yet.
	done    podSet // Pod IDs that have finished successfully.
	dead    podSet // Pod IDs that have failed to start.
}

func newK8DeployWatcher() *k8DeployWatcher {
	return &k8DeployWatcher{
		running: make(podSet, 5),
		done:    make(podSet, 5),
		dead:    make(podSet, 5),
	}
}

// watchIt watches events on a K8 pods with the appropriate `name` and
// `version` labels and returns an error if at least `expectedReplicas` are
// not deployed successfully.
func (kdw *k8DeployWatcher) watchIt(client *kubernetes.Clientset, name, version string, expectedReplicas int32, gen int64) error {
	podWatcher, err := client.CoreV1().
		Pods(k8Namespace).
		Watch(v1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%v,version=%v", name, version),
		})
	if err != nil {
		fmt.Println(err)
		return err
	}

	for event := range podWatcher.ResultChan() {
		pod, ok := event.Object.(*v1.Pod)
		if !ok {
			continue
		}
		switch kdw.inspectPodStatus(pod) {
		case statRunning:
			kdw.running[pod.ObjectMeta.Name] = true
		case statFailed:
			delete(kdw.running, pod.ObjectMeta.Name)
			kdw.dead[pod.ObjectMeta.Name] = true

			// This might be a little naive, but it should suffice.
			if int32(len(kdw.dead)) >= expectedReplicas {
				return ErrPodsFailedToStart
			}
		case statDone:
			delete(kdw.running, pod.ObjectMeta.Name)
			kdw.done[pod.ObjectMeta.Name] = true
			if int32(len(kdw.done)) >= expectedReplicas {
				return nil
			}
		}
	}

	return nil
}

// inspectPodStatus evaluates the pod's status and container condition's to
// determine if it has successfully started or not.
func (kdw *k8DeployWatcher) inspectPodStatus(pod *v1.Pod) deployStatus {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == v1.PodReady && cond.Status == v1.ConditionTrue {
			return statDone
		}
	}

	// If this pod just started it may not have "ContainerStatuses" set yet.
	// If so, considering it to still be "running".
	if len(pod.Status.ContainerStatuses) < 1 {
		return statRunning
	}

	for _, stat := range pod.Status.ContainerStatuses {
		if stat.State.Waiting != nil && kdw.isAcceptableWaitingState(stat.State.Waiting) {
			return statRunning
		} else if stat.State.Running != nil {
			return statDone
		} else if stat.State.Terminated != nil {
			return statFailed
		}
	}

	return statFailed
}

func (kdw *k8DeployWatcher) isAcceptableWaitingState(state *v1.ContainerStateWaiting) bool {
	for _, r := range acceptableWaitingReasons {
		if state.Reason == r {
			return true
		}
	}
	return false
}
