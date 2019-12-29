ARG IMG=golang:1.13.1

FROM ${IMG} as base
WORKDIR /opt/src
COPY . .

FROM base as test
ARG CODECOV_TOKEN
RUN go test $(go list ./... | grep -v vendor | grep -v mocks) -race -coverprofile=coverage.txt -covermode=atomic
RUN curl -s https://codecov.io/bash -o script.sh &&\
    bash ./script.sh -t ${CODECOV_TOKEN}

FROM base as build
RUN groupadd -g 1000 appuser &&\
    useradd -m -u 1000 -g appuser appuser
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /opt/app

FROM scratch
LABEL "repository"="https://github.com/anton-yurchenko/git-release"
LABEL "maintainer"="Anton Yurchenko <anton.doar@gmail.com>"
LABEL "version"="2.0.1"
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build --chown=1000:0 /opt/app /app
ENTRYPOINT [ "/app" ]