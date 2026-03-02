//go:build integration

package integration

// Log4ShellFixture contains expected data from log4j_cve.json fixture.
var Log4ShellFixture = struct {
	CVEName         string
	DeploymentCount int
	DeploymentNames []string
}{
	CVEName:         "CVE-2021-44228",
	DeploymentCount: 3,
	DeploymentNames: []string{"elasticsearch", "kafka-broker", "spring-boot-app"},
}

// AllClustersFixture contains expected data from all_clusters.json fixture.
var AllClustersFixture = struct {
	TotalCount   int
	ClusterNames []string
}{
	TotalCount: 5,
	ClusterNames: []string{
		"production-cluster",
		"staging-cluster",
		"staging-central-cluster",
		"development-cluster",
		"production-cluster-eu",
	},
}
