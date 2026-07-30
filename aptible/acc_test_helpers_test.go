package aptible

import (
	"context"
	"testing"

	aptibleapi "github.com/aptible/aptible-api-go/aptibleapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

// testEnvironment is a minimal struct to replace aptible.Environment from go-deploy.
// It provides the env ID for test config generation.
type testEnvironment struct {
	ID int64
}

// WithTestAccEnvironment creates a temporary environment for acceptance testing,
// runs the provided test function, then cleans up.
func WithTestAccEnvironment(t *testing.T, fn func(env testEnvironment)) {
	t.Helper()

	m := testAccProvider.Meta().(*providerMetadata)
	client := m.Client
	ctx := m.APIContext(context.Background())

	handle := "tf-acc-" + acctest.RandString(10)

	stackId := int32(testStackId)
	req := aptibleapi.NewCreateAccountRequest("development", handle, testOrganizationId)
	req.SetStackId(stackId)

	env, _, err := client.AccountsAPI.CreateAccount(ctx).CreateAccountRequest(*req).Execute()
	if err != nil {
		t.Fatalf("Unable to create test environment: %s", err.Error())
		return
	}

	defer func() {
		_, err := client.AccountsAPI.DeleteAccount(ctx, env.Id).Execute()
		if err != nil {
			t.Logf("Warning: failed to clean up test environment %d: %s", env.Id, err.Error())
		}
	}()

	fn(testEnvironment{ID: int64(env.Id)})
}
