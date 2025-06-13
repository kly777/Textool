package main

import (
	"textool/combine"
)

func main() {
	// config, err := config.GetConfig()
	// if err != nil {
	// 	panic(err)
	// }
	// key := config.Bl_api_key
	// tp := suit.NewTextProcessor(key, 10000, 50000, 0)
	// tp.ProcessFile("input.md", "./output/output2.md")
	combine.Combine(
		".",
		"./output/output3.txt",
		[]string{"."},
		[]string{"htm","md","txt","exe"})
}
