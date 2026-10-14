# Terraform Provider Tsuru [![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/tsuru/terraform-provider-tsuru?label=release)](https://github.com/tsuru/terraform-provider-tsuru/releases) [![license](https://img.shields.io/github/license/tsuru/terraform-provider-tsuru.svg)]()

- Usage
  - [Provider Documentation](https://registry.terraform.io/providers/tsuru/tsuru/latest/docs)
  - [Tsuru Documentation](https://docs.tsuru.io/)

The tsuru provider for Terraform is a plugin that enables lifecycle management of Tsuru resources. This provider is under early development by Tsuru Team.

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) 1.4.4+
-	[Go](https://golang.org/doc/install) 1.21.0+ (to build the provider plugin)

## Getting Started

### For macOS Users

#### 1. Install Prerequisites

First, install Go and Terraform using Homebrew:

```bash
# Install Go
brew install go

# Install Terraform
brew tap hashicorp/tap
brew install hashicorp/tap/terraform
```

Verify the installations:

```bash
go version
terraform version
```

#### 2. (Optional) Build and Install the Provider Locally

> 💡 **Note:** You don’t need to build the provider manually unless you want to contribute to its development or test local changes.  
> Regular users can install the provider automatically from the [Terraform Registry](https://registry.terraform.io/providers/tsuru/tsuru/latest).

If you want to build the provider from source:

```bash
# Clone the repository
git clone https://github.com/tsuru/terraform-provider-tsuru.git
cd terraform-provider-tsuru

# Build and install the provider locally
make install
```

This will compile the provider and install it to `~/.terraform.d/plugins/registry.terraform.io/tsuru/tsuru/<VERSION>/darwin_<arch>/`.

#### 3. Verify Installation

Create a test Terraform configuration:

```bash
mkdir test-provider
cd test-provider
cat > main.tf << 'EOF'
terraform {
  required_providers {
    tsuru = {
      source  = "registry.terraform.io/tsuru/tsuru"
      version = "2.15.8"
    }
  }
}

provider "tsuru" {
  host = "https://tsuru.example.com"  # Replace with your Tsuru API URL
}
EOF
```

Initialize Terraform to verify the provider is recognized:

```bash
terraform init
```

If successful, you should see: `Terraform has been successfully initialized!`

### For Linux Users

```bash
# Install Go (Ubuntu/Debian)
sudo apt update
sudo apt install golang-go

# Install Terraform
wget https://releases.hashicorp.com/terraform/1.4.4/terraform_1.4.4_linux_amd64.zip
unzip terraform_1.4.4_linux_amd64.zip
sudo mv terraform /usr/local/bin/

# Build and install the provider
make install
```

### Configuration

The provider can be configured using environment variables or directly in your Terraform configuration:

#### Environment Variables

```bash
export TSURU_HOST="https://tsuru.example.com"
export TSURU_TOKEN="your-api-token"
```

#### Terraform Configuration

```hcl
provider "tsuru" {
  host  = "https://tsuru.example.com"
  token = "your-api-token"
  
  # Optional: disable certificate verification (not recommended for production)
  skip_cert_verification = false
}
```

For more configuration options, see the [provider documentation](https://registry.terraform.io/providers/tsuru/tsuru/latest/docs).

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/tsuru/terraform-provider-tsuru.git
cd terraform-provider-tsuru

# Build the provider
make build

# Or build and install locally
make install
```

### Running Tests

```bash
# Run all tests
make test

# Note: Tests require access to a Tsuru API instance
# Set the required environment variables before running tests:
export TSURU_HOST="https://your-tsuru-instance.com"
export TSURU_TOKEN="your-token"
```

### Available Make Commands

- `make build` - Compile the provider binary
- `make install` - Build and install the provider locally
- `make test` - Run all tests
- `make lint` - Run linting checks
- `make generate-docs` - Generate documentation from code
- `make uninstall` - Remove locally installed provider

## Contributing

Contributions are welcome! Here's how you can help:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/my-improvement`
3. **Make your changes**
4. **Test your changes**: `make build` and verify it compiles
5. **Commit your changes**: `git commit -am 'Add some feature'`
6. **Push to the branch**: `git push origin feature/my-improvement`
7. **Open a Pull Request**

### Good First Issues

If you're new to the project, look for issues labeled:
- `good first issue` - Good for newcomers
- `documentation` - Documentation improvements
- `help wanted` - Extra attention is needed

### Development Tips for Beginners

- Start with documentation improvements - they're a great way to learn the project
- Read existing code before making changes
- Ask questions in issues if you're unsure
- Test your changes locally before submitting a PR

## Examples

See the [`examples/`](./examples/) directory for usage examples of various resources and data sources.

Basic example:

```hcl
terraform {
  required_providers {
    tsuru = {
      source = "tsuru/tsuru"
    }
  }
}

provider "tsuru" {
  host = "https://tsuru.example.com"
}

resource "tsuru_app" "example" {
  name        = "my-app"
  platform    = "python"
  pool        = "my-pool"
  team_owner  = "my-team"
  description = "My application managed by Terraform"
}
```

## License

This project is licensed under the BSD 3-Clause License - see the [LICENSE](LICENSE) file for details.
