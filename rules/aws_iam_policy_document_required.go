package rules

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// AwsIamPolicyDocumentRequiredRule enforces that aws_iam_policy resources
// set their policy attribute via a data.aws_iam_policy_document reference
// rather than jsonencode(), file(), or a raw string.
type AwsIamPolicyDocumentRequiredRule struct {
	tflint.DefaultRule
}

func (r *AwsIamPolicyDocumentRequiredRule) Name() string {
	return "aws_iam_policy_document_required"
}

func (r *AwsIamPolicyDocumentRequiredRule) Enabled() bool {
	return true
}

func (r *AwsIamPolicyDocumentRequiredRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *AwsIamPolicyDocumentRequiredRule) Link() string {
	return "https://github.com/lucidsoftware/tflint-lucid-conventions/blob/main/README.md#aws_iam_policy_document_required"
}

func (r *AwsIamPolicyDocumentRequiredRule) Check(runner tflint.Runner) error {
	body, err := runner.GetResourceContent("aws_iam_policy", &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{
			{Name: "policy"},
		},
	}, nil)
	if err != nil {
		return err
	}

	for _, resource := range body.Blocks {
		attr, exists := resource.Body.Attributes["policy"]
		if !exists {
			continue
		}
		if msg, bad := badPolicyExprMessage(attr.Expr); bad {
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf(`aws_iam_policy "%s": %s`, resource.Labels[1], msg),
				attr.Expr.StartRange(),
			); err != nil {
				return err
			}
		}
	}
	return nil
}

// badPolicyExprMessage reports whether expr is a disallowed policy value.
// Raw strings, string templates, jsonencode(), and file() are disallowed.
// Module outputs, variables, locals, and other references are permitted.
func badPolicyExprMessage(expr hcl.Expression) (string, bool) {
	switch e := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		return "policy must use a data.aws_iam_policy_document data source; raw string literals are not allowed", true
	case *hclsyntax.TemplateExpr:
		return "policy must use a data.aws_iam_policy_document data source; string templates are not allowed", true
	case *hclsyntax.FunctionCallExpr:
		if e.Name == "jsonencode" || e.Name == "file" {
			return fmt.Sprintf("policy must use a data.aws_iam_policy_document data source; %s() is not allowed", e.Name), true
		}
		return "", false
	default:
		return "", false
	}
}
