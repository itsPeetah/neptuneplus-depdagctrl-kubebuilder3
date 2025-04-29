package functionnodestest_test

import (
	"fmt"
)

// func main() {
// 	fmt.Println("Running test (1.a) build leaf first tree:")
// 	testBuildLeafFirstTree(NodesInput1)
// 	fmt.Println("Running test (1.b) build leaf first tree:")
// 	testBuildLeafFirstTree(NodesInput2)
// }

func TestBuildLeafFirstTree(nodesInput []FunctionNode) {

	nodesOutput := BuildLeafFirstTree(nodesInput)

	for idx, node := range nodesOutput {
		fmt.Println(idx, node.FunctionName)
	}

}
