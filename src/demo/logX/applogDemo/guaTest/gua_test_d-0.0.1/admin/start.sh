#!/bin/bash
# this file is generated automatically by /data/pkg/public-scripts/bin/functions.sh at 2017-12-11 18:40:48
#
# please don't edit.
#
# any problem , please contact linbobo (linbobo@yy.com)
#
#
set -x
set +H
shopt -s expand_aliases
export LANG=C
export LC_ALL=C
alias sudo='sudo -b'
source /data/pkg/public-scripts/bin/mypidof.sh
source /data/services/gua_test_d-0.0.1/admin/common-var.conf 
source /data/pkg/public-scripts/bin/limit-run.sh

# ！！！请根据具体需要更改该脚本，此处只是默认脚本，非 100% 通用 ！！！！

# 该处输入内容将保存为 /data/services/$PKG_NAME-$VERSION/admin/start.sh

# 如果你是使用 root 用户启动，则无需使用 sudo ，脚本默认以 root 身份运行 ！！！！

# 如果你是使用 user_00 用户启动，请注意去掉 sudo ，否则可能导致脚本运行失败 ！！！！

# 升级时调用当前内容启动新进程

# ！！！注意！！！

# 由于系统自带 pidof 命令在 cron 中会随机 coredump ，这里的 pidof 是自定义函数，

# 仅支持 pidof <name> 的用法 ！！！！

# 最后修改日期 : 2013-05-24 16:40

#---------------#
# 进程数量设置  #
#---------------#

count=1

#----------------#
# 初始化日志文件 #
#----------------#

log=$INSTALL_PATH/admin/start.log

true > $log

#-------------#
# 进程数检查  #
#-------------#

x=$(pidof $APP_NAME |wc -w)
y=$((count-x))
echo "delta=$y"

#--------------------------#
# 进程数大于 $count 就退出 #
#--------------------------#

if [ $y -le 0 ] ; then
   pidof $APP_NAME | xargs -r ps -lf
   echo "$APP_NAME num ($x) >= $count , no need to start , quit"
   exit 0
fi

#---------------#
# 启动进程      #
#---------------#

cd $INSTALL_PATH/bin || exit 1
#args=`cat $INSTALL_PATH/conf/args.ini`
for ((i=1;i<=$y;i++)); do
     echo "start #$i"
     nohup ./$APP_NAME >>$log 2>&1 &
     #nohup ./$APP_NAME ${args} >>$log 2>&1 &
     sleep 2
done

#---------------#
# 二次确认      #
#---------------#

if [ $(pidof $APP_NAME |wc -w) -eq $count ] ; then
     echo "start $APP_NAME ok"
     echo "output last 20 lines of $log"
     tail -n 20 $log
     echo "output last 20 lines of /data/yy/log/$APP_NAME/${APP_NAME}.log"
     tail -n 20 /data/yy/log/$APP_NAME/${APP_NAME}.log
     pidof $APP_NAME |xargs -r ps -lf
     exit 0
else
     echo "start $APP_NAME failed"
     echo "output last 20 lines of $log"
     tail -n 20 $log
     echo "output last 20 lines of /data/yy/log/$APP_NAME/${APP_NAME}.log"
     tail -n 20 /data/yy/log/$APP_NAME/${APP_NAME}.log
     pidof $APP_NAME |xargs -r ps -lf
     exit 1
fi
