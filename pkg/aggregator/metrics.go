package aggregator

import (
	"context"

	"github.com/itspeetah/neptune-depdag-controller/pkg/metrics"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/apis/custom_metrics/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Returns total sum of pod response times for function and the number of pods found for that function
func (a Aggregator) getFunctionMetric(name string, namespace string) (*v1beta2.MetricValue, int) {

	labeledPodList := &corev1.PodList{}
	a.client.List(context.TODO(), labeledPodList, client.MatchingLabels{
		"edgeautoscaler.polimi.it/function-name":      name,
		"edgeautoscaler.polimi.it/function-namespace": namespace,
	})
	podCount := len(labeledPodList.Items)
	klog.Infof("[%s Aggregator] Found %d pods.", name, podCount)

	var sum v1beta2.MetricValue
	for _, pod := range labeledPodList.Items {
		rt := a.getPodResponseTime(&pod)
		// klog.Infof("Response time for pod %s: %d", pod.Name, rt.Value.MilliValue())
		sum.Value.Add(rt.Value)
	}

	return &sum, podCount
}

func (a Aggregator) getPodResponseTime(pod *corev1.Pod) *v1beta2.MetricValue {
	rt, err := a.metricClient.PodMetrics(pod, metrics.ResponseTime)
	if err != nil {
		klog.Errorf("[%s] Could not retrieve pod metrics for pod: %s", pod.Name, err.Error())
		return &v1beta2.MetricValue{Value: *resource.NewMilliQuantity(0, resource.DecimalSI)}
	}
	return rt
}
