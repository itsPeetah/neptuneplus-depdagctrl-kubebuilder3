package aggregator

import (
	"strings"

	provisioningv1alpha1 "github.com/itspeetah/neptune-depdag-controller/api/v1alpha1"
	sametrics "github.com/itspeetah/neptune-depdag-controller/pkg/metrics"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DependencyGraph = provisioningv1alpha1.DependencyGraph
type FunctionNode = provisioningv1alpha1.FunctionNode
type NodeStatus = provisioningv1alpha1.NodeStatus

type PublishCallback = func(*DependencyGraph)

type Aggregator struct {
	Graph        DependencyGraph
	client       client.Client
	nodes        []FunctionNode
	metricClient sametrics.MetricGetter
	publishCb    PublishCallback
}

func NewAggregator(dag *DependencyGraph, client client.Client, metricClient sametrics.MetricGetter, publish PublishCallback) *Aggregator {
	return &Aggregator{
		Graph:        *dag.DeepCopy(),
		client:       client,
		metricClient: metricClient,
		nodes:        sortNodesByDependencies(dag.Spec.Nodes), // TODO: Maybe deepcopy?
		publishCb:    publish,
	}
}

func (a *Aggregator) Aggregate() {

	// Derive average function response time from available metrics
	avgFunctionRTs := make(map[string]int64)
	for _, node := range a.nodes {
		sum, count := a.getFunctionMetric(node.FunctionName, node.FunctionNamespace)
		avgFunctionRTs[node.FunctionNamespace+":"+node.FunctionName] = sum.Value.MilliValue() / int64(count)
	}

	// Calculate edge aggregate response time minding invocation mode
	avgEdgeRTs := make(map[int32]int64)
	for _, node := range a.nodes {
		for _, edge := range node.Invocations {
			currFunctionEdgeValue, ok := avgFunctionRTs[edge.FunctionNamespace+":"+edge.FunctionName]
			if !ok {
				currFunctionEdgeValue = 0
			}

			if val, ok := avgEdgeRTs[edge.EdgeId]; ok {
				// If edge id was already seen it means this is a parallel call, so we take the slower time
				avgEdgeRTs[edge.EdgeId] = max(val, currFunctionEdgeValue*int64(edge.EdgeMultiplier))
			} else {
				// This is either a sequential call or the first time we see a parallel call (therefore this is the slower so far)
				avgEdgeRTs[edge.EdgeId] = currFunctionEdgeValue * int64(edge.EdgeMultiplier)
			}
		}
	}

	// Calculate average external response time for every function
	functionExtRTs := make(map[string]int64)
	for _, node := range a.nodes {
		if len(node.Invocations) < 1 {
			// If the node is a leaf, external response time is zero :)
			functionExtRTs[node.FunctionNamespace+":"+node.FunctionName] = 0
			klog.Infof("[ERT %s:%s] External response time for function: 0", node.FunctionNamespace, node.FunctionName)
			continue
		}

		var sum int64 = 0
		for _, edge := range node.Invocations {
			sum += avgEdgeRTs[edge.EdgeId]
			// mark edge as already counted to avoid counting it multiple times for parallel invocations
			avgEdgeRTs[edge.EdgeId] = 0
		}
		functionExtRTs[node.FunctionNamespace+":"+node.FunctionName] = sum
		klog.Infof("[ERT %s:%s] External response time for function: %d", node.FunctionNamespace, node.FunctionName, sum)
	}

	// Phase 4: publish times
	nodeStatuses := make([]NodeStatus, 0)
	for functionNameAndNamespace, externalResponseTime := range functionExtRTs {
		parts := strings.Split(functionNameAndNamespace, ":")
		fNamespace := parts[0]
		fName := parts[1]

		nodeStatuses = append(nodeStatuses, NodeStatus{
			FunctionName:         fName,
			FunctionNamespace:    fNamespace,
			ExternalResponseTime: externalResponseTime,
		})
	}

	a.Graph.Status.Nodes = nodeStatuses
	a.publishCb(&a.Graph)
}
