// file objvalmo/sqlitelog.go

package objvalmo  // import "github.com/bstarynk/monimelt/objvalmo"

import (
	_ "github.com/mattn/go-sqlite3"
)

/*
#cgo CFLAGS: -O1 -g
#cgo pkg-config: sqlite3
#include <stdlib.h>
#include <stdio.h>
#include <stdarg.h>
#include <sqlite3.h>


const char monimelt_sqlitelog_timestamp[]= __DATE__ "@" __TIME__;

void monimeltSqlLogStderr(void *udp __attribute__((unused)), int err, const char *msg)
{
    fprintf(stderr, "SQLITE LOG: err#%d: %s\n", err, msg);
    fflush(stderr);
}

void goSqlite3EnableLog(void) {
   sqlite3_config(SQLITE_CONFIG_LOG, monimeltSqlLogStderr, NULL);
   fprintf(stderr, "SQLITE3 LOG ENABLED\n");
   fflush(stderr);
}
*/
import "C"

func EnableSqliteLog() {
	C.goSqlite3EnableLog()
}
