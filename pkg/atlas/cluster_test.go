package atlas

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCluster(t *testing.T) {
	expected := Cluster{
		Name:        "Cluster",
		StateName:   ClusterStateIdle,
		ClusterType: ClusterTypeReplicaSet,
	}

	atlas, server := setupTest(t, "/clusters", http.MethodPost, 200, expected)
	defer server.Close()

	cluster, err := atlas.CreateCluster(expected)

	assert.NoError(t, err)
	assert.Equal(t, &expected, cluster)
}

func TestCreateClusterExistingName(t *testing.T) {
	cluster := Cluster{
		Name:        "Cluster",
		StateName:   ClusterStateIdle,
		ClusterType: ClusterTypeReplicaSet,
	}

	atlas, server := setupTest(t, "/clusters", http.MethodPost, 400, errorResponse("DUPLICATE_CLUSTER_NAME"))
	defer server.Close()

	_, err := atlas.CreateCluster(cluster)

	assert.EqualError(t, err, ErrClusterAlreadyExists.Error())
}

func TestUpdateCluster(t *testing.T) {
	expected := Cluster{
		Name:        "Cluster",
		StateName:   ClusterStateIdle,
		ClusterType: ClusterTypeReplicaSet,
	}

	atlas, server := setupTest(t, "/clusters/"+expected.Name, http.MethodPatch, 200, expected)
	defer server.Close()

	cluster, err := atlas.UpdateCluster(expected)

	assert.NoError(t, err)
	assert.Equal(t, &expected, cluster)
}

func TestUpdateNonexistentCluster(t *testing.T) {
	expected := Cluster{
		Name:        "Cluster",
		StateName:   ClusterStateIdle,
		ClusterType: ClusterTypeReplicaSet,
	}

	atlas, server := setupTest(t, "/clusters/"+expected.Name, http.MethodPatch, 400, errorResponse("CLUSTER_NOT_FOUND"))
	defer server.Close()

	_, err := atlas.UpdateCluster(expected)

	assert.EqualError(t, err, ErrClusterNotFound.Error())
}

func TestGetCluster(t *testing.T) {
	expected := &Cluster{
		Name:        "Cluster",
		StateName:   ClusterStateIdle,
		ClusterType: ClusterTypeReplicaSet,
	}

	atlas, server := setupTest(t, "/clusters/"+expected.Name, http.MethodGet, 200, expected)
	defer server.Close()

	cluster, err := atlas.GetCluster(expected.Name)

	assert.NoError(t, err)
	assert.Equal(t, expected, cluster)
}

func TestGetNonexistentCluster(t *testing.T) {
	clusterName := "Cluster"
	atlas, server := setupTest(t, "/clusters/"+clusterName, http.MethodGet, 404, errorResponse("CLUSTER_NOT_FOUND"))
	defer server.Close()

	_, err := atlas.GetCluster(clusterName)

	assert.EqualError(t, err, ErrClusterNotFound.Error())
}

func TestTerminateCluster(t *testing.T) {
	clusterName := "Cluster"
	atlas, server := setupTest(t, "/clusters/"+clusterName, http.MethodDelete, 200, nil)
	defer server.Close()

	err := atlas.DeleteCluster(clusterName)
	assert.NoError(t, err)
}

func TestTerminateNonexistentCluster(t *testing.T) {
	clusterName := "Cluster"
	atlas, server := setupTest(t, "/clusters/"+clusterName, http.MethodDelete, 404, errorResponse("CLUSTER_NOT_FOUND"))
	defer server.Close()

	err := atlas.DeleteCluster(clusterName)

	assert.Equal(t, ErrClusterNotFound, err)
}
