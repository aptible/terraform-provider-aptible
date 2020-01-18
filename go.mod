module github.com/reggregory/terraform-provider-aptible

go 1.13

require (
	github.com/aptible/go-deploy v0.0.0-20191219200048-d7794dd95230
	github.com/go-openapi/runtime v0.19.9
	github.com/go-openapi/strfmt v0.19.4
	github.com/hashicorp/terraform v0.12.19
)

replace github.com/aptible/go-deploy => github.com/reggregory/go-deploy v0.0.0-20200117203405-01e8e3e42719
