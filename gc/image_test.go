// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package gc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/drone/drone-gc/mocks"

	"docker.io/go-docker/api/types"
	"github.com/golang/mock/gomock"
)

// This test verifies that images that have repoTags
// (and therefore listed in the UI as <repo>:<tag>) are removed by TAG
func TestCollectImages_ByTag(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockdf := types.DiskUsage{
		LayersSize: 850,
		Images: []*types.ImageSummary{
			{
				ID:         "a180b24e38ed",
				Created:    359596800,
				SharedSize: 50,
				Size:       300,
				RepoTags:   []string{"alpine:latest"},
			},
			{
				ID:         "4e38e38c8ce0",
				Created:    359596800,
				SharedSize: 50,
				Size:       300,
				RepoTags:   []string{"busybox:latest"},
			},
			// this image should not be removed since removal
			// of the above two images will put us below the
			// target threshold.
			{
				ID:         "481995377a04",
				Created:    359596800,
				SharedSize: 50,
				Size:       250,
				RepoTags:   []string{"hello-world:latest"},
			},
		},
	}
	mockImages := []types.ImageInspect{
		{ID: "a180b24e38ed", RepoTags: []string{"alpine:latest"}},
		{ID: "4e38e38c8ce0", RepoTags: []string{"busybox:latest"}},
		{ID: "481995377a04", RepoTags: []string{"hello-world:latest"}},
	}

	client := mocks.NewMockAPIClient(controller)
	client.EXPECT().DiskUsage(gomock.Any()).Return(mockdf, nil)
	client.EXPECT().ImageInspectWithRaw(gomock.Any(), mockImages[0].ID).Return(mockImages[0], nil, nil)
	client.EXPECT().ImageInspectWithRaw(gomock.Any(), mockImages[1].ID).Return(mockImages[1], nil, nil)
	// we DO NOT inspect image 481995377a04

	client.EXPECT().ImageRemove(gomock.Any(), mockImages[0].RepoTags[0], types.ImageRemoveOptions{}).Return(nil, nil)
	client.EXPECT().ImageRemove(gomock.Any(), mockImages[1].RepoTags[0], types.ImageRemoveOptions{}).Return(nil, nil)
	// we DO NOT remove image 481995377a04

	c := New(client, WithThreshold(500)).(*collector)
	err := c.collectImages(context.Background())
	if err != nil {
		t.Error(err)
	}
}

// This test verifies that images that have no repoTags but have repoDigests
// (and therefore listed in the UI as <repo>:<none>) are removed by DIGEST
func TestCollectImages_ByDigest(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockdf := types.DiskUsage{
		LayersSize: 850,
		Images: []*types.ImageSummary{
			{
				ID:          "a180b24e38ed",
				Created:     359596800,
				SharedSize:  50,
				Size:        300,
				RepoDigests: []string{"sha256:9a839e63dad54c3a6d1834e29692c8492d93f90c59c978c1ed79109ea4fb9a54"},
			},
			{
				ID:          "4e38e38c8ce0",
				Created:     359596800,
				SharedSize:  50,
				Size:        300,
				RepoDigests: []string{"sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042"},
			},
			// this image should not be removed since removal
			// of the above two images will put us below the
			// target threshold.
			{
				ID:          "481995377a04",
				Created:     359596800,
				SharedSize:  50,
				Size:        250,
				RepoDigests: []string{"sha256:8e3114318a995a1ee497790535e7b88365222a21771ae7e53687ad76563e8e76"},
			},
		},
	}
	mockImages := []types.ImageInspect{
		{ID: "a180b24e38ed", RepoDigests: []string{"sha256:9a839e63dad54c3a6d1834e29692c8492d93f90c59c978c1ed79109ea4fb9a54"}},
		{ID: "4e38e38c8ce0", RepoDigests: []string{"sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042"}},
		{ID: "481995377a04", RepoDigests: []string{"sha256:8e3114318a995a1ee497790535e7b88365222a21771ae7e53687ad76563e8e76"}},
	}

	client := mocks.NewMockAPIClient(controller)
	client.EXPECT().DiskUsage(gomock.Any()).Return(mockdf, nil)
	client.EXPECT().ImageInspectWithRaw(gomock.Any(), mockImages[0].ID).Return(mockImages[0], nil, nil)
	client.EXPECT().ImageInspectWithRaw(gomock.Any(), mockImages[1].ID).Return(mockImages[1], nil, nil)
	// we DO NOT inspect image 481995377a04

	client.EXPECT().ImageRemove(gomock.Any(), mockImages[0].RepoDigests[0], types.ImageRemoveOptions{}).Return(nil, nil)
	client.EXPECT().ImageRemove(gomock.Any(), mockImages[1].RepoDigests[0], types.ImageRemoveOptions{}).Return(nil, nil)
	// we DO NOT remove image 481995377a04

	c := New(client, WithThreshold(500)).(*collector)
	err := c.collectImages(context.Background())
	if err != nil {
		t.Error(err)
	}
}

// This test verifies that images that have no repoTags but have repoDigests
// (and therefore listed in the UI as <none>:<none>) are removed by ID
func TestCollectImages_ById(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockdf := types.DiskUsage{
		LayersSize: 850,
		Images: []*types.ImageSummary{
			{
				ID:         "a180b24e38ed",
				Created:    359596800,
				SharedSize: 50,
				Size:       300,
			},
			{
				ID:         "4e38e38c8ce0",
				Created:    359596800,
				SharedSize: 50,
				Size:       300,
			},
			// this image should not be removed since removal
			// of the above two images will put us below the
			// target threshold.
			{
				ID:         "481995377a04",
				Created:    359596800,
				SharedSize: 50,
				Size:       250,
			},
		},
	}
	mockImages := []types.ImageInspect{
		{ID: "a180b24e38ed"},
		{ID: "4e38e38c8ce0"},
		{ID: "481995377a04"},
	}

	client := mocks.NewMockAPIClient(controller)
	client.EXPECT().DiskUsage(gomock.Any()).Return(mockdf, nil)
	client.EXPECT().ImageInspectWithRaw(gomock.Any(), mockImages[0].ID).Return(mockImages[0], nil, nil)
	client.EXPECT().ImageInspectWithRaw(gomock.Any(), mockImages[1].ID).Return(mockImages[1], nil, nil)
	// we DO NOT inspect image 481995377a04

	client.EXPECT().ImageRemove(gomock.Any(), mockImages[0].ID, types.ImageRemoveOptions{}).Return(nil, nil)
	client.EXPECT().ImageRemove(gomock.Any(), mockImages[1].ID, types.ImageRemoveOptions{}).Return(nil, nil)
	// we DO NOT remove image 481995377a04

	c := New(client, WithThreshold(500)).(*collector)
	err := c.collectImages(context.Background())
	if err != nil {
		t.Error(err)
	}
}

// This test verifies that when an error is encountered we move to
// the next image in the list. Errors are aggregated and returned
// at the end of the loop.
func TestCollectImages_MutliError(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockdf := types.DiskUsage{
		LayersSize: 1,
		Images: []*types.ImageSummary{
			{
				ID:      "a180b24e38ed",
				Created: 359596800,
			},
			{
				ID:      "4e38e38c8ce0",
				Created: 359596800,
			},
		},
	}
	mockImages := []types.ImageInspect{
		{ID: "a180b24e38ed"},
		{ID: "4e38e38c8ce0"},
	}
	mockError := errors.New("cannot remove container")

	client := mocks.NewMockAPIClient(controller)
	client.EXPECT().DiskUsage(gomock.Any()).Return(mockdf, nil)
	client.EXPECT().ImageInspectWithRaw(gomock.Any(), mockImages[0].ID).Return(mockImages[0], nil, nil)
	client.EXPECT().ImageInspectWithRaw(gomock.Any(), mockImages[1].ID).Return(mockImages[1], nil, nil)

	client.EXPECT().ImageRemove(gomock.Any(), mockImages[0].ID, types.ImageRemoveOptions{}).Return(nil, mockError)
	client.EXPECT().ImageRemove(gomock.Any(), mockImages[1].ID, types.ImageRemoveOptions{}).Return(nil, nil)

	c := New(client).(*collector)
	err := c.collectImages(context.Background())
	if err == nil {
		t.Errorf("Expect multi-error returned")
	}
}

// this test verifies that we do not purge the image cache
// if the cache is already below the target threshold.
func TestCollectImages_BelowThreshold(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockdf := types.DiskUsage{
		LayersSize: 1, // 1 byte
	}
	client := mocks.NewMockAPIClient(controller)
	client.EXPECT().DiskUsage(gomock.Any()).Return(mockdf, nil)

	c := New(client, WithThreshold(2)).(*collector)
	err := c.collectImages(context.Background())
	if err != nil {
		t.Error(err)
	}
}

// this test verifies that a image SharedSize is only taken into consideration
// for removal size calculation if the ImageRemovalOption PruneChildren is set to true
func TestCollectImages_OnlyConsiderSharedSizeWhenPruneChildrenIsTrue(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockdf := types.DiskUsage{
		LayersSize: 600,
		Images: []*types.ImageSummary{
			{
				ID:         "a180b24e38ed",
				Created:    359596800,
				SharedSize: 50,
				Size:       300,
				RepoTags:   []string{"alpine:latest"},
			},
			// This image will only be removed if PruneChildren is True, since only in that case SharedSize
			// will be considered in the removal size calculation
			{
				ID:         "4e38e38c8ce0",
				Created:    359596800,
				SharedSize: 50,
				Size:       300,
				RepoTags:   []string{"busybox:latest"},
			},
		},
	}
	mockImages := []types.ImageInspect{
		{ID: "a180b24e38ed", RepoTags: []string{"alpine:latest"}},
		{ID: "4e38e38c8ce0", RepoTags: []string{"busybox:latest"}},
	}

	pruneChildrenImageOptions := types.ImageRemoveOptions{PruneChildren: true}

	client := mocks.NewMockAPIClient(controller)
	client.EXPECT().DiskUsage(gomock.Any()).Return(mockdf, nil).AnyTimes()

	client.EXPECT().ImageInspectWithRaw(gomock.Any(), mockImages[0].ID).Return(mockImages[0], nil, nil).Times(2)
	client.EXPECT().ImageRemove(gomock.Any(), mockImages[0].RepoTags[0], types.ImageRemoveOptions{}).Return(nil, nil)
	client.EXPECT().ImageRemove(gomock.Any(), mockImages[0].RepoTags[0], pruneChildrenImageOptions).Return(nil, nil)

	client.EXPECT().ImageInspectWithRaw(gomock.Any(), mockImages[1].ID).Return(mockImages[1], nil, nil)
	client.EXPECT().ImageRemove(gomock.Any(), mockImages[1].RepoTags[0], types.ImageRemoveOptions{}).Return(nil, nil)

	c := New(client, WithThreshold(300)).(*collector)
	err := c.collectImages(context.Background())
	if err != nil {
		t.Error(err)
	}

	c = New(client, WithThreshold(300), WithImageRemoveOptions(pruneChildrenImageOptions)).(*collector)
	err = c.collectImages(context.Background())
	if err != nil {
		t.Error(err)
	}
}

// this test verifies that we do not purge images that are not old enough
func TestCollectImages_SkipNotOldEnough(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockdf := types.DiskUsage{
		LayersSize: 1,
		Images: []*types.ImageSummary{
			// this image was created 20 mins ago
			{
				ID:      "a180b24e38ed",
				Created: time.Now().Add(-20 * time.Minute).Unix(),
			},
			// this image was created 40 mins ago
			{
				ID:      "481995377a04",
				Created: time.Now().Add(-40 * time.Minute).Unix(),
			},
		},
	}

	mockImages := []types.ImageInspect{
		{ID: "a180b24e38ed"},
		{ID: "481995377a04"},
	}

	client := mocks.NewMockAPIClient(controller)
	client.EXPECT().DiskUsage(gomock.Any()).Return(mockdf, nil)

	// We DO NOT expect image 0 to be removed since is not old enough
	client.EXPECT().ImageInspectWithRaw(gomock.Any(), mockImages[1].ID).Return(mockImages[1], nil, nil)
	client.EXPECT().ImageRemove(gomock.Any(), mockImages[1].ID, types.ImageRemoveOptions{}).Return(nil, nil)

	// Minimum image age set to 30 mins
	c := New(client, WithMinImageAge(time.Hour/2)).(*collector)
	err := c.collectImages(context.Background())
	if err != nil {
		t.Error(err)
	}
}

// this test verifies that we do not purge images that are
// in-use by the system or are newly created.
func TestCollectImages_Skip(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockdf := types.DiskUsage{
		LayersSize: 1,
		Containers: []*types.Container{
			{ImageID: "a180b24e38ed"},
		},
		Images: []*types.ImageSummary{
			// this image is in-use
			{
				ID:      "a180b24e38ed",
				Created: 359596800,
			},
			// this image is newly created
			{
				ID:      "481995377a04",
				Created: time.Now().Unix(),
			},
		},
	}
	client := mocks.NewMockAPIClient(controller)
	client.EXPECT().DiskUsage(gomock.Any()).Return(mockdf, nil)

	c := New(client, WithMinImageAge(time.Hour)).(*collector)
	err := c.collectImages(context.Background())
	if err != nil {
		t.Error(err)
	}
}

// this test verifies that we do not purge images that
// are whitelisted by the user.
func TestCollectImages_SkipWhitelist(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockdf := types.DiskUsage{
		LayersSize: 1,
		Images: []*types.ImageSummary{
			{
				ID:      "a180b24e38ed",
				Created: 359596800,
			},
		},
	}

	mockImageInspect := types.ImageInspect{
		ID: "a180b24e38ed",
		RepoTags: []string{
			"drone/drone:1.0.0",
			"drone/drone:1.0",
			"drone/drone:1",
			"drone/drone:latest",
		},
	}

	client := mocks.NewMockAPIClient(controller)
	client.EXPECT().DiskUsage(gomock.Any()).Return(mockdf, nil)
	client.EXPECT().ImageInspectWithRaw(gomock.Any(), mockImageInspect.ID).Return(mockImageInspect, nil, nil)

	c := New(client,
		WithImageWhitelist(
			[]string{"drone/drone:*"},
		),
	).(*collector)
	err := c.collectImages(context.Background())
	if err != nil {
		t.Error(err)
	}
}
