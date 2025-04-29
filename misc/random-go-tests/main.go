package main

import (
	"fmt"

	functionnodes "pgmp.me/test1/function-nodes"
)

func main() {
	fmt.Println("Running test (1.a) build leaf first tree:")
	testBuildLeafFirstTree(functionnodes.NodesInput1)
	fmt.Println("Running test (1.b) build leaf first tree:")
	testBuildLeafFirstTree(functionnodes.NodesInput2)
}

func testBuildLeafFirstTree(nodesInput []functionnodes.FunctionNode) {

	nodesOutput := functionnodes.BuildLeafFirstTree(nodesInput)

	for idx, node := range nodesOutput {
		fmt.Println(idx, node.FunctionName)
	}

}
