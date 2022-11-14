# neospring-bridge

This is a personal bridge program that cross publishes [sequences](https://brandur.org/sequences) to my [Spring '83 board](https://www.robinsloan.com/lab/specifying-spring-83/). Open-sourced for interest sake, but don't use this for anything.

Configure by copying `.envrc.sample` and adding appropriate keys to it:

    cp .envrc.sample .envrc
    # add SPRING_PRIVATE_KEY and SPRING_PUBLIC_KEY
    direnv allow

Build and run:

    go build . && ./neospring-bridge

## Development

Run the test suite:

    $ go test .

Run lint:

    $ golangci-lint --fix
