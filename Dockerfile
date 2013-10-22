FROM ubuntu:12.04
MAINTAINER Dan Buch <d.buch@modcloth.com>

ENV DEBIAN_FRONTEND noninteractive
ENV PATH /usr/local/rvm/wrappers/ruby-2.0.0-p247:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin
ENV HOOKWORM_VERSION v0.4.0

RUN apt-get update -yq && \
    apt-get install -yq curl && \
    curl -L -s https://godeb.s3.amazonaws.com/godeb-amd64.tar.gz | tar xzf - && \
    ./godeb install 1.1.2 && \
    curl -L -s https://get.rvm.io | bash -s stable --ruby=2.0.0-p247 && \
    gem install mail -v '2.5.4' --no-ri --no-rdoc && \
    cd / && \
    mkdir -p /data /public && \
    curl -L -s https://s3.amazonaws.com/modcloth-public-travis-artifacts/artifacts/binaries/linux/amd64/hookworm/$HOOKWORM_VERSION/hookworm.tar.bz2 | tar xjf - && \
    mkdir -p /hookworm/src && \
    cd /hookworm/src && \
    tar xjf /hookworm/hookworm-src.tar.bz2 && \
    ln -s /hookworm/src/worm.d /hookworm/worm.d

EXPOSE 9988
CMD ["/hookworm/hookworm-server"]
VOLUME ["/data", "/public"]
