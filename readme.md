1.启用容器:
docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -e COLLECTOR_OTLP_GRPC_HOST_PORT=:4317 \
  -e COLLECTOR_OTLP_HTTP_HOST_PORT=:4318 \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5775:5775/udp \
  -p 5778:5778/tcp \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 14250:14250 \
  docker.1ms.run/jaegertracing/all-in-one:latest

2.启动应用: go run main.go

3.访问: http://localhost:8080/hello?name=123&age=567

4.验证:http://localhost:16686/ 
    1.在Search中确认下有没有 gin-jaeger-demo 的service
    2.Operation 中过滤 /hello
    3.find traces
    4.能看到刚才的请求即成功
