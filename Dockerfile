FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    ca-certificates \
    golang \
    mysql-client  \
    curl
    
# 証明書関連の設定をコピー
COPY certs /etc/ssl/certs
COPY 04.go go.mod go.sum /apiserver/

#image内での作業ディレクトリの指定
WORKDIR /apiserver


# ソースコードをビルド
RUN go build 04.go

# ポートの公開
EXPOSE 8080

CMD ["/apiserver/04", "/bin/bash"]
