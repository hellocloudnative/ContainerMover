# `containerMover` 使用文档
`containerMover` 是一个用于在不同容器运行时之间迁移容器镜像的命令行工具。以下是如何使用 `containerMover` 的详细说明。
## 安装
在安装 `containerMover` 之前，请确保您的系统满足所有必要的依赖项。然后，您可以通过以下方式安装 `containerMover`：
```sh
# 假设您已经克隆了仓库并进入了项目目录
go build -o containerMover main.go
```
将生成的 `containerMover` 可执行文件移动到您的系统 `PATH` 中的适当位置。
## 使用方法
`containerMover` 命令的基本用法如下：
```sh
containerMover [命令] [选项] [参数]
```
### migrate 命令
`migrate` 命令用于迁移容器镜像。
#### images 子命令
`images` 子命令用于迁移指定的容器镜像。
```sh
containerMover migrate images [选项] [镜像名称...]
```
#### 选项
- `--src-type`: 源运行时类型（例如，docker）。
- `--dst-type`: 目标运行时类型（例如，containerd）。
- `--namespace`: 容器镜像所在的命名空间。
- `--all`: 迁移命名空间中的所有镜像。
- `--image-list`: 包含要迁移的镜像名称列表的文件，每行一个。
- `--hosts`: 以逗号分隔的远程主机地址列表。
- `--username`: 用于远程主机的用户名。
- `--password`: 用于远程主机的用户密码。
#### 示例
迁移 Docker 镜像到 Containerd：
```sh
containerMover migrate images --src-type docker --dst-type containerd --namespace A myimage:latest
```
迁移 Docker 镜像到 isulad：
```sh
containerMover migrate images --src-type docker --dst-type isulad --namespace B myimage:latest
```
迁移 Containerd 镜像到 Docker：
```sh
containerMover migrate images --src-type containerd --dst-type docker --namespace C myimage:latest
```
迁移命名空间中的所有镜像：
```sh
containerMover migrate images --src-type docker --dst-type containerd --namespace A --all
```
从文件中迁移 Docker 镜像列表到 Containerd：
```sh
containerMover migrate images --src-type docker --dst-type containerd --image-list imagelist.txt
```

迁移 本地Docker镜像到 远程 Containerd 主机：
```sh
以下是一个具体的 containerMover 命令示例，用于将两个 Docker 镜像 A 和 B 迁移到远程 Containerd 服务：

containerMover images --src-type docker --dst-type containerd --hosts 192.168.1.2,192.168.1.3 --username root --password 123456 A B
在这个例子中，A 和 B 是要迁移的 Docker 镜像的名称。这些镜像将被迁移到 IP 地址为 192.168.1.2 和 192.168.1.3 的远程服务器上，迁移后镜像的名称仍然是 A 和 B。

注意事项
请确保远程服务器的 Containerd 服务正在运行，并且可以访问。
出于安全考虑，避免在脚本或命令行中硬编码密码。考虑使用环境变量或密码管理工具。
在执行迁移之前，建议测试迁移过程，以确保一切按预期工作。
```

### 帮助信息
如果您需要查看任何命令的帮助信息，可以使用 `--help` 选项。
```sh
containerMover [命令] --help
```
这将显示该命令的详细帮助信息。
## 故障排除
如果您在使用 `containerMover` 时遇到任何问题，请查看日志文件或使用 `--debug` 选项来获取更详细的输出。
## 联系和支持
如果您需要帮助或有任何建议，请通过 [邮件列表](mailto:200922702@qq.com) 联系我们。
--- 
请注意，以上文档是根据提供的代码片段生成的，实际使用时可能需要根据具体情况进行调整。
