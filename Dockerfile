FROM golang:1.23.4-bookworm

ENV SERVICE_NAME ipcat

ENV ROOT /opt/$SERVICE_NAME

WORKDIR $ROOT

# Now add the entire source code tree
COPY . $ROOT
