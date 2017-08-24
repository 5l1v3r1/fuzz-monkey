//+build !test

package main

// Creates configuration objects for an endpoint and an attack. Useful for testing specific attacks.
func CreateTestEndpointAndAttackConfiguration(expectedStatus string) (EndpointConfig, AttackConfig) {
  endpoint := EndpointConfig{Name:"Test Endpoint", Protocol:"http",Host:"localhost",Port:"8080",Path:"/my-endpoint"}
  attack := AttackConfig{Type:"HTTP_SPAM",Concurrents:1,MessagesPerConcurrent:1,ExpectedStatus:expectedStatus,Method:"GET"}

  return endpoint, attack
}

// Creates a full config object with an endpoint and an attack. Useful for testing whole app.
func CreateFullTestConfiguration() Config {
  config := Config{}

  endpoint, attack := CreateTestEndpointAndAttackConfiguration("200")

  config.Endpoints = append(config.Endpoints, endpoint)
  config.Endpoints[0].Attacks = append(config.Endpoints[0].Attacks, attack)

  return config
}
