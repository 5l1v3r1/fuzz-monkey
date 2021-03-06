package main

import (
  "encoding/json"
  "io/ioutil"
  "fmt"
)

// Object representation of config file provided by the user.
type Config struct {
  Endpoints []EndpointConfig `json:"endpoints"`
}

// Configuration for a specific endpoint.
type EndpointConfig struct {
  Name string `json:"name"`
  Host string `json:"host"`
  Port string `json:"port"`
  Path string `json:"path"`
  Protocol string `json:"protocol"`
  Attacks []AttackConfig `json:"attacks"`
}

// Configuration for a specific attack on an endpoint.
type AttackConfig struct {
  Type string `json:"type"`
  ExpectedStatus string `json:"expectedStatus"`
  Concurrents int `json:"concurrents"`
  MessagesPerConcurrent int `json:"messagesPerConcurrent"`
  Method string `json:"method"`
  Parameters string `json:"parameters"`
}

func GetConfig(configPath string) (*Config) {
  fileContents := loadConfigFile(configPath)
  return mapFileToObject(fileContents)
}

func loadConfigFile(configPath string) ([]byte) {

  if configPath == "" {
    configPath = "fuzz-monkey.json"
  }

  file, err := ioutil.ReadFile(configPath)

  CheckError(err)

  return file
}

func IsValidConfig(config *Config) (bool, error) {

  if len(config.Endpoints) == 0 {
    return false, fmt.Errorf("⚠️ Endpoints can not be empty. The monkey needs victims. ⚠️")
  }

  for i,endpoint := range config.Endpoints {
    if endpoint.Name == "" {
      return false, fmt.Errorf("⚠️ Endpoint name can not be empty for endpoint #%d. The monkey is like Arya Stark. It needs a name. ⚠️", i + 1)
    }

    if endpoint.Host == "" {
      return false, fmt.Errorf("⚠️ Host can not be null for endpoint with name %s. The monkey needs an address to go after. ⚠️", endpoint.Name)
    }

    if len(endpoint.Attacks) == 0 {
      return false, fmt.Errorf("⚠️ Endpoint must have attacks associated with it. The monkey kills all it sees. ⚠️")
    }

    for j,attack := range endpoint.Attacks {
      if attack.Type == "" {
        return false, fmt.Errorf("⚠️ Attack config #%d for endpoint %s needs a type. Future versions will interpret this as an all access pass for the monkey. ⚠️", endpoint.Name, j + 1)
      }
    }
  }

  return true, nil
}

func mapFileToObject(contents []byte) (*Config) {
  config := &Config{}
  err := json.Unmarshal(contents, config)
  CheckError(err)

  valid, err := IsValidConfig(config)

  if !valid {
    panic(err)
  }

  return config
}
