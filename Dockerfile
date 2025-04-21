# Stage 1: Build environment
FROM gocv/opencv:4.10.0 AS builder

# Install additional build dependencies
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
    git \
    wget \
    unzip \
    tesseract-ocr \
    tesseract-ocr-slk \
    sqlite3 \
    pdftk \
    libx264-dev \
    libegl1-mesa-dev \
    libxkbcommon-dev \
    libwayland-dev \
    libvulkan-dev \
    libxkbcommon-x11-dev \
    libx11-xcb-dev \
    && rm -rf /var/lib/apt/lists/*
    

# Set Go environment
ENV GOPATH=/go \
    PATH=$PATH:/go/bin

# Build your application (replace with your actual build steps)
WORKDIR /go/src/app

# Copy go mod files first so Docker can cache dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Example build command - modify according to your project
RUN go build -o /app/main .

# Stage 2: Runtime environment
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
    libgl1 \
    libegl1-mesa \
    libxkbcommon0 \
    libxcursor1 \
    libxkbcommon-x11-0 \
    libwayland-cursor0 \
    libwayland-egl1 \
    libvulkan1 \
    libx11-xcb1 \
    libgtk2.0-0 \
    libdc1394-25 \
    libtesseract5 \
    libsqlite3-0 \
    pdftk \
    libx264-164 \
    libpcre3 \
    dbus-x11 \
    && rm -rf /var/lib/apt/lists/*

# Copy binaries and libraries from builder
COPY --from=builder /app/main /app/main
COPY --from=builder /usr/local/lib /usr/local/lib
# Copy required shared libraries from builder
COPY --from=builder /usr/lib/x86_64-linux-gnu /usr/lib/x86_64-linux-gnu

# Configure environment
ENV LD_LIBRARY_PATH=/usr/local/lib \
    PKG_CONFIG_PATH=/usr/local/lib/pkgconfig


WORKDIR /app
COPY ./database ./database
COPY ./assets ./assets
CMD ["/app/main"]