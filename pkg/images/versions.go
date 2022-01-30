/*

Copyright (C) 2019  Daniele Rondina <geaaru@sabayonlinux.org>
Credits goes also to Gogs authors, some code portions and re-implemented design
are also coming from the Gogs project, which is using the go-macaron framework
and was really source of ispiration. Kudos to them!

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/
package images

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	d_shared "github.com/lxc/distrobuilder/shared"
	lxd_streams "github.com/lxc/lxd/shared/simplestreams"
	v "github.com/spf13/viper"

	config "github.com/MottainaiCI/simplestreams-builder/pkg/config"
)

const BYTE_BUFFER_LEN = 256

type VersionsSSBuilderManifest struct {
	Name       string                                `json:"name"`
	SupportEOL string                                `json:"expiry,omitempty"`
	Versions   map[string]lxd_streams.ProductVersion `json:"versions"`
}

type BuildVersionsManifestOptions struct {
	ProductDir          string
	PrefixPath          string
	ForceExpireDuration string
	ImageFile           string
}

type CombinedSha256Builder struct {
	SquashFsIsPresent      bool
	TarXzIsPresent         bool
	CombinedRootxzSha256   hash.Hash
	CombinedSquashfsSha256 hash.Hash
}

func newCombinedSha256Builder() CombinedSha256Builder {
	return CombinedSha256Builder{
		SquashFsIsPresent:      false,
		TarXzIsPresent:         false,
		CombinedSquashfsSha256: sha256.New(),
		CombinedRootxzSha256:   sha256.New(),
	}
}

func BuildVersionsManifest(product *config.SimpleStreamsProduct,
	opts BuildVersionsManifestOptions) (*VersionsSSBuilderManifest, error) {
	var err error
	var files []os.FileInfo
	var productBasePath, itemDir, eolDuration string
	var item, lxdTarXzItem *lxd_streams.ProductVersionItem
	var ans *VersionsSSBuilderManifest = &VersionsSSBuilderManifest{
		Name:     product.Name,
		Versions: make(map[string]lxd_streams.ProductVersion),
	}
	var combined CombinedSha256Builder

	if opts.ImageFile != "" && opts.ForceExpireDuration == "" {
		var imageDef *d_shared.Definition
		imageDef, err = ReadImageFile(opts.ImageFile, opts.PrefixPath)
		if err != nil {
			fmt.Printf("Error on retrieve data from image file %s\n", opts.ImageFile)
			return nil, err
		}

		eolDuration = imageDef.Image.Expiry
	} else if opts.ForceExpireDuration != "" {
		eolDuration = opts.ForceExpireDuration
	}

	if eolDuration != "" {
		now := time.Now()
		eol := d_shared.GetExpiryDate(now, eolDuration)
		if !now.Equal(eol) {
			ans.SupportEOL = fmt.Sprintf("%d", eol.Unix())
			fmt.Printf("For product %s use SupportEOL = %s (%s)\n",
				product.Name, ans.SupportEOL, eolDuration)
		}
	}

	// Iterate for every sub-directories that match with regex
	files, err = ioutil.ReadDir(opts.ProductDir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		fmt.Println("Check directory " + f.Name())
		if !f.IsDir() {
			continue
		}

		if len(f.Name()) < 8 {
			continue
		}

		// It seems that is needed only that
		// directory start with format YYYYMMDD
		_, err = time.Parse("20060102", f.Name()[0:8])
		if err != nil {
			// Skip directory
			fmt.Println("Skipping directory " + f.Name())
			continue
		}

		version := lxd_streams.ProductVersion{
			Items: make(map[string]lxd_streams.ProductVersionItem),
		}

		productBasePath = path.Join(
			strings.TrimRight(opts.PrefixPath, "/"),
			path.Join(product.Directory, f.Name()))
		itemDir = path.Join(opts.ProductDir, f.Name())
		combined = newCombinedSha256Builder()
		fmt.Println(fmt.Sprintf("For product %s I use base path %s.",
			product.Name, productBasePath))

		lxdTarXzItem, _ = checkItem("lxd.tar.xz", itemDir, productBasePath, &combined)
		item, _ = checkItem("rootfs.squashfs", itemDir, productBasePath, &combined)
		if item != nil {
			version.Items["root.squashfs"] = *item
		}
		item, _ = checkItem("rootfs.tar.xz", itemDir, productBasePath, &combined)
		if item != nil {
			version.Items["root.tar.xz"] = *item
		}

		if lxdTarXzItem != nil {
			if combined.SquashFsIsPresent {
				(*lxdTarXzItem).LXDHashSha256SquashFs = hex.EncodeToString(
					combined.CombinedSquashfsSha256.Sum(nil),
				)
			}

			if combined.TarXzIsPresent {
				sha := hex.EncodeToString(combined.CombinedRootxzSha256.Sum(nil))
				(*lxdTarXzItem).LXDHashSha256RootXz = sha
				(*lxdTarXzItem).LXDHashSha256 = sha
			}
			version.Items["lxd.tar.xz"] = *lxdTarXzItem
		}

		ans.Versions[f.Name()] = version
	}

	return ans, nil
}

func checkItem(base, dir, productBasePath string, combined *CombinedSha256Builder) (*lxd_streams.ProductVersionItem, error) {
	var err error
	var ans *lxd_streams.ProductVersionItem
	var filePath string = fmt.Sprintf("%s/%s", dir, base)
	var ftype string
	var f os.FileInfo
	var fmd5 hash.Hash = md5.New()
	var fsha hash.Hash = sha256.New()
	var buf []byte = make([]byte, BYTE_BUFFER_LEN)
	var pb []byte
	var nBytes int

	if combined == nil {
		return nil, fmt.Errorf("Invalid combined struct")
	}

	if base == "rootfs.squashfs" {
		ftype = "squashfs"
		(*combined).SquashFsIsPresent = true
	} else if base == "lxd.tar.xz" {
		ftype = base
	} else if base == "rootfs.tar.xz" {
		(*combined).TarXzIsPresent = true
		ftype = "root.tar.xz"
	} else {
		return nil, fmt.Errorf("Unexpected file " + base)
	}

	fmt.Println("Check file " + base)

	if f, err = os.Stat(filePath); os.IsNotExist(err) {
		return nil, err
	}

	ans = &lxd_streams.ProductVersionItem{
		Path:     fmt.Sprintf("%s/%s", productBasePath, base),
		FileType: ftype,
		Size:     f.Size(),
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0665)
	if err != nil {
		fmt.Println("Error on read file " + filePath)
		return nil, err
	}
	defer file.Close()

	for {
		nBytes, err = file.Read(buf)

		if nBytes > 0 {
			pb = buf[0:nBytes]
			fmd5.Write(pb)
			fsha.Write(pb)
			if base == "lxd.tar.xz" {
				(*combined).CombinedRootxzSha256.Write(pb)
				(*combined).CombinedSquashfsSha256.Write(pb)
			} else if base == "rootfs.tar.xz" {
				(*combined).CombinedRootxzSha256.Write(pb)
			} else if base == "rootfs.squashfs" {
				(*combined).CombinedSquashfsSha256.Write(pb)
			}
		}

		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			fmt.Println("Error read bytes from file " + filePath)
			return nil, err
		}
	}

	ans.HashMd5 = hex.EncodeToString(fmd5.Sum(nil))
	ans.HashSha256 = hex.EncodeToString(fsha.Sum(nil))

	return ans, nil
}

func WriteVersionsManifestJson(manifest *VersionsSSBuilderManifest, out io.Writer) error {
	enc := json.NewEncoder(out)
	return enc.Encode(manifest)
}

func ReadVersionsManifestJsonFromUrl(url, apiKey string) (*VersionsSSBuilderManifest, error) {
	var ans *VersionsSSBuilderManifest = nil
	var req *http.Request = nil
	var err error

	transport := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		MaxIdleConns:    5,
		IdleConnTimeout: 30 * time.Second,
	}

	// TODO: To refactor
	skipVerifyCert := os.Getenv("SSBUILDER_INSECURE_SKIPVERIFY")
	if skipVerifyCert == "1" {
		fmt.Printf("SSBUILDER_INSECURE_SKIPVERIFY catched for url %s. You know what you do.\n", url)
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	var timeout int = 60

	httpTimeout := os.Getenv("SSBUILDER_HTTP_TIMEOUT")
	if httpTimeout != "" {
		t, err := strconv.Atoi(httpTimeout)
		if err == nil {
			fmt.Printf("SSBUILDER_HTTP_TIMEOUT available. Using %s\n", httpTimeout)
			timeout = t
		} else {
			fmt.Printf(
				"SSBUILDER_HTTP_TIMEOUT available. Ignoring wrong value %s\n",
				httpTimeout)
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeout) * time.Second,
	}

	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if apiKey != "" {
		req.Header.Add("Authorization", "token "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid response %d for url %s",
			resp.StatusCode, url)
	}

	byteValue, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ans = &VersionsSSBuilderManifest{}
	err = json.Unmarshal(byteValue, ans)
	if err != nil {
		return nil, err
	}

	return ans, nil
}

func ReadVersionsManifestJson(ssbPath string) (*VersionsSSBuilderManifest, error) {
	file, err := os.OpenFile(ssbPath, os.O_RDONLY, 0665)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var ans *VersionsSSBuilderManifest = &VersionsSSBuilderManifest{}
	err = json.Unmarshal(byteValue, ans)
	if err != nil {
		return nil, err
	}

	return ans, nil
}

func ReadImageFile(imageFile, prefixPath string) (*d_shared.Definition, error) {
	var err error
	var image string
	var ibytes []byte
	var ans d_shared.Definition

	if _, err = os.Stat(path.Join(prefixPath, imageFile)); os.IsNotExist(err) {
		// Check if exists a local image file or through path defined if there is ABS.
		if _, err = os.Stat(imageFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("Image file %s not found.", imageFile)
		}
		image = imageFile
	} else {
		image = path.Join(prefixPath, imageFile)
	}

	// Read configuration
	ibytes, err = ioutil.ReadFile(image)
	if err != nil {
		fmt.Printf("Error on read file %s!\n", image)
		return nil, err
	}

	viper := v.New()
	viper.SetConfigType("yaml")
	err = viper.ReadConfig(bytes.NewBuffer(ibytes))
	if err != nil {
		fmt.Printf("Error on process configuration file %s\n", image)
		return nil, err
	}

	err = viper.Unmarshal(&ans)
	if err != nil {
		fmt.Printf("Error on unmarshal file %s: %s\n", image, err.Error())
		return nil, err
	}

	return &ans, nil
}
