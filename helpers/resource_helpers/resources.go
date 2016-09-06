package resource_helpers

type AppsResource struct {
	Resources []appResourceStruct `json:"resources"`
}

type appResourceStruct struct {
	Name  string      `json:"name"`
	Guid  string      `json:"guid"`
	Links linksStruct `json:"links"`
}

type PackagesStruct struct {
	Resources []packageResourceStruct `json:"resources"`
}

type packageResourceStruct struct {
	Guid string `json:"guid"`
}

type DropletsResource struct {
	Resources []DropletResource `json:"resources"`
}

type DropletResource struct {
	Guid  string      `json:"guid"`
	State string      `json:"state"`
	Links linksStruct `json:"links"`
}

type TaskResource struct {
	Guid    string `json:"guid"`
	Command string `json:"command"`
	State   string `json:"state"`
	Name    string `json:"name"`
}

type routeResourceStruct struct {
	Metadata MetadataStruct `json:"metadata"`
}

type RoutesResource struct {
	Resources []routeResourceStruct `json:"resources"`
}

type linksStruct struct {
	Self    map[string]string `json:"self"`
	Droplet map[string]string `json:"droplet"`
}

type MetadataStruct struct {
	Guid string `json:"guid"`
}

type SyslogDrainUrls struct {
	Results map[string][]string `json:"results"`
}
