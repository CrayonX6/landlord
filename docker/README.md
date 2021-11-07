
### 环境安装步骤

* 使用 [`DaoCloud`](https://www.daocloud.io/mirror) 为 `Docker` 加速。

    > 注册后可获取加速 `ID` 用替换以下备用 `ID`。

    ```bash

    curl -sSL https://get.daocloud.io/daotools/set_mirror.sh | sh -s http://{ID}.m.daocloud.io

    # 例子
    curl -sSL https://get.daocloud.io/daotools/set_mirror.sh | sh -s http://8dd58468.m.daocloud.io
    sudo service docker restart
    ```

* 进入到相应项目目录。

    ```bash
    cd beginning/docker
    ```

* 编译 `Docker` 并启动。

    ```bash
    sudo docker-compose up --build
    ```

* 启动 `Docker` 并在后台运行。

    ```bash
    sudo docker-compose up -d
    ```

* 部分备用命令。

    ```
    # 删除所有的镜像
    sudo docker rmi $(sudo docker images -q)

    # 删除 `untagged` 镜像
    sudo docker rmi $(sudo docker images | grep "^<none>" | awk "{print $3}")

    # 删除所有的容器
    sudo docker rm $(sudo docker ps -a -q)

    # 启动/停止/重启指定容器（例如 beginning-redis 服务）
    sudo docker-compose start beginning-redis
    sudo docker-compose stop beginning-redis
    sudo docker-compose restart beginning-redis
    ```

* 安装提供的命令别名脚本。

    ```bash
    cd script
    chmod a+x *.sh
    # 逐一执行 ./install-xxx.sh 脚本
    source ~/.bash_profile
    ```
