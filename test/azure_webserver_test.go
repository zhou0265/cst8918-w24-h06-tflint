package test

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

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
			"InternalServerError":                           "Retrying due to Azure internal server error",
			"NetworkSecurityGroupOldReferencesNotCleanedUp": "Retrying due to NSG reference cleanup delay",
		},
		MaxRetries:         5,
		TimeBetweenRetries: 15 * time.Second,
	}

	defer func() {
		logger.Log(t, "Waiting 60 seconds before destroy to ensure Azure settles...")
		time.Sleep(60 * time.Second)
		terraform.Destroy(t, terraformOptions)
	}()
	terraform.InitAndApply(t, terraformOptions)

	resourceGroupName := terraform.Output(t, terraformOptions, "resource_group_name")
	vmName := terraform.Output(t, terraformOptions, "vm_name")
	nicName := terraform.Output(t, terraformOptions, "nic_name")
	logger.Log(t, "Resource Group Name: ", resourceGroupName)
	logger.Log(t, "VM Name: ", vmName)
	logger.Log(t, "NIC Name: ", nicName)

	logger.Log(t, "Waiting 30 seconds for Azure to settle...")
	time.Sleep(30 * time.Second)

	// Confirm VM exists
	cmd := exec.Command("cmd", "/c", "az", "vm", "show", "--name", vmName, "--resource-group", resourceGroupName, "--subscription", subscriptionID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Log(t, "Manual az vm show failed: ", err, string(output))
		t.Fatalf("VM check failed: %v", err)
	}
	logger.Log(t, "Manual az vm show succeeded: VM exists")

	// Confirm NIC exists and is connected to VM
	cmd = exec.Command("cmd", "/c", "az", "vm", "nic", "list", "--vm-name", vmName, "--resource-group", resourceGroupName, "--subscription", subscriptionID)
	output, err = cmd.CombinedOutput()
	if err != nil {
		logger.Log(t, "Manual az vm nic list failed: ", err, string(output))
		t.Fatalf("NIC check failed: %v", err)
	}
	var nicList []map[string]interface{}
	if err := json.Unmarshal(output, &nicList); err != nil {
		logger.Log(t, "Failed to parse NIC list JSON: ", err, string(output))
		t.Fatalf("NIC JSON parse failed: %v", err)
	}
	if len(nicList) == 0 {
		t.Fatal("No NICs found for VM")
	}
	id, ok := nicList[0]["id"]
	if !ok || id == nil {
		logger.Log(t, "NIC list missing 'id' field: ", string(output))
		t.Fatalf("NIC id is missing or nil in response")
	}
	idStr, ok := id.(string)
	if !ok {
		logger.Log(t, "NIC id is not a string: ", string(output))
		t.Fatalf("NIC id is not a string: %v", id)
	}
	// Extract NIC name from the id (last segment after '/')
	actualNicName := strings.Split(idStr, "/")[len(strings.Split(idStr, "/"))-1]
	assert.Equal(t, nicName, actualNicName)

	// Confirm VM image
	cmd = exec.Command("cmd", "/c", "az", "vm", "get-instance-view", "--name", vmName, "--resource-group", resourceGroupName, "--subscription", subscriptionID, "--query", "storageProfile.imageReference")
	output, err = cmd.CombinedOutput()
	if err != nil {
		logger.Log(t, "Manual az vm image check failed: ", err, string(output))
		t.Fatalf("Image check failed: %v", err)
	}
	var image map[string]string
	if err := json.Unmarshal(output, &image); err != nil {
		logger.Log(t, "Failed to parse image JSON: ", err, string(output))
		t.Fatalf("Image JSON parse failed: %v", err)
	}
	assert.Equal(t, "Canonical", image["publisher"])
	assert.Equal(t, "22_04-lts-gen2", image["sku"])
}
