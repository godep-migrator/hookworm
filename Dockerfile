FROM ubuntu:12.04
MAINTAINER Dan Buch <d.buch@modcloth.com>

ENV DEBIAN_FRONTEND noninteractive
ENV PATH /usr/local/rvm/wrappers/ruby-2.0.0-p353:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin
ENV HOOKWORM_VERSION v0.5.0

RUN apt-get update -yq
RUN apt-get install -yq curl
RUN curl -s -L https://go.googlecode.com/files/go1.2.linux-amd64.tar.gz | tar xzf - -C /usr/local && \
    ln -svf /usr/local/go/bin/* /usr/local/bin/
RUN curl -L -s https://get.rvm.io | bash -s stable --ruby=2.0.0-p353
RUN cd / && mkdir -p /data /public
RUN cd / && curl -L -s https://s3.amazonaws.com/modcloth-public-travis-artifacts/artifacts/binaries/linux/amd64/hookworm/$HOOKWORM_VERSION/hookworm.tar.bz2 | tar xjf -
RUN mkdir -p /hookworm/src && \
    cd /hookworm/src && \
    curl -L -s https://s3.amazonaws.com/modcloth-public-travis-artifacts/artifacts/binaries/linux/amd64/hookworm/$HOOKWORM_VERSION/hookworm.src.tar.bz2 | tar xjf -
RUN cd /hookworm/src && gem install -g Gemfile --no-ri --no-rdoc

EXPOSE 9988
CMD ["/hookworm/hookworm-server"]
VOLUME ["/data", "/public"]
