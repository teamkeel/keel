package config

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type ProjectData struct {
	Data struct {
		Projects struct {
			Projects []struct {
				ID           string `json:"id"`
				Name         string `json:"name"`
				Description  string `json:"description"`
				Environments struct {
					Environments []Environments `json:"environments"`
				} `json:"environments"`
			} `json:"projects"`
		} `json:"projects"`
	} `json:"data"`
}

var createProjectQueryJson = map[string]interface{}{
	"query": "query listProjectsAndEnvironments($pagination: PaginationInput!) {  projects(pagination: $pagination) {    projects {      id      name      description      environments(pagination: $pagination) {        environments {          id          name          }      }    }  }}",
	"variables": map[string]interface{}{
		"pagination": map[string]interface{}{
			"first": 20,
		},
	},
}

func (c *Config) SetProjectConfig() error {
	config, err := c.GetConfig()
	if err != nil {
		config = &RootConfig{}
	}

	graphQLEndpoint := "https://api.staging.keel.xyz/query"

	jsonValue, err := json.Marshal(createProjectQueryJson)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", graphQLEndpoint, bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", "Bearer "+config.User.Token)
	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var responseJson ProjectData
	err = json.Unmarshal(body, &responseJson)
	if err != nil {
		return err
	}

	projects := make(map[string]ProjectConfig)

	for _, project := range responseJson.Data.Projects.Projects {
		environments := make(map[string]Environments)
		for _, environment := range project.Environments.Environments {
			environments[environment.ID] = environment
		}

		projects[project.ID] = ProjectConfig{
			Project:      project.ID,
			Environments: environments,
		}
	}

	config.Projects = projects

	return c.SetConfig(config)
}
