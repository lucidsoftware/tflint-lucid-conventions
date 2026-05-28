package rules

import (
	"testing"

	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestAwsIamPolicyDocumentRequired(t *testing.T) {
	rule := &AwsIamPolicyDocumentRequiredRule{}

	cases := []struct {
		name       string
		content    string
		issueCount int
	}{
		{
			name: "pass: simple data source ref",
			content: `
resource "aws_iam_policy" "ok" {
  policy = data.aws_iam_policy_document.example.json
}`,
			issueCount: 0,
		},
		{
			name: "pass: indexed data source ref (count)",
			content: `
resource "aws_iam_policy" "ok" {
  policy = data.aws_iam_policy_document.example[0].json
}`,
			issueCount: 0,
		},
		{
			name: "pass: indexed data source ref (for_each)",
			content: `
resource "aws_iam_policy" "ok" {
  policy = data.aws_iam_policy_document.example["key"].json
}`,
			issueCount: 0,
		},
		{
			name: "pass: local reference",
			content: `
resource "aws_iam_policy" "ok" {
  policy = local.my_policy_json
}`,
			issueCount: 0,
		},
		{
			name: "pass: variable reference",
			content: `
resource "aws_iam_policy" "ok" {
  policy = var.policy_json
}`,
			issueCount: 0,
		},
		{
			name: "pass: module output",
			content: `
resource "aws_iam_policy" "ok" {
  policy = module.example.policy_json
}`,
			issueCount: 0,
		},
		{
			name: "fail: jsonencode",
			content: `
resource "aws_iam_policy" "bad" {
  policy = jsonencode({ Version = "2012-10-17", Statement = [] })
}`,
			issueCount: 1,
		},
		{
			name: "fail: file()",
			content: `
resource "aws_iam_policy" "bad" {
  policy = file("policy.json")
}`,
			issueCount: 1,
		},
		{
			name: "fail: raw string",
			content: `
resource "aws_iam_policy" "bad" {
  policy = "{\"Version\":\"2012-10-17\",\"Statement\":[]}"
}`,
			issueCount: 1,
		},
		{
			name: "multiple resources: mixed pass and fail",
			content: `
resource "aws_iam_policy" "ok" {
  policy = data.aws_iam_policy_document.example.json
}
resource "aws_iam_policy" "bad" {
  policy = jsonencode({ Version = "2012-10-17", Statement = [] })
}`,
			issueCount: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"main.tf": tc.content})
			if err := rule.Check(runner); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if len(runner.Issues) != tc.issueCount {
				t.Errorf("expected %d issue(s), got %d:", tc.issueCount, len(runner.Issues))
				for _, issue := range runner.Issues {
					t.Errorf("  - %s", issue.Message)
				}
			}
		})
	}
}
