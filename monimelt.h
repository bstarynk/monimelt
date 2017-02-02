// file monimelt.h       -*- C++ -*-
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


#define MOM_NEVERLOG_AT(Fil,Lin,Log) do { \
  if (false && mom_verboseflag)     \
    std::clog << "*MOM @" << Fil << ":" << Lin  \
              << " /" << __FUNCTION__ << ": " \
              << Log << std::endl;              \
 } while(0)
#define MOM_NEVERLOG_AT_BIS(Fil,Lin,Log) \
  MOM_NEVERLOG_AT(Fil,Lin,Log)
#define MOM_NEVERLOG(Log) \
  MOM_NEVERLOG_AT_BIS(__FILE__,__LINE__,Log)

// MOM_DO_NOT_LOG has the same length in characters as MOM_VERBOSELOG
#define MOM_DO_NOT_LOG(Log) MOM_NEVERLOG(Log)
//      MOM_VERBOSELOG has the same width


class MomRandom
{
  static thread_local MomRandom _rand_thr_;
  unsigned long _rand_count;
  std::mt19937 _rand_generator;
  uint32_t generate_32u(void)
  {
    if (MOM_UNLIKELY(_rand_count++ % 4096 == 0))
      {
        std::random_device randev;
        auto s1=randev(), s2=randev(), s3=randev(), s4=randev(),
             s5=randev(), s6=randev(), s7=randev();
        std::seed_seq seq {s1,s2,s3,s4,s5,s6,s7};
        _rand_generator.seed(seq);
      }
    return _rand_generator();
  };
  uint32_t generate_nonzero_32u(void)
  {
    uint32_t r = 0;
    do
      {
        r = generate_32u();
      }
    while (MOM_UNLIKELY(r==0));
    return r;
  };
  uint64_t generate_64u(void)
  {
    return (static_cast<uint64_t>(generate_32u())<<32) | static_cast<uint64_t>(generate_32u());
  };
public:
  static uint32_t random_32u(void)
  {
    return _rand_thr_.generate_32u();
  };
  static uint64_t random_64u(void)
  {
    return _rand_thr_.generate_64u();
  };
  static uint32_t random_nonzero_32u(void)
  {
    return _rand_thr_.generate_nonzero_32u();
  };
};        // end class MomRandom

#define MOM_B62DIGITS \
    "0123456789" "abcdefghijklmnopqrstuvwxyz" "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
class MomSerial63
{
  uint64_t _serial;
public:
  static constexpr const uint64_t _minserial_ = 1024;
  static constexpr const uint64_t _maxserial_ =
    (uint64_t)10 * 62*62*62 * 62*62*62 * 62*62*62 * 62;
  static constexpr const uint64_t _deltaserial_ =
    _maxserial_ - _minserial_;
  static constexpr const char* _b62digstr_ = MOM_B62DIGITS;
  static constexpr unsigned _nbdigits_ = 11;
  static constexpr unsigned _base_ = 62;
  static_assert(_maxserial_ < ((uint64_t)1<<63),
                "corrupted _maxserial_ in MomSerial63");
  static_assert(_deltaserial_ > ((uint64_t)1<<62),
                "corrupted _deltaserial_ in MomSerial63");
  static constexpr const unsigned _maxbucket_ = 10*62;
  inline MomSerial63(uint64_t n=0, bool nocheck=false);
  MomSerial63(std::nullptr_t) : _serial(0) {};
  ~MomSerial63()
  {
    _serial=0;
  };
  uint64_t serial() const
  {
    return _serial;
  };
  unsigned bucketnum() const
  {
    return _serial / (_deltaserial_ / _maxbucket_);
  };
  uint64_t buckoffset() const
  {
    return _serial % (_deltaserial_ / _maxbucket_);
  };
  std::string to_string(void) const;
  static const MomSerial63 make_from_cstr(const char*s, const char*&end, bool fail=false);
  static const MomSerial63 make_from_cstr(const char*s, bool fail=false)
  {
    const char*end=nullptr;
    return make_from_cstr(s,end,fail);
  };
  static const MomSerial63 make_random(void);
  static const MomSerial63 make_random_of_bucket(unsigned bun);
  MomSerial63(const MomSerial63&s) : _serial(s._serial) {};
  MomSerial63(MomSerial63&& s) : _serial(std::move(s._serial)) { };
  operator bool () const
  {
    return _serial != 0;
  };
  bool operator ! () const
  {
    return _serial == 0;
  };
  bool equal(const MomSerial63 r) const
  {
    return _serial == r._serial;
  };
  bool less(const MomSerial63 r) const
  {
    return _serial < r._serial;
  };
  bool less_equal(const MomSerial63 r) const
  {
    return _serial <= r._serial;
  };
  bool operator == (const MomSerial63 r) const
  {
    return equal(r);
  };
  bool operator != (const MomSerial63 r) const
  {
    return !equal(r);
  };
  bool operator < (const MomSerial63 r) const
  {
    return less(r);
  };
  bool operator <= (const MomSerial63 r) const
  {
    return less_equal(r);
  };
  bool operator > (const MomSerial63 r) const
  {
    return !less_equal(r);
  };
  bool operator >= (const MomSerial63 r) const
  {
    return !less(r);
  };
};        /* end class MomSerial63 */


typedef uint32_t MomHash_t;
typedef Json::Value MomJson;

//////////////// to ease debugging
class MomOut
{
  std::function<void(std::ostream&)> _fn_out;
public:
  MomOut(std::function<void(std::ostream&)> fout): _fn_out(fout) {};
  ~MomOut() = default;
  void out(std::ostream&os) const
  {
    _fn_out(os);
  };
};
inline std::ostream& operator << (std::ostream& os, const MomOut& bo)
{
  bo.out(os);
  return os;
};


class MomUtf8Out
{
  std::string _str;
  unsigned _flags;
public:
  MomUtf8Out(const std::string&str, unsigned flags=0) : _str(str), _flags(flags)
  {
    if (!utf8::is_valid(str.begin(), str.end()))
      {
        MOM_BACKTRACELOG("MomUtf8Out invalid str=" << str);
        throw std::runtime_error("MomUtf8Out invalid string");
      }
  };
  ~MomUtf8Out()
  {
    _str.clear();
    _flags=0;
  };
  MomUtf8Out(const MomUtf8Out&) = default;
  MomUtf8Out(MomUtf8Out&&) = default;
  void out(std::ostream&os) const;
};        // end class MomUtf8Out

inline std::ostream& operator << (std::ostream& os, const MomUtf8Out& bo)
{
  bo.out(os);
  return os;
};


class MomObject;
class MomVal;
class MomString;
class MomSequence;
class MomSet;
class MomTuple;


#define MOM_SIZE_MAX (INT32_MAX/3)
enum class MomVKind : std::uint8_t
{
  NoneK,
  IntK,
  /* we probably dont need doubles at first. But we want to avoid NaNs if we need them; we would use nil instead of boxed NaN */
  // DoubleK,
  StringK,
  RefobjK,
  SetK,
  TupleK,
  /* we don't need mix (of scalar values, e.g. ints, doubles, strings, objects) at first */
  // MixK,
};

class MomRefobj
{
  MomObject* _ptrobj;
public:
  MomRefobj(MomObject& ob) : _ptrobj(&ob) {};
  MomRefobj(MomObject* pob=nullptr) : _ptrobj(pob) {};
  ~MomRefobj()
  {
    _ptrobj = nullptr;
  };
  MomRefobj(const MomRefobj& ro) : _ptrobj(ro._ptrobj) {};
  MomRefobj(MomRefobj&& mo): _ptrobj(std::move(mo._ptrobj)) {};
  MomRefobj& operator = (const MomRefobj& ro)
  {
    _ptrobj = ro._ptrobj;
    return *this;
  };
  MomRefobj& operator = (MomRefobj&&mo)
  {
    _ptrobj = std::move(mo._ptrobj);
    return *this;
  };
  MomObject* get_const(void) const
  {
    if (!_ptrobj)
      {
        MOM_BACKTRACELOG("MomRefobj::get_const nil pointer @" << (void*)this);
        throw std::runtime_error("MomRefobj::get_const nil dereference");
      }
    return _ptrobj;
  };
  MomObject* get(void)
  {
    if (!_ptrobj)
      {
        MOM_BACKTRACELOG("MomRefobj::get nil pointer @" << (void*)this);
        throw std::runtime_error("MomRefobj::get nil dereference");
      }
    return _ptrobj;
  }
  MomObject* unsafe_get(void)
  {
    return _ptrobj;
  };
  MomObject* unsafe_get_const(void) const
  {
    return _ptrobj;
  };
  operator MomObject* () const
  {
    return get_const();
  };
  MomObject* operator * (void) const
  {
    return get_const();
  };
  MomObject* operator * (void)
  {
    return get();
  };
  MomObject* operator -> (void) const
  {
    return get_const();
  };
  MomObject* operator -> (void)
  {
    return get();
  };
  MomRefobj& unsafe_put(MomObject*pob)
  {
    _ptrobj = pob;
    return *this;
  };
  MomRefobj& put_non_nil(MomObject*pob)
  {
    if (pob==nullptr)
      {
        MOM_BACKTRACELOG("MomRefobj::put_non_nil got nil pointer @" << (void*)this);
        throw std::runtime_error("MomRefobj::put_non_nil with nil pointer");
      }
    return *this;
  }
  MomRefobj& operator = (MomObject*pob)
  {
    return put_non_nil(pob);
  };
  MomRefobj& operator = (std::nullptr_t)
  {
    _ptrobj=nullptr;
    return *this;
  };
  MomRefobj& clear(void)
  {
    _ptrobj = nullptr;
    return *this;
  };
  inline MomHash_t hash(void) const;
  inline bool less(const MomRefobj) const;
  inline bool less_equal(const MomRefobj) const;
  inline bool equal(const MomRefobj) const;
  bool operator < (const MomRefobj r) const
  {
    return less(r);
  };
  bool operator <= (const MomRefobj r) const
  {
    return less_equal(r);
  };
  bool operator > (const MomRefobj r) const
  {
    return r.less(*this);
  };
  bool operator >= (const MomRefobj r) const
  {
    return r.less_equal(*this);
  };
  bool operator == (const MomRefobj r) const
  {
    return equal(r);
  };
  bool operator != (const MomRefobj r) const
  {
    return !equal(r);
  };
  static void collect_vector(const std::vector<MomRefobj>&) {};
  static void really_collect_vector(std::vector<MomRefobj>&) {};
  static inline void collect_vector_sequence(std::vector<MomRefobj>&vec, const MomSequence&seq);
  static inline void collect_vector_refobj(std::vector<MomRefobj>&vec, const MomRefobj rob)
  {
    if (rob) vec.push_back(rob);
  }
  static inline void collect_vector_sequence(std::vector<MomRefobj>&vec, const MomSequence*pseq = nullptr)
  {
    if (pseq) collect_vector_sequence(vec, *pseq);
  };
  template<typename... Args>
  static void collect_vector(std::vector<MomRefobj>&vec,  Args... args)
  {
    vec.reserve(vec.size()+ 4*sizeof...(args)/3);
    really_collect_vector(vec, args...);
  };
  template<typename... Args>
  static void really_collect_vector(std::vector<MomRefobj>&vec, const MomRefobj rob, Args... args)
  {
    collect_vector_refobj(vec,rob);
    really_collect_vector(vec, args...);
  };
  template<typename... Args>
  static void really_collect_vector(std::vector<MomRefobj>&vec,MomObject*pob, Args... args)
  {
    if (pob) collect_vector_refobj(vec,MomRefobj{pob});
    really_collect_vector(vec, args...);
  }
  template<typename... Args>
  static void really_collect_vector(std::vector<MomRefobj>&vec, const MomSequence& seq, Args... args)
  {
    collect_vector_sequence(vec,seq);
    really_collect_vector(vec, args...);
  };
  template<typename... Args>
  static void really_collect_vector(std::vector<MomRefobj>&vec, const MomSequence* pseq, Args... args)
  {
    if (pseq) collect_vector_sequence(vec,pseq);
    really_collect_vector(vec, args...);
  };
};    // end class MomRefobj
static_assert(sizeof(MomRefobj)==sizeof(void*), "too wide MomRefobj");




struct MomHashRefobj
{
  inline size_t operator() (MomRefobj) const;
};

struct MomLessRefobj
{
  inline bool operator() (MomRefobj l, MomRefobj r) const;
};

class MomObject
{
public:
  typedef std::pair<const MomSerial63, const MomSerial63> pairid_t;
  static bool id_is_null(const pairid_t pi)
  {
    return pi.first.serial() == 0 && pi.second.serial() == 0;
  };
  static std::string id_to_string(const pairid_t);
  static const pairid_t id_from_cstr(const char*s, const char*&end, bool fail=false);
  static const pairid_t id_from_cstr(const char*s, bool fail=false)
  {
    const char*end = nullptr;
    return id_from_cstr(s,end,fail);
  };
  static inline const pairid_t random_id(void);
  static unsigned id_bucketnum(const pairid_t pi)
  {
    return pi.first.bucketnum();
  };
  static inline const pairid_t random_id_of_bucket(unsigned bun);
private:
  const pairid_t _serpair;
  static MomHash_t hash0pairid(const pairid_t pi);
public:
  static MomHash_t hash_id(const pairid_t pi)
  {
    if (MOM_UNLIKELY(id_is_null(pi))) return 0;
    auto ls = pi.first.serial();
    auto rs = pi.second.serial();
    MomHash_t h {(MomHash_t)(ls ^ (rs>>2))};
    if (MOM_UNLIKELY(h==0)) return hash0pairid(pi);
    return h;
  };
  const pairid_t ident() const
  {
    return _serpair;
  };
  const MomSerial63 hi_ident() const
  {
    return _serpair.first;
  };
  const MomSerial63 lo_ident() const
  {
    return _serpair.second;
  };
  uint64_t hi_serial() const
  {
    return _serpair.first.serial();
  };
  uint64_t lo_serial() const
  {
    return _serpair.second.serial();
  };
  MomHash_t hash() const
  {
    return hash_id(_serpair);
  };
  bool equal(const MomObject*r) const
  {
    return this==r;
  };
  bool equal(const MomRefobj rf) const
  {
    return this== rf.unsafe_get_const();
  };
  bool less(const MomObject*r) const
  {
    if (!r) return false;
    if (this == r) return false;
    return _serpair < r->_serpair;
  };
  bool less_equal(const MomObject*r) const
  {
    if (!r) return false;
    if (r==this) return true;
    return _serpair <= r->_serpair;
  };
  bool less_equal(const MomRefobj rf) const
  {
    return less_equal(rf.unsafe_get_const());
  };
  bool operator = (const MomObject*r) const
  {
    return equal(r);
  };
  bool operator != (const MomObject*r) const
  {
    return !equal(r);
  };
  bool operator < (const MomObject*r) const
  {
    return less(r);
  };
  bool operator <= (const MomObject*r) const
  {
    return less_equal(r);
  };
  bool operator > (const MomObject*r) const
  {
    return !less_equal(r);
  };
  bool operator >= (const MomObject*r) const
  {
    return !less(r);
  };
};    // end class MomObject




////////////////////////////////////////////////////////////////
class MomSequence:
  public std::enable_shared_from_this<MomSequence>
{
public:
  struct RawTag {};
  struct PlainTag {};
  struct CheckTag {};
  enum class SeqKind : std::uint8_t
  {
    NoneS=0,
    TupleS= (std::uint8_t)MomVKind::TupleK,
    SetS= (std::uint8_t)MomVKind::SetK,
  };
  static constexpr bool _check_sequence_ = true;
protected:
  const MomHash_t _hash;
  const unsigned _len;
  const MomRefobj* _seq;
  const SeqKind _skd;
  ~MomSequence()
  {
    *(const_cast<SeqKind*>(&_skd)) = SeqKind::NoneS;
    for (unsigned ix=0; ix<_len; ix++)
      (const_cast< MomRefobj*>(_seq))[ix].clear();
    delete _seq;
    _seq = nullptr;
    *(const_cast<MomHash_t*>(&_hash)) = 0;
    *(const_cast<unsigned*>(&_len)) = 0;
  }
  static unsigned check_length(unsigned ln)
  {
    if (ln >= MOM_SIZE_MAX)
      {
        MOM_BACKTRACELOG("MomSequence::check_length too big ln=" << ln);
        throw std::runtime_error("MomSequence::check_length too big ln");
      }
    return ln;
  }
  MomSequence(SeqKind k, MomHash_t h, unsigned ln, const MomRefobj*arr)
    : _hash(h), _len( check_length(ln)), _seq(new MomRefobj[ln]), _skd(k)
  {
    MOM_ASSERT(ln==0 || arr!=nullptr, "missing arr");
    for (unsigned ix=0; ix<ln; ix++) (const_cast< MomRefobj*>(_seq))[ix] = arr[ix];
  };
  MomSequence(SeqKind k, MomHash_t h, unsigned ln, const MomRefobj*rawarr, RawTag)
    : _hash(h), _len(ln), _seq(rawarr), _skd(k)
  {
    MOM_ASSERT(ln<=MOM_SIZE_MAX, "too big length=" << ln);
    MOM_ASSERT(ln==0 || rawarr!=nullptr, "missing rawarr");
  }
  MomSequence(SeqKind k, MomHash_t h, const MomSequence&sq)
    : MomSequence(k,h,sq._len,sq._seq)
  {
    MOM_ASSERT(sq._skd != SeqKind::NoneS, "bad original sequence");
  };
  MomSequence(SeqKind k, MomHash_t h, const std::vector<MomRefobj>& vec)
    : MomSequence(k, h, check_length(vec.size()), vec.data()) {};
  MomSequence(SeqKind k, MomHash_t h, const std::vector<MomRefobj>& vec, RawTag)
    : MomSequence(k, h, vec.size(), vec.data()) {};
  MomSequence(SeqKind k, MomHash_t h, std::initializer_list<MomRefobj> il)
    : MomSequence(k, h, check_length(il.size()), il.begin()) {};
  MomSequence(SeqKind k, MomHash_t h, std::initializer_list<MomRefobj> il, RawTag)
    : MomSequence(k, h, check_length(il.size()), il.begin()) {};
  static bool good_array_refobj(const MomRefobj* arr, unsigned len)
  {
    if (len>0 && !arr) return false;
    for (unsigned ix=0; ix<len; ix++)
      if (!arr[ix]) return false;
    return true;
  }
  static const MomRefobj* filter_array_refobj(const MomRefobj* arr, unsigned len)
  {
    if (!good_array_refobj(arr, len))
      {
        MOM_BACKTRACELOG("MomSequence::filter_array_refobj bad array " << arr << " of length:" << len);
        throw std::runtime_error("MomSequence::filter_array_refobj bad array");
      }
    return arr;
  }
  static bool good_vector_refobj(const std::vector<MomRefobj>& vec)
  {
    for (MomRefobj ro : vec)
      if (!ro) return false;
    return true;
  };
  static const std::vector<MomRefobj>& filter_vector_refobj(const std::vector<MomRefobj>& vec)
  {
    if (!good_vector_refobj(vec))
      {
        MOM_BACKTRACELOG("MomSequence::filter_array_refobj bad vector of size:" << vec.size());
        throw std::runtime_error("MomSequence::filter_vector_refobj bad vector");
      }
    return vec;
  }
  static bool good_initializer_list_refobj(std::initializer_list<MomRefobj> il)
  {
    for (MomRefobj ro : il)
      if (!ro) return false;
    return true;
  };
  static std::initializer_list<MomRefobj> filter_initializer_list_refobj(std::initializer_list<MomRefobj> il)
  {
    if (!good_initializer_list_refobj(il))
      {
        MOM_BACKTRACELOG("MomSequence::filter_initializer_list_refobj bad initializer_list of size:" << il.size());
        throw std::runtime_error("MomSequence::filter_initializer_list_refobj bad initializer_list");
      }
    return il;
  }
  static std::vector<MomRefobj> vector_real_refs(const std::vector<MomRefobj>& vec);
  static std::vector<MomRefobj> vector_real_refs(const std::initializer_list<MomRefobj> il);
  template<unsigned hinit, unsigned k1, unsigned k2, unsigned k3, unsigned k4, bool check=false>
  static MomHash_t  hash_vector_refobj(const std::vector<MomRefobj>& vec)
  {
    MomHash_t h = hinit;
    unsigned ln = vec.size();
    for (unsigned rk = 0; rk<ln; rk++)
      {
        auto ro = vec[rk];
        if (check)
          {
            if (!ro)
              {
                MOM_BACKTRACELOG("MomSequence::hash_vector_refobj null ref#" << rk << " ln="<< ln);
                throw std::runtime_error("MomSequence::hash_vector_refobj null ref");
              }
          }
        else
          MOM_ASSERT(ro, "hash_vector_refobj bad refobj#" << rk << ":" << ro);
        if (rk%2==0)
          h = (h*k1) + (k3*ro.hash());
        else
          h = (h*k2) ^ (k4*ro.hash());
      }
    if (MOM_UNLIKELY(h==0))
      {
        if (ln == 0)
          h = ((k1+7*k2+19*k3+317*k4)&0xfffff) + 11;
        else
          {
            auto firstro = vec[0];
            h = (((k1+19*k2)*ln) & 0xfffff) + 13*(((k3+13*k4)*firstro.hash()) & 0xffffff) + 1321;
          }
      }
    MOM_ASSERT(h!=0, "hash_vector_refobj zero h");
    return h;
  }; /* end hash_vector_refobj */
  template<unsigned hinit, unsigned k1, unsigned k2, unsigned k3, unsigned k4, bool check=false>
  static MomHash_t hash_initializer_list_refobj(const std::initializer_list<MomRefobj> il)
  {
    MomHash_t h = hinit;
    unsigned ln = il.size();
    unsigned rk = 0;
    for (MomRefobj ro : il)
      {
        if (check)
          {
            if (!ro)
              {
                MOM_BACKTRACELOG("MomSequence::hash_initializer_list_refobj null ref#" << rk << " ln="<< ln);
                throw std::runtime_error("MomSequence::hash_initializer_list_refobj null ref");
              }
          }
        else
          MOM_ASSERT(ro, "hash_initializer_list_refobj bad refobj#" << rk << ":" << ro);
        rk++;
        if (rk%2==0)
          h = (h*k1) + (k3*ro.hash());
        else
          h = (h*k2) ^ (k4*ro.hash());
      }
    if (MOM_UNLIKELY(h==0))
      {
        if (ln == 0)
          h = ((k1+7*k2+19*k3+317*k4)&0xfffff) + 11;
        else
          {
            auto firstro = *il.begin();
            h = (((k1+19*k2)*ln) & 0xfffff) + 13*(((k3+13*k4)*firstro.hash()) & 0xffffff) + 1321;
          }
      }
    MOM_ASSERT(h!=0, "hash_initializer_list_refobj zero h");
    return h;
  }; /* end hash_initializer_list_refobj */
  template<unsigned hinit, unsigned k1, unsigned k2, unsigned k3, unsigned k4, bool check=false>
  static MomHash_t  hash_array_refobj(const MomRefobj* arr, unsigned len)
  {
    if (len == 0) return hinit;
    MOM_ASSERT(arr != nullptr, "hash_array_refobj null arr");
    MomHash_t h = hinit;
    for (unsigned rk = 0; rk<len; rk++)
      {
        auto ro = arr[rk];
        if (check)
          {
            if (!ro)
              {
                MOM_BACKTRACELOG("MomSequence::hash_array_refobj null ref#" << rk << " len="<< len);
                throw std::runtime_error("MomSequence::hash_array_refobj null ref");
              }
          }
        else
          MOM_ASSERT(ro, "hash_array_refobj bad refobj#" << rk << ":" << ro);
        if (rk%2==0)
          h = (h*k1) + (k3*ro.hash());
        else
          h = (h*k2) ^ (k4*ro.hash());
      }
    if (MOM_UNLIKELY(h==0))
      {
        if (len == 0)
          h = ((k1+7*k2+19*k3+317*k4)&0xfffff) + 11;
        else
          {
            auto firstro = arr[0];
            h = (((k1+19*k2)*len) & 0xfffff) + 13*(((k3+13*k4)*firstro.hash()) & 0xffffff) + 1321;
          }
      }
    MOM_ASSERT(h!=0, "hash_array_refobj zero h");
    return h;
  }; /* end hash_array_refobj */
public:
  typedef const MomRefobj* iterator;
  MomHash_t hash() const
  {
    return _hash;
  };
  unsigned length() const
  {
    return _len;
  };
  unsigned size() const
  {
    return _len;
  };
  SeqKind skind() const
  {
    return _skd;
  };
  bool is_tuple() const
  {
    return _skd == SeqKind::TupleS;
  };
  bool is_set() const
  {
    return _skd == SeqKind::SetS;
  };
  const MomRefobj* data() const
  {
    return _seq;
  };
  const MomRefobj* begin() const
  {
    return _seq;
  };
  const MomRefobj* end() const
  {
    return _seq+_len;
  };
  MomRefobj nth(int rk) const
  {
    if (rk<0) rk += _len;
    if (rk>=0 && rk<(int)_len) return _seq[rk];
    return nullptr;
  }
  MomRefobj nth(int rk, CheckTag) const
  {
    int origrk = 0;
    if (rk<0) rk += _len;
    if (rk>=0 && rk<(int)_len) return _seq[rk];
    MOM_BACKTRACELOG("MomSequence::nth bad rk=" << origrk << " for length=" << _len);
    throw std::runtime_error("MomSequence::nth rank out of range");
  }
  MomRefobj at(unsigned ix, PlainTag) const
  {
    if (ix<_len) return _seq[ix];
    return nullptr;
  }
  MomRefobj at(unsigned ix) const
  {
    return at(ix,PlainTag{});
  };
  MomRefobj at(unsigned ix, RawTag) const
  {
    MOM_ASSERT(ix<_len, "MomSequence::at bad ix=" << ix << " for length=" << _len);
    return _seq[ix];
  }
  MomRefobj at(unsigned ix, CheckTag) const
  {
    if (ix<_len) return _seq[ix];
    MOM_BACKTRACELOG("MomSequence::at bad ix=" << ix << " for length=" << _len);
    throw std::runtime_error("MomSequence:at index out of range");
  }
};        // end class MomSequence


class MomTuple : public MomSequence
{
  static constexpr unsigned hinit = 100;
  static constexpr unsigned k1 = 233;
  static constexpr unsigned k2 = 1217;
  static constexpr unsigned k3 = 2243;
  static constexpr unsigned k4 = 139;
public:
  ~MomTuple() {};
  MomTuple(std::initializer_list<MomRefobj> il, PlainTag) :
    MomSequence(SeqKind::TupleS,hash_initializer_list_refobj<hinit,k1,k2,k3,k4,_check_sequence_>(il), il) {};
  MomTuple(std::initializer_list<MomRefobj> il, RawTag) :
    MomSequence(SeqKind::TupleS,hash_initializer_list_refobj<hinit,k1,k2,k3,k4>(il), il) {};
  MomTuple(std::initializer_list<MomRefobj> il, CheckTag) :
    MomSequence(SeqKind::TupleS,hash_initializer_list_refobj<hinit,k1,k2,k3,k4,_check_sequence_>(il), il) {};
  MomTuple(const std::vector<MomRefobj>&vec, PlainTag) :
    MomSequence(SeqKind::TupleS,hash_vector_refobj<hinit,k1,k2,k3,k4,_check_sequence_>(vec),vec) {};
  MomTuple(const std::vector<MomRefobj>&vec, RawTag) :
    MomSequence(SeqKind::TupleS,hash_vector_refobj<hinit,k1,k2,k3,k4>(vec),vec) {};
  MomTuple(const std::vector<MomRefobj>&vec, CheckTag) :
    MomSequence(SeqKind::TupleS,hash_vector_refobj<hinit,k1,k2,k3,k4,_check_sequence_>(vec),vec) {};
  MomTuple(const std::vector<MomRefobj>&vec) :  MomTuple(vec, CheckTag{}) {};
  MomTuple(const MomSequence&sq) :
    MomSequence(SeqKind::TupleS,
                sq.is_tuple()?sq.hash()
                :hash_array_refobj<hinit,k1,k2,k3,k4>(sq.data(),sq.size()),
                sq.size(),sq.data())  {};
  MomTuple(const MomSequence&sq, CheckTag) :
    MomSequence(SeqKind::TupleS,sq.is_tuple()?sq.hash()
                :hash_array_refobj<hinit,k1,k2,k3,k4,_check_sequence_>(sq.data(),sq.size()),
                sq.size(),sq.data())  {};
  template <typename... Args> MomTuple(CheckTag tag, Args ... args)
    : MomTuple(std::initializer_list<MomRefobj>
  {
    args...
  }, tag) {};
  template <typename... Args> MomTuple(RawTag tag, Args ... args)
    : MomTuple(std::initializer_list<MomRefobj>
  {
    args...
  }, tag) {};
};        // end class MomTuple


class MomVal
{
  /// these classes are subclasses of MomVal
  friend class MomVNone;
  friend class MomVInt;
  friend class MomVString;
  friend class MomVRef;
  friend class MomVSet;
  friend class MomVTuple;
  friend class MomRefobj;
public:
  struct TagNone {};
  struct TagInt {};
  struct TagString {};
  struct TagRefobj {};
  struct TagSet {};
  struct TagTuple {};
protected:
  const MomVKind _kind;
  union
  {
    void* _ptr;
    intptr_t _int;
    MomRefobj _ref;
    std::shared_ptr<const MomString> _str;
    std::shared_ptr<const MomSet> _set;
    std::shared_ptr<const MomTuple> _tup;
  };
  MomVal(TagNone, std::nullptr_t)
    : _kind(MomVKind::NoneK), _ptr(nullptr) {};
  MomVal(TagInt, intptr_t i)
    : _kind(MomVKind::IntK), _int(i) {};
  inline MomVal(TagString, const std::string& s);
  inline MomVal(TagString, const MomString*);
  inline MomVal(TagRefobj, const MomRefobj);
  inline MomVal(TagSet, const MomSet*pset);
  inline MomVal(TagTuple, const MomTuple*ptup);
public:
  MomVKind kind() const
  {
    return _kind;
  };
  MomVal() : MomVal(TagNone {}, nullptr) {};
  MomVal(std::nullptr_t) : MomVal(TagNone {}, nullptr) {};
  inline MomVal(const MomVal&v);
  inline MomVal(MomVal&&v);
  inline MomVal& operator = (const MomVal&);
  inline MomVal& operator = (MomVal&&);
  inline void clear();
  void reset(void)
  {
    clear();
  };
  ~MomVal()
  {
    reset();
  };
  inline bool equal(const MomVal&) const;
  bool operator == (const MomVal&r) const
  {
    return equal(r);
  };
  bool less(const MomVal&) const;
  bool less_equal(const MomVal&) const;
  bool operator < (const MomVal&v) const
  {
    return less(v);
  };
  bool operator <= (const MomVal&v) const
  {
    return less_equal(v);
  };
  inline MomHash_t hash() const;
  void out(std::ostream&os) const;
  /// the is_XXX methods are testing the kind
  /// the as_XXX methods may throw an exception
  /// the get_XXX methods may throw an exception or gives a raw non-null ptr
  /// the to_XXX methods make return a default
  bool is_null(void) const
  {
    return _kind == MomVKind::NoneK;
  };
  bool operator ! (void) const
  {
    return is_null();
  };
  operator bool (void) const
  {
    return !is_null();
  };
  inline std::nullptr_t as_null(void) const;
  //
  bool is_int(void) const
  {
    return  _kind == MomVKind::IntK;
  };
  inline intptr_t as_int (void) const;
  inline intptr_t to_int (intptr_t def=0) const
  {
    if (_kind != MomVKind::IntK) return def;
    return _int;
  };
  //
  bool is_string(void) const
  {
    return _kind == MomVKind::StringK;
  };
  inline std::shared_ptr<const MomString> as_bstring(void) const;
  inline std::shared_ptr<const MomString> to_bstring(const std::shared_ptr<const MomString>& def=nullptr) const;
  inline const MomString*get_bstring(void) const;
  inline std::string as_string(void) const;
  inline std::string to_string(const std::string& str="") const;
  //
  bool is_set(void) const
  {
    return _kind == MomVKind::SetK;
  };
  inline std::shared_ptr<const MomSet> as_set(void) const;
  inline std::shared_ptr<const MomSet> to_set(const std::shared_ptr<const MomSet> def=nullptr) const;
  inline const MomSet*get_set(void) const;
  //
  bool is_tuple(void) const
  {
    return _kind == MomVKind::TupleK;
  };
  inline std::shared_ptr<const MomTuple> as_tuple(void) const;
  inline std::shared_ptr<const MomTuple> to_tuple(const std::shared_ptr<const MomTuple> def=nullptr) const;
  inline const MomTuple*get_tuple(void) const;
  //
  bool is_sequence(void) const
  {
    return  _kind == MomVKind::SetK || _kind ==MomVKind::TupleK;
  };
  inline std::shared_ptr<const MomSequence> as_sequence(void) const;
  inline std::shared_ptr<const MomSequence> to_sequence(const std::shared_ptr<const MomSequence> def=nullptr) const;
  inline const MomSequence*get_sequence(void) const;
  //
  bool is_refobj(void) const
  {
    return _kind == MomVKind::RefobjK;
  };
  inline MomRefobj as_refobj(void) const;
  inline MomRefobj to_refobj(const MomRefobj def=nullptr) const;
  inline const MomRefobj get_refobj(void) const;
};    // end class MomVal

class MomVNone: public MomVal
{
public:
  MomVNone() : MomVal(TagNone {},nullptr) {};
  ~MomVNone() = default;
};        // end MomVNone

class MomVInt: public MomVal
{
public:
  MomVInt(int64_t i=0): MomVal(TagInt {},i) {};
  ~MomVInt() = default;
};        // end MomVInt


class MomVString: public MomVal
{
public:
  MomVString(const MomString&);
  MomVString(const char*s, int l= -1);
  MomVString(const std::string& str);
  ~MomVString() = default;
};        // end MomVString

class MomVSet: public MomVal
{
public:
  ~MomVSet() = default;
  inline MomVSet(void);
  inline MomVSet(const MomSet&bs);
  inline MomVSet(const std::set<const MomRefobj,MomLessRefobj>&);
  inline MomVSet(const std::unordered_set<const MomRefobj,MomHashRefobj>&);
  inline MomVSet(const std::vector<MomRefobj>&);
  inline MomVSet(const std::vector<MomObject*>&);
// MomVSet(const std::initializer_list<MomRefobj> il)
//   : MomVSet(std::vector<MomRefobj>(il)) {};
// MomVSet(const std::initializer_list<MomObject*>il)
//   : MomVSet(std::vector<MomObject*>(il)) {};
// template <typename... Args> MomVSet(MomObject*obp, Args ... args)
//   : MomVSet(std::initializer_list<MomObject*>
// {
//   obp, args...
// }) {};
// template <typename... Args> MomVSet(MomRefobj ref, Args ... args)
//   : MomVSet(std::initializer_list<const MomRefobj>
// {
//   ref, args...
// }) {};
};        // end MomVSet

class MomVTuple: public MomVal
{
public:
  ~MomVTuple() = default;
  inline MomVTuple(const MomTuple&);
  inline MomVTuple(void);
  MomVTuple(const std::vector<MomRefobj>&);
  MomVTuple(const std::vector<MomObject*>&);
// MomVTuple(std::initializer_list<MomRefobj> il)
//   : MomVTuple(std::vector<MomRefobj>(il)) {};
// MomVTuple(std::initializer_list<MomObject*>il)
//   : MomVTuple(std::vector<MomObject*>(il)) {};
// template <typename... Args> MomVTuple(MomRefobj pob,Args ... args)
//   : MomVTuple(std::initializer_list<MomRefobj>
// {
//   pob,args...
// }) {};
// template <typename... Args> MomVTuple(MomObject*obp,Args ... args)
//   : MomVTuple(std::initializer_list<MomObject*>
// {
//   obp,args...
// }) {};
};        // end MomVTuple



////////////////////////////////////////////////////////////////
/***************** INLINE FUNCTIONS ****************/
MomSerial63::MomSerial63(uint64_t n, bool nocheck) : _serial(n)
{
  if (nocheck || n==0) return;
  if (n<_minserial_)
    {
      MOM_BACKTRACELOG("MomSerial63 too small n:" << n);
      throw std::runtime_error("MomSerial63 too small n");
    };
  if (n>_maxserial_)
    {
      MOM_BACKTRACELOG("MomSerial63 too big n:" << n);
      throw std::runtime_error("MomSerial63 too big n");
    }
}      /* end MomSerial63::MomSerial63 */



/// see also http://stackoverflow.com/a/28613483/841108
void MomVal::clear()
{
  auto k = _kind;
  *const_cast<MomVKind*>(&_kind) = MomVKind::NoneK;
  switch(k)
    {
    case MomVKind::NoneK:
      break;
    case MomVKind::IntK:
      _int = 0;
      break;
    case MomVKind::StringK:
      _str.~shared_ptr<const MomString>();;
      break;
    case MomVKind::RefobjK:
      _ref.clear();
      break;
    case MomVKind::SetK:
      _set.~shared_ptr<const MomSet>();
    case MomVKind::TupleK:
      _tup.~shared_ptr<const MomTuple>();
      break;
    }
  _ptr = nullptr;
} // end MomVal::clear()

bool
MomRefobj::less(const MomRefobj r) const
{
  if (!r) return false;
  if (unsafe_get_const() == r.unsafe_get_const()) return false;
  if (!unsafe_get_const()) return true;
  return unsafe_get_const()->less(r);
} // end MomRefobj::less


MomHash_t
MomRefobj::hash(void) const
{
  auto pob = unsafe_get_const();
  if (!pob) return 0;
  return pob->hash();
} // end MomRefobj::hash

bool
MomRefobj::less_equal(const MomRefobj r) const
{
  if (unsafe_get_const() == r.unsafe_get_const()) return true;
  if (!unsafe_get_const()) return true;
  return unsafe_get_const()->less_equal(r);
} // end MomRefobj::less_equal

bool
MomRefobj::equal(const MomRefobj r) const
{
  return (unsafe_get_const() == r.unsafe_get_const());
} // end MomRefobj::equal

const MomObject::pairid_t
MomObject::random_id(void)
{
  return pairid_t{MomSerial63::make_random(),MomSerial63::make_random()};
} // end MomObject::random_id

const MomObject::pairid_t
MomObject::random_id_of_bucket(unsigned bucknum)
{
  return pairid_t{MomSerial63::make_random_of_bucket(bucknum),
                  MomSerial63::make_random()};
}      // end MomObject::random_id_of_bucket


void
MomRefobj::collect_vector_sequence(std::vector<MomRefobj>&vec, const MomSequence&seq)
{
  vec.reserve(vec.size()+seq.size());
  for (auto rob : seq)
    {
      MOM_ASSERT(rob, "collect_vector_sequence null rob");
      collect_vector_refobj(vec,rob);
    }
}      // end MomRefobj::collect_vector_sequence
#endif /*MONIMELT_HEADER*/
