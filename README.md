# registry-cli

公共参数
-u --username
-p --password
   --auth
-s --server
   --debug

默认加载 ~/.docker/config配置


默认加载　REGISTRY_SERVER 环境变量
# registrycli repos -s registry-url -n max -o json
REPOSITORY
abc/dxd
sdfa/ddd


# registrycli tags -s registry-url  -n max -r repo-name --tag-only --platform -o json -t type --all
REPOSITORY  TAG    DIGEST   TYPE   PLATFORM SIZE
v1
v2
v3

# registrycli get tag_or_digest -s registry-url -r repo-name --show-layers -o json
NAME    DIGEST   TYPE  PLATFORM  SIZE  LAYERS


# registrycli blob blob_id -s registry-url -r repo-name --destination --raw

# registrycli del tag_or_digest -s registry-url -r repo-name

