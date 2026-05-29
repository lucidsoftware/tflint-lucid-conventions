package naming

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// TerraformRemoteStateNamingRule enforces the local-name conventions for
// `data "terraform_remote_state"` blocks documented in the Lucid terraform repo.
//
// Two conventions:
//
//  1. Standard keys (match ^\d+/[a-z0-9-]+$): local name must equal the key
//     with the leading \d+/ stripped and `-` replaced with `_`.
//     e.g. key="1/com-preprod-us" => name="com_preprod_us"
//
//  2. Legacy keys (anything else): local name must start with the prefix
//     "legacy_".
//     e.g. key="tfstate" => name="legacy_bi" (or any legacy_* name)
type TerraformRemoteStateNamingRule struct {
	tflint.DefaultRule
}

var (
	standardKeyRe = regexp.MustCompile(`^\d+/[a-z0-9-]+$`)
	stripPrefixRe = regexp.MustCompile(`^\d+/`)
)

func (r *TerraformRemoteStateNamingRule) Name() string {
	return "terraform_remote_state_naming"
}

func (r *TerraformRemoteStateNamingRule) Enabled() bool {
	return false
}

func (r *TerraformRemoteStateNamingRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformRemoteStateNamingRule) Link() string {
	return "https://github.com/lucidsoftware/tflint-lucid-conventions/blob/main/README.md#terraform_remote_state_naming"
}

func (r *TerraformRemoteStateNamingRule) Check(runner tflint.Runner) error {
	content, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "data",
				LabelNames: []string{"type", "name"},
				Body: &hclext.BodySchema{
					Mode: hclext.SchemaJustAttributesMode,
				},
			},
		},
	}, nil)
	if err != nil {
		return err
	}

	for _, block := range content.Blocks {
		if block.Labels[0] != "terraform_remote_state" {
			continue
		}
		name := block.Labels[1]

		configAttr, ok := block.Body.Attributes["config"]
		if !ok {
			// Some declarations may use other shapes (e.g. backend-specific args
			// at the top level). If we can't find config.key, skip silently.
			continue
		}

		key, ok := extractKeyFromObjectExpr(configAttr.Expr)
		if !ok {
			// key not present, or not a static string (e.g. interpolation).
			continue
		}

		if standardKeyRe.MatchString(key) {
			expected := strings.ReplaceAll(stripPrefixRe.ReplaceAllString(key, ""), "-", "_")
			if name != expected {
				if err := runner.EmitIssue(
					r,
					fmt.Sprintf(
						"local name %q does not match the standard convention for key %q (expected %q)",
						name, key, expected,
					),
					block.DefRange,
				); err != nil {
					return err
				}
			}
		} else {
			if !strings.HasPrefix(name, "legacy_") {
				if err := runner.EmitIssue(
					r,
					fmt.Sprintf(
						"local name %q references legacy key %q and must start with the prefix \"legacy_\"",
						name, key,
					),
					block.DefRange,
				); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// extractKeyFromObjectExpr walks an object expression (the value of `config = { ... }`)
// and returns the value of its `key` entry if it's a static string literal.
// Returns ("", false) if the key entry is missing or non-static.
func extractKeyFromObjectExpr(expr hcl.Expression) (string, bool) {
	kvPairs, diags := hcl.ExprMap(expr)
	if diags.HasErrors() || kvPairs == nil {
		return "", false
	}
	for _, kv := range kvPairs {
		keyName, diags := kv.Key.Value(nil)
		if diags.HasErrors() || keyName.Type() != cty.String {
			continue
		}
		if keyName.AsString() != "key" {
			continue
		}
		val, diags := kv.Value.Value(nil)
		if diags.HasErrors() || val.Type() != cty.String {
			return "", false
		}
		return val.AsString(), true
	}
	return "", false
}
