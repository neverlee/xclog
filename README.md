# xclog
一个简单的日志库

## 设计目标构想
1. 包括stdout和stderr输出（有利于使用supervisor接管日志）
2. 尽量简单高效

## 日志模块设计
| 日志级别 | 描述                                |
|----------|-------------------------------------|
| FATAL    | 致命错误（打印此日志后程序会退出）  |
| CRIT     | 严重错误                            |
| ERR      | 错误                                |
| WARN     | 警告                                |
| NOTICE   | 提示                                |
| INFO     | 信息                                |
| VERBOSE  | 详细信息                            |
| DEBUG    | 调度信息                            |

## 日志级别配置模式
1. 标准输出与错误分位数（默认为WARN），小于此数的日志使用标准错误打印，大于此数的使用标准输出打印
2. 标准错误级别（默认为ERR），小于等于此数字的日志才会被打印到标准错误
3. 标准输出级别（默认为INFO），小于等于此数字的日志才会被打印到标准错误

注：默认配置下，标准输出会打印NOTICE,INFO日志，标准错误会打印FATAL,CRIT,ERR日志

## 日志格式
采用类似glog形式的日志格式，原glog日志格式及示例如下
```
Lmmdd hh:mm:ss.uuuuuu threadid file:line] msg...
E0425 15:50:26.308867 25487 main.go:10] hello world
L为日志级别的第一个字母
mmdd为月日
hh:mm:ss.uuuuuu为时间，并精确到微秒
threadid为记录此日志的线程id
file:line为当前所在的代码文件名及行号
msg...为日志消息内容
```
这里将其简化，时间只精确到秒，并去掉线程id。格式如下
```
Lmmdd hh:mm:ss file:line] msg...
E0425 15:50:26 main.go:10] hello world
```

## 接口
### C语言接口
```
typedef enum {
    XC_NONE=0,
    XC_FATAL,
    XC_CRIT,
    XC_ERROR,
    XC_WARN,
    XC_NOTICE,
    XC_INFO,
    XC_VERBOSE,
    XC_DEBUG,
} xclog_level;
// 初始化及设置接口
void xclog.initialize();
void xclog.initialize_with_args(int line_bufsize, xclog_level difflv, xclog_level errlv, xclog_level outlv);
void xclog.set_difflevel(xclog_level lv);
void xclog.set_errlevel(xclog_level lv);
void xclog.set_outlevel(xclog_level lv);
// 日志输出接口
void xcfatalf(FMT, ...)
void xccritf(FMT, ...)
void xcerrorf(FMT, ...)
void xcwarnf(FMT, ...)
void xcnoticef(FMT, ...)
void xcinfof(FMT, ...)
void xcverbosef(FMT, ...)
void xcdebugf(FMT, ...)
```

## 当前完成状态
C语言版本日志功能已经完成

## 之后的短期计划
下周内完成PHP和Go版本的日志模块

## 长期考虑计划
1. 目前fatal采用assert进行程序退出，考虑改成退出时打印栈信息
2. 目前使用锁保证线程安全，考虑支持使用无锁队列异步
3. 考虑开发C++版本，支持stream流模式日志输出
4. 考虑做tee文件功能
