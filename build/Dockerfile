FROM golang:1.21 as build

#!!!DO NOT REMOVE THE BELOW 3 LINES!!! Needed for github action to work
#GHAONLY:COPY dotssh/* /root/.ssh/
#GHAONLY:COPY dotgitconfig /root/
#GHAONLY:RUN cat /root/dotgitconfig >> /root/.gitconfig

# Download public key for github.com
RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts

# Use git protocol whenever possible
RUN echo '[url "git@github.com:"]\n\tinsteadOf = https://github.com/\n' >> ~/.gitconfig

WORKDIR /build

# Put the go modules in own layer to speed up build
COPY go.mod .
COPY go.sum .
RUN --mount=type=ssh go mod download

# Add remaining code
COPY . .

ARG location=github.com/fluffy-bunny/fluffycore-starterkit/internal/version
ARG commit=AFFE
ARG version=0.0.0
ARG date=19700101
ARG THEBINARY=server
ENV THEBINARY=${THEBINARY}

ENV GOOS=linux
ENV GOARCH=${TARGETARCH}
ENV CGO_ENABLED=0

ENV EXTLDFLAGS="-static"
ENV LDFLAGS="-X ${location}.date=${date} -X ${location}.version=${version} -X ${location}.commit=${commit} -s -w -extldflags '${EXTLDFLAGS}'"
RUN --mount=type=ssh go build -a -tags netgo -ldflags "$LDFLAGS" ./cmd/$THEBINARY


FROM alpine:latest as stage1

RUN apk add -U --no-cache ca-certificates

RUN apk add bash curl && \
  curl -fSsL https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh -o /wait-for-it.sh && \
  chmod +x /wait-for-it.sh && \
  apk del curl


FROM stage1

LABEL maintainer="info@fluffybunny.com"
LABEL description="example backend service"

USER fluffybunnyuser

ARG THEBINARY=server
ENV THEBINARY=${THEBINARY}

WORKDIR /app
COPY --from=build /build/${THEBINARY} ./

ENV PATH=/app:${PATH}
ENV PORT="50051"
ENV REST_PORT="50052"
ENV OAUTH2_PORT="50053"
ENV GRPC_GATEWAY_ENABLED="true"

ENV ENABLE_GRPC_SERVER_REFLECTION="true"
ENV JWT_VALIDATORS__ISSUERS="@@@REPLACE@@@"
ENV JWT_VALIDATORS__JWKS_URLS="@@@REPLACE@@@"
          

# If necessary, CMD can be used to overwrite the mode in docker-compose to be "--gtm.mode=server" or "--gtm.mode=worker" 
ENTRYPOINT ["/app/server"]


 