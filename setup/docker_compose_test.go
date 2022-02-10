package setup

import (
	"encoding/json"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
	config     = basepath + "/docker-compose.yaml"
	client     = &DockerComposeClient{path: config}
)

func TestCanRunCommand(t *testing.T) {
	is := is.New(t)

	// Stop any already running containers, which might cause a conflict
	stopAllContainers := exec.Command("/bin/bash", "-c", "docker ps $(docker stop -aq)")
	is.NoErr(stopAllContainers.Run())

	_, err := client.run("up -d")
	is.NoErr(err)

	out, err := client.run("ps --format=json")
	is.NoErr(err)

	type container struct {
		ID      string
		Service string
	}

	var containers []container
	is.NoErr(json.Unmarshal(out, &containers))

	var cont *container
	for _, container := range containers {
		if container.Service == "db" {
			cont = &container
		}
	}

	if cont == nil {
		is.Fail()
	}

	is.Equal(cont.Service, "db")

	// Tear down docker-compose stack
	_, err = client.run("down --rmi=all")

	// Query for a list of running containers again
	out, err = client.run("ps --format=json")
	is.NoErr(err)
	is.NoErr(json.Unmarshal(out, &containers))

	// Check there are none running
	is.True(len(containers) == 0)
}

func TestCanFindContainerByName(t *testing.T) {
	is := is.New(t)

	is.NoErr(client.Up())

	container, err := client.FindContainer("db")
	is.NoErr(err)
	is.Equal(container.Service, "db")

	// Teardown
	is.NoErr(client.Down())
}

func TestCanStartAndStopContainer(t *testing.T) {
	is := is.New(t)

	is.NoErr(client.Up())

	container, err := client.FindContainer("db")
	is.NoErr(err)
	is.Equal(container.Service, "db")
	is.NoErr(client.Down())
}

func TestCanWaitForContainer(t *testing.T) {
	is := is.New(t)
	is.NoErr(client.Up())

	// Wait for the db service to be up and running
	awaitRunning, errs := client.Wait("db", 10, time.Millisecond*500)

	for {
		select {
		case container := <-awaitRunning:
			is.Equal(container.Service, "db")
			is.NoErr(client.Down())
			return
		case <-errs:
			is.Fail()
			return
		}
	}
}

func TestCanTimeoutOnWait(t *testing.T) {
	is := is.New(t)
	is.NoErr(client.Up())

	// Wait for the db service to be up and running
	awaitRunning, errs := client.Wait("db", 1, time.Millisecond*1)
	defer is.NoErr(client.Down())
	for {
		select {
		case <-awaitRunning:
			is.Fail()
			return
		case err := <-errs:
			is.Equal(err, ErrTimeOut)
			return
		}
	}
}

func TestCanWaitMany(t *testing.T) {
	is := is.New(t)
	is.NoErr(client.Up())

	// Wait for the db service to be up and running
	waitFor := []string{"db", "cache"}
	done := client.WaitMany(waitFor, 60, time.Second*1)
	is.NoErr(<-done)
	is.NoErr(client.Down())
}

func TestCanGetRootConfigPath(t *testing.T) {
	is := is.New(t)
	path, err := GetRootConfigPath("docker-compose.yml")
	is.NoErr(err)
	log.Println("path: ", path)
	is.True(strings.Contains(path, "mediamagic-platform/docker-compose.yml"))
}
