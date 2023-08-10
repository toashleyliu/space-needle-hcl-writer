# Schema Generation Workaround

When the QuickSight dashboard and analysis definition attribute is present in the CloudFormation schema during compilation, an issue occurs with the AWSCC provider module that fails due to OOM. This workaround reduces the complexity of the generated objects and makes use swap space to generate a complete schema locally. 

## Steps to Reproduce

1. Clone the [Terraform Provider AWSCC](https://github.com/hashicorp/terraform-provider-awscc) repository.

```
$ mkdir -p $HOME/development/terraform-providers/; cd $HOME/development/terraform-providers/
$ git clone git@github.com:terraform-providers/terraform-provider-awscc
```

2. Apply the git patch file. This will remove `StringAttributes` from the schema map and into a list. The map will then reference that list, effectively reducing the complexity of the map objects. Additionally, only QuickSight resource schemas will be generated to reduce execution time.

```
$ git apply $HOME/space-needle-hcl-writer/space-needle-hcl-writer/schema-gen-workaround/workaround.patch
```

3. Replace [AWS_QuickSight_Analysis.json](https://github.com/hashicorp/terraform-provider-awscc/blob/main/internal/service/cloudformation/schemas/AWS_QuickSight_Analysis.json) with the latest [analysis schema](https://code.amazon.com/packages/AWSCloudFormationResourceSchemas/blobs/mainline/--/resources/schemas/aws-quicksight-analysis.json) and [AWS_QuickSight_Dashboard.json](https://github.com/hashicorp/terraform-provider-awscc/blob/main/internal/service/cloudformation/schemas/AWS_QuickSight_Dashboard.json)` with the latest [dashboard schema](https://code.amazon.com/packages/AWSCloudFormationResourceSchemas/blobs/mainline/--/resources/schemas/aws-quicksight-dashboard.json). This will add the Definition attribute keep any schema changes up to date. 

4. Within the `AWS_QuickSight_Analysis.json` and `AWS_QuickSight_Dashboard.json`, search for and replace the regex pattern `"^[^\\u0000-\\u00FF]$"` with `"^.*$"`. This will prevent the build from rejecting the JSON. 

5. Add [swap space](https://repost.aws/knowledge-center/ec2-memory-swap-file), replacing step 1 with 

```
$ sudo dd if=/dev/zero of=/swapfile bs=128M count=256
```

6. Enter the provider directory, install tools for provider then generate resources.

```
$ make tools
$ make all
```

7. Follow [Using the Provider](https://github.com/hashicorp/terraform-provider-awscc/blob/main/contributing/DEVELOPMENT.md#using-the-provider) to apply development overrides.

Now running `terraform providers schema -json` in a directory containing a .tf file with an AWSCC [provider configuration](https://developer.hashicorp.com/terraform/language/providers/configuration) will generate the complete schema for AWSCC QuickSight resources.
