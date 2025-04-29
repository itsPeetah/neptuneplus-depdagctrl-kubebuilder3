package functionnodes

type InvocationEdge struct {
	// FunctionName is the name of the invoked function, used as a pod/service selector. It should match the function name in another node in the graph.
	FunctionName string `json:"functionName"`
	// Id of the invocation. Edges with the same id are invoked concurrently, different ids imply the invocations happen sequentially.
	EdgeId int32 `json:"edgeId"`
	// Multiplier describes how many invocations to this function are performed by the caller function.
	EdgeMultiplier int32 `json:"edgeMultiplier"`
}

type FunctionNode struct {
	// FunctionName represents what function this node is assigned to and it is used as a selector for the pods running said function.
	FunctionName string `json:"functionName"`
	// Invocations is the list of out-edges from the node to invoked functions.
	Invocations []InvocationEdge `json:"invocations"`
}
