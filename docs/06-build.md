# 自行编译

当前源码编译依赖环境如下:

- macOS (Intel 或 Apple Silicon)
- Go 1.26+
- Xcode Command Line Tools（提供 `lipo`，用于合并 universal binary）

在环境配置完成后在本项目根目录下执行编译即可

``` sh
# 编译 darwin/amd64、darwin/arm64 并合并 universal，输出在 dist 目录
make

# 仅仅安装到 $(go env GOPATH)/bin 中
make install
```

编译产物如下:

- `dist/mmh_darwin_amd64`: Intel 芯片
- `dist/mmh_darwin_arm64`: Apple Silicon 芯片
- `dist/mmh_darwin_universal`: 同时支持两种芯片
- `dist/checksums.txt`: SHA-256 校验值

[首页](.) | [上一页](05-usage) | [下一页](07-example)
