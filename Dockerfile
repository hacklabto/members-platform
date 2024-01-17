FROM alpine:latest AS builder

RUN apk add go

WORKDIR /build
COPY . .
RUN go build -o hl-web ./cmd/web
# RUN go build -o hl-apply ./cmd/apply
# RUN go build -o hl-memberizer ./cmd/memberizer

FROM alpine:latest
COPY --from=builder /build/hl-web /usr/local/bin/hl-web
