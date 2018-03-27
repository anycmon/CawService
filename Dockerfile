#
# Super simple example of a Dockerfile
#
FROM ubuntu:latest
MAINTAINER Anycmon "anycmon@gmail.com"

RUN apt-get update

COPY UserService /home
WORKDIR /home

EXPOSE 9090
ENTRYPOINT /home/UserService
