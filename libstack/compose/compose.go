package compose

import (
	"context"
	"fmt"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose-cli/pkg/api"
	"github.com/docker/compose-cli/pkg/compose"
)

// Up create and start containers
func Up(filePaths []string, host, projectName, envFilePath string) ([]byte, error) {
	service, err := prepareService(host)
	if err != nil {
		return nil, fmt.Errorf("failed creating compose service: %w", err)
	}

	project, err := prepareProject(filePaths, projectName, envFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed preparing project: %w", err)
	}

	err = service.Up(context.Background(), project, api.UpOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed deploying: %w", err)
	}

	return []byte{}, nil
}

// Down stop and remove containers
func Down(filePaths []string, host, projectName string) ([]byte, error) {
	service, err := prepareService(host)
	if err != nil {
		return nil, fmt.Errorf("failed creating compose service: %w", err)
	}

	err = service.Down(context.Background(), projectName, api.DownOptions{RemoveOrphans: true})
	if err != nil {
		return nil, fmt.Errorf("failed removing: %w", err)
	}

	return []byte{}, nil
}

func prepareService(host string) (api.Service, error) {
	// compose-go outputs to stdout/stderr anyway, there's no way to overwrite it for now.
	dockercli, err := command.NewDockerCli(command.WithStandardStreams())
	if err != nil {
		return nil, fmt.Errorf("error creating client %w", err)
	}

	initOpts := flags.NewClientOptions()
	if host != "" {
		initOpts = &flags.ClientOptions{Common: &flags.CommonOptions{Hosts: []string{host}}}
	}

	err = dockercli.Initialize(initOpts)
	if err != nil {
		return nil, fmt.Errorf("error init client %w", err)
	}

	return compose.NewComposeService(dockercli.Client(), &configfile.ConfigFile{}), nil
}

func prepareProject(filePaths []string, projectName, envFilePath string) (*types.Project, error) {
	additionalOpts := []cli.ProjectOptionsFn{cli.WithName(projectName)}

	if envFilePath != "" {
		additionalOpts = append(additionalOpts, cli.WithEnvFile(envFilePath))
	}

	projectOpts, err := cli.NewProjectOptions(filePaths, additionalOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed creating project options: %w", err)
	}

	project, err := cli.ProjectFromOptions(projectOpts)
	if err != nil {
		return nil, fmt.Errorf("error loading files: %w", err)
	}

	return project, nil
}