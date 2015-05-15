package rancher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/utils"
	"github.com/rancherio/rancher-compose/librcompose/project"
)

type Uploader interface {
	Upload(p *project.Project, name string, reader io.ReadSeeker, hash string) (string, string, error)
	Name() string
}

func getUploader() Uploader {
	// This should do some factory magic
	return &S3Uploader{}
}

func Upload(p *project.Project, name string) (string, string, error) {
	uploader := getUploader()
	logrus.Infof("Uploading build for %s using provider %s", name, uploader.Name())

	content, hash, err := createBuildArchive(p, name)
	if err != nil {
		return "", "", err
	}

	return uploader.Upload(p, name, content, hash)
}

func createBuildArchive(p *project.Project, name string) (io.ReadSeeker, string, error) {
	tar, err := createTar(p, name)
	if err != nil {
		return nil, "", err
	}
	defer tar.Close()

	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, "", err
	}

	if err := os.Remove(tempFile.Name()); err != nil {
		tempFile.Close()
		return nil, "", err
	}

	digest := sha256.New()
	output := io.MultiWriter(tempFile, digest)

	_, err = io.Copy(output, tar)
	if err != nil {
		tempFile.Close()
		return nil, "", err
	}

	hexString := hex.EncodeToString(digest.Sum([]byte{}))
	_, err = tempFile.Seek(0, 0)
	if err != nil {
		tempFile.Close()
		return nil, "", err
	}

	return tempFile, hexString, nil
}

func createTar(p *project.Project, name string) (io.ReadCloser, error) {
	// This code was ripped off from docker/api/client/build.go

	serviceConfig := p.Configs[name]
	root := serviceConfig.Build
	dockerfileName := filepath.Join(root, serviceConfig.Dockerfile)

	fmt.Println("!!!!!!!!!2 " + root + " " + dockerfileName)

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	filename := dockerfileName

	if dockerfileName == "" {
		// No -f/--file was specified so use the default
		dockerfileName = api.DefaultDockerfileName
		filename = filepath.Join(absRoot, dockerfileName)

		// Just to be nice ;-) look for 'dockerfile' too but only
		// use it if we found it, otherwise ignore this check
		if _, err = os.Lstat(filename); os.IsNotExist(err) {
			tmpFN := path.Join(absRoot, strings.ToLower(dockerfileName))
			if _, err = os.Lstat(tmpFN); err == nil {
				dockerfileName = strings.ToLower(dockerfileName)
				filename = tmpFN
			}
		}
	}

	origDockerfile := dockerfileName // used for error msg
	if filename, err = filepath.Abs(filename); err != nil {
		return nil, err
	}

	fmt.Println("!!!!!f " + filename)

	// Now reset the dockerfileName to be relative to the build context
	dockerfileName, err = filepath.Rel(absRoot, filename)
	if err != nil {
		return nil, err
	}

	fmt.Println("!!!!!d " + dockerfileName)
	// And canonicalize dockerfile name to a platform-independent one
	dockerfileName, err = archive.CanonicalTarNameForPath(dockerfileName)
	if err != nil {
		return nil, fmt.Errorf("Cannot canonicalize dockerfile path %s: %v", dockerfileName, err)
	}

	if _, err = os.Lstat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("Cannot locate Dockerfile: %s", origDockerfile)
	}
	var includes = []string{"."}

	excludes, err := utils.ReadDockerIgnore(path.Join(root, ".dockerignore"))
	if err != nil {
		return nil, err
	}

	// If .dockerignore mentions .dockerignore or the Dockerfile
	// then make sure we send both files over to the daemon
	// because Dockerfile is, obviously, needed no matter what, and
	// .dockerignore is needed to know if either one needs to be
	// removed.  The deamon will remove them for us, if needed, after it
	// parses the Dockerfile.
	keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
	keepThem2, _ := fileutils.Matches(dockerfileName, excludes)
	if keepThem1 || keepThem2 {
		includes = append(includes, ".dockerignore", dockerfileName)
	}

	if err := utils.ValidateContextDirectory(root, excludes); err != nil {
		return nil, fmt.Errorf("Error checking context is accessible: '%s'. Please check permissions and try again.", err)
	}

	options := &archive.TarOptions{
		Compression:     archive.Uncompressed,
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
	}

	return archive.TarWithOptions(root, options)
}
