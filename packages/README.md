# KubeZero Packages

The packages are pre-defined resources could be used as they are in the registry.

## How to use

Copy (or link) the package you want to the `registry` directory, and KubeZero will pick it up and enable it using GitOps.

Example:

```shell
cp -a packages/virtual-management registry/virtual-management
```

Then you can bootstrap for local testing.

You can also commit it:

```shell
git add registry/virtual-management
git commit -m "feat: enable virtual-management cluster"
```
