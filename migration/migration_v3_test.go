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

var _ = Describe("Using v3 endpoints", func() {
	Context("For apps pushed using v2 before the migration", func() {
		var appGuid string
		var appName string

		BeforeEach(func() {
			appName = os.Getenv("APP_TO_RESTART_AND_RESTAGE_WITH_V3")
			appGuid = string(cf.Cf("app", appName, "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents())
		})

		It("can restart the app", func() {
			Expect(cf.Cf("curl", "/v3/apps/"+strings.TrimSpace(appGuid)+"/stop", "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

			Expect(cf.Cf("curl", "/v3/apps/"+strings.TrimSpace(appGuid)+"/start", "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora"))
		})

		It("can restage the app", func() {
			packageJsonResponse := cf.Cf("curl", "/v3/apps/"+strings.TrimSpace(appGuid)+"/packages").Wait(DEFAULT_TIMEOUT).Out.Contents()
			packages := PackagesStruct{}

			err := json.Unmarshal(packageJsonResponse, &packages)
			Expect(err).NotTo(HaveOccurred())

			Expect(packages.Resources).To(HaveLen(1))
			packageGuid := packages.Resources[0].Guid

			droplet := DropletResource{}
			dropletJsonResponse := cf.Cf("curl", "/v3/packages/"+strings.TrimSpace(string(packageGuid))+"/droplets", "-X", "POST").Wait(DEFAULT_TIMEOUT).Out.Contents()

			err = json.Unmarshal(dropletJsonResponse, &droplet)
			Expect(err).NotTo(HaveOccurred())

			dropletPath := droplet.Links.Self["href"]

			Eventually(func() *Session {
				session := cf.Cf("curl", dropletPath).Wait(DEFAULT_TIMEOUT)
				Expect(session).NotTo(Say("FAILED"))
				return session
			}, CF_PUSH_TIMEOUT).Should(Say("STAGED"))
		})

		It("can add routes", func() {
			space := "test"
			domain := "bosh-lite.com"
			host := "banana"
			newRoute := fmt.Sprintf("http://%v.%v", host, domain)

			Expect(cf.Cf("create-route", space, domain, "-n", host).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			getRoutePath := fmt.Sprintf("/v2/routes?q=host:%s", host)
			routeBody := cf.Cf("curl", getRoutePath).Wait(DEFAULT_TIMEOUT).Out.Contents()

			var routeJSON RoutesResource
			json.Unmarshal([]byte(routeBody), &routeJSON)
			routeGuid := routeJSON.Resources[0].Metadata.Guid

			addRouteBody := fmt.Sprintf(`
		{
			"relationships": {
				"app":   {"guid": "%s"},
				"route": {"guid": "%s"}
			}
		}`, strings.TrimSpace(appGuid), routeGuid)

			Expect(cf.Cf("curl", "/v3/route_mappings/", "-X", "POST", "-d", addRouteBody).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			Eventually(func() string {
				return string(helpers.Curl(newRoute).Wait(DEFAULT_TIMEOUT).Out.Contents())
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora"))
		})

		It("can bind a service", func() {
			serviceInstance := os.Getenv("SERVICE")

			Expect(cf.Cf("v3-bind-service", appName, serviceInstance).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			Expect(cf.Cf("curl", "/v3/apps/"+strings.TrimSpace(appGuid)+"/stop", "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

			Expect(cf.Cf("curl", "/v3/apps/"+strings.TrimSpace(appGuid)+"/start", "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora"))

			var appEnvs string
			Eventually(func() string {
				appEnvs = helpers.CurlApp(appName, "/env")

				return appEnvs
			}, DEFAULT_TIMEOUT).Should(ContainSubstring(serviceInstance))

			Expect(appEnvs).To(ContainSubstring("credentials"))
		})

		It("can get a syslog drain url", func() {
			syslogDrainService := os.Getenv("SYSLOG_DRAIN_SERVICE")
			apiEndpoint := os.Getenv("API_ENDPOINT")
			expectedSyslogDrainUrl := "logit.io/drain/here"
			appGuid = strings.TrimSpace(appGuid)

			Expect(cf.Cf("v3-bind-service", appName, syslogDrainService).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			Expect(cf.Cf("curl", "/v3/apps/"+appGuid+"/stop", "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

			Expect(cf.Cf("curl", "/v3/apps/"+appGuid+"/start", "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora"))

			Eventually(func() string {
				return helpers.CurlApp(appName, "/env")
			}, DEFAULT_TIMEOUT).Should(ContainSubstring(expectedSyslogDrainUrl))

			syslogDrainUrlsResponse := helpers.Curl("bulk_api:bulk-password@" + apiEndpoint + "/v2/syslog_drain_urls").Wait(DEFAULT_TIMEOUT).Out.Contents()
			var syslogDrainUrls SyslogDrainUrls

			json.Unmarshal([]byte(syslogDrainUrlsResponse), &syslogDrainUrls)

			Expect(syslogDrainUrls.Results[appGuid][0]).To(Equal(expectedSyslogDrainUrl))
		})
	})

	Context("Updating the migrated app's package", func() {
		var appName string
		var appGuid string

		It("can push and run new buildpack bits", func() {
			appName = os.Getenv("BUILDPACK_APP_TO_REPUSH_WITH_V3")

			Expect(helpers.CurlAppRoot(appName)).To(ContainSubstring("Hi, I'm Dora"))

			appGuid = strings.TrimSpace(string(cf.Cf("app", appName, "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()))
			StopApp(appGuid)

			packageGuid := CreatePackage(appGuid)
			apiEndpoint := os.Getenv("API_ENDPOINT")
			uploadUrl := fmt.Sprintf("http://%s/v3/packages/%s/upload", apiEndpoint, packageGuid)

			UploadPackage(uploadUrl, "../assets/updated_dora.zip", GetAuthToken())
			WaitForPackageToBeReady(packageGuid)

			dropletGuid := StageBuildpackPackage(packageGuid, "ruby_buildpack")
			WaitForDropletToStage(dropletGuid)

			AssignDropletToApp(appGuid, dropletGuid)

			processes := GetProcesses(appGuid, appName)
			webProcess := GetProcessByType(processes, "web")
			Expect(webProcess.Guid).ToNot(BeEmpty())

			StartApp(appGuid)

			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("Goodbye, I'm Dora"))
		})

		It("can push and run a new docker image", func() {
			type envStruct struct {
				Port string `json:"PORT", json:"port"`
			}

			appName = os.Getenv("DOCKER_APP_TO_REPUSH_WITH_V3")
			appGuid = strings.TrimSpace(string(cf.Cf("app", appName, "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()))

			env := envStruct{}
			envStr := helpers.CurlApp(appName, "/env")
			err := json.Unmarshal([]byte(envStr), &env)
			Expect(err).NotTo(HaveOccurred())

			Expect(env.Port).To(Equal("8080"))

			StopApp(appGuid)

			packageGuid := CreateDockerPackage(appGuid, "cloudfoundry/diego-docker-app-custom:latest")

			dropletGuid := StageDockerPackage(packageGuid)
			WaitForDropletToStage(dropletGuid)

			AssignDropletToApp(appGuid, dropletGuid)

			processes := GetProcesses(appGuid, appName)
			webProcess := GetProcessByType(processes, "web")
			Expect(webProcess.Guid).ToNot(BeEmpty())

			StartApp(appGuid)

			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, CF_PUSH_TIMEOUT).Should(Equal("0"))

			envStr = helpers.CurlApp(appName, "/env")
			err = json.Unmarshal([]byte(envStr), &env)
			Expect(err).NotTo(HaveOccurred())
			Expect(env.Port).To(Equal("7070"))
		})

		It("can stage and run a previously unstaged app's package", func() {
			appName = os.Getenv("UNSTAGED_APP_TO_STAGE_AND_START_WITH_V3")
			appGuid = strings.TrimSpace(string(cf.Cf("app", appName, "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()))

			packageJsonResponse := cf.Cf("curl", "/v3/apps/"+appGuid+"/packages").Wait(DEFAULT_TIMEOUT).Out.Contents()
			packages := PackagesStruct{}

			err := json.Unmarshal(packageJsonResponse, &packages)
			Expect(err).NotTo(HaveOccurred())

			Expect(packages.Resources).To(HaveLen(1))
			packageGuid := packages.Resources[0].Guid

			dropletGuid := StageBuildpackPackage(packageGuid, "ruby_buildpack")
			WaitForDropletToStage(dropletGuid)

			AssignDropletToApp(appGuid, dropletGuid)

			processes := GetProcesses(appGuid, appName)
			webProcess := GetProcessByType(processes, "web")
			Expect(webProcess.Guid).ToNot(BeEmpty())

			StartApp(appGuid)

			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora"))
		})

		It("can run a previously stopped app", func() {
			appName = os.Getenv("STOPPED_APP_TO_START_WITH_V3")
			appGuid = strings.TrimSpace(string(cf.Cf("app", appName, "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()))

			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

			StartApp(appGuid)

			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora"))
		})
	})

	Context("Tasks", func() {
		var appName string
		var appGuid string

		It("can run a task against an existing droplet", func() {
			appName = os.Getenv("TASK_APP")
			appGuid = string(cf.Cf("app", appName, "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents())

			postBody := `{"command": "echo 0", "name": "mreow"}`
			taskJsonResponse := cf.Cf("curl", "/v3/apps/"+strings.TrimSpace(appGuid)+"/tasks", "-X", "POST", "-d", postBody).Wait(DEFAULT_TIMEOUT)

			var task TaskResource
			Expect(taskJsonResponse).To(Exit(0))
			err := json.Unmarshal(taskJsonResponse.Out.Contents(), &task)
			Expect(err).NotTo(HaveOccurred())
			Expect(task.Command).To(Equal("echo 0"))
			Expect(task.Name).To(Equal("mreow"))
			Expect(task.State).To(Equal("RUNNING"))
		})
	})

	Context("Apps pushed using V3 before the migration", func() {
		It("should truncate all pre-migration v3 app data", func() {
			appName := os.Getenv("V3_APP")
			apps := cf.Cf("curl", "/v3/apps/").Wait(DEFAULT_TIMEOUT).Out.Contents()

			Expect(apps).ToNot(ContainSubstring(appName))
		})
	})
})
