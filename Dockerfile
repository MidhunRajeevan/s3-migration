FROM image-registry.openshift-image-registry.svc:5000/tars-images/golang:alpine AS builder

RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/github.com/MidhunRajeevan/s3-migration
COPY . .
RUN go get -d -v
RUN export CGO_ENABLED=0 && go build -o /go/bin/tugs3

FROM image-registry.openshift-image-registry.svc:5000/tars-images/alpine:latest

ENV APP_TENANT_STRING=
ENV APP_UPLOAD_LIMIT=
ENV APP_ALLOW_INSECURE=
ENV S3_LOCATION=
ENV S3_ENDPOINT=
ENV S3_ACCESS_KEY=
ENV S3_SECRET_KEY=
ENV S3_USE_SSL=

COPY --from=builder /go/bin/tugs3 /go/bin/tugs3
ENTRYPOINT ["/go/bin/tugs3"]
EXPOSE 9090
