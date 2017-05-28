FROM ubuntu:latest
MAINTAINER JW Bell <bjwbell@gmail.com>
RUN apt-get update && apt-get install -y \
    ca-certificates
ADD . /renfish
WORKDIR /renfish
