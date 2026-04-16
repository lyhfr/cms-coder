FROM artifacts.cmschina.com.cn/public-docker-rdm-remote/golang:1.26.2-bookworm
ENV ARTIFACTSDOMAIN http://172.253.70.26
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
# 默认上下文目录 . 应为项目根目录
COPY . /app
RUN go env -w GO111MODULE=on && go env -w GOPROXY=${ARTIFACTSDOMAIN}/artifactory/api/go/public-go-rdm-virtual && cd /app && go mod tidy && go build main.go
WORKDIR /app
RUN chmod +x main
# 运行时需要挂载config.toml文件到/app/config/config.toml
CMD ./main
