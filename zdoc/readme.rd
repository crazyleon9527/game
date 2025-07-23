服务器类型：
用于运行 rk-api 的服务器可以选择计算优化的实例，比如 AWS 的 c5.xlarge，以满足处理大量请求的需要。
对于 rk-admin 和 rk-front 所在的服务器，
由于 rk-front 是承载静态资源，rk-admin 流量不大，可以选择一台标准的服务器即可，例如 AWS 的 t3.medium 或 t3.large。
如果你预计 rk-front 会有大量的流量，那么可能需要选择一个内存优化的实例，比如 AWS 的 r5.large