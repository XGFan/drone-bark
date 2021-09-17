FROM golang:alpine AS builder
WORKDIR /app
ADD . ./
ARG GOPROXY
RUN if [[ -z "$GOPROXY" ]] ; then echo GOPROXY not provided ; else export GOPROXY=$GOPROXY ; fi
RUN GOOS=linux GOARCH=amd64 go build -o bark .

FROM alpine:3
COPY --from=builder /app/bark /bin/bark
USER root
WORKDIR /app/
RUN chmod +x /bin/bark
ENTRYPOINT ["/bin/bark"]