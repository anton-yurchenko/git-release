ARG BUILDER=golang:1.13.1-alpine

FROM ${BUILDER} as build
WORKDIR /opt/src
COPY . .
RUN addgroup -g 1000 appuser &&\
    adduser -D -u 1000 -G appuser appuser
RUN apk add --update --no-cache alpine-sdk
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /opt/app

FROM scratch
LABEL "repository"="https://github.com/anton-yurchenko/git-release"
LABEL "maintainer"="Anton Yurchenko <anton.doar@gmail.com>"
LABEL "version"="2.0.0"
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build --chown=1000:0 /opt/app /app
ENTRYPOINT [ "/app" ]