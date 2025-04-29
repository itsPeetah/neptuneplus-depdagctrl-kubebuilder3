package functionnodestest_test

/*
	tree
		  / --- C
	A --- B
		  \ --- D --- E

	list:
	A, B, C, D, E

	expected output:
	[C, E], D, B, A
*/

var NodesInput1 = []FunctionNode{
	{
		FunctionName: "A",
		Invocations: []InvocationEdge{
			{
				FunctionName:   "B",
				EdgeMultiplier: 1,
				EdgeId:         1,
			},
		},
	},
	{
		FunctionName: "B",
		Invocations: []InvocationEdge{
			{
				FunctionName:   "C",
				EdgeMultiplier: 1,
				EdgeId:         1,
			},
			{
				FunctionName:   "D",
				EdgeMultiplier: 1,
				EdgeId:         1,
			},
		},
	},
	{
		FunctionName: "C",
		Invocations:  []InvocationEdge{},
	},
	{
		FunctionName: "D",
		Invocations: []InvocationEdge{
			{
				FunctionName:   "E",
				EdgeMultiplier: 1,
				EdgeId:         1,
			},
		},
	},
	{
		FunctionName: "E",
		Invocations:  []InvocationEdge{},
	},
}

/*
	tree
				/ --- F
		  / --- C --- G
	A --- B			  / --- H
		  \ --- D --- E --- I --- M
					  \ --- L
	list:
	A, B, C, D, E, F, G, H, I, L, M

	expected outputs: [x,y,z] means in any order
	[F, G, H, M, L], [C, I], E, D, B, A
*/

var NodesInput2 = []FunctionNode{
	{
		FunctionName: "A",
		Invocations: []InvocationEdge{
			{
				FunctionName:   "B",
				EdgeMultiplier: 1,
				EdgeId:         1,
			},
		},
	},
	{
		FunctionName: "B",
		Invocations: []InvocationEdge{
			{
				FunctionName:   "C",
				EdgeMultiplier: 1,
				EdgeId:         1,
			},
			{
				FunctionName:   "D",
				EdgeMultiplier: 1,
				EdgeId:         1,
			},
		},
	},
	{
		FunctionName: "C",
		Invocations: []InvocationEdge{
			{
				FunctionName:   "F",
				EdgeMultiplier: 1,
				EdgeId:         1,
			},
			{
				FunctionName:   "G",
				EdgeMultiplier: 1,
				EdgeId:         1,
			},
		},
	},
	{
		FunctionName: "D",
		Invocations: []InvocationEdge{
			{
				FunctionName:   "E",
				EdgeMultiplier: 1,
				EdgeId:         1,
			},
		},
	},
	{
		FunctionName: "E",
		Invocations: []InvocationEdge{
			{
				FunctionName:   "H",
				EdgeMultiplier: 1,
				EdgeId:         1,
			},
			{
				FunctionName:   "I",
				EdgeMultiplier: 1,
				EdgeId:         1,
			},
			{
				FunctionName:   "L",
				EdgeMultiplier: 1,
				EdgeId:         1,
			},
		},
	},
	{
		FunctionName: "F",
		Invocations:  []InvocationEdge{},
	},
	{
		FunctionName: "G",
		Invocations:  []InvocationEdge{},
	},
	{
		FunctionName: "H",
		Invocations:  []InvocationEdge{},
	},
	{
		FunctionName: "I",
		Invocations: []InvocationEdge{
			{
				FunctionName:   "M",
				EdgeMultiplier: 1,
				EdgeId:         1,
			},
		},
	},
	{
		FunctionName: "L",
		Invocations:  []InvocationEdge{},
	},
	{
		FunctionName: "M",
		Invocations:  []InvocationEdge{},
	},
}
