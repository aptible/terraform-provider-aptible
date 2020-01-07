resource "example_server" "hello_world" {
    name = "hello_world"
}

output "msg" {
  value = "hello, world!"
}