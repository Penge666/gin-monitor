# 应用示例

gin-metrics 是一个 Gin Web 框架的中间件，无缝集成 Prometheus，开箱即用地收集并暴露应用性能指标，帮助开发者高效监控应用运行状况。

## 文件目录

```bash
penge@DESKTOP-B0LM74I MINGW64 /g/GoProject/test
$ ls
ReadMe/  ReadMe.md  docker-compose.yml  go.mod  go.sum  main.go  prometheus.yml
```

`prometheus.yml`

- **`static_configs`**: 这部分定义了静态配置，指定了需要抓取的目标地址。在这里，它使用了`targets`来指定Prometheus抓取的目标服务的地址。
  - **`targets: ['10.195.52.140:8080']`**: 这是一个目标列表，表示Prometheus会抓取IP地址为`10.195.52.140`，端口为`8080`的服务。这个服务应该暴露出Prometheus可以读取的监控数据，例如使用`webserver_exporter`（或类似的exporter）来提供Web服务器的指标。

## 运行步骤

运行命令：

```bash
docker-compose up -d
```

![image-20250321095559410](ReadMe/image-20250321095559410.png)

Note:要将这两个中间件放在同一个网桥里面，要不然网络无法通信。



>  Step1：先添加数据源

http://localhost:3000/

![image-20250321095205035](ReadMe/image-20250321095205035.png)

![image-20250321095235524](ReadMe/image-20250321095235524.png)

>  Step2：导入模板内容

![image-20250321095256171](ReadMe/image-20250321095256171.png)

这样，就可以通过granafa导入数据进行展示了

![image-20250321095113041](ReadMe/image-20250321095113041.png)

同时也可以通过prometheus：http://localhost:9090/ 对其进行访问

![image-20250321100626295](ReadMe/image-20250321100626295.png)