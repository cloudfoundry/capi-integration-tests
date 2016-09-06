package migration_test

import (
	"encoding/json"
	"os"
	"strings"

	. "code.cloudfoundry.org/capi-integration-tests/helpers/v3_helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("V2 behavior with DEA backend", func() {
	It("can restart the app", func() {
		Expect(cf.Cf("stop", os.Getenv("APP_NAME")).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(os.Getenv("APP_NAME"))
		}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

		Expect(cf.Cf("start", os.Getenv("APP_NAME")).Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(os.Getenv("APP_NAME"))
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora!"))
	})

	It("persists environment variables", func() {
		type AppResource struct {
			Entity struct {
				EnvironmentJson map[string]string `json:"environment_json"`
			} `json:"entity"`
		}

		var appResource AppResource
		appGuid := cf.Cf("app", os.Getenv("APP_WITH_ENV_VARS"), "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()

		cfResponse := cf.Cf("curl", "/v2/apps/"+strings.TrimSpace(string(appGuid))).Wait(DEFAULT_TIMEOUT).Out.Contents()
		json.Unmarshal(cfResponse, &appResource)

		envVars := appResource.Entity.EnvironmentJson

		Expect(envVars).Should(HaveKeyWithValue("kittens", "cutest"))
	})

	It("can restage the app", func() {
		Expect(cf.Cf("restage", os.Getenv("APP_NAME")).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(os.Getenv("APP_NAME"))
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora!"))
	})

	It("persists multiple routes", func() {
		type AppRoutesResponse struct {
			TotalResults int `json:"total_results"`
		}

		appGuid := cf.Cf("app", os.Getenv("APP_WITH_MULTIPLE_ROUTES_NAME"), "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()

		var appRoutesResponse AppRoutesResponse
		cfResponse := cf.Cf("curl", "/v2/apps/"+strings.TrimSpace(string(appGuid))+"/routes").Wait(DEFAULT_TIMEOUT).Out.Contents()
		json.Unmarshal(cfResponse, &appRoutesResponse)

		Expect(appRoutesResponse.TotalResults).To(Equal(3))
	})

	It("persists environment variables", func() {
		type AppResource struct {
			Entity struct {
				EnvironmentJson map[string]string `json:"environment_json"`
			} `json:"entity"`
		}

		var appResource AppResource
		appGuid := cf.Cf("app", os.Getenv("APP_WITH_ENV_VARS"), "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()

		cfResponse := cf.Cf("curl", "/v2/apps/"+strings.TrimSpace(string(appGuid))).Wait(DEFAULT_TIMEOUT).Out.Contents()
		json.Unmarshal(cfResponse, &appResource)

		envVars := appResource.Entity.EnvironmentJson

		Expect(envVars).Should(HaveKeyWithValue("kittens", "cutest"))
	})

	It("service bindings are available in env", func() {
		Expect(cf.Cf("stop", os.Getenv("APP_WITH_SERVICE_BINDING_NAME")).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(os.Getenv("APP_WITH_SERVICE_BINDING_NAME"))
		}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

		Expect(cf.Cf("start", os.Getenv("APP_WITH_SERVICE_BINDING_NAME")).Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(os.Getenv("APP_WITH_SERVICE_BINDING_NAME"))
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora!"))

		appEnv := cf.Cf("env", os.Getenv("APP_WITH_SERVICE_BINDING_NAME")).Wait(DEFAULT_TIMEOUT).Out.Contents()

		Expect(appEnv).To(ContainSubstring("credentials"))
	})

	It("persists the syslog drain url", func() {
		appEnv := cf.Cf("env", os.Getenv("APP_WITH_SYSLOG_DRAIN_URL_NAME")).Wait(DEFAULT_TIMEOUT).Out.Contents()

		Expect(appEnv).To(ContainSubstring("logit.io/drain/here"))
	})

	It("repushes a buildpack app successfully", func() {
		appName := os.Getenv("BUILDPACK_APP_TO_REPUSH")
		Expect(cf.Cf("push", appName, "-p", "../assets/dora").Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora!"))
	})

	It("repushes a docker app successfully", func() {
		type envStruct struct {
			Port string `json:"PORT", json:"port"`
		}

		appName := os.Getenv("DOCKER_APP_TO_REPUSH")

		envStr := helpers.CurlApp(appName, "/env")

		env := envStruct{}
		err := json.Unmarshal([]byte(envStr), &env)
		Expect(err).NotTo(HaveOccurred())

		Expect(env.Port).To(Equal("8080"))

		Expect(cf.Cf("push", appName,
			"-o", "cloudfoundry/diego-docker-app-custom:latest",
		).Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, CF_PUSH_TIMEOUT).Should(Equal("0"))

		envStr = helpers.CurlApp(appName, "/env")
		err = json.Unmarshal([]byte(envStr), &env)
		Expect(err).NotTo(HaveOccurred())

		Expect(env.Port).To(Equal("7070"))
	})
})
