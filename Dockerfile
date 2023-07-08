FROM node:18-slim as node
WORKDIR /
COPY . .
RUN npm ci --only=production \
    && npm run minify \
    && sed -i '' server/static/**/*


FROM golang:1.20-bookworm as go
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Set up directory structures
WORKDIR /root/
RUN mkdir -p .
COPY . .
COPY --from=node ./server/static ./server/static

# App-specific setup
RUN make bins

FROM debian:bookworm-slim
LABEL maintainer="git@albertyw.com"
EXPOSE 5003
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl                                        `: Basic-packages` \
    && rm -rf /var/lib/apt/lists/*
HEALTHCHECK --interval=5s --timeout=3s CMD ./healthcheck.sh || exit 1

WORKDIR /root/
COPY --from=go /root/reaction-pics .
COPY --from=go /root/bin/healthcheck.sh .
RUN mkdir -p /root/logs/app

CMD ["./reaction-pics"]
