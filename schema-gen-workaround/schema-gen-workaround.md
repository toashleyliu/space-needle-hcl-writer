# Schema Generation Workaround

When the QuickSight dashboard and analysis definition attribute is present in the CloudFormation schema during compilation, an issue occurs with the AWSCC provider module that fails due to OOM. This workaround reduces the complexity of the generated objects and makes use swap space to generate a complete schema locally. 

## Steps to Reproduce

1. Clone [Terraform Provider AWSCC](https://github.com/hashicorp/terraform-provider-awscc) repository 

```
$ mkdir -p $HOME/development/terraform-providers/; cd $HOME/development/terraform-providers/
$ git clone git@github.com:terraform-providers/terraform-provider-awscc
```

2. Apply the git patch file

```
$ git apply $HOME/space-needle-hcl-writer/space-needle-hcl-writer/schema-gen-workaround/workaround.patch
```

3. Add [swap space](https://repost.aws/knowledge-center/ec2-memory-swap-file), replacing step 1 with 

```
$ sudo dd if=/dev/zero of=/swapfile bs=128M count=256
```

4. Enter the provider directory and install tools for provider then generate resources

```
$ make tools
$ make all
```

5. Follow [Using the Provider](https://github.com/hashicorp/terraform-provider-awscc/blob/main/contributing/DEVELOPMENT.md#using-the-provider) to apply development overrides

Now running `terraform providers schema -json` in a directory containing a .tf file with a [provider configuration](https://developer.hashicorp.com/terraform/language/providers/configuration) will generate the complete schema for AWSCC QuickSight resources.
