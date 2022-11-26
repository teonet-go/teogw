
# Copyright 2022 Kirill Scherba <kirill@scherba.ru>.  All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
#
# Teonet teogw server application docker file
#
# Docker build (included private repositories):
# --------------------------------------------
#
#   docker build --build-arg github_user="${USERNAME}" --build-arg github_personal_token="${TOKEN_FOR_REPOSITORIES}" -t teogw -f ./.github/Dockerfile .
#
# Docker test run:
# ---------------
#
#   docker run --network host -it --rm teogw teogw
#
# Publish to github:
# -----------------
#
#   docker login ghcr.io -u USERNAME -p TOKEN_FOR_PACKAGES
#   #
#   docker tag teogw ghcr.io/teonet-go/teogw:0.0.0
#   docker push ghcr.io/teonet-go/teogw:0.0.0
#
# Run from gihub repo:
# -------------------
#
#   # Copy static and tmpl folders to the host $HOME/.config/teonet/teogw
#   # cp -r static $HOME/.config/teonet/teogw
#   # cp -r tmpl $HOME/.config/teonet/teogw
#   docker run --network=host -it --restart=always --name teogw -p 7701:7701/udp -v $HOME/.config/teonet/teogw:/root/.config/teonet/teogw ghcr.io/teonet-go/teogw:0.0.0 teogw -tp 7701 -loglevel=debugv
#

# Build stage
#
FROM golang:alpine AS builder
# RUN apk add --no-cache git build-base
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .

# Add the keys from arguments
# ARG github_user
# ENV github_user=$github_user
# ARG github_personal_token
# ENV github_personal_token=$github_personal_token
ARG version
ENV version=$version

# Change github url
# RUN git config \
#     --global \
#     url."https://${github_user}:${github_personal_token}@github.com".insteadOf \
#     "https://github.com"

RUN go get
RUN go install ./cmd/teogw

# Final stage
#
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/* /usr/local/bin/
CMD "teogw"
LABEL Name=teogw Version=$(version)
