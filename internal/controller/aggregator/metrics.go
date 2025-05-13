package aggregator

import (
	"context"

	"github.com/itspeetah/neptune-depdag-controller/internal/controller/metrics"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/apis/custom_metrics/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (a Aggregator) getFunctionResponseTime(name string, namespace string) int64 {

	labeledPodList := &corev1.PodList{}
	a.client.List(context.TODO(), labeledPodList, client.MatchingLabels{
		"edgeautoscaler.polimi.it/function-name":      name,
		"edgeautoscaler.polimi.it/function-namespace": namespace,
	})

	klog.Infof("Found %d pods for function %s:%s", len(labeledPodList.Items), namespace, name)

	var sum v1beta2.MetricValue
	for _, pod := range labeledPodList.Items {
		klog.Infof("Fetching response time metric for pod %s", pod.Name)
		rt := a.getPodResponseTime(&pod)
		klog.Infof("Response time for pod %s: %d", pod.Name, rt.Value.MilliValue())
		sum.Value.Add(rt.Value)
	}

	average := float64(sum.Value.MilliValue()) / float64(labeledPodList.Size())

	return int64(average)
}

func (a Aggregator) getPodResponseTime(pod *corev1.Pod) *v1beta2.MetricValue {
	klog.Info("Before pod metrics fetch")
	rt, err := a.metricClient.PodMetrics(pod, metrics.ResponseTime)
	klog.Info("After pod metrics fetch")
	if err != nil {
		klog.Error("ERROR WHILE FETCHING METRICS", err)
		return &v1beta2.MetricValue{Value: *resource.NewMilliQuantity(0, rt.Value.Format)}
	}
	return rt
}
