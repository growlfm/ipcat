FROM golang:1.22.5-bookworm

ENV SERVICE_NAME ipcat

ENV ROOT /opt/$SERVICE_NAME

WORKDIR $ROOT

# Now add the entire source code tree
COPY . $ROOT
