package main

import (
	"github.com/lucidsoftware/tflint-lucid-conventions/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &tflint.BuiltinRuleSet{
			Name:    "lucid-conventions",
			Version: "0.1.0",
			Rules: []tflint.Rule{
				&rules.AwsIamPolicyDocumentRequiredRule{},
			},
		},
	})
}
