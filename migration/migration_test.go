package migration_test

import (
	"encoding/json"
	"os"
	"strings"

	. "code.cloudfoundry.org/capi-integration-tests/helpers/resource_helpers"
	. "code.cloudfoundry.org/capi-integration-tests/helpers/v3_helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("V2 behavior with DEA backend", func() {
	It("can restart the app", func() {
		appName := os.Getenv("APP")
		Expect(cf.Cf("stop", appName).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

		Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
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
		appName := os.Getenv("APP")
		Expect(cf.Cf("restage", appName).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora!"))
	})

	It("persists multiple routes", func() {
		type AppRoutesResponse struct {
			TotalResults int `json:"total_results"`
		}

		appGuid := cf.Cf("app", os.Getenv("APP_WITH_MULTIPLE_ROUTES"), "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()

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
		appName := os.Getenv("APP_WITH_SERVICE_BINDING")
		Expect(cf.Cf("stop", appName).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

		Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora!"))

		appEnv := cf.Cf("env", appName).Wait(DEFAULT_TIMEOUT).Out.Contents()

		Expect(appEnv).To(ContainSubstring("credentials"))
	})

	It("persists the syslog drain url", func() {
		appName := os.Getenv("APP_WITH_SYSLOG_DRAIN_URL")
		appEnv := cf.Cf("env", appName).Wait(DEFAULT_TIMEOUT).Out.Contents()
		appGuid := strings.TrimSpace(string(cf.Cf("app", appName, "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()))
		apiEndpoint := os.Getenv("API_ENDPOINT")
		expectedSyslogDrainUrl := "logit.io/drain/here"

		Expect(appEnv).To(ContainSubstring(expectedSyslogDrainUrl))

		syslogDrainUrlsResponse := helpers.Curl("bulk_api:bulk-password@" + apiEndpoint + "/v2/syslog_drain_urls").Wait(DEFAULT_TIMEOUT).Out.Contents()
		var syslogDrainUrls SyslogDrainUrls

		json.Unmarshal([]byte(syslogDrainUrlsResponse), &syslogDrainUrls)

		Expect(syslogDrainUrls.Results[appGuid][0]).To(Equal(expectedSyslogDrainUrl))
	})

	It("repushes a buildpack app successfully", func() {
		appName := os.Getenv("BUILDPACK_APP_TO_REPUSH")
		Expect(cf.Cf("push", appName, "-p", "../assets/dora").Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora!"))
	})
})
