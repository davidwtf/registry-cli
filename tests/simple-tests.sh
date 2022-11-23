#!/bin/bash
WORKSPACE=${WORKSPACE:-"/workspace"}

REGISTRY_VERSION=${REGISTRY_VERSION:-"v2.8.1"}

T="${WORKSPACE}/output/registrycli"

set -xe

function prepare_registry() {
    local url="https://github.com/distribution/distribution/releases/download/${REGISTRY_VERSION}/registry_${REGISTRY_VERSION:1}_linux_$(go env GOARCH).tar.gz"
    curl -L "${url}" | tar xvz -C /usr/bin/ --exclude LICENSE --exclude READEME.md
    mkdir -p "/var/lib/registry"
    tar xzf ${WORKSPACE}/tests/registry.tgz -C "/var/lib/registry"

    /usr/bin/registry serve "${WORKSPACE}/tests/registry.yaml" &
    sleep 10
}

function test_repos() {
    ${T} repos 127.0.0.1:5000 --plain-http
}

function test_tags() {
    ${T} tags 127.0.0.1:5000/repo1 --plain-http
}

function test_inspect() {
    ${T} inspect 127.0.0.1:5000/repo1:v1.0 --plain-http
}

function test_layer() {
    ${T} layer 127.0.0.1:5000/repo1@sha256:36842a4bab9b581f82e33fc5af9caa57f977c591fd02a6e0047887ad3ab424c3 --plain-http
}

function test_del() {
    ${T} del 127.0.0.1:5000/repo1:v1.0 --plain-http
}


function main() {
    prepare_registry
    test_repos
    test_tags
    test_inspect
    test_layer
    test_del
}

main