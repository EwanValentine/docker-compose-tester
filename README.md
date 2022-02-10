# Test Helpers

Test helpers contains a list of useful tools for writing integration tests.

## Docker Compose Client

Integration tests should use the `docker-compose.yml` set-up in the root of the project.

In order to automate testing using the `docker-compose.yml` file, this package exposes a series of useful helper methods.

Example docker-compose file:

```yaml
version: '3.9'

services:
  auth-service:
    image: auth-service 

  deployment-service:
    image: deployment-service
```

```golang
package main

import (
	"time"
	"testing"
	"github.com/matryer/is"
	"github.com/EwanValentine/docker-compose-tester/setup"
)

func TestCanDoSomething(t *testing.T) {
	is := is.New(t)
	client := setup.NewDockerComposeClient()
	
	// Gets the absolute path of the root directory + 'docker-compose.yml'
	// I.e. /Users/ewanvalentine/development/mediamagic-platform/docker-compose.yml'
	// This is useful for testing, as your tests may execute from various different 
	// directories and contexts
	path, err := setup.GetRootConfigPath("docker-compose.yml")
  is.NoErr(err)
	
	// `$ docker compose up`
	err = client.Up(path)
	is.NoErr(err)

	// This runs `$ docker compose ps` on an interval until the containers specific have started successfully
	retries := 10
	interval := time.Second * 3
	done := client.WaitMany([]string{"auth-service", "deployment-service"}, retries, interval)
  is.NoErr(<-done)
	
	// Do things with your containers
}
```
