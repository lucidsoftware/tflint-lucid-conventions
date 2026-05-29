package naming

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

var remoteStateNamingConfig string = `
rule "terraform_remote_state_naming" {
  enabled = true
}
`

func Test_TerraformRemoteStateNamingRule(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "standard key with correct name - no issue",
			Content: `
data "terraform_remote_state" "com_preprod_us" {
  backend = "s3"
  config = {
    bucket = "*************-spacelift-states-******"
    key    = "1/com-preprod-us"
    region = "us-east-1"
  }
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "standard key with wrong name - issue",
			Content: `
data "terraform_remote_state" "preprod_us" {
  backend = "s3"
  config = {
    bucket = "*************-spacelift-states-******"
    key    = "1/com-preprod-us"
    region = "us-east-1"
  }
}`,
			Expected: helper.Issues{
				{
					Rule:    &TerraformRemoteStateNamingRule{},
					Message: `local name "preprod_us" does not match the standard convention for key "1/com-preprod-us" (expected "com_preprod_us")`,
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 43},
					},
				},
			},
		},
		{
			Name: "standard key with multi-segment hyphens - correct name",
			Content: `
data "terraform_remote_state" "services_customer_migration_com_preprod_us" {
  backend = "s3"
  config = {
    bucket = "*************-spacelift-states-******"
    key    = "1/services-customer-migration-com-preprod-us"
    region = "us-east-1"
  }
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "standard key bare 'tags' name - issue",
			Content: `
data "terraform_remote_state" "tags" {
  backend = "s3"
  config = {
    bucket = "*************-spacelift-states-******"
    key    = "1/com-tags"
    region = "us-east-1"
  }
}`,
			Expected: helper.Issues{
				{
					Rule:    &TerraformRemoteStateNamingRule{},
					Message: `local name "tags" does not match the standard convention for key "1/com-tags" (expected "com_tags")`,
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 37},
					},
				},
			},
		},
		{
			Name: "legacy bare key with legacy_ prefix - no issue",
			Content: `
data "terraform_remote_state" "legacy_bi" {
  backend = "s3"
  config = {
    bucket = "lucid-terraform"
    key    = "tfstate"
    region = "us-east-1"
  }
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "legacy bare key without legacy_ prefix - issue",
			Content: `
data "terraform_remote_state" "utility_infrastructure" {
  backend = "s3"
  config = {
    bucket = "lucid-terraform"
    key    = "tfstate"
    region = "us-east-1"
  }
}`,
			Expected: helper.Issues{
				{
					Rule:    &TerraformRemoteStateNamingRule{},
					Message: `local name "utility_infrastructure" references legacy key "tfstate" and must start with the prefix "legacy_"`,
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 55},
					},
				},
			},
		},
		{
			Name: "legacy key with hyphens but no leading digit - needs legacy_ prefix",
			Content: `
data "terraform_remote_state" "gov_production" {
  backend = "s3"
  config = {
    bucket = "lucid-terraform-gov"
    key    = "lucid-gov-production-global"
    region = "us-gov-west-1"
  }
}`,
			Expected: helper.Issues{
				{
					Rule:    &TerraformRemoteStateNamingRule{},
					Message: `local name "gov_production" references legacy key "lucid-gov-production-global" and must start with the prefix "legacy_"`,
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 47},
					},
				},
			},
		},
		{
			Name: "dynamic key with var interpolation - skipped (no issue)",
			Content: `
variable "account_name" {
  type = string
}

data "terraform_remote_state" "account" {
  backend = "s3"
  config = {
    bucket = "*************-spacelift-states-******"
    key    = "1/dev-${var.account_name}-account"
    region = "us-east-1"
  }
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "other data source type - not affected",
			Content: `
data "aws_caller_identity" "preprod_us" {}`,
			Expected: helper.Issues{},
		},
		{
			Name: "multiple blocks - mixed correctness",
			Content: `
data "terraform_remote_state" "com_preprod_us" {
  backend = "s3"
  config = {
    bucket = "*************-spacelift-states-******"
    key    = "1/com-preprod-us"
    region = "us-east-1"
  }
}

data "terraform_remote_state" "tags" {
  backend = "s3"
  config = {
    bucket = "*************-spacelift-states-******"
    key    = "1/com-tags"
    region = "us-east-1"
  }
}

data "terraform_remote_state" "legacy_bi" {
  backend = "s3"
  config = {
    bucket = "lucid-terraform"
    key    = "tfstate"
    region = "us-east-1"
  }
}`,
			Expected: helper.Issues{
				{
					Rule:    &TerraformRemoteStateNamingRule{},
					Message: `local name "tags" does not match the standard convention for key "1/com-tags" (expected "com_tags")`,
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 11, Column: 1},
						End:      hcl.Pos{Line: 11, Column: 37},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			rule := &TerraformRemoteStateNamingRule{}

			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": test.Content,
				".tflint.hcl": remoteStateNamingConfig,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, test.Expected, runner.Issues)
		})
	}
}
