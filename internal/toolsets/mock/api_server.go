package mock

import (
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"
	v1 "github.com/stackrox/rox/generated/api/v1"
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
