FROM stackbrew/ubuntu:13.10
MAINTAINER Dan Buch <d.buch@modcloth.com>

ENV DEBIAN_FRONTEND noninteractive
ENV HOOKWORM_VERSION v0.5.0

RUN apt-get update -yq
RUN apt-get install --no-install-recommends -yq curl ca-certificates python ruby2.0 ruby-switch
RUN ruby-switch --set ruby2.0
RUN update-ca-certificates --fresh
RUN curl -s http://nodejs.org/dist/v0.10.24/node-v0.10.24-linux-x64.tar.gz | tar xzf - --strip-components=1 -C /
RUN mkdir -p /data /public /hookworm/src
RUN cd / && curl -L -s https://s3.amazonaws.com/modcloth-public-travis-artifacts/artifacts/binaries/linux/amd64/hookworm/$HOOKWORM_VERSION/hookworm.tar.bz2 | tar xjf -
RUN cd /hookworm/src && curl -L -s https://s3.amazonaws.com/modcloth-public-travis-artifacts/artifacts/binaries/linux/amd64/hookworm/$HOOKWORM_VERSION/hookworm.src.tar.bz2 | tar xjf -
RUN cd /hookworm/src && gem install -g Gemfile --no-ri --no-rdoc

EXPOSE 9988
CMD ["/hookworm/hookworm-server"]
VOLUME ["/data", "/public"]
