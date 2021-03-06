package testing

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
)

type ControllerTestFixture struct {
	*FakeCluster

	Clusters           map[string]*FakeCluster
	ClusterClientStore *FakeClusterClientStore

	Recorder *record.FakeRecorder
}

func NewControllerTestFixture() *ControllerTestFixture {
	const recorderBufSize = 42

	store := NewFakeClusterClientStore()
	fakeCluster := NewNamedFakeCluster("mgmt")

	return &ControllerTestFixture{
		FakeCluster: fakeCluster,

		Clusters:           make(map[string]*FakeCluster),
		ClusterClientStore: store,

		Recorder: record.NewFakeRecorder(recorderBufSize),
	}
}

func NewManagementControllerTestFixture(
	mgmtClusterObjects []runtime.Object,
	appClusterShipperObjects map[string][]runtime.Object,
) *ControllerTestFixture {
	f := NewControllerTestFixture()

	for clusterName, objects := range appClusterShipperObjects {
		cluster := f.AddNamedCluster(clusterName)

		for _, object := range objects {
			cluster.ShipperClient.Tracker().Add(object)
		}
	}

	for _, object := range mgmtClusterObjects {
		f.ShipperClient.Tracker().Add(object)
	}

	return f
}

func (f *ControllerTestFixture) AddNamedCluster(name string) *FakeCluster {
	cluster := NewNamedFakeCluster(name)

	f.Clusters[cluster.Name] = cluster
	f.ClusterClientStore.AddCluster(cluster)

	return cluster
}

func (f *ControllerTestFixture) AddCluster() *FakeCluster {
	name := fmt.Sprintf("cluster-%d", len(f.Clusters))
	return f.AddNamedCluster(name)
}

func (f *ControllerTestFixture) Run(stopCh chan struct{}) {
	f.KubeInformerFactory.Start(stopCh)
	f.KubeInformerFactory.WaitForCacheSync(stopCh)

	f.ShipperInformerFactory.Start(stopCh)
	f.ShipperInformerFactory.WaitForCacheSync(stopCh)

	f.ClusterClientStore.Run(stopCh)
}
