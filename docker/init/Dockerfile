FROM golang:alpine as base

COPY pre.sh .

RUN apk --no-cache add ca-certificates git jq curl
RUN ./pre.sh

FROM base

VOLUME /tasks

WORKDIR /xmc-init
COPY scripts ./scripts
COPY init.sh .

ENTRYPOINT /xmc-init/init.sh
