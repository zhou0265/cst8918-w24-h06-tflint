package test

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

var subscriptionID string = "f13ce856-164c-44af-a9ea-d8c5f46ac222"

func TestAzureLinuxVMCreation(t *testing.T) {
	logger.Log(t, "Current PATH before update: ", os.Getenv("PATH"))
	os.Setenv("PATH", os.Getenv("PATH")+";C:\\Program Files\\Microsoft SDKs\\Azure\\CLI2\\wbin")
	logger.Log(t, "Updated PATH: ", os.Getenv("PATH"))

	terraformOptions := &terraform.Options{
		TerraformDir: "../",
		Vars: map[string]interface{}{
			"label_prefix": "zhou0265",
		},
		RetryableTerraformErrors: map[string]string{
			"InternalServerError": "Retrying due to Azure internal server error",
		},
		MaxRetries:         3,
		TimeBetweenRetries: 5 * time.Second,
	}

	defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	resourceGroupName := terraform.Output(t, terraformOptions, "resource_group_name")
	vmName := terraform.Output(t, terraformOptions, "vm_name")
	nicName := terraform.Output(t, terraformOptions, "nic_name")
	logger.Log(t, "Resource Group Name: ", resourceGroupName)
	logger.Log(t, "VM Name: ", vmName)
	logger.Log(t, "NIC Name: ", nicName)

	// Wait for Azure to propagate the resource group
	logger.Log(t, "Waiting 30 seconds for Azure to settle...")
	time.Sleep(30 * time.Second)

	// Confirm VM exists with manual az call (with retry)
	var output []byte
	var err error
	for i := 0; i < 3; i++ {
		cmd := exec.Command("cmd", "/c", "az", "vm", "show", "--name", vmName, "--resource-group", resourceGroupName, "--subscription", subscriptionID)
		output, err = cmd.CombinedOutput()
		if err == nil {
			break
		}
		logger.Log(t, "Attempt", i+1, "failed: ", err, string(output))
		time.Sleep(10 * time.Second) // Wait before retry
	}
	if err != nil {
		logger.Log(t, "Manual az vm show failed after retries: ", err, string(output))
		t.Fatalf("VM check failed: %v", err)
	} else {
		logger.Log(t, "Manual az vm show succeeded: VM exists")
	}

	actualNicNames := azure.GetVirtualMachineNics(t, vmName, resourceGroupName, subscriptionID)
	assert.Equal(t, nicName, actualNicNames[0])

	vmImage := azure.GetVirtualMachineImage(t, vmName, resourceGroupName, subscriptionID)
	expectedOSPublisher := "Canonical"
	expectedOSVersion := "22_04-lts-gen2"
	assert.Equal(t, expectedOSPublisher, vmImage.Publisher)
	assert.Equal(t, expectedOSVersion, vmImage.SKU)
}
