package controller

import (
	provisioningv1alpha1 "github.com/itspeetah/thesis-test/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DependencyGraphSpec = provisioningv1alpha1.DependencyGraphSpec
type FunctionNode = provisioningv1alpha1.FunctionNode

type Aggregator struct {
	dependencyGraphNodes []FunctionNode
}

func NewAggregator(dag DependencyGraphSpec) *Aggregator {

	return &Aggregator{
		dependencyGraphNodes: sortNodesByDependencies(dag.Nodes), // TODO: Maybe deepcopy?
	}
}

// I don't think this is particularly optimized, but it's not running often and the code that I got Gemini to generate for me was utter trash
func sortNodesByDependencies(nodes []provisioningv1alpha1.FunctionNode) []provisioningv1alpha1.FunctionNode {

	// NodeName -> NodeIndex
	remainingNodeIndices := make(map[string]int)
	for index, node := range nodes {
		remainingNodeIndices[node.FunctionName] = index
	}

	calculateOutDegrees := func() map[string]int {
		outDegrees := make(map[string]int)
		// initialize at 0
		for nodeName, indexInArray := range remainingNodeIndices {
			node := nodes[indexInArray]
			outDegrees[nodeName] = 0
			// increase outdegree only if invoked node has not been "sorted" yet
			for _, edge := range node.Invocations {
				_, ok := remainingNodeIndices[edge.FunctionName]
				if ok {
					outDegrees[nodeName]++
				}
			}
		}
		return outDegrees
	}

	// Get "current" leaves (nodes with outdegree 0 in the current scenario, i.e. with already sorted nodes excluded from the pool)
	getCurrentLeaves := func(outDegreeMap map[string]int) []string {
		leaves := []string{}
		for nodeName, degree := range outDegreeMap {
			if degree == 0 {
				leaves = append(leaves, nodeName)
			}
		}
		return leaves
	}

	sortedNodes := []provisioningv1alpha1.FunctionNode{}

	iterations := 0 // just to make sure that I don't run into an infinite loop: there should never be more iterations than nodes
	limit := len(nodes)
	for len(sortedNodes) < limit && iterations <= limit {
		outDegree := calculateOutDegrees()
		leaves := getCurrentLeaves(outDegree)

		for _, leafName := range leaves {
			sortedNodes = append(sortedNodes, nodes[remainingNodeIndices[leafName]])
			// remove current node from pool of nodes that still need to be sorted
			delete(remainingNodeIndices, leafName)
		}

		iterations++
	}

	if len(sortedNodes) != len(nodes) {
		return []provisioningv1alpha1.FunctionNode{}
	}

	return sortedNodes
}

func (a *Aggregator) aggregate(client client.Client) {

	// Phase 1: get average pod response time for each function in the graph
	functionResponseTimes := make(map[string]float64)
	functionPodCount := make(map[string]int)
	for _, node := range a.dependencyGraphNodes {
		functionResponseTimes[node.FunctionName] = 0
		functionPodCount[node.FunctionName] = 0

		// TODO Implement metric getter

		// Get Pods for Service{node.FunctionName}
		// // Foreach Pod: get latest response time
	}
	for functioName, _ := range functionResponseTimes {
		functionResponseTimes[functioName] /= float64(functionPodCount[functioName])
	}

	// Phase 2: aggregate edge times
	graphEdgeAggregations := make(map[int32]float64)
	for _, node := range a.dependencyGraphNodes {
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
	for _, node := range a.dependencyGraphNodes {
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
