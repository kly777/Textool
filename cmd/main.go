package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"textool/internal/combine"
	"textool/internal/config"
	"textool/internal/divider"
	"textool/internal/suit"
	"textool/internal/translate"
)

func main() {
	// 定义子命令
	divideCmd := flag.NewFlagSet("divide", flag.ExitOnError)
	combineCmd := flag.NewFlagSet("combine", flag.ExitOnError)
	mdifyCmd := flag.NewFlagSet("mdify", flag.ExitOnError)
	configCmd := flag.NewFlagSet("config", flag.ExitOnError)
	translateCmd := flag.NewFlagSet("translate", flag.ExitOnError)

	// 验证参数
	if len(os.Args) < 2 {
		fmt.Println("可用命令: divide, combine, process, config")
		fmt.Println("使用 'textool [command] -help' 查看具体命令的帮助")
		os.Exit(0)
	}

	// 路由命令
	switch os.Args[1] {
	case "divide":
		divideInput := divideCmd.String("i", "", "输入文件路径")
		divideOutput := divideCmd.String("o", "", "输出目录路径")
		if err := divideCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("参数解析失败: %v\n", err)
			os.Exit(1)
		}

		if *divideInput == "" || *divideOutput == "" {
			fmt.Println("错误: 必须提供输入文件和输出目录")
			divideCmd.Usage()
			os.Exit(1)
		}

		if err := divider.DivideMDFile(*divideInput, *divideOutput); err != nil {
			fmt.Printf("分割失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("分割完成")

	case "combine":
		rootPath := combineCmd.String("r", ".", "根目录路径")
		outputPath := combineCmd.String("o", "combined.txt", "输出文件路径")
		excludePrefixes := combineCmd.String("ep", "", "排除前缀(逗号分隔)")
		excludeSuffixes := combineCmd.String("es", "", "排除后缀(逗号分隔)")

		if err := combineCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("参数解析失败: %v\n", err)
			os.Exit(1)
		}

		excludeP := []string{}
		excludeS := []string{}

		if *excludePrefixes != "" {
			excludeP = strings.Split(*excludePrefixes, ",")
		}
		if *excludeSuffixes != "" {
			excludeS = strings.Split(*excludeSuffixes, ",")
		}

		combine.Combine(*rootPath, *outputPath, excludeP, excludeS)
		fmt.Println("合并完成")

	case "mdify":
		inputPath := mdifyCmd.String("i", "", "输入文件路径")
		outputPath := mdifyCmd.String("o", "", "输出文件路径")
		if err := mdifyCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("参数解析失败: %v\n", err)
			os.Exit(1)
		}

		if *inputPath == "" {
			if *inputPath = os.Args[2]; *inputPath == "" {
				fmt.Println("错误: 必须提供输入文件路径")
				mdifyCmd.Usage()
				os.Exit(1)
			}
		}
		if *outputPath == "" {
			*outputPath = *inputPath + ".md"
			fmt.Println("未指定输出文件路径，将输出到 ", *outputPath)
		}

		cfg, err := config.GetConfig()
		if err != nil {
			fmt.Printf("获取配置失败: %v\n", err)
			os.Exit(1)
		}

		tp := suit.NewMdiftConfig(cfg.Bl_api_key, 10000, 20000, 0)
		tp.ProcessFile(*inputPath, *outputPath)
		fmt.Println("处理完成")

	case "translate":
		inputPath := translateCmd.String("i", "", "输入文件路径")
		outputPath := translateCmd.String("o", "", "输出文件路径")
		if err := translateCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("参数解析失败: %v\n", err)
			os.Exit(1)
		}

		if *inputPath == "" || *outputPath == "" {
			fmt.Println("错误: 必须提供输入和输出文件路径")
			translateCmd.Usage()
			os.Exit(1)
		}

		cfg, err := config.GetConfig()
		if err != nil {
			fmt.Printf("获取配置失败: %v\n", err)
			os.Exit(1)
		}

		translator := translate.NewTranslator(cfg.Ds_api_key)
		if err := translator.ProcessFile(*inputPath, *outputPath); err != nil {
			fmt.Printf("翻译失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("翻译完成")

	case "config":
		if err := configCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("参数解析失败: %v\n", err)
			os.Exit(1)
		}
		cfg, err := config.GetConfig()
		if err != nil {
			fmt.Printf("获取配置失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("当前配置:\nDeepSeek API Key: %s\nBaiLian API Key: %s\n", cfg.Ds_api_key, cfg.Bl_api_key)

	default:
		fmt.Println("未知命令:", os.Args[1])
		fmt.Println("可用命令: divide, combine, process, translate, config")
		fmt.Println("使用 'textool [command] -help' 查看具体命令的帮助")
		os.Exit(1)
	}
}
