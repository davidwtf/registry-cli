# registry-cli

`registrycli` 是 [Docker Registry](https://github.com/distribution/distribution) 的客户端工具。

提供如下功能：

* 列出所有仓库
* 列出仓库的标签
* 查看 Manifest 详情, 支持对 docker image 和 oci chart 做解析
* 删除 Manifest
* 下载 Blob

## 公共参数
 | 参数 | 默认值 | 说明 |
 | - | - | - |
 | -s 或 --server | | 镜像仓库服务器的地址 |
 | -u 或 --username | | 登录用户名 |
 | -p 或 --password | | 登录密码 |
 | --auth | | 使用认证auth登录，通常为 base64(username:password) |
 | --insecure | false | 使用不安全的 TLS 通信 |
 | --plain-http | false | 使用 HTTP 协议|
 | -h 或　--help | false | 查看帮助 |
 | -v 或　--version | false | 查看版本 |
 | --debug | false | 输出调试信息 |

* 注: 默认会加载 ~/.docker/config 中配置的认证信息，优先使用参数中的登录信息。

## 子命令

### repos
### 列出所有仓库
 | 参数 | 默认值 | 说明 |
 | - | - | - |
 | -o 或 --output | text | 输出格式，选项：json text |

* 示例:
   ```bash
   registrycli repos -s 127.0.0.1:5000
   ```

### tags
### 列出所有tag

 | 参数 | 默认值 | 说明 |
 | - | - | - |
 | -o 或 --output | text | 输出格式，选项：json text |
 | -r 或 --repository | | 仓库名 |
 | --all | false | 查询所有仓库名 |
 | --parellel | 20 | 并行查询的线程数 |

* 示例:
   ```bash
   registrycli tags -s 127.0.0.1:5000 -r repo1
   ```

### inspect TAG_OR_DIGEST
### 查看 tag 或 digest 详情，支持对 docker image 和 oci chart 做解析

 | 参数 | 默认值 | 说明 |
 | - | - | - |
 | -o 或 --output | text | 输出格式，选项：json text |
 | -r 或 --repository | | 仓库名 |

* 示例:
   ```bash
   registrycli inspect v1.0 -s 127.0.0.1:5000 -r repo1
   ```

### del TAG_OR_DIGEST
### 删除 tag 或 digest

 | 参数 | 默认值 | 说明 |
 | - | - | - |
 | -r 或 --repository | | 仓库名 |

 注: 按 tag 删除是 docker registry 在 3.0 中新增的功能。

* 示例:
   ```bash
   registrycli del sha256:74f5f150164eb49b3e6f621751a353dbfbc1dd114eb9b651ef8b1b4f5cc0c0d5 -s 127.0.0.1:5000 -r repo1
   ```

### blob BLOB_ID
### 下载 blob 内容

 | 参数 | 默认值 | 说明 |
 | - | - | - |
 | -r 或 --repository | | 仓库名 |

* 示例:
   ```bash
   registrycli　blob sha256:275b2e73e3dc5cbf88c41ba15962045f0d36eeaf09dfe01f259ff2a12d3326af -s 127.0.0.1:5000 -r repo1
   ```
