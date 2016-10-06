package migration_test

import (
	"os"

	. "code.cloudfoundry.org/capi-integration-tests/helpers/v3_helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Detected start commands persist", func() {

	It("restarts JAVA apps", func() {
		appName := os.Getenv("JAVA_APP")
		Expect(cf.Cf("stop", appName).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

		Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Hello, from your friendly neighborhood Java JSP!"))
	})

	It("restarts NODE apps", func() {
		appName := os.Getenv("NODE_APP")
		Expect(cf.Cf("stop", appName).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

		Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Hello from a node app!"))
	})

	It("restarts GOLANG apps", func() {
		appName := os.Getenv("GOLANG_APP")
		Expect(cf.Cf("stop", appName).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

		Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("go, world"))
	})

	It("restarts PHP apps", func() {
		appName := os.Getenv("PHP_APP")
		Expect(cf.Cf("stop", appName).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

		Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Hello from php"))
	})

	It("restarts PYTHON apps", func() {
		appName := os.Getenv("PYTHON_APP")
		Expect(cf.Cf("stop", appName).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

		Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("python, world"))
	})
})
