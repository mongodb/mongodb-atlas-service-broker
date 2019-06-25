package atlas

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCluster(t *testing.T) {
	expected := Cluster{
		Name:  "Cluster",
		State: ClusterStateIdle,
		Type:  ClusterTypeReplicaSet,
	}

	atlas := setupTest(t, "/clusters", http.MethodPost, 200, expected)

	cluster, err := atlas.CreateCluster(expected)

	assert.NoError(t, err)
	assert.Equal(t, &expected, cluster)
}

func TestCreateClusterExistingName(t *testing.T) {
	cluster := Cluster{
		Name:  "Cluster",
		State: ClusterStateIdle,
		Type:  ClusterTypeReplicaSet,
	}

	atlas := setupTest(t, "/clusters", http.MethodPost, 400, errorResponse("DUPLICATE_CLUSTER_NAME"))

	_, err := atlas.CreateCluster(cluster)

	assert.EqualError(t, err, ErrClusterAlreadyExists.Error())
}

func TestGetCluster(t *testing.T) {
	expected := &Cluster{
		Name:  "Cluster",
		State: ClusterStateIdle,
		Type:  ClusterTypeReplicaSet,
	}

	atlas := setupTest(t, "/clusters/"+expected.Name, http.MethodGet, 200, expected)

	cluster, err := atlas.GetCluster(expected.Name)

	assert.NoError(t, err)
	assert.Equal(t, expected, cluster)
}

func TestGetNonexistentCluster(t *testing.T) {
	clusterName := "Cluster"
	atlas := setupTest(t, "/clusters/"+clusterName, http.MethodGet, 404, errorResponse("CLUSTER_NOT_FOUND"))

	_, err := atlas.GetCluster(clusterName)

	assert.EqualError(t, err, ErrClusterNotFound.Error())
}

func TestTerminateCluster(t *testing.T) {
	clusterName := "Cluster"
	atlas := setupTest(t, "/clusters/"+clusterName, http.MethodDelete, 200, nil)

	err := atlas.TerminateCluster(clusterName)
	assert.NoError(t, err)
}

func TestTerminateNonexistentCluster(t *testing.T) {
	clusterName := "Cluster"
	atlas := setupTest(t, "/clusters/"+clusterName, http.MethodDelete, 404, errorResponse("CLUSTER_NOT_FOUND"))

	err := atlas.TerminateCluster(clusterName)

	assert.Equal(t, ErrClusterNotFound, err)
}
