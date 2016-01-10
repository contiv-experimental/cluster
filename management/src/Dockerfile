# This file describes the environment setup to build cluster-manager
FROM golang:1.5

ENV GOPATH=/go

ARG user
ARG uid
# give permissions on gopath dir to the build user and then set it as current user
RUN useradd --uid ${uid} --password foo ${user}
RUN chown -R ${uid}:${uid} ${GOPATH}
USER ${user}

# Pre-install dependencies
RUN go get github.com/tools/godep
RUN go get github.com/golang/lint/golint
RUN go get golang.org/x/tools/cmd/stringer
RUN go get github.com/golang/mock/gomock
RUN go get github.com/golang/mock/mockgen

# set GOBIN to make go install cluster binaries there
ARG work_dir
ENV GOBIN=${work_dir}/bin

CMD bash
