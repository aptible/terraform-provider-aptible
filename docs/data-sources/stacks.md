# Stack Data Source

[Stacks](https://deploy-docs.aptible.com/docs/stacks)
are the underlying virtualized infrastructure (EC2 instances, private network, etc.) your resources
are deployed on.

## Example Usage

### Determining the Stack ID

If you have an Stack with the name "test-stack" you can
create the data source:

```hcl
data "aptible_stack" "test-stack" {
    name = "test-stack"
}
```

Once defined, you can use this data source in your resource definitions.
For example, when defining an environment:

```hcl
resource "aptible_environment" "test-env" {
    stack_id = data.aptible_stack.test-stack.stack_id
    org_id = data.aptible_stack.test-stack.org_id
    name = "test-env"
}
```

## Argument Reference

- `name` - The name of the Stack

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique ID for an Stack suitable for use in `stack_id` attributes
- `org_id` - If the stack is a [dedicated stack](https://deploy-docs.aptible.com/docs/shared-dedicated#dedicated-stacks),
you will also receive an id that corresponds to that organization. If it is a shared stack, this value will be empty.
