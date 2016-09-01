package migration_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	. "code.cloudfoundry.org/capi-integration-tests/v3_helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

type appsStruct struct {
	Resources []appResourceStruct `json:"resources"`
}

type appResourceStruct struct {
	Name  string      `json:"name"`
	Guid  string      `json:"guid"`
	Links linksStruct `json:"links"`
}

type packagesStruct struct {
	Resources []packageResourceStruct `json:"resources"`
}

type packageResourceStruct struct {
	Guid string `json:"guid"`
}

type dropletsStruct struct {
	Resources []dropletResourceStruct `json:"resources"`
}

type dropletResourceStruct struct {
	Guid  string      `json:"guid"`
	State string      `json:"state"`
	Links linksStruct `json:"links"`
}

type taskResourceStruct struct {
	Guid    string `json:"guid"`
	Command string `json:"command"`
	State   string `json:"state"`
	Name    string `json:"name"`
}

type routeResourceStruct struct {
	Metadata MetadataStruct `json:"metadata"`
}

type routesStruct struct {
	Resources []routeResourceStruct `json:"resources"`
}

type linksStruct struct {
	Self    map[string]string `json:"self"`
	Droplet map[string]string `json:"droplet"`
}

type MetadataStruct struct {
	Guid string `json:"guid"`
}

func GetV3AppGuid(appName string) string {
	jsonResponse := cf.Cf("curl", "v3/apps").Wait(DEFAULT_TIMEOUT).Out.Contents()

	apps := appsStruct{}
	err := json.Unmarshal(jsonResponse, &apps)
	Expect(err).NotTo(HaveOccurred())

	var appGuid string
	for _, app := range apps.Resources {
		if app.Name == appName {
			appGuid = app.Guid
			break
		}
	}

	return appGuid
}

var _ = Describe("Using v3 endpoints", func() {
	Context("V3 App", func() {
		var appGuid string
		var appName string

		BeforeEach(func() {
			appName = os.Getenv("V3_APP")
			appGuid = GetV3AppGuid(appName)
		})

		It("can restart the app", func() {
			// Use this app guid when actually running the migration before this test
			// v2 app guid should be migrated to be the v3 app guid also

			// appGuid := cf.Cf("app", os.Getenv("V3_APP"), "--guid").Wait(DEFAULT_TIMEOUT).Out.Contents()
			Expect(cf.Cf("curl", "/v3/apps/"+strings.TrimSpace(string(appGuid))+"/stop", "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

			Expect(cf.Cf("curl", "/v3/apps/"+strings.TrimSpace(string(appGuid))+"/start", "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora"))
		})

		It("can restage the app", func() {
			packageJsonResponse := cf.Cf("curl", "/v3/apps/"+strings.TrimSpace(string(appGuid))+"/packages").Wait(DEFAULT_TIMEOUT).Out.Contents()
			packages := packagesStruct{}

			err := json.Unmarshal(packageJsonResponse, &packages)
			Expect(err).NotTo(HaveOccurred())

			Expect(packages.Resources).To(HaveLen(1))
			packageGuid := packages.Resources[0].Guid

			droplet := dropletResourceStruct{}
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

			var routeJSON routesStruct
			json.Unmarshal([]byte(routeBody), &routeJSON)
			routeGuid := routeJSON.Resources[0].Metadata.Guid

			addRouteBody := fmt.Sprintf(`
		{
			"relationships": {
				"app":   {"guid": "%s"},
				"route": {"guid": "%s"}
			}
		}`, appGuid, routeGuid)

			Expect(cf.Cf("curl", "/v3/route_mappings/", "-X", "POST", "-d", addRouteBody).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			Eventually(func() string {
				return string(helpers.Curl(newRoute).Wait(DEFAULT_TIMEOUT).Out.Contents())
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora"))
		})

		It("can bind a service", func() {
			serviceInstance := os.Getenv("SERVICE")

			Expect(cf.Cf("v3-bind-service", appName, serviceInstance).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			// Uncomment this when running actual test
			// Expect(cf.Cf("restart", appName).Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

			Expect(cf.Cf("curl", "/v3/apps/"+strings.TrimSpace(string(appGuid))+"/stop", "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

			Expect(cf.Cf("curl", "/v3/apps/"+strings.TrimSpace(string(appGuid))+"/start", "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
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

			Expect(cf.Cf("v3-bind-service", appName, syslogDrainService).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			// Uncomment this when running actual test
			// Expect(cf.Cf("restart", appName).Wait(CF_PUSH_TIMEOUT)).To(Exit(0))

			Expect(cf.Cf("curl", "/v3/apps/"+strings.TrimSpace(string(appGuid))+"/stop", "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("404"))

			Expect(cf.Cf("curl", "/v3/apps/"+strings.TrimSpace(string(appGuid))+"/start", "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora"))

			Eventually(func() string {
				return helpers.CurlApp(appName, "/env")
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("logit.io/drain/here"))
		})
	})

	Context("Updating the migrated app's package", func() {
		var appName string
		var appGuid string

		It("can push and run new buildpack bits", func() {
			appName = os.Getenv("BUILDPACK_V3_APP_TO_REPUSH")
			appGuid = GetV3AppGuid(appName)

			Expect(helpers.CurlAppRoot(appName)).To(ContainSubstring("Hi, I'm Dora"))

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

			space := os.Getenv("SPACE")
			CreateAndMapRoute(appGuid, space, os.Getenv("API_DOMAIN"), webProcess.Name)

			StartApp(appGuid)

			Eventually(func() string {
				return helpers.CurlAppRoot(appName)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("Goodbye, I'm Dora"))
		})

		It("can push and run a new docker image", func() {
		})
	})

	Context("Tasks", func() {
		var appName string
		var appGuid string

		It("can run a task against an existing droplet", func() {
			appName = os.Getenv("TASK_APP")
			appGuid = GetV3AppGuid(appName)

			postBody := `{"command": "echo 0", "name": "mreow"}`
			taskJsonResponse := cf.Cf("curl", "/v3/apps/"+strings.TrimSpace(string(appGuid))+"/tasks", "-X", "POST", "-d", postBody).Wait(DEFAULT_TIMEOUT)

			var task taskResourceStruct
			Expect(taskJsonResponse).To(Exit(0))
			err := json.Unmarshal(taskJsonResponse.Out.Contents(), &task)
			Expect(err).NotTo(HaveOccurred())
			Expect(task.Command).To(Equal("echo 0"))
			Expect(task.Name).To(Equal("mreow"))
			Expect(task.State).To(Equal("RUNNING"))
		})
	})
})
