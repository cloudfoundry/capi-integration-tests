package migration_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	. "code.cloudfoundry.org/capi-integration-tests/helpers/resource_helpers"
	. "code.cloudfoundry.org/capi-integration-tests/helpers/v3_helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("V2 behavior with diego backend", func() {
	It("can restart the app", func() {
		appName := os.Getenv("DIEGO_APP")
		Expect(cf.Cf("stop", appName).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

		Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora!"))
	})

	It("can restage the app", func() {
		appName := os.Getenv("DIEGO_APP")
		Expect(cf.Cf("restage", appName).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora!"))
	})

	It("persists multiple routes", func() {
		type AppRoutesResponse struct {
			TotalResults int `json:"total_results"`
		}

		appGuid := cf.Cf("app", os.Getenv("DIEGO_APP_WITH_MULTIPLE_ROUTES"), "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()

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
		appGuid := cf.Cf("app", os.Getenv("DIEGO_APP_WITH_ENV_VARS"), "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()

		cfResponse := cf.Cf("curl", "/v2/apps/"+strings.TrimSpace(string(appGuid))).Wait(DEFAULT_TIMEOUT).Out.Contents()
		json.Unmarshal(cfResponse, &appResource)

		envVars := appResource.Entity.EnvironmentJson

		Expect(envVars).Should(HaveKeyWithValue("kittens", "cutest"))
	})

	It("service bindings are available in env", func() {
		appName := os.Getenv("DIEGO_APP_WITH_SERVICE_BINDING")
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
		appName := os.Getenv("DIEGO_APP_WITH_SYSLOG_DRAIN_URL")
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
		appName := os.Getenv("DIEGO_BUILDPACK_APP_TO_REPUSH")

		Expect(cf.Cf("push", appName, "-p", "../assets/dora").Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora!"))
	})

	It("repushes a docker app successfully", func() {
		type envStruct struct {
			Port string `json:"PORT", json:"port"`
		}
		appName := os.Getenv("DIEGO_DOCKER_APP_TO_REPUSH")

		env := envStruct{}
		envStr := helpers.CurlApp(appName, "/env")
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

	It("can ssh to a diego app", func() {
		appName := os.Getenv("DIEGO_APP")

		envCmd := cf.Cf("ssh", "-v", appName, "-c", "/usr/bin/env && /usr/bin/env >&2")
		Expect(envCmd.Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		output := string(envCmd.Out.Contents())
		stdErr := string(envCmd.Err.Contents())

		Expect(string(output)).To(MatchRegexp(fmt.Sprintf(`VCAP_APPLICATION=.*"application_name":"%s"`, appName)))
		Expect(string(output)).To(MatchRegexp("INSTANCE_INDEX=0"))

		Expect(string(stdErr)).To(MatchRegexp(fmt.Sprintf(`VCAP_APPLICATION=.*"application_name":"%s"`, appName)))
		Expect(string(stdErr)).To(MatchRegexp("INSTANCE_INDEX=0"))

		Eventually(cf.Cf("logs", appName, "--recent"), DEFAULT_TIMEOUT).Should(Say("Successful remote access"))
		Eventually(cf.Cf("events", appName), DEFAULT_TIMEOUT).Should(Say("audit.app.ssh-authorized"))
	})

	It("persists multiple app ports exposed on an app with health check none", func() {
		appName := os.Getenv("DIEGO_APP_WITH_MULTIPLE_PORTS")
		appGuid := cf.Cf("app", appName, "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()

		app := V2App{}
		cfResponse := cf.Cf("curl", "/v2/apps/"+strings.TrimSpace(string(appGuid))).Wait(DEFAULT_TIMEOUT).Out.Contents()
		err := json.Unmarshal([]byte(cfResponse), &app)
		Expect(err).NotTo(HaveOccurred())

		Expect(app.Entity.Ports).To(ConsistOf(9191, 9090))

		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Lattice"))
	})
})
