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
	"fmt"
	"io/ioutil"
	"os"
	exec "os/exec"
	"path"
	"sort"
	"time"

	config "github.com/MottainaiCI/simplestreams-builder/pkg/config"
	tools "github.com/MottainaiCI/simplestreams-builder/pkg/tools"
)

type BuildProductOpts struct {
	BuildLxc        bool
	BuildLxd        bool
	PurgeOldImages  bool
	BuildScriptHook string
}

func NewBuildProductOpts() *BuildProductOpts {
	return &BuildProductOpts{
		BuildLxc:        true,
		BuildLxd:        true,
		PurgeOldImages:  true,
		BuildScriptHook: "",
	}
}

func BuildProduct(product *config.SimpleStreamsProduct, targetDir, imageFile string, opts *BuildProductOpts) error {
	var fileInfo *os.FileInfo = nil
	var err error
	var productDir, dateDir, tmpDir, cacheDir string

	// Hereinafter, a summary of all operations to do:
	// 1. Create target directory if doesn't exist (with date)
	// 2. Check if exists TMPDIR else create it.
	// 3. Create cachedir (CACHEDIR) if doesn't exist.
	// 4. Run distrobuilder build-lxc if option BuildLxc is true
	// 5. Run distrobuilder build-lxd if option BuildLxd is true
	// 6. Purge old images if option PurgeOldImages is true
	// 7. Create images manifest if option CreateImagesManifest is true

	productDir = path.Join(targetDir, product.Directory)
	tmpDir = os.Getenv("TMPDIR")
	cacheDir = os.Getenv("CACHEDIR")

	if tmpDir == "" {
		tmpDir = "/tmp"
	}
	if cacheDir == "" {
		cacheDir = "/tmp/cachedir"
	}

	// Make target directory
	fileInfo, err = tools.MkdirIfNotExist(productDir, 0760)
	if err != nil {
		return err
	} else if fileInfo != nil && !(*fileInfo).IsDir() {
		return fmt.Errorf("Path %s is not a directory.", targetDir)
	}

	// Create temporary directory used by distrobuilder if doesn't exist
	_, err = tools.MkdirIfNotExist(tmpDir, 0760)
	if err != nil {
		fmt.Println("Error on create tmp directory: " + err.Error())
		return err
	}
	// Create cachedir directory
	_, err = tools.MkdirIfNotExist(cacheDir, 0760)
	if err != nil {
		fmt.Println("Error on create cachedir directory: " + err.Error())
		return err
	}

	if opts.BuildLxc || opts.BuildLxd {
		// Directory must be in the format: "20190407_13:00"
		now := time.Now()
		dateDir = path.Join(productDir,
			fmt.Sprintf("%4d%02d%02d_%02d:%02d",
				now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()),
		)
		_, err = tools.MkdirIfNotExist(dateDir, 0760)
		if err != nil {
			return err
		}

		fmt.Println(fmt.Sprintf("Created directory %s for image of the product %s.",
			dateDir, product.Name))
	}

	if product.BuildScriptHook != "" || opts.BuildScriptHook != "" {

		hookScript := product.BuildScriptHook
		if opts.BuildScriptHook != "" {
			// Override hookScript
			hookScript = opts.BuildScriptHook
		}

		// Create rootfs directory
		rootfsDir := path.Join(dateDir, "staging")
		buildDirCommand := exec.Command("distrobuilder",
			"build-dir", imageFile, rootfsDir,
			"--cache-dir", cacheDir)
		defer tools.RemoveDirIfNotExist(rootfsDir)

		buildDirCommand.Stdout = os.Stdout
		buildDirCommand.Stderr = os.Stderr

		err = buildDirCommand.Run()
		if err != nil {
			return err
		}

		runHookCommand := exec.Command(hookScript)
		// Prepare env for hook
		runHookCommand.Env = append(os.Environ(),
			fmt.Sprintf("%s_STAGING_DIR=%s", config.SSB_ENV_PREFIX, rootfsDir),
			fmt.Sprintf("%s_BUILD_PRODUCT=%s", config.SSB_ENV_PREFIX, product.Name))

		runHookCommand.Stdout = os.Stdout
		runHookCommand.Stderr = os.Stderr

		err = runHookCommand.Run()
		if err != nil {
			fmt.Println(fmt.Sprintf("Error on execute hook %s: %s",
				hookScript, err.Error()))
			return err
		}

		// Create LXC package
		if opts.BuildLxc {
			err = packImage(imageFile, rootfsDir, dateDir, cacheDir, "pack-lxc")
			if err != nil {
				return err
			}
		}

		// Create LXD package
		if opts.BuildLxd {
			err = packImage(imageFile, rootfsDir, dateDir, cacheDir, "pack-lxd")
			if err != nil {
				return err
			}
		}

	} else {

		if opts.BuildLxc {
			buildLxcCommand := exec.Command("distrobuilder",
				"build-lxc", imageFile, dateDir,
				"--cache-dir", cacheDir)

			// https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html
			buildLxcCommand.Stdout = os.Stdout
			buildLxcCommand.Stderr = os.Stderr

			err = buildLxcCommand.Run()
			if err != nil {
				return err
			}

			// Create cache dir cleanup by distrobuilder
			_, err = tools.MkdirIfNotExist(cacheDir, 0760)
			if err != nil {
				return err
			}
		}

		if opts.BuildLxd {
			buildLxdCommand := exec.Command("distrobuilder",
				"build-lxd", imageFile, dateDir,
				"--cache-dir", cacheDir)

			buildLxdCommand.Stdout = os.Stdout
			buildLxdCommand.Stderr = os.Stderr

			err = buildLxdCommand.Run()
			if err != nil {
				return err
			}
		}

	}

	if opts.PurgeOldImages {
		err = purgeOldImages(productDir, product)
	}

	return nil
}

func packImage(imageFile, rootfsDir, dateDir, cacheDir, subCommand string) error {
	var err error

	// Create cache dir cleanup by distrobuilder
	_, err = tools.MkdirIfNotExist(cacheDir, 0760)
	if err != nil {
		return err
	}

	fmt.Printf("Executing %s command...\n", subCommand)

	packCommand := exec.Command("distrobuilder",
		subCommand, imageFile, rootfsDir, dateDir,
		"--cache-dir", cacheDir)

	packCommand.Stdout = os.Stdout
	packCommand.Stderr = os.Stderr

	err = packCommand.Run()
	if err != nil {
		return err
	}

	return nil
}

func purgeOldImages(productDir string, product *config.SimpleStreamsProduct) error {
	var err error
	var files []os.FileInfo
	var dates []time.Time
	var date time.Time
	var dateDir string

	// Iterate for every directory
	files, err = ioutil.ReadDir(productDir)
	if err != nil {
		return err
	}

	fmt.Println("Purge directory " + productDir + "...")

	for _, f := range files {

		if !f.IsDir() || len(f.Name()) < 8 {
			continue
		}

		date, err = time.Parse("20060102_15:04", f.Name())
		if err != nil {
			fmt.Println(fmt.Sprintf("Skipping directory %s: %s",
				f.Name(), err.Error()))
			continue
		}

		// Retrieve hour and minutes

		dates = append(dates, date)
	}

	// Sort dates
	sort.Sort(tools.TimeSorter(dates))

	fmt.Println(fmt.Sprintf("Found %d dates.", len(dates)))

	// PRE: I consider to have only one image for day.
	for len(dates) > product.Days {
		// Remove directory old

		dateDir = path.Join(productDir,
			fmt.Sprintf("%4d%02d%02d_%02d:%02d",
				dates[0].Year(),
				dates[0].Month(),
				dates[0].Day(),
				dates[0].Hour(),
				dates[0].Minute()),
		)

		fmt.Println(fmt.Sprintf("Removing directory %s...", dateDir))
		err = os.RemoveAll(dateDir)
		if err != nil {
			fmt.Println(fmt.Sprintf("ERROR on remove directory %s: %s",
				dateDir, err.Error()))
		}

		dates = dates[1:len(dates)]
	}

	return nil
}
