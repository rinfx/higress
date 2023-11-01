set -e
set -x

make build-gateway-local
make build-istio-local
make docker-build

tag=$(docker images | grep higress/proxyv2 | head -n 1 | awk '{print $2}')
docker tag higress-registry.cn-hangzhou.cr.aliyuncs.com/higress/proxyv2:${tag} liuxr25/proxyv2:latest
docker tag higress-registry.cn-hangzhou.cr.aliyuncs.com/higress/proxyv2:${tag} liuxr25/gateway:latest
docker push liuxr25/proxyv2:latest
docker push liuxr25/gateway:latest

tag=$(docker images | grep higress/pilot | head -n 1 | awk '{print $2}')
docker tag higress-registry.cn-hangzhou.cr.aliyuncs.com/higress/pilot:${tag} liuxr25/pilot:latest
docker push liuxr25/pilot:latest

tag=$(docker images | grep higress/higress | head -n 1 | awk '{print $2}')
docker tag higress-registry.cn-hangzhou.cr.aliyuncs.com/higress/higress:${tag} liuxr25/higress:latest
docker push liuxr25/higress:latest

# make helm reinstall