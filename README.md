# sequences-bridge

Configure by copying `.envrc.sample` and adding appropriate keys to it:

    cp .envrc.sample .envrc
    # add SPRING_PRIVATE_KEY and SPRING_PUBLIC_KEY
    direnv allow

Build and run:

    go build . && ./sequences-bridge