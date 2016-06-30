BUILDDIR=/tmp
WORKDIR=github.com/harukasan/go-libwebp

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

.PHONY: \
	all \
	test

