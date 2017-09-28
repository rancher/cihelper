package model

//ServiceUpgrade config
type ServiceUpgrade struct {
	ServiceSelector map[string]string `json:"serviceSelector,omitempty" mapstructure:"serviceSelector"`
	Tag             string            `json:"tag,omitempty" mapstructure:"tag"`
	BatchSize       int64             `json:"batchSize,omitempty" mapstructure:"batchSize"`
	IntervalMillis  int64             `json:"intervalMillis,omitempty" mapstructure:"intervalMillis"`
	StartFirst      bool              `json:"startFirst,omitempty" mapstructure:"startFirst"`
	Type            string            `json:"type,omitempty" mapstructure:"type"`
}

//StackUpgrade config
type StackUpgrade struct {
	CattleUrl       string
	AccessKey       string
	SecretKey       string
	ToLatestCatalog bool
	StackName       string
	DockerCompose   string
	RancherCompose  string
	ExternalId      string
	Environment     map[string]interface{}
}

//CatalogUpgrade config
type CatalogUpgrade struct {
	GitUrl             string
	GitBranch          string
	TemplateFolderName string
	TemplateIsSystem   bool
	CacheRoot          string
	GitUser            string
	GitPassword        string
	DockerCompose      string
	RancherCompose     string
	Readme             string
}
