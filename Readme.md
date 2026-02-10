# 实现功能
- [x] 监控工厂产物的产能及其AE库存的产物数量
- [ ] 监控AE2网络的能源情况
- [ ] 监控AE2网络的库存情况

# 使用方法
## 前端
```bash
cd mineCCT-web
npm install
npm run dev
```
## 后端
```bash
go run main.go
```
## mc端
更简单的方式是只在第一次安装时下载启动器，之后它会在每次启动时自动更新脚本，无需重复下载。

在主计算机输入
```bash
wget http://127.0.0.1:8080/lua/startup.lua startup/startup.lua
```
在海龟端输入
```bash
wget http://127.0.0.1:8080/lua/startup_meter.lua startup/startup_meter.lua
```
然后重启计算机/海龟即可
