ARG BUILD_IMAGE=golang:1.19-bullseye
ARG RUN_TEST=true

FROM ${BUILD_IMAGE} AS build
ARG RUN_TEST

ADD . /workspace

WORKDIR /workspace

ENV WORKSPACE=/workspace

RUN make all

RUN mv output/$(go env GOOS)/$(go env GOARCH)/registrycli output/

RUN [ "${RUN_TEST}" != "true" ] || bash /workspace/tests/simple-tests.sh

FROM scratch

COPY --from=build /workspace/output/registrycli /
