FROM golang:alpine
RUN apk add build-base

WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go install -trimpath \
  -ldflags "-X github.com/redhat-developer/app-services-cli/internal/build.BuildSource="local" \
  -X github.com/redhat-developer/app-services-cli/internal/build.MASSSORedirectPath="mas-sso-callback" \
  -X github.com/redhat-developer/app-services-cli/internal/build.SSORedirectPath="sso-redhat-callback" \
  -X github.com/redhat-developer/app-services-cli/internal/build.DefaultPageNumber="1" \
  -X github.com/redhat-developer/app-services-cli/internal/build.DefaultPageSize="10" \
  -X github.com/redhat-developer/app-services-cli/internal/build.DynamicConfigURL="https://raw.githubusercontent.com/redhat-developer/app-services-ui/main/static/configs/service-constants.json" \
  -X github.com/redhat-developer/app-services-cli/internal/build.RepositoryName="app-services-cli" \
  -X github.com/redhat-developer/app-services-cli/internal/build.RepositoryOwner="redhat-developer" \
  -X github.com/redhat-developer/app-services-cli/internal/build.Version="dev" " ./cmd/rhoas

ENTRYPOINT ["rhoas"]