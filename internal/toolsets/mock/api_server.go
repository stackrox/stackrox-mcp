package mock

import (
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"
	v1 "github.com/stackrox/rox/generated/api/v1"
	v2 "github.com/stackrox/rox/generated/api/v2"
	"github.com/stackrox/rox/generated/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufferSize = 1024 * 1024

// SetupAPIServer creates an in-memory gRPC Central server.
func SetupAPIServer(
	deploymentService v1.DeploymentServiceServer,
	imageService v1.ImageServiceServer,
	nodeService v1.NodeServiceServer,
	clusterService v1.ClustersServiceServer,
) (*grpc.Server, *bufconn.Listener) {
	buffer := bufferSize
	listener := bufconn.Listen(buffer)

	grpcServer := grpc.NewServer()
	v1.RegisterDeploymentServiceServer(grpcServer, deploymentService)
	v1.RegisterImageServiceServer(grpcServer, imageService)
	v1.RegisterNodeServiceServer(grpcServer, nodeService)
	v1.RegisterClustersServiceServer(grpcServer, clusterService)

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	return grpcServer, listener
}

// SetupNodeServer creates an in-memory gRPC server with node services.
func SetupNodeServer(nodeService v1.NodeServiceServer) (*grpc.Server, *bufconn.Listener) {
	return SetupAPIServer(
		v1.UnimplementedDeploymentServiceServer{},
		v1.UnimplementedImageServiceServer{},
		nodeService,
		v1.UnimplementedClustersServiceServer{},
	)
}

// SetupDeploymentServer creates an in-memory gRPC server with deployment services.
func SetupDeploymentServer(mockService v1.DeploymentServiceServer) (*grpc.Server, *bufconn.Listener) {
	return SetupAPIServer(
		mockService,
		v1.UnimplementedImageServiceServer{},
		v1.UnimplementedNodeServiceServer{},
		v1.UnimplementedClustersServiceServer{},
	)
}

// SetupClusterServer creates an in-memory gRPC server with cluster services.
func SetupClusterServer(mockService v1.ClustersServiceServer) (*grpc.Server, *bufconn.Listener) {
	return SetupAPIServer(
		v1.UnimplementedDeploymentServiceServer{},
		v1.UnimplementedImageServiceServer{},
		v1.UnimplementedNodeServiceServer{},
		mockService,
	)
}

// ClustersService implements v1.ClustersServiceServer for testing.
type ClustersService struct {
	v1.UnimplementedClustersServiceServer

	clusters []*storage.Cluster
	err      error

	lastCallQuery string
}

// NewClustersServiceMock return mock to cluster service.
func NewClustersServiceMock(clusters []*storage.Cluster, err error) *ClustersService {
	return &ClustersService{clusters: clusters, err: err}
}

// GetLastCallQuery returns query used for the last call.
func (cs *ClustersService) GetLastCallQuery() string {
	return cs.lastCallQuery
}

// GetClusters implements v1.ClustersServiceServer.GetClusters for testing.
func (cs *ClustersService) GetClusters(
	_ context.Context,
	req *v1.GetClustersRequest,
) (*v1.ClustersList, error) {
	cs.lastCallQuery = req.GetQuery()

	if cs.err != nil {
		return nil, cs.err
	}

	return &v1.ClustersList{
		Clusters: cs.clusters,
	}, nil
}

// NodeService implements v1.NodeServiceServer for testing.
type NodeService struct {
	v1.UnimplementedNodeServiceServer

	nodes []*storage.Node
	err   error

	lastCallQuery string
}

// NewNodeServiceMock return mock to node service.
func NewNodeServiceMock(nodes []*storage.Node, err error) *NodeService {
	return &NodeService{nodes: nodes, err: err}
}

// GetLastCallQuery returns query used for the last call.
func (ns *NodeService) GetLastCallQuery() string {
	return ns.lastCallQuery
}

// ExportNodes implements v1.NodeServiceServer.ExportNodes for testing.
func (ns *NodeService) ExportNodes(
	req *v1.ExportNodeRequest,
	stream grpc.ServerStreamingServer[v1.ExportNodeResponse],
) error {
	ns.lastCallQuery = req.GetQuery()

	if ns.err != nil {
		return ns.err
	}

	// Send all nodes through the stream.
	for _, node := range ns.nodes {
		resp := &v1.ExportNodeResponse{Node: node}
		if err := stream.Send(resp); err != nil {
			return errors.Wrap(err, "sending node over stream failed")
		}
	}

	return nil
}

// DeploymentService implements v1.DeploymentServiceServer for testing.
type DeploymentService struct {
	v1.UnimplementedDeploymentServiceServer

	deployments []*storage.ListDeployment
	err         error

	// Mock call information.
	lastCallQuery  string
	lastCallLimit  int32
	lastCallOffset int32
}

// NewDeploymentServiceMock returns mock for deployment service.
func NewDeploymentServiceMock(deployments []*storage.ListDeployment, err error) *DeploymentService {
	return &DeploymentService{
		deployments: deployments,
		err:         err,
	}
}

// GetLastCallQuery returns query used for the last call.
func (ds *DeploymentService) GetLastCallQuery() string {
	return ds.lastCallQuery
}

// GetLastCallLimit returns limit used for the last call.
func (ds *DeploymentService) GetLastCallLimit() int32 {
	return ds.lastCallLimit
}

// GetLastCallOffset returns offset used for the last call.
func (ds *DeploymentService) GetLastCallOffset() int32 {
	return ds.lastCallOffset
}

// ListDeployments implements v1.DeploymentServiceServer.ListDeployments for testing.
func (ds *DeploymentService) ListDeployments(
	_ context.Context,
	query *v1.RawQuery,
) (*v1.ListDeploymentsResponse, error) {
	ds.lastCallQuery = query.GetQuery()
	ds.lastCallLimit = query.GetPagination().GetLimit()
	ds.lastCallOffset = query.GetPagination().GetOffset()

	if ds.err != nil {
		return nil, ds.err
	}

	return &v1.ListDeploymentsResponse{
		Deployments: ds.deployments,
	}, nil
}

// ImageService implements v1.ImageServiceServer for testing.
type ImageService struct {
	v1.UnimplementedImageServiceServer

	images map[string][]*storage.ListImage // keyed by deploymentID
	err    error

	// We are requesting images in parallel requests.
	lock sync.Mutex

	// Mock call information.
	lastCallQuery string
	lastCallLimit int32
	callCount     int
}

// NewImageServiceMock returns mock for image service.
func NewImageServiceMock(images map[string][]*storage.ListImage, err error) *ImageService {
	return &ImageService{
		images: images,
		err:    err,
	}
}

// GetLastCallQuery returns query used for the last call.
func (is *ImageService) GetLastCallQuery() string {
	return is.lastCallQuery
}

// GetLastCallLimit returns limit used for the last call.
func (is *ImageService) GetLastCallLimit() int32 {
	return is.lastCallLimit
}

// GetCallCount returns count off all calls.
func (is *ImageService) GetCallCount() int {
	return is.callCount
}

// ListImages implements v1.ImageServiceServer.ListImages for testing.
func (is *ImageService) ListImages(
	_ context.Context,
	query *v1.RawQuery,
) (*v1.ListImagesResponse, error) {
	is.lock.Lock()
	defer is.lock.Unlock()

	is.callCount++
	is.lastCallQuery = query.GetQuery()
	is.lastCallLimit = query.GetPagination().GetLimit()

	if is.err != nil {
		return nil, is.err
	}

	// Extract deployment ID from query.
	// Query format: CVE:"CVE-2021-44228"+Deployment ID:"dep-1"
	deploymentID := extractDeploymentIDFromQuery(query.GetQuery())

	return &v1.ListImagesResponse{
		Images: is.images[deploymentID],
	}, nil
}

// extractDeploymentIDFromQuery extracts deployment ID from the query string.
func extractDeploymentIDFromQuery(query string) string {
	const prefix = "Deployment ID:\""

	start := strings.Index(query, prefix)
	if start == -1 {
		return ""
	}

	start += len(prefix)

	end := strings.Index(query[start:], "\"")
	if end == -1 {
		return ""
	}

	return query[start : start+end]
}

// SetupComplianceServer creates an in-memory gRPC server with V2 compliance services.
func SetupComplianceServer(
	profileService v2.ComplianceProfileServiceServer,
	resultsService v2.ComplianceResultsServiceServer,
	scanConfigService v2.ComplianceScanConfigurationServiceServer,
) (*grpc.Server, *bufconn.Listener) {
	listener := bufconn.Listen(bufferSize)

	grpcServer := grpc.NewServer()

	if profileService == nil {
		profileService = v2.UnimplementedComplianceProfileServiceServer{}
	}

	if resultsService == nil {
		resultsService = v2.UnimplementedComplianceResultsServiceServer{}
	}

	if scanConfigService == nil {
		scanConfigService = v2.UnimplementedComplianceScanConfigurationServiceServer{}
	}

	v2.RegisterComplianceProfileServiceServer(grpcServer, profileService)
	v2.RegisterComplianceResultsServiceServer(grpcServer, resultsService)
	v2.RegisterComplianceScanConfigurationServiceServer(grpcServer, scanConfigService)

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	return grpcServer, listener
}

// ComplianceProfileService implements v2.ComplianceProfileServiceServer for testing.
type ComplianceProfileService struct {
	v2.UnimplementedComplianceProfileServiceServer

	profiles []*v2.ComplianceProfile
	err      error
}

// NewComplianceProfileServiceMock returns mock for compliance profile service.
func NewComplianceProfileServiceMock(profiles []*v2.ComplianceProfile, err error) *ComplianceProfileService {
	return &ComplianceProfileService{profiles: profiles, err: err}
}

// ListComplianceProfiles implements v2.ComplianceProfileServiceServer.ListComplianceProfiles for testing.
func (s *ComplianceProfileService) ListComplianceProfiles(
	_ context.Context,
	_ *v2.ProfilesForClusterRequest,
) (*v2.ListComplianceProfilesResponse, error) {
	if s.err != nil {
		return nil, s.err
	}

	return &v2.ListComplianceProfilesResponse{
		Profiles:   s.profiles,
		TotalCount: int32(len(s.profiles)), //nolint:gosec // test-only mock, overflow impossible
	}, nil
}

// ComplianceResultsService implements v2.ComplianceResultsServiceServer for testing.
type ComplianceResultsService struct {
	v2.UnimplementedComplianceResultsServiceServer

	scanResults       *v2.ListComplianceResultsResponse
	scanConfigResults *v2.ListComplianceResultsResponse
	checkResult       *v2.ListComplianceCheckClusterResponse
	err               error
	checkResultErr    error
}

// NewComplianceResultsServiceMock returns mock for compliance results service.
func NewComplianceResultsServiceMock(
	scanResults *v2.ListComplianceResultsResponse,
	scanConfigResults *v2.ListComplianceResultsResponse,
	err error,
) *ComplianceResultsService {
	return &ComplianceResultsService{
		scanResults:       scanResults,
		scanConfigResults: scanConfigResults,
		err:               err,
	}
}

// SetCheckResult sets the response for GetComplianceProfileCheckResult.
func (s *ComplianceResultsService) SetCheckResult(result *v2.ListComplianceCheckClusterResponse) {
	s.checkResult = result
}

// SetCheckResultError sets the error for GetComplianceProfileCheckResult.
func (s *ComplianceResultsService) SetCheckResultError(err error) {
	s.checkResultErr = err
}

// GetComplianceScanResults implements v2.ComplianceResultsServiceServer for testing.
func (s *ComplianceResultsService) GetComplianceScanResults(
	_ context.Context,
	_ *v2.RawQuery,
) (*v2.ListComplianceResultsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.scanResults, nil
}

// GetComplianceScanConfigurationResults implements v2.ComplianceResultsServiceServer for testing.
func (s *ComplianceResultsService) GetComplianceScanConfigurationResults(
	_ context.Context,
	_ *v2.ComplianceScanResultsRequest,
) (*v2.ListComplianceResultsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.scanConfigResults, nil
}

// GetComplianceProfileCheckResult implements v2.ComplianceResultsServiceServer for testing.
func (s *ComplianceResultsService) GetComplianceProfileCheckResult(
	_ context.Context,
	_ *v2.ComplianceProfileCheckRequest,
) (*v2.ListComplianceCheckClusterResponse, error) {
	if s.checkResultErr != nil {
		return nil, s.checkResultErr
	}

	return s.checkResult, nil
}

// ComplianceScanConfigurationService implements v2.ComplianceScanConfigurationServiceServer for testing.
type ComplianceScanConfigurationService struct {
	v2.UnimplementedComplianceScanConfigurationServiceServer

	configurations []*v2.ComplianceScanConfigurationStatus
	err            error
}

// NewComplianceScanConfigurationServiceMock returns mock for compliance scan configuration service.
func NewComplianceScanConfigurationServiceMock(
	configurations []*v2.ComplianceScanConfigurationStatus,
	err error,
) *ComplianceScanConfigurationService {
	return &ComplianceScanConfigurationService{configurations: configurations, err: err}
}

// ListComplianceScanConfigurations implements v2.ComplianceScanConfigurationServiceServer for testing.
func (s *ComplianceScanConfigurationService) ListComplianceScanConfigurations(
	_ context.Context,
	_ *v2.RawQuery,
) (*v2.ListComplianceScanConfigurationsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}

	return &v2.ListComplianceScanConfigurationsResponse{
		Configurations: s.configurations,
		TotalCount:     int32(len(s.configurations)), //nolint:gosec // test-only mock, overflow impossible
	}, nil
}
