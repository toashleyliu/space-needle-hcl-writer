# SpaceNeedleHCLWriter

## Install

```
$ mkdir space-needle-hcl-writer
$ cd space-needle-hcl-writer
$ git clone https://github.com/toashleyliu/space-needle-hcl-writer.git
$ cd space-needle-hcl-writer
```

## Run
```
To convert all files in folder:

./build/bin/SpaceNeedleHCLWriter --src <source path> --dest <destination path> --aws-account-id <AWS account ID> --region <region> [--override <"property name, property resource ARN">...] [--prefix-for-all-resources] [--terraform-file-name <converted file name>]
```
