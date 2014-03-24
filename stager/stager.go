package stager

import (
	"errors"
	"fmt"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry-incubator/runtime-schema/router"
	"github.com/cloudfoundry/gunk/urljoiner"
	"strings"
	"time"
)

type Stager interface {
	Stage(models.StagingRequestFromCC, string) error
}

type stager struct {
	stagerBBS bbs.StagerBBS
	compilers map[string]string
}

func NewStager(stagerBBS bbs.StagerBBS, compilers map[string]string) Stager {
	return &stager{
		stagerBBS: stagerBBS,
		compilers: compilers,
	}
}

var ErrNoFileServerPresent = errors.New("no available file server present")
var ErrMissingAppId = errors.New("missing app id")
var ErrNoCompilerDefined = errors.New("no compiler defined for requested stack")

func (stager *stager) Stage(request models.StagingRequestFromCC, replyTo string) error {
	if len(request.AppId) == 0 {
		return ErrMissingAppId
	}

	fileServerURL, err := stager.stagerBBS.GetAvailableFileServer()
	if err != nil {
		return ErrNoFileServerPresent
	}

	compilerURL, err := stager.compilerDownloadURL(request, fileServerURL)
	if err != nil {
		return err
	}

	buildpacksOrder := []string{}
	for _, buildpack := range request.Buildpacks {
		buildpacksOrder = append(buildpacksOrder, buildpack.Key)
	}

	smeltingConfig := models.NewLinuxSmeltingConfig(buildpacksOrder)

	actions := []models.ExecutorAction{}

	actions = append(actions, models.ExecutorAction{
		models.DownloadAction{
			From:    compilerURL,
			To:      smeltingConfig.CompilerPath(),
			Extract: true,
		},
	})

	actions = append(actions, models.ExecutorAction{
		models.DownloadAction{
			From:    request.AppBitsDownloadUri,
			To:      smeltingConfig.AppDir(),
			Extract: true,
		},
	})

	for _, buildpack := range request.Buildpacks {
		actions = append(actions, models.ExecutorAction{
			models.DownloadAction{
				From:    buildpack.Url,
				To:      smeltingConfig.BuildpackPath(buildpack.Key),
				Extract: true,
			},
		})
	}

	actions = append(actions, models.ExecutorAction{
		models.RunAction{
			Script:  smeltingConfig.Script(),
			Env:     request.Environment,
			Timeout: 15 * time.Minute,
		},
	})

	uploadURL, err := stager.dropletUploadURL(request, fileServerURL)
	if err != nil {
		return err
	}

	actions = append(actions, models.ExecutorAction{
		models.UploadAction{
			From: smeltingConfig.DropletArchivePath(),
			To:   uploadURL,
		},
	})

	actions = append(actions, models.ExecutorAction{
		models.FetchResultAction{
			File: smeltingConfig.ResultJsonPath(),
		},
	})

	err = stager.stagerBBS.DesireRunOnce(&models.RunOnce{
		Guid:            strings.Join([]string{request.AppId, request.TaskId}, "-"),
		Stack:           request.Stack,
		ReplyTo:         replyTo,
		FileDescriptors: request.FileDescriptors,
		MemoryMB:        request.MemoryMB,
		DiskMB:          request.DiskMB,
		Actions:         actions,
		Log: models.LogConfig{
			Guid:       request.AppId,
			SourceName: "STG",
		},
	})

	return err
}

func (stager *stager) compilerDownloadURL(request models.StagingRequestFromCC, fileServerURL string) (string, error) {
	compilerPath, ok := stager.compilers[request.Stack]
	if !ok {
		return "", ErrNoCompilerDefined
	}

	staticRoute, ok := router.NewFileServerRoutes().RouteForHandler(router.FS_STATIC)
	if !ok {
		return "", errors.New("couldn't generate the compiler download path")
	}

	return urljoiner.Join(fileServerURL, staticRoute.Path, compilerPath), nil
}

func (stager *stager) dropletUploadURL(request models.StagingRequestFromCC, fileServerURL string) (string, error) {
	staticRoute, ok := router.NewFileServerRoutes().RouteForHandler(router.FS_UPLOAD_DROPLET)
	if !ok {
		return "", errors.New("couldn't generate the compiler download path")
	}

	path, err := staticRoute.PathWithParams(map[string]string{
		"guid": request.AppId,
	})

	if err != nil {
		return "", fmt.Errorf("failed to build droplet upload URL: %s", err)
	}

	return urljoiner.Join(fileServerURL, path), nil
}
