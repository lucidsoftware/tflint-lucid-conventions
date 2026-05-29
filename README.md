# TFLint Lucid Conventions
[![Build Status](https://github.com/lucidsoftware/tflint-lucid-conventions/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/lucidsoftware/tflint-lucid-conventions/actions)

Custom [tflint](https://github.com/terraform-linters/tflint) rules that enforce Lucid's Terraform/OpenTofu coding conventions.

## Requirements

- TFLint v0.46+
- Go v1.25+

## Installation

You can install the plugin with `tflint --init`. Declare it in `.tflint.hcl`:

```hcl
plugin "lucid-conventions" {

  enabled = true

  version = "0.1.0"
  source  = "github.com/lucidsoftware/tflint-lucid-conventions"

  # Optionally omit this to use Keyless verification
  signing_key = <<-KEY
    -----BEGIN PGP PUBLIC KEY BLOCK-----

    mG8EaXACCRMFK4EEACIDAwQdSSKnORcu1YozK8MQMrLJ4LBN171J/Zf3G//FUxX8
    hvlh1CyPvcTgi1UYuj8wWCF19L2GazNv32MmPDk9ueGzfmTsp5ONHddg4Tiu6SZV
    zgfkfyrhJfq9h4A1FTYq0oC0JUx1Y2lkIFNvZnR3YXJlLCBJbmMuICh0ZmxpbnQg
    c2lnbmluZymIswQTEwkAOxYhBC/Y2UK9yTIieUACoT8p7Wh2KOJWBQJpcAIJAhsD
    BQsJCAcCAiICBhUKCQgLAgQWAgMBAh4HAheAAAoJED8p7Wh2KOJWUV4BgIuR7YuX
    ON5YSi9+XPbg6zxEihHRp9NWs76ipYHJdd5eXKRYeW69MrQgr7TY5TyDFwF/Q2Be
    1Xy0JwzT5zmg6vnwPc3+9I5oq7rWEEbKAP4PZd0pYLsH5MqghYrcE1FCXf7+
    =LNHa
    -----END PGP PUBLIC KEY BLOCK-----
  KEY
}
```

## Rules

|Name|Description|Severity|Enabled|
| --- | --- | --- | --- |
|aws_iam_policy_document_required|Requires `aws_iam_policy.policy` to be set from a `data.aws_iam_policy_document` reference rather than `jsonencode`, `file`, or a raw string|ERROR|-|
|terraform_remote_state_naming|Enforces local-name conventions on `data "terraform_remote_state"` declarations|ERROR|-|

### aws_iam_policy_document_required

Enforces that `aws_iam_policy` resources set their `policy` attribute via a `data.aws_iam_policy_document` reference rather than `jsonencode()`, `file()`, or a raw string literal. This keeps policy authoring consistent across the repo and benefits from the data source's validation.

**Configuration:**

```hcl
rule "aws_iam_policy_document_required" {
  enabled = true
}
```

### terraform_remote_state_naming

Enforces local-name conventions on `data "terraform_remote_state"` declarations.

**Two conventions:**

1. **Standard keys** (match `^\d+/[a-z0-9-]+$`): the local name must equal the key with the leading `\d+/` stripped and `-` replaced with `_`.
2. **Legacy keys** (anything else): the local name must start with the prefix `legacy_`.

Keys that aren't static string literals (e.g. with `${var.x}` interpolation) are skipped.

**Configuration:**

No options — just enable the rule.

```hcl
rule "terraform_remote_state_naming" {
  enabled = true
}
```

**Example of valid code:**

```hcl
# Standard key: name derived mechanically.
data "terraform_remote_state" "com_preprod_us" {
  backend = "s3"
  config = {
    bucket = "*************-spacelift-states-******"
    key    = "1/com-preprod-us"
    region = "us-east-1"
  }
}

# Legacy key: name must start with legacy_.
data "terraform_remote_state" "legacy_bi" {
  backend = "s3"
  config = {
    bucket = "lucid-terraform"
    key    = "tfstate"
    region = "us-east-1"
  }
}
```

**Example of invalid code:**

```hcl
# Standard key but the local name doesn't match.
data "terraform_remote_state" "preprod_us" {
  backend = "s3"
  config = {
    key    = "1/com-preprod-us"
    bucket = "*************-spacelift-states-******"
    region = "us-east-1"
  }
}
# Error: local name "preprod_us" does not match the standard convention for key "1/com-preprod-us" (expected "com_preprod_us")

# Legacy key but missing the legacy_ prefix.
data "terraform_remote_state" "utility_infrastructure" {
  backend = "s3"
  config = {
    key    = "tfstate"
    bucket = "lucid-terraform"
    region = "us-east-1"
  }
}
# Error: local name "utility_infrastructure" references legacy key "tfstate" and must start with the prefix "legacy_"
```

## Building the plugin

```
$ make
```

Install locally for testing:

```
$ make install
```

Then run tflint with the plugin enabled:

```
$ cat << EOS > .tflint.hcl
plugin "lucid-conventions" {
  enabled = true
}
EOS
$ tflint
```
