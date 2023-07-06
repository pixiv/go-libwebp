repo = github.com/pixiv/go-libwebp
build_dir = /tmp
cur_dir = $(shell pwd)
libwebp_so = ${LIBWEBP_PREFIX}/lib/libwebp.so
LIBWEBP_VERSION ?= 0.5.1

all: test

test:
	go test -v $(repo)/...

libwebp: $(libwebp_so)

$(libwebp_so):
	cd $(build_dir) \
	&& wget http://downloads.webmproject.org/releases/webp/libwebp-$(LIBWEBP_VERSION).tar.gz \
	&& tar xf libwebp-$(LIBWEBP_VERSION).tar.gz \
	&& cd libwebp-$(LIBWEBP_VERSION) \
	&& ./configure --prefix=$(LIBWEBP_PREFIX) \
	&& make \
	&& make install

docker-test:
	docker run -v $(cur_dir):/go/src/$(repo) -it go-libwebp:$(LIBWEBP_VERSION)

docker-sh:
	docker run -v $(cur_dir):/go/src/$(repo) -it go-libwebp:$(LIBWEBP_VERSION) sh

docker-build:
	docker build -t go-libwebp:$(LIBWEBP_VERSION) .

docker-clean:
	docker rm $$(docker ps -a -q -f "ancestor=go-libwebp")

.PHONY: \
	all \
	test \
	libwebp \
	docker-test \
	docker-sh \
	docker-build \
	docker-clean
