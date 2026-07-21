# ─────────────────────────────────────────────
# Stage 1: 依赖缓存层（仅 go.mod / go.sum）
# 只要依赖不变，此层永久命中缓存
# ─────────────────────────────────────────────
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS deps

WORKDIR /src

# 单独复制依赖文件，充分利用 layer cache
COPY go.mod go.sum ./
RUN go mod download && go mod verify


# ─────────────────────────────────────────────
# Stage 2: 编译
# ─────────────────────────────────────────────
FROM deps AS builder

# 构建参数：支持多平台交叉编译
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build \
      -trimpath \
      -ldflags="-s -w \
        -X main.version=${VERSION} \
        -X main.commit=${COMMIT} \
        -X main.buildDate=${BUILD_DATE}" \
      -o /out/vertiv_exporter \
      ./cmd/vertiv_exporter


# ─────────────────────────────────────────────
# Stage 3: 最终镜像
# distroless/static：无 shell、无包管理器、无 libc
# ─────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

# ARG 作用域仅限当前 Stage，最终镜像必须重新声明才能在 LABEL 中引用
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# OCI 标准镜像元数据
LABEL org.opencontainers.image.title="vertiv-exporter" \
      org.opencontainers.image.description="Prometheus exporter for Vertiv modular cabinet" \
      org.opencontainers.image.source="https://github.com/MarismeCom/vertiv_exporter" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}"

WORKDIR /app

# 仅拷贝二进制，config 由运行环境挂载
COPY --from=builder /out/vertiv_exporter /app/vertiv_exporter

# distroless:nonroot 内置 uid=65532(nonroot)，无需手动 adduser
USER nonroot:nonroot

# Prometheus 标准 exporter 端口
EXPOSE 9101

ENTRYPOINT ["/app/vertiv_exporter"]
CMD ["--config.file=/app/config.yaml"]
