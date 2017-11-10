package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/rancher/go-rancher/catalog"

	log "github.com/Sirupsen/logrus"
	"github.com/gitlawr/cihelper/model"
	"github.com/pkg/errors"
	"github.com/rancher/go-rancher/v2"
)

var regTag = regexp.MustCompile(`^[\w]+[\w.-]*`)

func UpgradeServices(apiClient *client.RancherClient, config *model.ServiceUpgrade, pushedImage string) {
	var key, value string
	var secondaryPresent, primaryPresent bool
	serviceSelector := make(map[string]string)

	for key, value = range config.ServiceSelector {
		serviceSelector[key] = value
	}
	batchSize := config.BatchSize
	intervalMillis := config.IntervalMillis
	startFirst := config.StartFirst
	allService := []client.Service{}
	services, err := apiClient.Service.List(&client.ListOpts{})
	allService = append(allService, services.Data...)
	for nextservices, err := services.Next(); nextservices != nil; services = nextservices {
		allService = append(allService, nextservices.Data...)
		if err != nil {
			log.Fatalf("Error %v in listing services", err)
		}
	}
	if err != nil {
		log.Fatalf("Error %v in listing services", err)
	}
	toUpgradeCount := 0
	for _, service := range allService {
		secondaryPresent = false
		primaryPresent = false
		primaryLabels := service.LaunchConfig.Labels
		secConfigs := []client.SecondaryLaunchConfig{}
		for _, secLaunchConfig := range service.SecondaryLaunchConfigs {
			labels := secLaunchConfig.Labels
			found := true
			for k, v := range serviceSelector {
				if val, ok := labels[k]; !ok || !strings.EqualFold(val.(string), v) {
					found = false
					break
				}
			}
			if found {
				secLaunchConfig.ImageUuid = "docker:" + pushedImage
				secLaunchConfig.Labels["io.rancher.container.pull_image"] = "always"
				secConfigs = append(secConfigs, secLaunchConfig)
				secondaryPresent = true
			}
		}

		newLaunchConfig := service.LaunchConfig
		found := true
		for k, v := range serviceSelector {
			if val, ok := primaryLabels[k]; !ok || !strings.EqualFold(val.(string), v) {
				found = false
				break
			}
		}
		if found {
			primaryPresent = true
			newLaunchConfig.ImageUuid = "docker:" + pushedImage
			newLaunchConfig.Labels["io.rancher.container.pull_image"] = "always"
		}

		if !primaryPresent && !secondaryPresent {
			continue
		}
		toUpgradeCount++
		logrus.Infof("start upgrade service '%s'", service.Name)
		func(service client.Service, apiClient *client.RancherClient, newLaunchConfig *client.LaunchConfig,
			secConfigs []client.SecondaryLaunchConfig, primaryPresent bool, secondaryPresent bool) {
			upgStrategy := &client.InServiceUpgradeStrategy{
				BatchSize:      batchSize,
				IntervalMillis: intervalMillis * 1000,
				StartFirst:     startFirst,
			}
			if primaryPresent && secondaryPresent {
				upgStrategy.LaunchConfig = newLaunchConfig
				upgStrategy.SecondaryLaunchConfigs = secConfigs
			} else if primaryPresent && !secondaryPresent {
				upgStrategy.LaunchConfig = newLaunchConfig
			} else if !primaryPresent && secondaryPresent {
				upgStrategy.SecondaryLaunchConfigs = secConfigs
			}

			upgradedService, err := apiClient.Service.ActionUpgrade(&service, &client.ServiceUpgrade{
				InServiceStrategy: upgStrategy,
			})
			if err != nil {
				log.Fatalf("upgrading service '%s' error: %v", service.Name, err)
				return
			}

			if err := wait(apiClient, upgradedService); err != nil {
				log.Fatal(err)
				return
			}

			if upgradedService.State != "upgraded" {
				return
			}
			_, err = apiClient.Service.ActionFinishupgrade(upgradedService)
			if err != nil {
				log.Fatalf("Error %v in finishUpgrade of service %s", err, upgradedService.Name)
				return
			}
			log.Infof("upgrade service '%s' success", upgradedService.Name)
		}(service, apiClient, newLaunchConfig, secConfigs, primaryPresent, secondaryPresent)
	}
	if toUpgradeCount == 0 {
		jsonString, _ := json.Marshal(serviceSelector)
		log.Fatalf("No service matches the selectors:%v, marked as failure", string(jsonString))
	}
	log.Infof("Upgraded %d services", toUpgradeCount)
}

//UpgradeStack currently works for catalog stack only
func UpgradeStack(apiClient *client.RancherClient, config *model.StackUpgrade) error {
	stackName := config.StackName
	var toUpgradeStack *client.Stack
	stacks, err := apiClient.Stack.List(&client.ListOpts{})
	if err != nil {
		log.Errorf("Error %v in listing stacks", err)
		return err
	}
	for _, stack := range stacks.Data {
		if stack.Name == stackName {
			toUpgradeStack = &stack
			break
		}
	}
	if toUpgradeStack == nil {
		return fmt.Errorf("Stack %v is not found.", stackName)
	}

	if config.ToLatestCatalog {
		log.Infof("trying to upgrade stack '%s' to latest catalog version", stackName)
		if toUpgradeStack.ExternalId == "" {
			log.Error("stack is not deployed from catalog")
			return errors.New("stack is not deployed from catalog")
		}
		log.Infof("current catalog templat: %s", toUpgradeStack.ExternalId)
		log.Infoln("refreshing catalog templates...")
		if err = refreshCatalog(apiClient, config); err != nil {
			return err
		}
		log.Infoln("refresh catalog templates done")
		if config.ExternalId == "" {
			latestExtId, err := getTemplateLatestVersion(config, toUpgradeStack.ExternalId)
			if err != nil {
				return err
			}
			config.ExternalId = latestExtId
			template, err := getTemplateVersion(config, latestExtId)
			if err != nil {
				return err
			}
			for k, v := range template.Files {
				if strings.HasPrefix(k, "docker-compose") && config.DockerCompose == "" {
					config.DockerCompose = v
				} else if strings.HasPrefix(k, "rancher-compose") && config.RancherCompose == "" {
					config.RancherCompose = v
				}
			}
			if config.Environment == nil && toUpgradeStack.Environment != nil {
				log.Infoln("using previous environment.")
				config.Environment = toUpgradeStack.Environment
			}
		}

		if config.ExternalId == toUpgradeStack.ExternalId {
			log.Infoln("Got latest template '%s', latest template version already...\nDo no upgrade and end.", toUpgradeStack.ExternalId)
			return nil
		}
	}
	log.Infof("upgrading stack '%s' to '%s'", stackName, config.ExternalId)
	stackUpgrade := &client.StackUpgrade{
		DockerCompose:  config.DockerCompose,
		RancherCompose: config.RancherCompose,
		ExternalId:     config.ExternalId,
		Environment:    config.Environment,
	}
	stack, err := apiClient.Stack.ActionUpgrade(toUpgradeStack, stackUpgrade)
	serviceIds := stack.ServiceIds

	for _, id := range serviceIds {
		service, err := apiClient.Service.ById(id)
		if err != nil {
			log.Fatalf("Error %v in upgrading stacks", err)
			return err
		}
		if err := wait(apiClient, service); err != nil {
			log.Fatal(err)
			return err
		}
	}

	if err := apiClient.Reload(&stack.Resource, stack); err != nil {
		return err
	}

	if err := waitStack(apiClient, stack); err != nil {
		log.Error(err.Error())
		return err
	}

	if stack.State != "upgraded" {
		log.Errorf("expected upgraded stack state but got:'%s'", stack.State)
		return errors.New("upgrade stack failed")
	}

	_, err = apiClient.Stack.ActionFinishupgrade(stack)
	if err != nil {
		log.Errorf("Error %v in finishUpgrade of stack %s", err, stack.Name)
		return err
	}

	log.Infof("upgrade stack '%s' success", stack.Name)
	return nil
}

func getProjId(config *model.StackUpgrade) (string, error) {

	client := &http.Client{}

	requestURL := config.CattleUrl

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Infoln("Cannot connect to the rancher server. Please check the rancher server URL")
		return "", err
	}
	req.SetBasicAuth(config.AccessKey, config.SecretKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Infoln("Cannot connect to the rancher server. Please check the rancher server URL")
		return "", err
	}
	defer resp.Body.Close()
	userid := resp.Header.Get("X-Api-User-Id")
	if userid == "" {
		log.Infoln("Cannot get userid")
		err := errors.New("Forbidden")
		return "Forbidden", err

	}
	return userid, nil
}

func getTemplateVersion(config *model.StackUpgrade, externalId string) (*catalog.TemplateVersion, error) {
	u, err := url.Parse(config.CattleUrl)
	if err != nil {
		return &catalog.TemplateVersion{}, err
	}
	projId, err := getProjId(config)
	if err != nil {
		return &catalog.TemplateVersion{}, err
	}
	trimExternalId := externalId[strings.LastIndex(externalId, "/")+1:]
	templateUrl := fmt.Sprintf("%s://%s/v1-catalog/templates/%s?projectId=%s", u.Scheme, u.Host, trimExternalId, projId)

	client := &http.Client{}

	req, err := http.NewRequest("GET", templateUrl, nil)
	if err != nil {
		log.Infoln("Cannot connect to the rancher server. Please check the rancher server URL")
		return &catalog.TemplateVersion{}, err
	}
	req.SetBasicAuth(config.AccessKey, config.SecretKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Infoln("Cannot connect to the rancher server. Please check the rancher server URL")
		return &catalog.TemplateVersion{}, err
	}
	defer resp.Body.Close()

	byteContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &catalog.TemplateVersion{}, err
	}
	tempObj := &catalog.TemplateVersion{}
	if err := json.Unmarshal(byteContent, tempObj); err != nil {
		return &catalog.TemplateVersion{}, errors.Wrap(err, fmt.Sprintf("getTemplateLatestVersion error, Failed to parse: %s", byteContent))
	}
	return tempObj, nil
}

func getTemplateLatestVersion(config *model.StackUpgrade, externalId string) (string, error) {

	tempObj, err := getTemplateVersion(config, externalId)
	if err != nil {
		return "", err
	}
	//no upgrade version
	if tempObj.UpgradeVersionLinks == nil || len(tempObj.UpgradeVersionLinks) == 0 {
		return externalId, nil
	}
	retV := ""
	retRev := 0
	for _, v := range tempObj.UpgradeVersionLinks {
		extId := v[strings.LastIndex(v, "/")+1:]
		_, _, _, Rev, _ := TemplateURLPath(extId)
		RevI, err := strconv.Atoi(Rev)
		if err != nil {
			return "", err
		}
		if RevI > retRev {
			retRev = RevI
			retV = extId
		}
	}
	return "catalog://" + retV, nil
}

func refreshCatalog(apiClient *client.RancherClient, config *model.StackUpgrade) error {
	u, err := url.Parse(config.CattleUrl)
	if err != nil {
		return err
	}
	projId, err := getProjId(config)
	if err != nil {
		return err
	}
	refreshUrl := fmt.Sprintf("%s://%s/v1-catalog/templates?action=refresh&projectId=%s", u.Scheme, u.Host, projId)

	if err := apiClient.Post(refreshUrl, nil, nil); err != nil {
		return err
	}
	return nil
}

func wait(apiClient *client.RancherClient, service *client.Service) error {
	logrus.Debugf("waiting for service '%s' finishing...", service.Name)
	for {
		if err := apiClient.Reload(&service.Resource, service); err != nil {
			return err
		}
		if service.Transitioning != "yes" {
			break
		}
		time.Sleep(5 * time.Second)
	}

	switch service.Transitioning {
	case "yes":
		return fmt.Errorf("Service %s upgrade timeout", service.Name)
	case "no":
		return nil
	default:
		return fmt.Errorf("Service %s upgrade failed: %s", service.Name, service.TransitioningMessage)
	}
}

func waitStack(apiClient *client.RancherClient, stack *client.Stack) error {
	logrus.Debugf("waiting for stack '%s' finishing...", stack.Name)
	for {
		if err := apiClient.Reload(&stack.Resource, stack); err != nil {
			return err
		}
		if stack.Transitioning != "yes" {
			break
		}
		time.Sleep(5 * time.Second)
	}

	switch stack.Transitioning {
	case "yes":
		return fmt.Errorf("Timeout upgrade stack %s", stack.Name)
	case "no":
		return nil
	default:
		return fmt.Errorf("Stack %s upgrade failed: %s", stack.Name, stack.TransitioningMessage)
	}
}

func TemplateURLPath(path string) (string, string, string, string, bool) {
	pathSplit := strings.Split(path, ":")
	switch len(pathSplit) {
	case 2:
		catalog := pathSplit[0]
		template := pathSplit[1]
		templateSplit := strings.Split(template, "*")
		templateBase := ""
		switch len(templateSplit) {
		case 1:
			template = templateSplit[0]
		case 2:
			templateBase = templateSplit[0]
			template = templateSplit[1]
		default:
			return "", "", "", "", false
		}
		return catalog, template, templateBase, "", true
	case 3:
		catalog := pathSplit[0]
		template := pathSplit[1]
		revisionOrVersion := pathSplit[2]
		templateSplit := strings.Split(template, "*")
		templateBase := ""
		switch len(templateSplit) {
		case 1:
			template = templateSplit[0]
		case 2:
			templateBase = templateSplit[0]
			template = templateSplit[1]
		default:
			return "", "", "", "", false
		}
		return catalog, template, templateBase, revisionOrVersion, true
	default:
		return "", "", "", "", false
	}
}
