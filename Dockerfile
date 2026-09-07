FROM golang:1.24-alpine AS build-stage

# Set working directory
WORKDIR /app

# Update OS packages
RUN apk update && apk upgrade

# Install common tools for build stage
RUN apk add openssl ca-certificates

# pre-copy/cache go.mod for pre-downloading dependencies and only re-downloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy all files
COPY . .

# Build the application
RUN go build -v -o /usr/local/bin/app

# --- Production stage ---
FROM golang:1.24-alpine AS production-stage

# Set constant environment variables area
ENV TZ="Asia/Bangkok"
ENV ENV="production"

WORKDIR /app

# Copy executable file from build-stage
COPY --from=build-stage /usr/local/bin/app /usr/local/bin/app

# Copy env file from build-stage
COPY --from=build-stage /app/.env /app/.env

# Update OS packages
RUN apk update && apk upgrade

# Install tools
RUN apk add \
    bash \
    zip unzip \
    openssl \
    ca-certificates \
    git \
    curl \
    busybox-extras \
    tzdata

# Add dumb-init for support prefork mode
RUN apk add dumb-init

# Non-root user
RUN adduser -D 1001
USER 1001

# Opening ports
EXPOSE 8000
EXPOSE 8001

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["app"]

# Configure a healthcheck to validate that everything is up & running
HEALTHCHECK --timeout=10s CMD curl --silent --fail http://127.0.0.1:8000/health
