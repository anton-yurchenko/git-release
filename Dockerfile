FROM golang:1.16 as builder
WORKDIR /opt/src
COPY . .
RUN groupadd -g 1000 appuser &&\
    useradd -m -u 1000 -g appuser appuser
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /opt/app

FROM scratch
LABEL "repository"="https://github.com/anton-yurchenko/git-release"
LABEL "maintainer"="Anton Yurchenko <anton.doar@gmail.com>"
LABEL "version"="4.0.0-rc"
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY LICENSE.md /LICENSE.md
COPY --from=builder --chown=1000:0 /opt/app /app
ENTRYPOINT [ "/app" ]