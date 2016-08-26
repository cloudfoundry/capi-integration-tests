package migration_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	APP_NAME                      = "meow"
	APP_WITH_ENV_VARS             = "meow"
	APP_WITH_SERVICE_BINDING_NAME = "service-app"
	APP_WITH_MULTIPLE_ROUTES_NAME = "woof"
	DEFAULT_TIMEOUT               = 30 * time.Second
	CF_PUSH_TIMEOUT               = 2 * time.Minute
)

func TestMigration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Migration Suite")
}
