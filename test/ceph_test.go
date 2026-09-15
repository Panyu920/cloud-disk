package test

import (
	"testing"

	"github.com/Panyu920/cloud-disk/store/ceph"
)

func TestCeph(t *testing.T) {
	err := ceph.CreateBucket("cloud-disk")
	if err != nil {
		t.Errorf("ceph createBucket error: %v", err)
	}

	err = ceph.UploadObject("cloud-disk", "test", "../uploads/d3dcompiler_47.dll")
	if err != nil {
		t.Errorf("ceph uploadObject error: %v", err)
		return
	}
	bucket, err := ceph.GetBucket("cloud-disk")
	if err != nil {
		t.Errorf("ceph getBucket error: %v", err)
	}
	t.Logf("bucket name: %s", *bucket.Name)
	objects, err := ceph.ListObjects("cloud-disk")
	if err != nil {
		t.Errorf("ceph listObjects error: %v", err)
	}
	for i, _ := range objects {
		t.Logf("%s: %d", *objects[i].Key, *objects[i].Size)
	}
}
