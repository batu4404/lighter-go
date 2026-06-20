### Local builds

build-darwin-local:
    go mod vendor
    go build -buildmode=c-shared -trimpath -o ./build/lighter-darwin-arm64.dylib ./sharedlib/main.go

# Note: build-linux-local does not append -arm or amd64 at end
build-linux-local:
    go mod vendor
    CGO_ENABLED=1 go build -buildmode=c-shared -trimpath -o ./build/lighter-linux.so ./sharedlib/main.go

# Note: build-windows-local does not append -arm or amd64 at end
# Windows build (requires gcc from msys2: choco install msys2)
# CMD:        set PATH=C:\msys64\mingw64\bin;%PATH% && set CGO_ENABLED=1 && go mod vendor && go build -buildmode=c-shared -trimpath -o ./build/signer-amd64.dll ./sharedlib/main.go
# PowerShell: $env:Path='C:\msys64\mingw64\bin;'+$env:Path; $env:CGO_ENABLED='1'; go mod vendor; go build -buildmode=c-shared -trimpath -o ./build/signer-amd64.dll ./sharedlib/main.go
build-windows-local:
    go mod vendor
    $env:Path='C:\msys64\mingw64\bin;'+$env:Path; $env:CGO_ENABLED='1'; go build -buildmode=c-shared -trimpath -o ./build/lighter-windows.dll ./sharedlib/main.go

### Docker builds

# Note: I don't think this works TBH
#build-darwin-arm64-docker:
#    docker run --rm -v ${PWD}:/go/src/sdk -w /go/src/sdk golang:1.23.2-bullseye bash -c " \
#      cd /go/src/sdk && \
#      go build -buildmode=c-shared -trimpath -o ./build/lighter-darwin-arm64.dylib ./sharedlib"

build-linux-amd64-docker:
    go mod vendor
    docker run --rm --platform linux/amd64 -v ${PWD}:/go/src/sdk -w /go/src/sdk golang:1.23.2-bullseye /bin/sh -c " \
      CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildmode=c-shared -buildvcs=false -trimpath -o ./build/lighter-linux-amd64.so ./sharedlib"

build-linux-arm64-docker:
    go mod vendor
    docker run --rm --platform linux/arm64 -v ${PWD}:/go/src/sdk -w /go/src/sdk golang:1.23.2-bullseye /bin/sh -c " \
      CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -buildmode=c-shared -buildvcs=false -trimpath -o ./build/lighter-linux-arm64.so ./sharedlib"

build-windows-amd64-docker:
    go mod vendor
    docker run --rm --platform linux/amd64 -v ${PWD}:/go/src/sdk -w /go/src/sdk golang:1.23.2-bullseye bash -c " \
      apt-get update && \
      apt-get install -y gcc-mingw-w64-x86-64 && \
      CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -buildvcs=false -trimpath -o ./build/lighter-windows-amd64.dll ./sharedlib"

### WASM builds

build-wasm:
    go mod vendor
    GOOS=js GOARCH=wasm go build -trimpath -o ./build/lighter.wasm ./wasm/
### Verification / upstream parity

# Run the full Go test suite.
test:
    go test ./... -count=1

# Byte-for-byte signer parity vs the lighter-python native signer.
# Run after any signer change or upstream lighter-python bump.
parity-signer:
    CGO_ENABLED=1 python3 tools/parity_check.py

# Live WebSocket conformance vs the running Lighter server (the only WS source
# of truth -- there is no WS schema in openapi.json). Public channels run
# unauthenticated; set LIGHTER_WS_AUTH + LIGHTER_WS_ACCOUNT to also exercise
# the account/pool channels.
ws-conformance:
    go test -tags wsconformance -run TestWS -v ./client/ws/ -count=1

# Regenerate the OpenAPI model types after a lighter-python openapi.json bump.
gen-models:
    go generate ./types/apimodels
