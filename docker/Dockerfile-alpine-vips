FROM alpine:3.5

WORKDIR /tmp
COPY vips-8.3.1.tar.gz /tmp
RUN apk add --no-cache --virtual .build-deps \
  gcc g++ make libc-dev \
  curl \
  automake \
  libtool \
  tar \
  gettext && \
  apk add --no-cache --virtual .libdev-deps \
  glib-dev \
  libpng-dev \
  libwebp-dev \
  libexif-dev \
  libxml2-dev \
  libxml2-dev \
  pango-dev   \
  libjpeg-turbo-dev && \
  apk add --no-cache --virtual .run-deps \
  glib \
  libpng \
  libwebp \
  libexif \
  libxml2 \
  pango \
  libjpeg-turbo && \
  tar zvxf vips-8.3.1.tar.gz && \
  cd vips-8.3.1 && \
  ./configure --without-python --without-gsf && \
  make -j4 && \
  make install && \
  rm -rf /tmp/vips-* && \
  apk del .build-deps && \
  apk del .libdev-deps && \
  rm -rf /var/cache/apk/* && \
  rm -rf /tmp/vips-*

RUN apk add  wqy-zenhei --update-cache --repository http://dl-3.alpinelinux.org/alpine/edge/testing/ --allow-untrusted

ENV CPATH /usr/local/include
ENV LIBRARY_PATH /usr/local/lib
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:$PKG_CONFIG_PATH

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
