// file momain.cc
#include "monimelt.h"

#include <cxxabi.h>
#include <QProcess>
#include <QCoreApplication>
#include <QApplication>
#include <QCommandLineParser>
#include "sqlite3.h"

thread_local MomRandom MomRandom::_rand_thr_;
bool mom_verboseflag;
void* mom_dlh;

void mom_abort(void)
{
  fflush(NULL);
  abort();
} // end of mom_abort

char *
mom_strftime_centi (char *buf, size_t len, const char *fmt, double ti)
{
  struct tm tm;
  time_t tim = (time_t) ti;
  memset (&tm, 0, sizeof (tm));
  if (!buf || !fmt || !len)
    return NULL;
  strftime (buf, len, fmt, localtime_r (&tim, &tm));
  char *dotundund = strstr (buf, ".__");
  if (dotundund)
    {
      double ind = 0.0;
      double fra = modf (ti, &ind);
      char minibuf[16];
      memset (minibuf, 0, sizeof (minibuf));
      snprintf (minibuf, sizeof (minibuf), "%.02f", fra);
      strncpy (dotundund, strchr (minibuf, '.'), 3);
    }
  return buf;
} // end mom_strftime_centi

std::string
mom_demangled_typename(const std::type_info &ti)
{
  int dstat = -1;
  char*dnam = abi::__cxa_demangle(ti.name(), 0, 0, &dstat);
  if (dstat == 0 && dnam != nullptr)
    {
      std::string ns {dnam};
      free (dnam);
      return ns;
    }
  if (dnam) free(dnam), dnam=nullptr;
  return "??";
} // end mom_demangled_typename

/************************* backtrace *************************/

/* A callback function passed to the backtrace_full function.  */

#define MOM_MAX_CALLBACK_DEPTH 64
static int
mom_bt_callback (void *data, uintptr_t pc, const char *filename, int lineno,
                 const char *function)
{
  int *pcount = (int *) data;

  /* If we don't have any useful information, don't print
     anything.  */
  if (filename == NULL && function == NULL)
    return 0;

  /* Print up to MOM_MAX_CALLBACK_DEPTH functions.    */
  if (*pcount >= MOM_MAX_CALLBACK_DEPTH)
    {
      /* Returning a non-zero value stops the backtrace.  */
      fprintf (stderr, "...etc...\n");
      return 1;
    }
  ++*pcount;

  int demstatus = -1;
  char* demfun = abi::__cxa_demangle(function, nullptr, nullptr, &demstatus);
  if (demstatus != 0)
    {
      if (demfun)
        free(demfun);
      demfun = nullptr;
    };
  fprintf (stderr, "MoniMelt[0x%lx] %s\n\t%s:%d\n",
           (unsigned long) pc,
           demfun?demfun:(function == NULL ? "???" : function),
           filename == NULL ? "???" : filename, lineno);
  if (demfun)
    {
      free(demfun);
      demfun = nullptr;
    }
  return 0;
}                               /* end mom_bt_callback */

/* An error callback function passed to the backtrace_full function.  This is
   called if backtrace_full has an error.  */

static void
mom_bt_err_callback (void *data MOM_UNUSED, const char *msg, int errnum)
{
  if (errnum < 0)
    {
      /* This means that no debug info was available.  Just quietly
         skip printing backtrace info.  */
      return;
    }
  fprintf (stderr, "%s%s%s\n", msg, errnum == 0 ? "" : ": ",
           errnum == 0 ? "" : strerror (errnum));
}                               /* end mom_bt_err_callback */


void mom_backtracestr_at (const char*fil, int lin, const std::string&str)
{
  double nowti = mom_clock_time (CLOCK_REALTIME);
  char thrname[24];
  char buf[256];
  char timbuf[64];
  memset (buf, 0, sizeof (buf));
  memset (thrname, 0, sizeof (thrname));
  memset (timbuf, 0, sizeof (timbuf));
  pthread_getname_np (pthread_self (), thrname, sizeof (thrname) - 1);
  fflush (NULL);
  mom_strftime_centi (timbuf, sizeof(timbuf), "%Y-%b-%d %H:%M:%S.__ %Z", nowti);
  fprintf (stderr, "MONIMELT BACKTRACE @%s:%d <%s:%d> %s\n* %s\n",
           fil, lin, thrname, (int) mom_gettid (), timbuf, str.c_str());
  fflush (NULL);
  struct backtrace_state *btstate =
    backtrace_create_state (NULL, 0, mom_bt_err_callback, NULL);
  if (btstate != NULL)
    {
      int count = 0;
      backtrace_full (btstate, 1, mom_bt_callback, mom_bt_err_callback,
                      (void *) &count);
    }
} // end of mom_backtracestr_at



static struct timespec start_realtime_ts_mom;

static void
check_updated_binary_mom(void)
{
  // should run make -C monimelt_directory -q MOM_PROGBINARY
  QProcess makeproc;
  QStringList makeargs;
  makeargs << "-C" << monimelt_directory << "-q" << MOM_PROGBINARY;
  makeproc.start("make",makeargs);
  makeproc.waitForFinished(-1);
  if (makeproc.exitStatus() != QProcess::NormalExit || makeproc.exitCode() != 0)
    {
      MOM_BACKTRACELOG("check_updated_binary binary  " << MOM_PROGBINARY << " in " << monimelt_directory << " is obsolete");
      exit(EXIT_FAILURE);
    }
} // end check_updated_binary_mom

static void show_size_mom(void)
{
  printf("sizeof intptr_t : %zd (align %zd)\n",
         sizeof(intptr_t), alignof(intptr_t));
} // end show_size_mom





// for SQLITE_CONFIG_LOG
static void
mom_sqlite_errorlog (void *pdata MOM_UNUSED, int errcode, const char *msg)
{
  MOM_BACKTRACELOG("Sqlite Error errcode="<< errcode << " msg=" << msg);
} // end mom_sqlite_errorlog



int
main (int argc_main, char **argv_main)
{
  clock_gettime (CLOCK_REALTIME, &start_realtime_ts_mom);
  check_updated_binary_mom();
  mom_dlh = dlopen(nullptr, RTLD_NOW|RTLD_GLOBAL);
  if (!mom_dlh)
    {
      fprintf(stderr, "%s failed to dlopen main program (%s)\n",
              argv_main[0], dlerror());
      exit(EXIT_FAILURE);
    }
  bool nogui = false;
  for (int ix=1; ix<argc_main; ix++)
    {
      if (!strcmp("--no-gui", argv_main[ix]) || !strcmp("-N", argv_main[ix]) || !strcmp("--batch", argv_main[ix]))
        nogui = true;
      if (!strcmp("-V", argv_main[ix]) || !strcmp("--verbose",argv_main[ix]))
        mom_verboseflag = true;
    }
  sqlite3_config (SQLITE_CONFIG_LOG, mom_sqlite_errorlog, NULL);
  {
    unsigned bn = getpid() % MomSerial63::_maxbucket_;
    auto s = MomSerial63::make_random_of_bucket(bn);
    auto s2 = MomSerial63::make_random_of_bucket(bn);
    auto s3 = MomSerial63::make_random_of_bucket(2*MomSerial63::_maxbucket_/3 + getpid() % (MomSerial63::_maxbucket_/4));
    MOM_ASSERT(s.bucketnum() == bn, "corrupted bucketnum");
    MOM_VERBOSELOG("s=" << s << " s2=" << s2 << "=="
                   << MomSerial63::make_from_cstr(s2.to_string().c_str())
		   << " s3=" << s3 << "=="
		   << MomSerial63::make_from_cstr(s3.to_string().c_str()));
    auto t1 = MomSerial63::make_random();
    auto t2 = MomSerial63::make_random();
    auto t3 = MomSerial63::make_random();
    MOM_VERBOSELOG("t1=" << t1 << " t2=" << t2 << " t3=" << t3);
    auto id1 = MomObject::random_id();
    auto id2 = MomObject::random_id();
    auto id3 = MomObject::random_id_of_bucket(bn);
    MOM_VERBOSELOG("id1=" << id1 << ";h=" << MomObject::hash_id(id1)
		   << " id2=" << id2 << " id3=" << id3);
    MOM_VERBOSELOG("id1==" << MomObject::id_from_cstr(MomObject::id_to_string(id1).c_str()));
    MOM_VERBOSELOG("id2==" << MomObject::id_from_cstr(MomObject::id_to_string(id2).c_str()) << ";h=" << MomObject::hash_id(id2));
  }
} // end main

double
mom_elapsed_real_time (void)
{
  struct timespec curts = { 0, 0 };
  clock_gettime (CLOCK_REALTIME, &curts);
  return 1.0 * (curts.tv_sec - start_realtime_ts_mom.tv_sec)
         + 1.0e-9 * (curts.tv_nsec - start_realtime_ts_mom.tv_nsec);
} // end mom_elapsed_real_time

double
mom_process_cpu_time (void)
{
  struct timespec curts = { 0, 0 };
  clock_gettime (CLOCK_PROCESS_CPUTIME_ID, &curts);
  return 1.0 * (curts.tv_sec) + 1.0e-9 * (curts.tv_nsec);
} // end mom_process_cpu_time

double
mom_thread_cpu_time (void)
{
  struct timespec curts = { 0, 0 };
  clock_gettime (CLOCK_THREAD_CPUTIME_ID, &curts);
  return 1.0 * (curts.tv_sec) + 1.0e-9 * (curts.tv_nsec);
} // end mom_thread_cpu_time


const MomSerial63
MomSerial63::make_random(void)
{
  uint64_t s = 0;
  do
    {
      s = MomRandom::random_64u() & (((uint64_t)1<<63)-1);
    }
  while (s<=_minserial_ || s>=_maxserial_);
  return MomSerial63{s};
} // end MomSerial63::make_random


const MomSerial63
MomSerial63::make_random_of_bucket(unsigned bucknum)
{
  if (MOM_UNLIKELY(bucknum >= _maxbucket_))
    {
      MOM_BACKTRACELOG("MomSerial63::random_of_bucket too big bucknum="
                       << bucknum);
      throw std::runtime_error("random_of_bucket too big bucknum");
    }
  uint64_t ds = MomRandom::random_64u() % (_deltaserial_ / _maxbucket_);
  uint64_t s = (bucknum * (_deltaserial_ / _maxbucket_)) + ds + _minserial_;
  MOM_ASSERT(s>=_minserial_ && s<=_maxserial_,
             "good s=" << s << " between _minserial_=" << _minserial_
             << " and _maxserial_=" << _maxserial_
             << " with ds=" << ds << " and bucknum=" << bucknum
             << " and _deltaserial_=" << _deltaserial_
             << " and _maxbucket_=" << _maxbucket_);
  MOM_DO_NOT_LOG("ds=" << ds << " bucknum=" << bucknum
                 << " _deltaserial_=" << _deltaserial_
                 << " _maxbucket_=" << _maxbucket_
                 << " _minserial_=" << _minserial_
                 << " _maxserial_=" << _maxserial_
                 << " s=" << s);
  return MomSerial63{s};
} // end of MomSerial63::make_random_of_bucket

//constexpr const char MomSerial63::_b62digits_[] = MOM_B62DIGITS;
std::string MomSerial63::to_string(void) const
{
  static_assert(sizeof(MOM_B62DIGITS)==_base_+1, "bad MOM_B62DIGITS in MomSerial63");
  char buf[24];
  static_assert(sizeof(buf)>_nbdigits_, "too small buf");
  memset (buf, 0, sizeof(buf));
  buf[0] = '_';
  memset (buf+1, '0', _nbdigits_);
  uint64_t n = _serial;
  char*pc = buf+_nbdigits_;
  while (n != 0)
    {
      unsigned d = n % _base_;
      n = n / _base_;
      *pc = _b62digstr_[d];
      pc--;
    }
  MOM_ASSERT(pc>=buf, "to_string bad pc - buffer underflow");
  MOM_ASSERT(strlen(buf) == _nbdigits_ + 1, "to_string bad buf=" << buf);
  return std::string{buf};
} // end  MomSerial63::to_string


const MomSerial63
MomSerial63::make_from_cstr(const char*s, const char*&end, bool fail)
{
  uint64_t n = 0;
  if (!s)
    goto failure;
  if (s[0] != '_')
    goto failure;
  if (!isdigit(s[1]))
    goto failure;
  for (auto i=0U; i<_nbdigits_; i++)
    {
      if (!s[i+1])
        goto failure;
      auto p = strchr(_b62digstr_,s[i+1]);
      if (!p)
        goto failure;
      n = n*_base_ + (p-_b62digstr_);
    }
  if (n!=0 && n<_minserial_)
    goto failure;
  if (n>_maxserial_)
    goto failure;
  end = s+_nbdigits_+1;
  return MomSerial63{n};
failure:
  if (fail)
    {
      std::string str{s};
      if (str.size() > _nbdigits_+2)
        str.resize(_nbdigits_+2);
      MOM_BACKTRACELOG("make_from_cstr failure str=" << str);
      throw std::runtime_error("MomSerial63::make_from_cstr failure");
    }
  end = s;
  return MomSerial63{nullptr};
} // end MomSerial63::make_from_cstr



void
MomUtf8Out::out(std::ostream&os) const
{
  uint32_t uc = 0;
  auto it = _str.begin();
  auto end = _str.end();
  while ((uc=utf8::next(it, end)) != 0)
    {
      switch (uc)
        {
        case 0:
          os << "\\0";
          break;
        case '\"':
          os << "\\\"";
          break;
        case '\\':
          os << "\\\\";
          break;
        case '\a':
          os << "\\a";
          break;
        case '\b':
          os << "\\b";
          break;
        case '\f':
          os << "\\f";
          break;
        case '\n':
          os << "\\n";
          break;
        case '\r':
          os << "\\r";
          break;
        case '\t':
          os << "\\t";
          break;
        case '\v':
          os << "\\v";
          break;
        case '\033':
          os << "\\e";
          break;
        default:
          if (uc<127 && ::isprint((char)uc))
            os << (char)uc;
          else if (uc<=0xffff)
            {
              char buf[8];
              snprintf(buf, sizeof(buf), "\\u%04x", (int)uc);
              os << buf;
            }
          else
            {
              char buf[16];
              snprintf(buf, sizeof(buf), "\\U%08x", (int)uc);
              os << buf;
            }
        }
    }
} // end of MomUtf8Out::out

