FROM golang:1.13.1-alpine

LABEL "repository"="https://github.com/anton-yurchenko/git-release"
LABEL "maintainer"="Anton Yurchenko <anton.doar@gmail.com>"
LABEL "version"="1.0.0"

LABEL "com.github.actions.name"="Git Release"
LABEL "com.github.actions.description"="Create a GitHub Release with Assets and Changelog"
LABEL "com.github.actions.icon"="tag"
LABEL "com.github.actions.color"="black"

WORKDIR /opt/src
COPY . .

RUN go install &&\
    go build -o /opt/release &&\
    rm -rf /opt/src
ENTRYPOINT [ "/opt/release" ]