/*
Copyright (C) 2019-2023  Daniele Rondina <geaaru@gmail.com>
*/

package simplestreams

// Stream represents the base structure of index.json.
type Stream struct {
	Index   map[string]StreamIndex `json:"index"`
	Updated string                 `json:"updated,omitempty"`
	Format  string                 `json:"format"`
}

// StreamIndex represents the Index entry inside index.json.
type StreamIndex struct {
	DataType string   `json:"datatype"`
	Path     string   `json:"path"`
	Updated  string   `json:"updated,omitempty"`
	Products []string `json:"products"`
	Format   string   `json:"format,omitempty"`
}

// Products represents the base of download.json.
type Products struct {
	ContentID string             `json:"content_id"`
	DataType  string             `json:"datatype"`
	Format    string             `json:"format"`
	License   string             `json:"license,omitempty"`
	Products  map[string]Product `json:"products"`
	Updated   string             `json:"updated,omitempty"`
}

// Product represents a single product inside download.json.
type Product struct {
	Aliases           string                    `json:"aliases"`
	Architecture      string                    `json:"arch"`
	OperatingSystem   string                    `json:"os"`
	LXDRequirements   map[string]string         `json:"lxd_requirements,omitempty"`
	IncusRequirements map[string]string         `json:"incus_requirements,omitempty"`
	Release           string                    `json:"release"`
	ReleaseCodename   string                    `json:"release_codename,omitempty"`
	ReleaseTitle      string                    `json:"release_title"`
	Supported         bool                      `json:"supported,omitempty"`
	SupportedEOL      string                    `json:"support_eol,omitempty"`
	Version           string                    `json:"version,omitempty"`
	Versions          map[string]ProductVersion `json:"versions"`

	// Non-standard fields (only used on some image servers).
	Variant string `json:"variant,omitempty"`
}

// ProductVersion represents a particular version of a product.
type ProductVersion struct {
	Items      map[string]ProductVersionItem `json:"items"`
	Label      string                        `json:"label,omitempty"`
	PublicName string                        `json:"pubname,omitempty"`
}

// ProductVersionItem represents a file/item of a particular ProductVersion.
type ProductVersionItem struct {
	CombinedHashSha256DiskImg     string `json:"combined_disk1-img_sha256,omitempty"`
	CombinedHashSha256DiskKvmImg  string `json:"combined_disk-kvm-img_sha256,omitempty"`
	CombinedHashSha256DiskUefiImg string `json:"combined_uefi1-img_sha256,omitempty"`
	CombinedHashSha256RootXz      string `json:"combined_rootxz_sha256,omitempty"`
	CombinedHashSha256            string `json:"combined_sha256,omitempty"`
	CombinedHashSha256SquashFs    string `json:"combined_squashfs_sha256,omitempty"`
	FileType                      string `json:"ftype"`
	HashMd5                       string `json:"md5,omitempty"`
	Path                          string `json:"path"`
	HashSha256                    string `json:"sha256,omitempty"`
	Size                          int64  `json:"size"`
	DeltaBase                     string `json:"delta_base,omitempty"`
}
