BUILDDIR=/tmp
WORKDIR=github.com/harukasan/go-libwebp
CURDIR=$(shell pwd)

all: test

test:
	go test -v ${WORKDIR}/...

libwebp:
	test -e ${LIBWEBP_PREFIX}/lib/libwebp.so || ( \
		cd ${BUILDDIR} \
		&& wget http://downloads.webmproject.org/releases/webp/libwebp-${LIBWEBP_VERSION}.tar.gz \
		&& tar xf libwebp-${LIBWEBP_VERSION}.tar.gz \
		&& cd libwebp-${LIBWEBP_VERSION} \
		&& ./configure --prefix=${LIBWEBP_PREFIX} \
		&& make \
		&& make install \
	)

docker-test:
	docker run -v ${CURDIR}:/go/src/github.com/harukasan/go-libwebp -it go-libwebp

docker-sh:
	docker run -v ${CURDIR}:/go/src/github.com/harukasan/go-libwebp -it go-libwebp sh

docker-build:
	docker build  -t go-libwebp .


.PHONY: \
	all \
	test \
	docker-test \
	docker-sh \
	docker-build

