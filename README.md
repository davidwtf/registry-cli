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
   registrycli repos 127.0.0.1:5000
   ```

### tags
### 列出所有 Tag, 返回：仓库名, Tag名, 适用平台, 资源大小, 创建时间, 资源类型, Digest

 | 参数 | 默认值 | 说明 |
 | - | - | - |
 | -o 或 --output | text | 输出格式，选项: json text |
 | --sort | tag | 排序方式，选项: tag size created |
 | --show-type | false | 以 text 格式输出时显示资源类型 |
 | --show-digest | false | 以 text 格式输出时显示 Digest |
 | --show-sum | true | 以 text 格式输出时，显示总量统计 |


* 示例:
   ```bash
   registrycli tags 127.0.0.1:5000/repo1
   ```

### inspect TAG_OR_DIGEST
### 根据 tag 或 digest 查看 manifest 详情，支持对 docker image 和 oci chart 做解析

 | 参数 | 默认值 | 说明 |
 | - | - | - |
 | -o 或 --output | text | 输出格式，选项：json text |

* 示例:
   ```bash
   registrycli inspect 127.0.0.1:5000/repo1:v1.0
   ```

### del TAG_OR_DIGEST
### 根据 tag 或 digest 删除 manifest

 | 参数 | 默认值 | 说明 |
 | - | - | - |
 | --untag | false | 仅删除 tag，不删除对应的 manifest |

 注: 按 tag 删除是 docker registry 在 3.0 中新增的功能。

* 示例:
   ```bash
   registrycli del 127.0.0.1:5000/repo1@sha256:74f5f150164eb49b3e6f621751a353dbfbc1dd114eb9b651ef8b1b4f5cc0c0d5
   registrycli del 127.0.0.1:5000/repo1@sha256:v1
   ```

### layer
### 下载 layer 内容


* 示例:
   ```bash
   registrycli layer 127.0.0.1:5000/repo1@sha256:275b2e73e3dc5cbf88c41ba15962045f0d36eeaf09dfe01f259ff2a12d3326af
   ```
