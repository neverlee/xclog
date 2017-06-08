#include "xclog.h"

int main() {
    xclog.initialize();
    xclog.xcerrorf("err %s %s", "dd", "中文");
    xclog.xcwarnf("warn %s", "dd");
    xclog.xcinfof("info %s %s", "dd", "中文");
    xclog.xcdebugf("debug %s", "dd");
    return 0;
}
