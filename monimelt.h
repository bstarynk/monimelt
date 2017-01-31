// file monimelt.h -*- C++ -*-
#ifndef MOMIMELT_HEADER
#define MONIMELT_HEADER "monimelt.h"

#include <features.h>           // GNU things
#include <stdexcept>
#include <cstdint>
#include <climits>
#include <cmath>
#include <cstring>
#include <memory>
#include <algorithm>
#include <iostream>
#include <sstream>
#include <fstream>
#include <set>
#include <initializer_list>
#include <map>
#include <vector>
#include <deque>
#include <unordered_map>
#include <unordered_set>
#include <random>
#include <typeinfo>

// libbacktrace from GCC 6, i.e. libgcc-6-dev package
#include <backtrace.h>

#include <unistd.h>
#include <sys/syscall.h>
#include <pthread.h>
#include <sched.h>
#include <syslog.h>
#include <stdlib.h>
#include <dlfcn.h>

#include <utf8.h>

#include "jsoncpp/json/json.h"

// common prefix mom

// mark unlikely conditions to help optimization
#ifdef __GNUC__
#define MOM_UNLIKELY(P) __builtin_expect(!!(P),0)
#define MOM_LIKELY(P) !__builtin_expect(!(P),0)
#define MOM_UNUSED __attribute__((unused))
#define MOM_OPTIMIZEDFUN __attribute__((optimize("O2")))
#else
#define MOM_UNLIKELY(P) (P)
#define MOM_LIKELY(P) (P)
#define MOM_UNUSED
#define MOM_OPTIMIZEDFUN
#endif /*__GNUC__*/


// from generated _timestamp.c
extern "C" const char monimelt_timestamp[];
extern "C" const char monimelt_lastgitcommit[];
extern "C" const char monimelt_lastgittag[];
extern "C" const char*const monimelt_cxxsources[];
extern "C" const char*const monimelt_csources[];
extern "C" const char*const monimelt_shellsources[];
extern "C" const char monimelt_directory[];
extern "C" const char monimelt_statebase[];

#define MOM_PROGBINARY "monimelt"

/// the dlopen handle for the whole program
extern "C" void* mom_dlh;


static inline pid_t
mom_gettid (void)
{
  return syscall (SYS_gettid, 0L);
}

// time measurement, in seconds
// query a clock
static inline double
mom_clock_time (clockid_t cid)
{
  struct timespec ts = { 0, 0 };
  if (clock_gettime (cid, &ts))
    return NAN;
  else
    return (double) ts.tv_sec + 1.0e-9 * ts.tv_nsec;
}

static inline struct timespec
mom_timespec (double t)
{
  struct timespec ts = { 0, 0 };
  if (std::isnan (t) || t < 0.0)
    return ts;
  double fl = floor (t);
  ts.tv_sec = (time_t) fl;
  ts.tv_nsec = (long) ((t - fl) * 1.0e9);
  // this should not happen
  if (MOM_UNLIKELY (ts.tv_nsec < 0))
    ts.tv_nsec = 0;
  while (MOM_UNLIKELY (ts.tv_nsec >= 1000 * 1000 * 1000))
    {
      ts.tv_sec++;
      ts.tv_nsec -= 1000 * 1000 * 1000;
    };
  return ts;
}


extern "C" double mom_elapsed_real_time (void);    /* relative to start of program */
extern "C" double mom_process_cpu_time (void);
extern "C" double mom_thread_cpu_time (void);

// call strftime on ti, but replace .__ with centiseconds for ti
extern "C" char *mom_strftime_centi (char *buf, size_t len, const char *fmt, double ti)
__attribute__ ((format (strftime, 3, 0)));

#define MOM_EMPTY_SLOT ((void*)(2*sizeof(void*)))

extern "C" void mom_backtracestr_at (const char*fil, int lin, const std::string&msg);

#define MOM_BACKTRACELOG_AT(Fil,Lin,Log) do {   \
    std::ostringstream _out_##Lin;              \
    _out_##Lin << Log << std::flush;            \
    mom_backtracestr_at((Fil), (Lin),           \
      _out_##Lin.str());                        \
  } while(0)
#define MOM_BACKTRACELOG_AT_BIS(Fil,Lin,Log) \
  MOM_BACKTRACELOG_AT(Fil,Lin,Log)
#define MOM_BACKTRACELOG(Log) \
  MOM_BACKTRACELOG_AT_BIS(__FILE__,__LINE__,Log)

extern "C" void mom_abort(void) __attribute__((noreturn));
#ifndef NDEBUG
#define MOM_ASSERT_AT(Fil,Lin,Prop,Log) do {    \
 if (MOM_UNLIKELY(!(Prop))) {                   \
   MOM_BACKTRACELOG_AT(Fil,Lin,                 \
           "**MOM_ASSERT FAILED** " #Prop ":"   \
           " @ " <<__PRETTY_FUNCTION__          \
                       <<  std::endl            \
                       << "::" << Log);         \
   mom_abort();                                 \
 }                                              \
} while(0)
#else
#define MOM_ASSERT_AT(Fil,Lin,Prop,Log)  do {   \
    if (false && !(Prop))                       \
      MOM_BACKTRACELOG_AT(Fil,Lin,Log);         \
} while(0)
#endif  // NDEBUG
#define MOM_ASSERT_AT_BIS(Fil,Lin,Prop,Log) \
  MOM_ASSERT_AT(Fil,Lin,Prop,Log)
#define MOM_ASSERT(Prop,Log) \
  MOM_ASSERT_AT_BIS(__FILE__,__LINE__,Prop,Log)


extern "C" bool mom_verboseflag;
#define MOM_VERBOSELOG_AT(Fil,Lin,Log) do {     \
  if (mom_verboseflag)                          \
    std::clog << "*MOM @" << Fil << ":" << Lin  \
              << " /" << __FUNCTION__ << ": " \
              << Log << std::endl;              \
 } while(0)
#define MOM_VERBOSELOG_AT_BIS(Fil,Lin,Log) \
  MOM_VERBOSELOG_AT(Fil,Lin,Log)
#define MOM_VERBOSELOG(Log) \
  MOM_VERBOSELOG_AT_BIS(__FILE__,__LINE__,Log)

class MomSerial {
};				/* end class MomSerial */

#endif /*MONIMELT_HEADER*/
