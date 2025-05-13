package aggregator

import (
	provisioningv1alpha1 "github.com/itspeetah/neptune-depdag-controller/api/v1alpha1"
	sametrics "github.com/itspeetah/neptune-depdag-controller/pkg/metrics"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DependencyGraph = provisioningv1alpha1.DependencyGraph
type FunctionNode = provisioningv1alpha1.FunctionNode

type Aggregator struct {
	client       client.Client
	nodes        []FunctionNode
	metricClient sametrics.MetricGetter
}

func NewAggregator(dag *DependencyGraph, client client.Client, metricClient sametrics.MetricGetter) *Aggregator {
	return &Aggregator{
		client:       client,
		metricClient: metricClient,
		nodes:        sortNodesByDependencies(dag.Spec.Nodes), // TODO: Maybe deepcopy?
	}
}

func (a *Aggregator) Aggregate() {

	klog.Info("Aggregating graph times")

	// Phase 1: get average pod response time for each function in the graph
	functionResponseTimes := make(map[string]int64)
	for _, node := range a.nodes {
		averageRt := a.getFunctionResponseTime(node.FunctionName, node.FunctionNamespace)
		functionResponseTimes[node.FunctionName] = averageRt
	}

	// Phase 2: aggregate edge times
	graphEdgeAggregations := make(map[int32]int64)
	for _, node := range a.nodes {
		for _, edge := range node.Invocations {
			currFunctionEdgeValue := functionResponseTimes[edge.FunctionName] * int64(edge.EdgeMultiplier)
			if val, ok := graphEdgeAggregations[edge.EdgeId]; ok {
				// If edge id was already seen it means this is a parallel call, so we take the slower time
				graphEdgeAggregations[edge.EdgeId] = max(val, currFunctionEdgeValue)
			} else {
				// This is either a sequential call or the first time we see a parallel call (therefore this is the slower so far)
				graphEdgeAggregations[edge.EdgeId] = currFunctionEdgeValue
			}
		}
	}

	// Phase 3: calculate external times
	// TODO: this could be integrated in the same loop for phase 2
	nodeExternalResponseTimes := make(map[string]int64)
	for _, node := range a.nodes {
		// If the node is a leaf, external response time is zero :)
		if len(node.Invocations) < 1 {
			nodeExternalResponseTimes[node.FunctionName] = 0
			continue
		}

		var sum int64 = 0
		for _, edge := range node.Invocations {
			sum += graphEdgeAggregations[edge.EdgeId]
		}
		nodeExternalResponseTimes[node.FunctionName] = sum
	}

	// Phase 4: publish times
	for functionName, externalResponseTime := range nodeExternalResponseTimes {
		// TODO: missing metric publisher implementation (var block just to dismiss error for now)
		var (
			_ = functionName
			_ = externalResponseTime
		)

		// TODO: there are three options
		/*
			1 - publish as service metric
			2 - publish as pod metric
			3 - publish as something different

			this influences how the metric is accessed by the kosmos recommender at
			pkg/pod-autoscaler/pkg/recommender/controller.go [line 288]
		*/
	}
}
