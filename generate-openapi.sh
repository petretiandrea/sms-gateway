#!/usr/bin/env sh
docker run --rm \
  -v ${PWD}:/local openapitools/openapi-generator-cli generate \
  -i /local/api-spec/openapi.yaml \
  -g go-gin-server \
  -o /local/internal/generated \
  --global-property apiTests=false \
  --global-property apiDocs=false \
  --global-property modelTests=false \
  --global-property apis,models,supportingFiles=routers.go\
  --additional-properties=interfaceOnly=true,packageName=openapi,apiPath=openapi