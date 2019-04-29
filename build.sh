#!/bin/bash

set -xe

s3_location="s3://bucket-name"

my_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
temp_dir=$(mktemp -d)
function cleanup() {
    rm -rf "${temp_dir}"
}
trap cleanup EXIT

cd ${my_dir}/example-service
mvn package
cp ./target/SpringBootMavenExample-2.1.1.RELEASE.jar "${temp_dir}"
cp -r ./scripts "${temp_dir}"
cp ./appspec.yml "${temp_dir}"

cd ${my_dir}
go build
cp ./fi-proxy "${temp_dir}"
cd "${temp_dir}"
ls -l
aws deploy push --application-name TestApplication --s3-location "${s3_location}"