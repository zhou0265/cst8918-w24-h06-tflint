# Define output values for later reference
output "resource_group_name" {
  value = azurerm_resource_group.rg.name
  description = "The name of the Azure resource group."
}

output "vm_name" {
  value = azurerm_linux_virtual_machine.webserver.name
  description = "The name of the virtual machine."
}

output "nic_name" {
  value = azurerm_network_interface.webserver.name
  description = "The name of the network interface."
}

output "public_ip" {
  value = azurerm_linux_virtual_machine.webserver.public_ip_address
  description = "The public IP address assigned to the virtual machine."
}
