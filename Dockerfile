FROM alpine:3.12

RUN apk add --no-cache g++ make go

RUN mkdir -p /tmp/go-libwebp
COPY Makefile /tmp/go-libwebp/Makefile

ENV LIBWEBP_PREFIX="/usr/local" \
    LIBWEBP_VERSION="0.5.1"
RUN cd /tmp/go-libwebp && make libwebp

ENV GOPATH="/go" \
    WORKDIR="/workspace/go-libwebp" \
    PATH="/go/bin:/usr/local/go/bin:$PATH" \
    CGO_CFLAGS="-I /usr/local/include"  \
    CGO_LDFLAGS="-L /usr/local/lib" \
    LD_LIBRARY_PATH="/usr/local/lib:$LD_LIBRARY_PATH"

RUN mkdir -p $GOPATH
RUN mkdir -p $WORKDIR
VOLUME $WORKDIR
WORKDIR $WORKDIR

CMD ["make", "test"]
