package setup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/integralist/go-findroot/find"
	"os"
	"os/exec"
	"time"
)

var (
	ErrTimeOut error = errors.New("error waiting for container to start, timed out")
)

// DockerComposeClient -
type DockerComposeClient struct {
	path string
}

// NewDockerComposeClient -
func NewDockerComposeClient(path string) *DockerComposeClient {
	return &DockerComposeClient{path}
}

func (c *DockerComposeClient) run(command string) ([]byte, error) {
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("docker compose -f %s %s", c.path, command))

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run command: %v", err)
	}

	return out.Bytes(), nil
}

// FindContainer -
func (c *DockerComposeClient) FindContainer(name string) (*Container, error) {
	var containers []Container
	out, err := c.run("ps --format=json")
	if err != nil {
		return nil, fmt.Errorf("error finding container by name %s, %v", name, err)
	}

	if err := json.Unmarshal(out, &containers); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	for _, c := range containers {
		if c.Service == name {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("no container found by the name: %s", name)
}

// Up runs a docker compose up on the given docker-compose file,
// note, if you need to wait for containers to have started up
// then use Wait, or WaitMany
func (c *DockerComposeClient) Up() error {
	_, err := c.run("up -d")
	if err != nil {
		return fmt.Errorf("error running docker compose up: %v", err)
	}
	return nil
}

// Stop a container by name, note that `name` refers to the name of the key of
// your service/container under the `services` list in your docker-compose. For
// example:
// services:
//   <NAME>:
//     image: redis:alpine
//
// Name does not refer to the generated name, or the `name` field.
func (c *DockerComposeClient) Stop(name string) error {
	_, err := c.FindContainer(name)
	if err != nil {
		return fmt.Errorf("error finding this container by name %s: %v", name, err)
	}

	_, err = c.run("stop " + name)
	if err != nil {
		return fmt.Errorf("error bringing down service: %v", err)
	}

	return nil
}

// Down will bring down the docker-compose project
func (c *DockerComposeClient) Down() error {
	_, err := c.run("down")
	if err != nil {
		return fmt.Errorf("error bringing project down: %v", err)
	}

	return nil
}

// Wait -
func (c *DockerComposeClient) Wait(name string, retries int, interval time.Duration) (<-chan *Container, <-chan error) {
	wait := make(chan *Container)
	errs := make(chan error)
	count := 0
	go func() {
		for {
			count++
			if count >= retries {
				errs <- ErrTimeOut
				close(errs)
				close(wait)
				return
			}

			container, err := c.FindContainer(name)
			if container != nil && err == nil && container.ExitCode == 0 && container.State == "running" {
				wait <- container
				close(wait)
				close(errs)
				return
			}

			time.Sleep(interval)
		}
	}()
	return wait, errs
}

// WaitMany -
func (c *DockerComposeClient) WaitMany(containers []string, retries int, interval time.Duration) <-chan error {
	done := make(chan error)
	for _, container := range containers {
		container := container
		go func() {
			awaiting, e := c.Wait(container, retries, interval)
			for {
				select {
				case err := <-e:
					done <- err
					return
				case <-awaiting:
					return
				}
			}
		}()
	}
	go func() { done <- nil }()
	return done
}

// GetRootConfigPath finds the docker-compose file in the root of the project
func GetRootConfigPath(name string) (string, error) {
	root, err := find.Repo()
	if err != nil {
		return "", err
	}

	// Check the file exists in the root
	_, err = os.ReadFile(root.Path + "/" + name)
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("file does not exist by the name: %s", name)
	}

	return root.Path + "/" + name, nil
}
