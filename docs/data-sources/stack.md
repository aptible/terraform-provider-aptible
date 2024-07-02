# Stack Data Source

[Stacks](https://www.aptible.com/docs/core-concepts/architecture/stacks)
are the underlying virtualized infrastructure (EC2 instances, private network,
etc.) your resources are deployed on.

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

__Note__ - If your environment is meant to be created on a dedicated stack you can omit the `org_id`
field in the example above.

## Argument Reference

- `name` - The name of the Stack

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `stack_id` - The unique ID for an Stack suitable for use in `stack_id` attributes
- `org_id` - If the stack is a [dedicated stack](https://www.aptible.com/docs/core-concepts/architecture/stacks#dedicated-stacks-isolated),
you will also receive an id that corresponds to that organization. If it is a shared stack, this value will be empty.
