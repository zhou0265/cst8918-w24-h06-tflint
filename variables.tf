# Define config variables
variable "label_prefix" {
  type        = string
  description = "Your college username. This will form the beginning of various resource names."
}

variable "region" {
  description = "The Azure region where resources will be deployed"
  type    = string
  default = "canadacentral"  # westus3 not support the vm size of Standard_B1s
}

variable "admin_username" {
  type        = string
  default     = "azureadmin"
  description = "The username for the local user account on the VM."
}
