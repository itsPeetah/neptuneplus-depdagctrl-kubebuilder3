package aggregator

import (
	provisioningv1alpha1 "github.com/itspeetah/neptune-depdag-controller/api/v1alpha1"
	sametrics "github.com/itspeetah/neptune-depdag-controller/pkg/metrics"
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

	// Derive average function response time from available metrics
	avgFunctionRTs := make(map[string]int64)
	for _, node := range a.nodes {
		sum, count := a.getFunctionMetric(node.FunctionName, node.FunctionNamespace)
		avgFunctionRTs[node.FunctionName] = sum.Value.MilliValue() / int64(count)
	}

	// Calculate edge aggregate response time minding invocation mode
	avgEdgeRTs := make(map[int32]int64)
	for _, node := range a.nodes {
		for _, edge := range node.Invocations {
			currFunctionEdgeValue := avgFunctionRTs[edge.FunctionName] * int64(edge.EdgeMultiplier)
			if val, ok := avgEdgeRTs[edge.EdgeId]; ok {
				// If edge id was already seen it means this is a parallel call, so we take the slower time
				avgEdgeRTs[edge.EdgeId] = max(val, currFunctionEdgeValue)
			} else {
				// This is either a sequential call or the first time we see a parallel call (therefore this is the slower so far)
				avgEdgeRTs[edge.EdgeId] = currFunctionEdgeValue
			}
		}
	}

	// Calculate average external response time for every function
	functionExtRTs := make(map[string]int64)
	for _, node := range a.nodes {
		if len(node.Invocations) < 1 {
			// If the node is a leaf, external response time is zero :)
			functionExtRTs[node.FunctionName] = 0
			continue
		}

		var sum int64 = 0
		for _, edge := range node.Invocations {
			sum += avgEdgeRTs[edge.EdgeId]
			// mark edge as already counted to avoid counting it multiple times for parallel invocations
			avgEdgeRTs[edge.EdgeId] = 0
		}
		functionExtRTs[node.FunctionName] = sum
	}

	// Phase 4: publish times
	for functionName, externalResponseTime := range functionExtRTs {
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
