package hclwriter

import (
	"fmt"
	"os"
	"os/exec"
)

func InitTerraform() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	cmdInit := exec.Command("terraform", "init")
	fmt.Println("Executing command:", cmdInit.String())
	cmdInit.Dir = dir

	if err := cmdInit.Start(); err != nil {
		fmt.Println("Error running terraform init: ", err)
		return err
	}
	if err := cmdInit.Wait(); err != nil {
		fmt.Println("Error waiting for terraform init: ", err)
		return err
	}

	return nil
}

func FormatFile(destPath, hclFileName string) error {
	cmdFmt := exec.Command("terraform", "fmt", hclFileName)
	cmdFmt.Dir = destPath
	fmt.Println("Executing command:", cmdFmt.String())

	if err := cmdFmt.Start(); err != nil {
		return fmt.Errorf("error running terraform fmt: %v", err)
	}
	if err := cmdFmt.Wait(); err != nil {
		return fmt.Errorf("error waiting for terraform fmt: %v", err)
	}

	return nil
}
