package service

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"

	"github.com/docker/docker/api/types"
	dclient "github.com/docker/docker/client"
	"github.com/rancher/go-rancher/v2"
	"golang.org/x/net/context"
)

const DEFAULT_REGISTRY = "index.docker.io"

//AuthAndPush find registry credential and push the image
func AuthAndPush(apiClient *client.RancherClient, image string) error {
	username, password, err := getRegistryAuth(apiClient, image)
	if err != nil {
		return err
	}
	return pushImage(image, username, password)
}

func getRegistryAuth(apiClient *client.RancherClient, image string) (string, string, error) {
	opt := &client.ListOpts{}
	regCollection, err := apiClient.Registry.List(opt)
	if err != nil {
		return "", "", err
	}

	hostName, _ := splitHostName(image)
	var regToPush *client.Registry
	for _, reg := range regCollection.Data {
		if reg.ServerAddress == hostName {
			regToPush = &reg
		}
	}
	username, password := "", ""
	if regToPush == nil {
		logrus.Warningf("Cannot find registry credential for '%v', You probably need to add it in registries configuration.", image)
	} else {

		var regCredToPush *client.RegistryCredential
		regCredCollection, err := apiClient.RegistryCredential.List(opt)
		if err != nil {
			return "", "", err
		}
		for _, regCred := range regCredCollection.Data {
			if regCred.RegistryId == regToPush.Id {
				regCredToPush = &regCred
			}
		}
		if regCredToPush == nil {
			logrus.Warningf("Cannot find registry credential for '%v', You probably need to add it in registries configuration.", image)
		} else {
			username = regCredToPush.PublicValue
			password = regCredToPush.SecretValue
		}
	}
	return username, password, nil

}

func pushImage(image string, username string, password string) error {
	ctx := context.Background()
	cli, err := dclient.NewEnvClient()
	if err != nil {
		return err
	}

	authConfig := types.AuthConfig{
		Username: username,
		Password: password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic(err)
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	out, err := cli.ImagePush(ctx, image, types.ImagePushOptions{RegistryAuth: authStr})
	defer out.Close()
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, out)
	return nil
}

// encodeAuthToBase64 serializes the auth configuration as JSON base64 payload
func encodeAuthToBase64(authConfig types.AuthConfig) (string, error) {
	buf, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}
func splitHostName(image string) (string, string) {
	i := strings.Index(image, "/")
	if i == -1 || (!strings.ContainsAny(image[:i], ".:") && image[:i] != "localhost") {
		return DEFAULT_REGISTRY, image
	}
	return image[:i], image[i+1:]
}
