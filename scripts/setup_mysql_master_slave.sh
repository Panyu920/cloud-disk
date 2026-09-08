#!/bin/bash
set -e  # 遇到错误立即退出

# 参数定义：$1=密码, $2=主库主机名（默认 mysql-master）
PASSWORD=${1:-panyu}
MASTER_HOST=${2:-mysql-master}

echo "==> 开始设置 MySQL 主从复制..."
echo "==> 1. 在主库授予复制权限给 panyu 用户"
docker exec -i mysql-master mysql -uroot -p"$PASSWORD" -e "GRANT REPLICATION SLAVE ON *.* TO 'panyu'@'%'; FLUSH PRIVILEGES;"

echo "==> 2. 获取主库当前的二进制日志信息"
MASTER_INFO=$(docker exec -i mysql-master mysql -uroot -p"$PASSWORD" -e "SHOW MASTER STATUS\G" | grep -E 'File|Position' | awk '{print $2}' | tr '\n' ' ')
if [ -z "$MASTER_INFO" ]; then
    echo "错误：无法获取主库状态，请确保主库已启动且 binlog 已启用。"
    exit 1
fi
FILE=$(echo "$MASTER_INFO" | awk '{print $1}')
POS=$(echo "$MASTER_INFO" | awk '{print $2}')
echo "  主库当前文件: $FILE, 位置: $POS"

echo "==> 3. 在从库配置并启动复制..."
docker exec -i mysql-slave mysql -uroot -p"$PASSWORD" -e "
STOP SLAVE;
RESET SLAVE ALL;
CHANGE MASTER TO
  MASTER_HOST='$MASTER_HOST',
  MASTER_USER='panyu',
  MASTER_PASSWORD='$PASSWORD',
  MASTER_LOG_FILE='$FILE',
  MASTER_LOG_POS=$POS,
  GET_MASTER_PUBLIC_KEY=1;
START SLAVE;"

echo "==> 4. 验证复制状态"
docker exec -i mysql-slave mysql -uroot -p"$PASSWORD" -e "SHOW SLAVE STATUS\G" | grep -E 'Slave_IO_Running|Slave_SQL_Running|Last_IO_Error|Last_SQL_Error'

echo "==> 完成！请检查上述输出中的 Slave_IO_Running 和 Slave_SQL_Running 是否为 Yes。"