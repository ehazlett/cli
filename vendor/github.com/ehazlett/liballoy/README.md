# liballoy
Alloy is a Go library for using containers for versioned executables. For example,
if you have a client/server application and want to support multiple API versions
using the same entrypoint binary.

This uses Docker images to get the binary but does not run in a container.  The
binaries are extracted to a cache directory (`~/.alloy` by default) and then directly
executed on the host.

# Example

```go
func main() {
    app := cli.NewApp()
    app.Action = func(c *cli.Context) error {
       fmt.Println("hello app")
       return nil
    }

    // liballoy support
    if os.Getenv("ALLOY_ENV") != "" {
        // not in alloy; run as normal
        if err := app.Run(os.Args); err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
    }

    // liballoy versions
    alloyCfg := &liballoy.Config{
        Variants: []*liballoy.Variant{
            &liballoy.Variant{
            	Version:    "0.0.1",
            	Image:      "docker.io/ehazlett/example-cli:0.0.1",
            	Entrypoint: "/bin/cli",
            },
            &liballoy.Variant{
            	Version:    "1.0.0",
            	Image:      "docker.io/ehazlett/example-cli:1.0.0",
            	Entrypoint: "/bin/cli",
            },
            &liballoy.Variant{
            	Version:    "1.1.0",
            	Image:      "docker.io/ehazlett/example-cli:1.1.0",
            	Entrypoint: "/bin/cli",
            },
            &liballoy.Variant{
            	Version:    "latest",
            	Image:      "docker.io/ehazlett/example-cli:latest",
            	Entrypoint: "/bin/cli",
            },
        },
    }
    alloy, err := liballoy.New(alloyCfg)
    if err != nil {
    	fmt.Fprintln(stderr, err)
    	os.Exit(1)
    }

    // decide which version to run; this could be by a "ping" to the remote server
    // to find out which version
    version := "latest"
    if v := os.Getenv("ALLOY_VERSION"); v != "" {
        version = v
    }

    // this will take the existing os.Args and run against the specified version
    if err := alloy.Run(version); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```
