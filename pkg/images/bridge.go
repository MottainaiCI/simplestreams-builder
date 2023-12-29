/*
Copyright (C) 2019-2023  Daniele Rondina <geaaru@gmail.com>
*/
package images

import (
	streams "github.com/MottainaiCI/simplestreams-builder/pkg/simplestreams"
)

// In order to have a Simplestreams file working with LXD and Incus together
// we need translate ftype "lxd.tar.xz" as "incus.tar.xz" or viceversa.
// This will work until both files will be compliants.
func BridgeIncusLXDVersionsItems(omap map[string]streams.ProductVersion) map[string]streams.ProductVersion {
	ans := make(map[string]streams.ProductVersion, 0)

	for k, v := range omap {
		items := v.Items

		pviLxd, hasLxd := v.Items["lxd.tar.xz"]
		pviIncus, hasIncus := v.Items["incus.tar.xz"]

		if hasLxd && !hasIncus {
			pviIncus = pviLxd
			pviIncus.FileType = "incus.tar.xz"
			items["incus.tar.xz"] = pviIncus

		} else if hasIncus && !hasLxd {
			pviLxd = pviIncus
			pviLxd.FileType = "lxd.tar.xz"
			items["lxd.tar.xz"] = pviLxd
		}

		v.Items = items
		ans[k] = v

	}

	return ans
}
