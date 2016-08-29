# This file describes the environment setup to build cluster-manager
FROM golang:1.7.0

ENV GOPATH=/go

ARG user
ARG uid
# give permissions on gopath dir to the build user and then set it as current user
RUN getent passwd ${user}; if [ "$?" != "0" ]; then useradd --uid ${uid} --password foo ${user}; fi
RUN chown -R ${uid}:${uid} ${GOPATH}
USER ${user}

# Pre-install dependencies
RUN go get github.com/tools/godep
RUN go get github.com/golang/lint/golint
RUN go get golang.org/x/tools/cmd/stringer
RUN go get github.com/golang/mock/gomock
RUN go get github.com/golang/mock/mockgen
RUN go get github.com/client9/misspell/cmd/misspell
RUN go get github.com/fzipp/gocyclo

# set GOBIN to make go install cluster binaries there
ARG work_dir
ENV GOBIN=${work_dir}/bin

CMD /bin/bash
