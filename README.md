# henkan

Generate Terraform templates from your existiong SAKURA Cloud resources(Inspired by [former2](https://github.com/iann0036/former2))

## Installation

You can download from ownload and extract files for your own platform from [GitHub Releases](https://github.com/RyoMasumura1201/henkan/releases)

## Usage

### generate

The `generate` command will generate Terraform template from all discovered resources and write them to the filename specified.

```
henkan generate \
  --output "terraform.hcl" \
  --filter "myapp" \
  --services "server"
```

#### Options

```
Options:
  -e, --exclude-services strings   list of services to exclude (can be comma separated)
  -f, --filter string              search filter for discovered resources (can be comma separated)
  -h, --help                       help for generate
  -o, --output string              filename for Terraform output (default "output.tf")
  -s, --services strings           list of services to include (can be comma separated (default: ALL))
```

#### Service Names

Below is a list of services for use with the `--services` and `--exclude-services` argument:

- server
- disk
- switch
- internet
