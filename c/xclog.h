// Copyright (c) 1999, Google Inc.
// Created: 2017-04-26
// Author: neverlee
// Email: listarmb@gmail.com

#ifndef XCLOG_H
#define XCLOG_H

#include <stdio.h>
#include <malloc.h>
#include <pthread.h>
#include <assert.h>

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

static const int minimum_bufsize = 512;

static const char severity_char[] = " FCEWNIVD";

static int initialized = 0;

static xclog_level difflevel = XC_WARN;
static xclog_level errlevel = XC_ERROR;
static xclog_level outlevel = XC_INFO;

static int log_line_bufsize = 512;
static char *log_buf = NULL;
static pthread_mutex_t output_lock;

void xclog_set_difflevel(xclog_level lv);
void xclog_set_errlevel(xclog_level lv);
void xclog_set_outlevel(xclog_level lv);

void xclog_initialize();
void xclog_initialize_with_args(int line_bufsize, xclog_level difflv, xclog_level errlv, xclog_level outlv);

void xclog_line(const char *fname, const int line, const xclog_level lv, const char *fmt, ...);
static void xclog_write(const char *fname, const int line, FILE *out, const xclog_level lv, const char remove_chinese, const char *fmt, va_list args);

typedef void (*__initalize_t)();
struct {
    void (*initialize)();
    void (*initialize_with_args)(int line_bufsize, xclog_level difflv, xclog_level errlv, xclog_level outlv);
    void (*set_difflevel)(xclog_level lv);
    void (*set_errlevel)(xclog_level lv);
    void (*set_outlevel)(xclog_level lv);
    void (*__xclog_raw)(const char *fname, const int line, const xclog_level lv, const char *fmt, ...);
} xclog = {
    xclog_initialize,
    xclog_initialize_with_args,
    xclog_set_difflevel,
    xclog_set_errlevel,
    xclog_set_outlevel,
    xclog_line,
};

// #define xclog(LEVEL, RC, FMT, ...) xclog_line(__FILE__, __LINE__, LEVEL, RC, FMT, ##__VA_ARGS__)

#define xcfatalf(FMT, ...)   __xclog_raw(__FILE__, __LINE__, XC_FATAL,   FMT, ##__VA_ARGS__)
#define xccritf(FMT, ...)    __xclog_raw(__FILE__, __LINE__, XC_CRIT,    FMT, ##__VA_ARGS__)
#define xcerrorf(FMT, ...)   __xclog_raw(__FILE__, __LINE__, XC_ERROR,   FMT, ##__VA_ARGS__)
#define xcwarnf(FMT, ...)    __xclog_raw(__FILE__, __LINE__, XC_WARN,    FMT, ##__VA_ARGS__)
#define xcnoticef(FMT, ...)  __xclog_raw(__FILE__, __LINE__, XC_NOTICE,  FMT, ##__VA_ARGS__)
#define xcinfof(FMT, ...)    __xclog_raw(__FILE__, __LINE__, XC_INFO,    FMT, ##__VA_ARGS__)
#define xcverbosef(FMT, ...) __xclog_raw(__FILE__, __LINE__, XC_VERBOSE, FMT, ##__VA_ARGS__)
#define xcdebugf(FMT, ...)   __xclog_raw(__FILE__, __LINE__, XC_DEBUG,   FMT, ##__VA_ARGS__)

void xclog_initialize() {
    log_buf = malloc(log_line_bufsize + 1);
    assert(log_buf != NULL);
    pthread_mutex_init(&output_lock, NULL);
    initialized = 1;
}

void xclog_initialize_with_args(int line_bufsize, xclog_level difflv, xclog_level errlv, xclog_level outlv) {
    assert(line_bufsize > minimum_bufsize);
    log_line_bufsize = line_bufsize;
    difflevel = difflv;
    errlevel = errlv;
    outlevel = outlv;
    xclog_initialize();
}

void xclog_set_difflevel(const xclog_level lv) {
    assert(initialized == 0);
    difflevel = lv;
}
void xclog_set_errlevel(const xclog_level lv) {
    assert(initialized == 0);
    errlevel = lv;
}
void xclog_set_outlevel(const xclog_level lv) {
    outlevel = lv;
    assert(initialized == 0);
}

void xclog_line(const char *fname, const int line, const xclog_level lv, const char remove_chinese, const char *fmt, ...) {
    assert(initialized == 1);
    FILE *out = NULL;
    if (lv <= difflevel && lv <= errlevel) {
        out = stderr;
    }
    if (lv > difflevel && lv <=outlevel) {
        out = stdout;
    }
    if (out && lv != XC_NONE) {
        va_list args;
        va_start(args, fmt);
        xclog_write(fname, line, out, lv, remove_chinese, fmt, args);
        va_end(args);
    }
    assert(lv != XC_FATAL);
}

void xclog_write(const char *fname, const int line, FILE *out, const xclog_level lv, const char remove_chinese, const char *fmt, va_list args) {
    struct tm tm_local;
    time_t t;
    t = time(NULL);
    localtime_r(&t, &tm_local);

    int prefix_len, remain_size;

    pthread_mutex_lock(&output_lock);
    prefix_len = sprintf(log_buf, "%c%02d%02d %02d:%02d:%02d %s:%d] "
            , severity_char[lv]
            , 1 + tm_local.tm_mon, tm_local.tm_mday
            , tm_local.tm_hour, tm_local.tm_min, tm_local.tm_sec
            , fname, line
            );

    remain_size = log_line_bufsize - prefix_len;

    size_t log_len = 0;
    int fmt_result;

    fmt_result = vsnprintf(log_buf+prefix_len, remain_size, fmt, args);

    if ((fmt_result > -1) && (fmt_result + prefix_len <= remain_size)) {
        log_len = fmt_result + prefix_len;
    } else {
        log_len = log_line_bufsize;
    }
    log_buf[log_line_bufsize] = 0;

    fputs(log_buf, out);
    fputc('\n', out);
    pthread_mutex_unlock(&output_lock);
}

#endif//XCLOG_H
