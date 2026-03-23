//go:build integration

package integration

// Log4ShellFixture contains expected data from wiremock/fixtures/deployments/log4j_cve.json fixture.
var Log4ShellFixture = struct {
	CVEName         string
	DeploymentCount int
	DeploymentNames []string
	ExpectedJSON    string
}{
	CVEName:         "CVE-2021-44228",
	DeploymentCount: 3,
	DeploymentNames: []string{"elasticsearch", "kafka-broker", "spring-boot-app"},
	ExpectedJSON: `{
  "deployments": [
    {
      "name": "elasticsearch",
      "namespace": "logging",
      "clusterId": "cluster-prod-01",
      "clusterName": "production-cluster"
    },
    {
      "name": "kafka-broker",
      "namespace": "messaging",
      "clusterId": "cluster-prod-02",
      "clusterName": "production-cluster-eu"
    },
    {
      "name": "spring-boot-app",
      "namespace": "applications",
      "clusterId": "cluster-staging-01",
      "clusterName": "staging-cluster"
    }
  ],
  "nextCursor": ""
}`,
}

// AllClustersFixture contains expected data from wiremock/fixtures/clusters/all_clusters.json fixture.
var AllClustersFixture = struct {
	TotalCount   int
	ClusterNames []string
	ExpectedJSON string
}{
	TotalCount: 5,
	ClusterNames: []string{
		"production-cluster",
		"staging-cluster",
		"staging-central-cluster",
		"development-cluster",
		"production-cluster-eu",
	},
	ExpectedJSON: `{
  "clusters": [
    {
      "id": "cluster-prod-01",
      "name": "production-cluster",
      "type": "KUBERNETES_CLUSTER"
    },
    {
      "id": "cluster-staging-01",
      "name": "staging-cluster",
      "type": "KUBERNETES_CLUSTER"
    },
    {
      "id": "65673bd7-da6a-4cdc-a5fc-95765d1b9724",
      "name": "staging-central-cluster",
      "type": "OPENSHIFT4_CLUSTER"
    },
    {
      "id": "cluster-dev-01",
      "name": "development-cluster",
      "type": "KUBERNETES_CLUSTER"
    },
    {
      "id": "cluster-prod-02",
      "name": "production-cluster-eu",
      "type": "KUBERNETES_CLUSTER"
    }
  ],
  "totalCount": 5,
  "offset": 0,
  "limit": 0
}`,
}

// EmptyDeploymentsJSON represents the expected JSON response when no deployments are found.
const EmptyDeploymentsJSON = `{"deployments": [], "nextCursor": ""}`

// EmptyClustersForCVEJSON represents the expected JSON response from get_clusters_with_orchestrator_cve when no clusters are found.
const EmptyClustersForCVEJSON = `{"clusters": []}`
