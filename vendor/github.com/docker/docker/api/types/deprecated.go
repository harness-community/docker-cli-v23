package types // import "github.com/harness-community/docker-v23/api/types"

import "github.com/harness-community/docker-v23/api/types/volume"

// Volume volume
//
// Deprecated: use github.com/harness-community/docker-v23/api/types/volume.Volume
type Volume = volume.Volume

// VolumeUsageData Usage details about the volume. This information is used by the
// `GET /system/df` endpoint, and omitted in other endpoints.
//
// Deprecated: use github.com/harness-community/docker-v23/api/types/volume.UsageData
type VolumeUsageData = volume.UsageData
