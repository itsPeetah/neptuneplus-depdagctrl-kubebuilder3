package aggregator

import (
	provisioningv1alpha1 "github.com/itspeetah/neptune-depdag-controller/api/v1alpha1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DependencyGraph = provisioningv1alpha1.DependencyGraph
type FunctionNode = provisioningv1alpha1.FunctionNode

type Aggregator struct {
	client client.Client
	nodes  []FunctionNode
}

func NewAggregator(dag *DependencyGraph, client client.Client) *Aggregator {
	return &Aggregator{
		client: client,
		nodes:  sortNodesByDependencies(dag.Spec.Nodes), // TODO: Maybe deepcopy?
	}
}

func (a *Aggregator) Aggregate() {

	klog.Info("Aggregating graph times")

	// Phase 1: get average pod response time for each function in the graph
	functionResponseTimes := make(map[string]float64)
	functionPodCount := make(map[string]int)
	for _, node := range a.nodes {
		functionResponseTimes[node.FunctionName] = 0
		functionPodCount[node.FunctionName] = 0

		// Get Pods for Service{node.FunctionName}
		// // Foreach Pod: get latest response time
	}
	for functioName, _ := range functionResponseTimes {
		functionResponseTimes[functioName] /= float64(functionPodCount[functioName])
	}

	// Phase 2: aggregate edge times
	graphEdgeAggregations := make(map[int32]float64)
	for _, node := range a.nodes {
		for _, edge := range node.Invocations {
			currFunctionEdgeValue := functionResponseTimes[edge.FunctionName] * float64(edge.EdgeMultiplier)
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
	nodeExternalResponseTimes := make(map[string]float64)
	for _, node := range a.nodes {
		// If the node is a leaf, external response time is zero :)
		if len(node.Invocations) < 1 {
			nodeExternalResponseTimes[node.FunctionName] = 0
			continue
		}

		sum := 0.0
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
