package main

import (
	"fmt"
	"os"

	"golang.a2z.com/SpaceNeedleHCLWriter/hclwriter"
)

func main() {
	err := hclwriter.InitTerraform()
	if err != nil {
		os.Exit(1)
	}

	args, help, err := hclwriter.ReadCmdLineArgs()
	if err != nil {
		fmt.Println("Failed to convert files: ", err)
		os.Exit(1)
	}
	if help {
		os.Exit(0)
	}

	err = hclwriter.ProcessFiles(args.AccountId, args.SrcPath, args.Name, args.DestPath, args.OverrideProperties, args.PrefixForAllResources)
	if err != nil {
		fmt.Println("Failed to convert Files", err)
		os.Exit(1)
	}

	fmt.Println("Successfully created datasource and dataset")
}
