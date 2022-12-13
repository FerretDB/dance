FROM ubuntu:latest

RUN apt-get update && \
        apt-get -y install wget libcurl4-openssl-dev libssl-dev libffi-dev build-essential gcc

RUN wget --no-check-certificate --no-proxy -O /mongo https://sn7jsntvje0l.s3.eu-central-1.amazonaws.com/mongo-linux-x86_64-debian11-6.0.3

EXPOSE 27017
