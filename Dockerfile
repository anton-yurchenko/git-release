FROM golang:1.13.1-alpine

LABEL "repository"="https://github.com/anton-yurchenko/git-release"
LABEL "maintainer"="Anton Yurchenko <anton.doar@gmail.com>"
LABEL "version"="1.0.0"

WORKDIR /opt/src
COPY . .

RUN go install &&\
    go build -o /opt/release &&\
    rm -rf /opt/src
ENTRYPOINT [ "/opt/release" ]