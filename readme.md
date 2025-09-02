# README

运行需要管理员模式

采集数据需要游戏处于1080p分辨率且界面位于交易中心

数据文件存于用户根目录下xhgm_prices.json，可以自行去/good/templates/items.json修改对应的产出和配方

## 依赖
[golang](https://go.dev/)

[wails](https://wails.io/zh-Hans/docs/gettingstarted/installation)

## 代码运行

go get github.com/go-vgo/robotgo (首次运行)

go mod tidy (首次运行)


wails dev

## 构建命令
### windows
wails build -clean -o srpt.exe

### 交叉平台
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ wails build -clean -o srpt.exe -platform windows/amd64
