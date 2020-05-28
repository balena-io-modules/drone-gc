// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package gc

import (
	"context"
	"time"

	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/filters"
	"github.com/docker/go-units"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
)

func (c *collector) collectDanglingImages(ctx context.Context) error {
	logger := log.Ctx(ctx)
	logger.Debug().
		Msg("prune dangling images")

	report, err := c.client.ImagesPrune(ctx, imagePruneArgs)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("cannot prune dangling images")
		return err
	}
	logger.Debug().
		Msg("images pruned")

	for _, image := range report.ImagesDeleted {
		logger.Info().
			Str("untagged", image.Untagged).
			Str("deleted", image.Deleted).
			Msg("deleted image")
	}
	return nil
}

func (c *collector) collectImages(ctx context.Context) error {
	var result error
	var logger = log.Ctx(ctx)

	df, err := c.client.DiskUsage(ctx)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("cannot get disk usage")
		return err
	}
	size := df.LayersSize

	if size < c.threshold {
		logger.Debug().
			Str("size", units.HumanSize(
				float64(df.LayersSize),
			)).
			Str("threshold", units.HumanSize(
				float64(c.threshold),
			)).
			Msg("image cache below threshold")
		return nil
	}

	logger.Debug().
		Msg("pruning named images")

	now := time.Now()
	for _, image := range df.Images {
		if isImageUsed(image, df.Containers) {
			continue
		}
		if time.Unix(image.Created, 0).Add(c.minImageAge).After(now) {
			continue
		}

		info, _, err := c.client.ImageInspectWithRaw(ctx, image.ID)
		if err != nil {
			logger.Error().
				Err(err).
				Str("name", image.ID).
				Msg("cannot find image")
			result = multierror.Append(result, err)
			continue
		}

		if matchPatterns(info.RepoTags, c.reserved) {
			continue
		}

		logger.Debug().
			Str("id", image.ID).
			Str("size", units.HumanSize(
				float64(image.Size),
			)).
			Int64("created", image.Created).
			Strs("repoTags", info.RepoTags).
			Strs("repoDigests", info.RepoDigests).
			Msg("remove image")

		err = c.removeImage(ctx, info)
		if err != nil {
			logger.Error().
				Err(err).
				Str("id", image.ID).
				Strs("image", info.RepoTags).
				Msg("cannot remove image")
			result = multierror.Append(result, err)
			continue
		}

		logger.Info().
			Str("id", image.ID).
			Strs("image", info.RepoTags).
			Msg("image removed")

		size = size - image.Size
		if shouldConsiderSharedSpace(c) {
			size -= image.SharedSize
		}

		if size < c.threshold {
			break
		}
	}

	logger.Debug().
		Str("size", units.HumanSize(
			float64(size),
		)).
		Msg("done pruning named images")

	return result
}

func shouldConsiderSharedSpace(c *collector) bool {
	return c.imageRemoveOptions.PruneChildren
}

func (c *collector) removeImage(ctx context.Context, imageInspect types.ImageInspect) error {
	var err error
	var removalAttributes = []string{imageInspect.ID}
	if len(imageInspect.RepoTags) > 0 {
		removalAttributes = imageInspect.RepoTags
	} else if len(imageInspect.RepoDigests) > 0 {
		removalAttributes = imageInspect.RepoDigests
	}

	for _, attribute := range removalAttributes {
		_, err = c.client.ImageRemove(ctx, attribute, c.imageRemoveOptions)
		if err != nil {
			break
		}
	}

	return err
}

var imagePruneArgs = filters.NewArgs(
	filters.KeyValuePair{
		Key:   "until",
		Value: "1h",
	},
)

func isImageUsed(image *types.ImageSummary, containers []*types.Container) bool {
	for _, container := range containers {
		if container.ImageID == image.ID ||
			container.ImageID == image.ParentID {
			return true
		}
	}
	return false
}
