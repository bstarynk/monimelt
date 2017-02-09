// file monimelt.h       -*- C++ -*-
#ifndef MOMIMELT_HEADER
#define MONIMELT_HEADER "monimelt.h"

#include <features.h> // GNU things
#include <algorithm>
#include <climits>
#include <cmath>
#include <cstdint>
#include <cstring>
#include <deque>
#include <fstream>
#include <initializer_list>
#include <iostream>
#include <map>
#include <memory>
#include <random>
#include <set>
#include <sstream>
#include <stdexcept>
#include <typeinfo>
#include <unordered_map>
#include <unordered_set>
#include <vector>
#include <mutex>
#include <shared_mutex>

// libbacktrace from GCC 6, i.e. libgcc-6-dev package
#include <backtrace.h>

#include <dlfcn.h>
#include <pthread.h>
#include <sched.h>
#include <stdlib.h>
#include <sys/syscall.h>
#include <syslog.h>
#include <unistd.h>

#include <utf8.h>

#include "jsoncpp/json/json.h"

// common prefix mom

// mark unlikely conditions to help optimization
#ifdef __GNUC__
#define MOM_UNLIKELY(P) __builtin_expect(!!(P), 0)
#define MOM_LIKELY(P) !__builtin_expect(!(P), 0)
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
extern "C" const char *const monimelt_cxxsources[];
extern "C" const char *const monimelt_csources[];
extern "C" const char *const monimelt_shellsources[];
extern "C" const char monimelt_directory[];
extern "C" const char monimelt_statebase[];

#define MOM_PROGBINARY "monimelt"

/// the dlopen handle for the whole program
extern "C" void *mom_dlh;

static inline pid_t mom_gettid(void)
{
  return syscall(SYS_gettid, 0L);
}

// time measurement, in seconds
// query a clock
static inline double mom_clock_time(clockid_t cid)
{
  struct timespec ts = {0, 0};
  if (clock_gettime(cid, &ts))
    return NAN;
  else
    return (double)ts.tv_sec + 1.0e-9 * ts.tv_nsec;
}

static inline struct timespec mom_timespec(double t)
{
  struct timespec ts = {0, 0};
  if (std::isnan(t) || t < 0.0)
    return ts;
  double fl = floor(t);
  ts.tv_sec = (time_t)fl;
  ts.tv_nsec = (long)((t - fl) * 1.0e9);
  // this should not happen
  if (MOM_UNLIKELY(ts.tv_nsec < 0))
    ts.tv_nsec = 0;
  while (MOM_UNLIKELY(ts.tv_nsec >= 1000 * 1000 * 1000))
    {
      ts.tv_sec++;
      ts.tv_nsec -= 1000 * 1000 * 1000;
    };
  return ts;
}

extern "C" double
mom_elapsed_real_time(void); /* relative to start of program */
extern "C" double mom_process_cpu_time(void);
extern "C" double mom_thread_cpu_time(void);

// call strftime on ti, but replace .__ with centiseconds for ti
extern "C" char *mom_strftime_centi(char *buf, size_t len, const char *fmt,
                                    double ti)
__attribute__((format(strftime, 3, 0)));

#define MOM_EMPTY_SLOT ((void *)(2 * sizeof(void *)))

extern "C" void mom_backtracestr_at(const char *fil, int lin,
                                    const std::string &msg);

#define MOM_BACKTRACELOG_AT(Fil, Lin, Log)                     \
  do {                                                         \
    std::ostringstream _out_##Lin;                             \
    _out_##Lin << Log << std::flush;                           \
    mom_backtracestr_at((Fil), (Lin), _out_##Lin.str());       \
  } while (0)
#define MOM_BACKTRACELOG_AT_BIS(Fil, Lin, Log)   \
  MOM_BACKTRACELOG_AT(Fil, Lin, Log)
#define MOM_BACKTRACELOG(Log) MOM_BACKTRACELOG_AT_BIS(__FILE__, __LINE__, Log)

extern "C" void mom_abort(void) __attribute__((noreturn));
#ifndef NDEBUG
#define MOM_ASSERT_AT(Fil, Lin, Prop, Log)                         \
  do {                                                             \
    if (MOM_UNLIKELY(!(Prop))) {                                   \
      MOM_BACKTRACELOG_AT(Fil, Lin,                                \
                          "**MOM_ASSERT FAILED** " #Prop ":"       \
                          " @ "                                    \
                          << __PRETTY_FUNCTION__                   \
        << std::endl  << "::" << Log);           \
      mom_abort();                                                 \
    }                                                              \
  } while (0)
#else
#define MOM_ASSERT_AT(Fil, Lin, Prop, Log)         \
  do {                                             \
    if (false && !(Prop))                          \
      MOM_BACKTRACELOG_AT(Fil, Lin, Log);          \
  } while (0)
#endif // NDEBUG
#define MOM_ASSERT_AT_BIS(Fil, Lin, Prop, Log)                                 \
  MOM_ASSERT_AT(Fil, Lin, Prop, Log)
#define MOM_ASSERT(Prop, Log) MOM_ASSERT_AT_BIS(__FILE__, __LINE__, Prop, Log)

extern "C" bool mom_verboseflag;
#define MOM_VERBOSELOG_AT(Fil, Lin, Log)                                       \
  do {                                                                         \
    if (mom_verboseflag)                                                       \
      std::clog << "*MOM @" << Fil << ":" << Lin << " /" << __FUNCTION__       \
                << ": " << Log << std::endl;                                   \
  } while (0)
#define MOM_VERBOSELOG_AT_BIS(Fil, Lin, Log) MOM_VERBOSELOG_AT(Fil, Lin, Log)
#define MOM_VERBOSELOG(Log) MOM_VERBOSELOG_AT_BIS(__FILE__, __LINE__, Log)

#define MOM_NEVERLOG_AT(Fil, Lin, Log)    \
  do {            \
    if (false && mom_verboseflag)   \
      std::clog  << "@-: " << Log << std::endl; \
  } while (0)
#define MOM_NEVERLOG_AT_BIS(Fil, Lin, Log) MOM_NEVERLOG_AT(Fil, Lin, Log)
#define MOM_NEVERLOG(Log) MOM_NEVERLOG_AT_BIS(__FILE__, __LINE__, Log)

// MOM_DO_NOT_LOG has the same length in characters as MOM_VERBOSELOG
#define MOM_DO_NOT_LOG(Log) MOM_NEVERLOG(Log)
//      MOM_VERBOSELOG has the same width

std::string mom_demangled_typename(const std::type_info &ti);

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
        auto s1 = randev(), s2 = randev(), s3 = randev(), s4 = randev(),
             s5 = randev(), s6 = randev(), s7 = randev();
        std::seed_seq seq{s1, s2, s3, s4, s5, s6, s7};
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
    while (MOM_UNLIKELY(r == 0));
    return r;
  };
  uint64_t generate_64u(void)
  {
    return (static_cast<uint64_t>(generate_32u()) << 32) |
           static_cast<uint64_t>(generate_32u());
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
}; // end class MomRandom

#define MOM_B62DIGITS                   \
  "0123456789"                          \
  "abcdefghijklmnopqrstuvwxyz"          \
  "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

class MomSerial63
{
  uint64_t _serial;

public:
  static constexpr const uint64_t _minserial_ = 62*62; /// 3884
  static constexpr const uint64_t _maxserial_ = /// 8392993658683402240, about 8.392994e+18
    (uint64_t)10 * 62 * (62 * 62 * 62) * (62 * 62 * 62) * (62 * 62 * 62);
  static constexpr const uint64_t _deltaserial_ = _maxserial_ - _minserial_;
  static constexpr const char *_b62digstr_ = MOM_B62DIGITS;
  static constexpr unsigned _nbdigits_ = 11;
  static constexpr unsigned _base_ = 62;
  static_assert(_maxserial_ < ((uint64_t)1 << 63),
                "corrupted _maxserial_ in MomSerial63");
  static_assert(_deltaserial_ > ((uint64_t)1 << 62),
                "corrupted _deltaserial_ in MomSerial63");
  static constexpr const unsigned _maxbucket_ = 10 * 62;
  inline MomSerial63(uint64_t n = 0, bool nocheck = false);
  MomSerial63(std::nullptr_t) : _serial(0) {};
  ~MomSerial63()
  {
    _serial = 0;
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
  static const MomSerial63 make_from_cstr(const char *s, const char *&end,
                                          bool fail = false);
  static const MomSerial63 make_from_cstr(const char *s, bool fail = false)
  {
    const char *end = nullptr;
    return make_from_cstr(s, end, fail);
  };
  static const MomSerial63 make_random(void);
  static const MomSerial63 make_random_of_bucket(unsigned bun);
  MomSerial63(const MomSerial63 &s) : _serial(s._serial) {};
  MomSerial63(MomSerial63 &&s) : _serial(std::move(s._serial)) {};
  operator bool() const
  {
    return _serial != 0;
  };
  bool operator!() const
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
  bool operator==(const MomSerial63 r) const
  {
    return equal(r);
  };
  bool operator!=(const MomSerial63 r) const
  {
    return !equal(r);
  };
  bool operator<(const MomSerial63 r) const
  {
    return less(r);
  };
  bool operator<=(const MomSerial63 r) const
  {
    return less_equal(r);
  };
  bool operator>(const MomSerial63 r) const
  {
    return !less_equal(r);
  };
  bool operator>=(const MomSerial63 r) const
  {
    return !less(r);
  };
}; /* end class MomSerial63 */

inline std::ostream &operator<<(std::ostream &os, const MomSerial63 s)
{
  os << s.to_string();
  return os;
} // end <<

typedef uint32_t MomHash_t;
typedef Json::Value MomJson;

//////////////// to ease debugging
class MomOut
{
  std::function<void(std::ostream &)> _fn_out;

public:
  MomOut(std::function<void(std::ostream &)> fout) : _fn_out(fout) {};
  ~MomOut() = default;
  void out(std::ostream &os) const
  {
    _fn_out(os);
  };
};
inline std::ostream &operator<<(std::ostream &os, const MomOut &bo)
{
  bo.out(os);
  return os;
};

class MomUtf8Out
{
  std::string _str;
  unsigned _flags;

public:
  MomUtf8Out(const std::string &str, unsigned flags = 0)
    : _str(str), _flags(flags)
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
    _flags = 0;
  };
  MomUtf8Out(const MomUtf8Out &) = default;
  MomUtf8Out(MomUtf8Out &&) = default;
  void out(std::ostream &os) const;
}; // end class MomUtf8Out

inline std::ostream &operator<<(std::ostream &os, const MomUtf8Out &bo)
{
  bo.out(os);
  return os;
};

class MomAllocObj;    // object allocator
class MomJsonParser;  // Json Parser
class MomJsonEmitter;  // Json Emitter
class MomObject;
class MomVal;
class MomString;
class MomSequence;
class MomSet;
class MomTuple;

#define MOM_SIZE_MAX (INT32_MAX / 3)
enum class
MomVKind : std::uint8_t
{
  NoneK,
  IntK,
  /* we probably dont need doubles at first. But we want to avoid NaNs
     if we need them; we would use nil instead of boxed NaN */
  // DoubleK,
  StringK,
  RefobjK,
  ColoRefK,
  SetK,
  TupleK,
  /* we don't need mix (of scalar values, e.g. ints, doubles, strings,
     objects) at first */
  // MixK,
};

class MomRefobj
{
  MomObject *_ptrobj;

public:
  MomRefobj(MomObject &ob) : _ptrobj(&ob) {};
  MomRefobj(MomObject *pob = nullptr) : _ptrobj(pob) {};
  ~MomRefobj()
  {
    _ptrobj = nullptr;
  };
  MomRefobj(const MomRefobj &ro) : _ptrobj(ro._ptrobj) {};
  MomRefobj(MomRefobj &&mo) : _ptrobj(std::move(mo._ptrobj)) {};
  MomRefobj &operator=(const MomRefobj &ro)
  {
    _ptrobj = ro._ptrobj;
    return *this;
  };
  MomRefobj &operator=(MomRefobj &&mo)
  {
    _ptrobj = std::move(mo._ptrobj);
    return *this;
  };
  MomObject *get_const(void) const
  {
    if (!_ptrobj)
      {
        MOM_BACKTRACELOG("MomRefobj::get_const nil pointer @" << (void *)this);
        throw std::runtime_error("MomRefobj::get_const nil dereference");
      }
    return _ptrobj;
  };
  MomObject *get(void)
  {
    if (!_ptrobj)
      {
        MOM_BACKTRACELOG("MomRefobj::get nil pointer @" << (void *)this);
        throw std::runtime_error("MomRefobj::get nil dereference");
      }
    return _ptrobj;
  }
  MomObject *unsafe_get(void)
  {
    return _ptrobj;
  };
  MomObject *unsafe_get_const(void) const
  {
    return _ptrobj;
  };
  operator MomObject *() const
  {
    return get_const();
  };
  MomObject *operator*(void)const
  {
    return get_const();
  };
  MomObject *operator*(void)
  {
    return get();
  };
  MomObject *operator->(void)const
  {
    return get_const();
  };
  MomObject *operator->(void)
  {
    return get();
  };
  MomRefobj &unsafe_put(MomObject *pob)
  {
    _ptrobj = pob;
    return *this;
  };
  MomRefobj &put_non_nil(MomObject *pob)
  {
    if (pob == nullptr)
      {
        MOM_BACKTRACELOG("MomRefobj::put_non_nil got nil pointer @"
                         << (void *)this);
        throw std::runtime_error("MomRefobj::put_non_nil with nil pointer");
      }
    return *this;
  }
  MomRefobj &operator=(MomObject *pob)
  {
    return put_non_nil(pob);
  };
  MomRefobj &operator=(std::nullptr_t)
  {
    _ptrobj = nullptr;
    return *this;
  };
  MomRefobj &clear(void)
  {
    _ptrobj = nullptr;
    return *this;
  };
  inline MomHash_t hash(void) const;
  inline bool less(const MomRefobj) const;
  inline bool less_equal(const MomRefobj) const;
  inline bool equal(const MomRefobj) const;
  bool operator<(const MomRefobj r) const
  {
    return less(r);
  };
  bool operator<=(const MomRefobj r) const
  {
    return less_equal(r);
  };
  bool operator>(const MomRefobj r) const
  {
    return r.less(*this);
  };
  bool operator>=(const MomRefobj r) const
  {
    return r.less_equal(*this);
  };
  bool operator==(const MomRefobj r) const
  {
    return equal(r);
  };
  bool operator!=(const MomRefobj r) const
  {
    return !equal(r);
  };
  static std::vector<MomRefobj> reserve_vector(size_t sz)
  {
    std::vector<MomRefobj> vec;
    vec.reserve(sz);
    return vec;
  };
  static std::vector<MomRefobj> &collect_vector(std::vector<MomRefobj> &vec)
  {
    return vec;
  };
  static void really_collect_vector(std::vector<MomRefobj> &) {};
  static inline void collect_vector_sequence(std::vector<MomRefobj> &vec,
      const MomSequence &seq);
  static inline void collect_vector_refobj(std::vector<MomRefobj> &vec,
      const MomRefobj rob)
  {
    if (rob)
      vec.push_back(rob);
  }
  static inline void
  collect_vector_sequence(std::vector<MomRefobj> &vec,
                          const MomSequence *pseq = nullptr)
  {
    if (pseq)
      collect_vector_sequence(vec, *pseq);
  };
  template <typename... Args>
  static std::vector<MomRefobj> &collect_vector(std::vector<MomRefobj> &vec,
      Args... args)
  {
    vec.reserve(vec.size() + 4 * sizeof...(args) / 3);
    really_collect_vector(vec, args...);
    return vec;
  };
  template <typename... Args>
  static void really_collect_vector(std::vector<MomRefobj> &vec,
                                    const MomRefobj rob, Args... args)
  {
    collect_vector_refobj(vec, rob);
    really_collect_vector(vec, args...);
  };
  template <typename... Args>
  static void really_collect_vector(std::vector<MomRefobj> &vec, MomObject *pob,
                                    Args... args)
  {
    if (pob)
      collect_vector_refobj(vec, MomRefobj{pob});
    really_collect_vector(vec, args...);
  }
  template <typename... Args>
  static void really_collect_vector(std::vector<MomRefobj> &vec,
                                    const MomSequence &seq, Args... args)
  {
    collect_vector_sequence(vec, seq);
    really_collect_vector(vec, args...);
  };
  template <typename... Args>
  static void really_collect_vector(std::vector<MomRefobj> &vec,
                                    const MomSequence *pseq, Args... args)
  {
    if (pseq)
      collect_vector_sequence(vec, pseq);
    really_collect_vector(vec, args...);
  };
  static std::set<MomRefobj> make_empty_set(void)
  {
    return std::set<MomRefobj>();
  };
  static std::set<MomRefobj> make_set(void)
  {
    return std::set<MomRefobj>();
  };
  static void add_set_refobj(std::set<MomRefobj> &set, const MomRefobj rob)
  {
    if (rob)
      set.insert(rob);
  };
  static inline void add_set_sequence(std::set<MomRefobj> &set,
                                      const MomSequence &seq);
  static void add_set_sequence(std::set<MomRefobj> &set,
                               const MomSequence *pseq)
  {
    if (pseq)
      add_set_sequence(set, *pseq);
  };
  static void add_set(std::set<MomRefobj> &) {};
  template <typename... Args>
  static void add_set(std::set<MomRefobj> &set, const MomRefobj rob,
                      Args... args)
  {
    add_set_refobj(set, rob);
    add_set(set, args...);
  };
  template <typename... Args>
  static void add_set(std::set<MomRefobj> &set, const MomSequence &seq,
                      Args... args)
  {
    add_set_sequence(set, seq);
    add_set(set, args...);
  };
  template <typename... Args>
  static std::set<MomRefobj> make_set(Args... args)
  {
    auto set = make_empty_set();
    add_set(set, args...);
    return set;
  }
  inline std::size_t longhash() const;
}; // end class MomRefobj
static_assert(sizeof(MomRefobj) == sizeof(void *), "too wide MomRefobj");

struct MomHashRefobj
{
  size_t operator()(MomRefobj ro) const
  {
    return ro.longhash();
  };
};

typedef std::unordered_set<MomRefobj, MomHashRefobj> MomUnorderedSetRefobj;

struct MomLessRefobj
{
  bool operator() (MomRefobj l, MomRefobj r) const
  {
    return l<r;
  };
};
typedef std::set<MomRefobj, MomLessRefobj> MomSetRefobj;

typedef std::pair<const MomSerial63, const MomSerial63> MomPairid;
inline bool operator ! (const MomPairid pi)
{
  return !pi.first && !pi.second;
};

inline std::ostream&operator << (std::ostream&, const MomPairid);
namespace std
{
template<> struct hash<MomPairid>
{
  std::size_t operator() (const MomPairid& p) const
  {
    return (std::size_t)((11*p.first) ^ (p.second >> 3));
  }
};
};


////////////////

class MomVal
{
  /// these classes are subclasses of MomVal
  friend class MomVNone;
  friend class MomVInt;
  friend class MomVString;
  friend class MomVRef;
  friend class MomVSet;
  friend class MomVTuple;
  friend class MomVColoRef;
  friend class MomRefobj;
  friend class MomSet;
  friend class MomTuple;

public:
  struct TagNone {};
  struct TagInt {};
  struct TagString {};
  struct TagRefobj {};
  struct TagColoRef {};
  struct TagSet {};
  struct TagTuple {};
  struct TagCheck {};
  struct TagRaw {};
  struct TagJson {};
  struct ColoRefObj
  {
    MomRefobj _cobref;
    MomRefobj _colorob;
  };
protected:
  MomVKind _kind;
  union
  {
    void *_ptr;
    void *_bothptr [2];
    intptr_t _int;
    MomRefobj _ref;
    ColoRefObj _coloref;
    std::shared_ptr<const MomString> _str;
    std::shared_ptr<const MomSet> _set;
    std::shared_ptr<const MomTuple> _tup;
    std::shared_ptr<const MomSequence> _seq;
  };
  MomVal(TagNone, std::nullptr_t) : _kind(MomVKind::NoneK), _ptr(nullptr)
  {
    _bothptr[1] = nullptr;
  };
  MomVal(TagInt, intptr_t i) : _kind(MomVKind::IntK), _int(i) {};
  MomVal(TagString, const MomString *s) : _kind(s?MomVKind::StringK:MomVKind::NoneK), _str(s)
  {
  };
  MomVal(TagString, const MomString *s, TagCheck) : _kind(MomVKind::StringK), _str(s)
  {
    if (!s)
      {
        MOM_BACKTRACELOG("MomVal no MomString");
        throw std::runtime_error("MomVal no MomString");
      }
  };
  MomVal(TagString, const MomString&s): _kind(MomVKind::StringK), _str(&s) {};
  inline MomVal(TagString, const std::string &s);
  MomVal(TagRefobj, const MomRefobj ro) : _kind(ro?MomVKind::RefobjK:MomVKind::NoneK), _ref(ro)
  {
  };
  MomVal(TagRefobj, const MomRefobj ro, TagCheck) : _kind(MomVKind::RefobjK), _ref(ro)
  {
    if (!ro)
      {
        MOM_BACKTRACELOG("MomVal no MomRefobj");
        throw std::runtime_error("MomVal no MomRefobj");
      }
  };
  MomVal(TagColoRef,  MomObject& ob, MomObject& colorob) :
    _kind(MomVKind::ColoRefK), _coloref{&ob,&colorob} {};
  MomVal(TagColoRef, const ColoRefObj& col) :
    _kind(MomVKind::ColoRefK), _coloref{col} {};
  MomVal(TagColoRef, const MomRefobj ob, const MomRefobj colorob):
    _kind(ob?(colorob?MomVKind::ColoRefK:MomVKind::RefobjK):MomVKind::NoneK),
    _coloref{ob,colorob} {};
  MomVal(TagColoRef, const MomRefobj ob, const MomRefobj colorob, TagRaw)
    : _kind(MomVKind::ColoRefK), _coloref{ob,colorob}
  {
    MOM_ASSERT(ob, "MomVal missing ob for TagColoRef");
    MOM_ASSERT(colorob, "MomVal missing colorob for TagColoRef");
  };
  MomVal(TagColoRef, const MomRefobj ob, const MomRefobj colorob, TagCheck)
    : _kind(MomVKind::ColoRefK), _coloref{ob,colorob}
  {
    if (!ob)
      {
        MOM_BACKTRACELOG("MomVal without ob for TagColoRef");
        throw std::runtime_error("MomVal without ob for TagColoRef");
      }
    if (!colorob)
      {
        MOM_BACKTRACELOG("MomVal without colorob for TagColoRef");
        throw std::runtime_error("MomVal without colorob for TagColoRef");
      }
  };
  MomVal(TagSet, const MomSet *pset)  : _kind(pset?MomVKind::SetK:MomVKind::NoneK), _set(pset)
  {
  }
  MomVal(TagSet, const MomSet *pset, TagCheck) : _kind(MomVKind::SetK), _set(pset)
  {
    if (!pset)
      {
        MOM_BACKTRACELOG("MomVal no MomSet");
        throw std::runtime_error("MomVal no MomSet");
      }
  }
  MomVal(TagSet, const MomSet& set) : _kind(MomVKind::SetK), _set(&set) {};
  MomVal(TagTuple, const MomTuple *ptup) : _kind(ptup?MomVKind::TupleK:MomVKind::NoneK), _tup(ptup)
  {
  }
  MomVal(TagTuple, const MomTuple *ptup, TagCheck) : _kind(MomVKind::TupleK), _tup(ptup)
  {
    if (!ptup)
      {
        MOM_BACKTRACELOG("MomVal no MomTuple");
        throw std::runtime_error("MomVal no MomTuple");
      }
  }
  MomVal(TagTuple, const MomTuple& tup) : _kind(MomVKind::TupleK), _tup(&tup) {};
  static MomVal parse_json(const MomJson&js, MomJsonParser&jp);
public:
  MomJson emit_json(MomJsonEmitter&je) const;
  // the scanning stops as soon as f returns true; the result is true if the value has been fully scanned
  bool scan_objects(const std::function<bool(MomRefobj)>&f) const;
  MomVal(TagJson, const MomJson&js, MomJsonParser&jp)
    : MomVal(std::move(parse_json(js,jp))) {};
  MomVKind kind() const
  {
    return _kind;
  };
  MomVal() : MomVal(TagNone{}, nullptr) {};
  MomVal(std::nullptr_t) : MomVal(TagNone{}, nullptr) {};
  inline MomVal(const MomVal &v);
  inline MomVal(MomVal &&v);
  inline MomVal &operator=(const MomVal &);
  inline MomVal &operator=(MomVal &&);
  inline void clear();
  void reset(void)
  {
    clear();
  };
  ~MomVal()
  {
    reset();
  };
  inline bool equal(const MomVal &) const;
  bool operator==(const MomVal &r) const
  {
    return equal(r);
  };
  bool less(const MomVal &) const;
  bool less_equal(const MomVal &) const;
  bool operator<(const MomVal &v) const
  {
    return less(v);
  };
  bool operator<=(const MomVal &v) const
  {
    return less_equal(v);
  };
  inline MomHash_t hash() const;
  void out(std::ostream &os) const;
  /// the is_XXX methods are testing the kind
  /// the as_XXX methods may throw an exception
  /// the get_XXX methods may throw an exception or gives a raw non-null ptr
  /// the to_XXX methods make return a default
  bool is_null(void) const
  {
    return _kind == MomVKind::NoneK;
  };
  bool operator!(void)const
  {
    return is_null();
  };
  operator bool(void) const
  {
    return !is_null();
  };
  inline std::nullptr_t as_null(void) const;
  //
  bool is_int(void) const
  {
    return _kind == MomVKind::IntK;
  };
  inline intptr_t as_int(void) const;
  inline intptr_t to_int(intptr_t def = 0) const
  {
    if (_kind != MomVKind::IntK)
      return def;
    return _int;
  };
  inline intptr_t unsafe_int() const
  {
    return _int;
  };
  //
  bool is_string(void) const
  {
    return _kind == MomVKind::StringK;
  };
  inline std::shared_ptr<const MomString> as_bstring(void) const;
  inline std::shared_ptr<const MomString>
  to_bstring(const std::shared_ptr<const MomString> &def = nullptr) const;
  inline const MomString *get_bstring(void) const;
  inline const std::string as_string(void) const;
  inline const std::string to_string(const std::string &str = "") const;
  inline const char* to_cstr(const char*defcstr = nullptr) const;
  inline const MomString* unsafe_bstring(void) const;
  //
  bool is_set(void) const
  {
    return _kind == MomVKind::SetK;
  };
  inline std::shared_ptr<const MomSet> as_set(void) const;
  inline std::shared_ptr<const MomSet>
  to_set(const std::shared_ptr<const MomSet> def = nullptr) const;
  inline const MomSet *get_set(void) const
  {
    if (is_set()) return _set.get();
    return nullptr;
  };
  inline const MomSet *unsafe_set(void) const
  {
    return _set.get();
  };
  //
  bool is_tuple(void) const
  {
    return _kind == MomVKind::TupleK;
  };
  inline std::shared_ptr<const MomTuple> as_tuple(void) const;
  inline std::shared_ptr<const MomTuple>
  to_tuple(const std::shared_ptr<const MomTuple> def = nullptr) const;
  inline const MomTuple *get_tuple(void) const
  {
    if (is_tuple()) return _tup.get();
    return nullptr;
  }
  inline const MomTuple *unsafe_tuple(void) const
  {
    return _tup.get();
  };
  //
  bool is_sequence(void) const
  {
    return _kind == MomVKind::SetK || _kind == MomVKind::TupleK;
  };
  inline std::shared_ptr<const MomSequence> as_sequence(void) const;
  inline std::shared_ptr<const MomSequence>
  to_sequence(const std::shared_ptr<const MomSequence> def = nullptr) const;
  inline const MomSequence *get_sequence(void) const
  {
    if (_kind==MomVKind::TupleK||_kind==MomVKind::SetK) return _seq.get();
    return nullptr;
  }
  inline const MomSequence *unsafe_sequence(void) const
  {
    return _seq.get();
  };
  //
  bool is_refobj(void) const
  {
    return _kind == MomVKind::RefobjK;
  };
  inline MomRefobj as_refobj(void) const;
  inline MomRefobj to_refobj(const MomRefobj def = nullptr) const;
  inline const MomRefobj get_refobj(void) const
  {
    return _kind==MomVKind::RefobjK?_ref:nullptr;
  }
  inline const MomRefobj unsafe_refobj(void) const
  {
    return _ref;
  };
  //
  bool is_coloref(void) const
  {
    return _kind == MomVKind::ColoRefK;
  }
  inline const MomRefobj get_colorefobj(void) const
  {
    return _kind == MomVKind::ColoRefK?_coloref._cobref:nullptr;
  }
  inline const MomRefobj get_colorob(void) const
  {
    return _kind == MomVKind::ColoRefK?_coloref._colorob:nullptr;
  };
  inline const MomRefobj unsafe_colorefobj(void) const
  {
    return _coloref._cobref;
  };
  inline const MomRefobj unsafe_colorob(void) const
  {
    return _coloref._colorob;
  };
}; // end class MomVal


////////////////
class MomPayload;

enum class MomSpace : std::uint8_t
{
  NoneSp,
  PredefinedSp,
  GlobalSp
};

inline std::ostream&operator << (std::ostream&os, const MomObject& ob);
class MomObject ///
{
  friend class MomPayload;
private:
  const MomPairid _obserpair;
  std::shared_timed_mutex _obmtx;
  MomSpace _obspace;
  std::unordered_map<MomRefobj,MomVal,MomHashRefobj> _obattrmap;
  std::vector<MomVal> _obcompvec;
  std::unique_ptr<MomPayload> _obpayload;
public:
  static bool id_is_null(const MomPairid pi)
  {
    return pi.first.serial() == 0 && pi.second.serial() == 0;
  };
  static std::string id_to_string(const MomPairid);
  static const MomPairid id_from_cstr(const char *s, const char *&end,
                                      bool fail = false);
  static const MomPairid id_from_cstr(const char *s, bool fail = false)
  {
    const char *end = nullptr;
    return id_from_cstr(s, end, fail);
  };
  static inline const MomPairid random_id(void);
  static unsigned id_bucketnum(const MomPairid pi)
  {
    return pi.first.bucketnum();
  };
  static inline const MomPairid random_id_of_bucket(unsigned bun);
private:
  class ObjBucket
  {
    mutable std::mutex _bumtx;
    std::unordered_map<MomPairid,MomObject*> _bumap;
  public:
    MomObject*find_object_in_bucket(const MomPairid id) const;
    void register_object_in_bucket(MomObject*ob);
    void unregister_object_in_bucket(MomObject*ob);
    ObjBucket() : _bumtx(), _bumap() {};
    ~ObjBucket()
    {
      _bumap.clear();
    };
    ObjBucket(ObjBucket&&) = default;
  };
  static MomHash_t hash0pairid(const MomPairid pi);
  static std::array<ObjBucket,MomSerial63::_maxbucket_> _buckarr_;
public:
  bool unsync_scan_inside_objects(const std::function<bool(MomRefobj)>&f) const;
  MomPayload*get_payload_ptr(void) const
  {
    return _obpayload.get();
  };
  template<class PayloadClass>
  PayloadClass* dyncast_payload_ptr(void) const
  {
    return dynamic_cast<PayloadClass*>(_obpayload.get());
  }
  template<class PayloadClass>
  PayloadClass* checkcast_payload_ptr(void) const
  {
    auto py = _obpayload.get();
    if (!py)
      {
        MOM_BACKTRACELOG("checkcast_payload_ptr no payload in " << *this);
        throw std::runtime_error("checkcast_payload_ptr no payload");
      }
    auto p = dynamic_cast<PayloadClass*>(py);
    if (!p)
      {
        MOM_BACKTRACELOG("checkcast_payload_ptr fail on " << *this << " for "
                         << typeid(PayloadClass).name());
        throw std::runtime_error("checkcast_payload_ptr fail");
      }
    return p;
  }
  template <class PaylClass, typename... Args> PaylClass* put_payload(Args... args)
  {
    auto py = new PaylClass(this, args...);
    _obpayload.reset(py);
    return py;
  }
  void reset_payload()
  {
    _obpayload.reset();
  };
  static MomObject* find_object_of_id(const MomPairid pi)
  {
    if (!pi) return nullptr;
    return _buckarr_[id_bucketnum(pi)].find_object_in_bucket(pi);
  }
  static MomHash_t hash_id(const MomPairid pi)
  {
    if (MOM_UNLIKELY(id_is_null(pi)))
      return 0;
    auto ls = pi.first.serial();
    auto rs = pi.second.serial();
    MomHash_t h{(MomHash_t)(ls ^ (rs >> 2))};
    if (MOM_UNLIKELY(h == 0))
      return hash0pairid(pi);
    return h;
  };
  std::string idstr() const
  {
    return id_to_string(_obserpair);
  };
  const MomPairid ident() const
  {
    return _obserpair;
  };
  const MomSerial63 hi_ident() const
  {
    return _obserpair.first;
  };
  const MomSerial63 lo_ident() const
  {
    return _obserpair.second;
  };
  uint64_t hi_serial() const
  {
    return _obserpair.first.serial();
  };
  uint64_t lo_serial() const
  {
    return _obserpair.second.serial();
  };
  MomHash_t hash() const
  {
    return hash_id(_obserpair);
  };
  bool equal(const MomObject *r) const
  {
    return this == r;
  };
  bool equal(const MomRefobj rf) const
  {
    return this == rf.unsafe_get_const();
  };
  bool less(const MomObject *r) const
  {
    if (!r)
      return false;
    if (this == r)
      return false;
    return _obserpair < r->_obserpair;
  };
  bool less_equal(const MomObject *r) const
  {
    if (!r)
      return false;
    if (r == this)
      return true;
    return _obserpair <= r->_obserpair;
  };
  bool less_equal(const MomRefobj rf) const
  {
    return less_equal(rf.unsafe_get_const());
  };
  bool operator=(const MomObject *r) const
  {
    return equal(r);
  };
  bool operator!=(const MomObject *r) const
  {
    return !equal(r);
  };
  bool operator<(const MomObject *r) const
  {
    return less(r);
  };
  bool operator<=(const MomObject *r) const
  {
    return less_equal(r);
  };
  bool operator>(const MomObject *r) const
  {
    return !less_equal(r);
  };
  bool operator>=(const MomObject *r) const
  {
    return !less(r);
  };
  inline std::size_t longhash() const;
}; // end class MomObject


namespace std
{
template<> struct hash<MomObject>
{
  std::size_t operator() (const MomObject& ob) const
  {
    return ob.longhash();
  }
};
};

class MomPayload ////
{
  friend class MomObject;
  MomObject* _pyowner;
protected:
  MomPayload(MomObject&ob): _pyowner(&ob) {};
  MomPayload(MomObject*pob): _pyowner(pob)
  {
    MOM_ASSERT(pob != nullptr, "null pob for MomPayload");
  };
public:
  virtual ~MomPayload();
  virtual const char*payload_name() const =0;
  // the scanning stops as soon as f returns true; the result is true if the value has been fully scanned
  virtual bool scan_objects(const std::function<bool(MomRefobj)>&f) const =0;
};    // end class MomPayload

////////////////////////////////////////////////////////////////
class MomSequence : public std::enable_shared_from_this<MomSequence>
{
public:
  struct AnyTag {};
  struct RawTag {};
  struct PlainTag {};
  struct CheckTag {};
  enum class SeqKind : std::uint8_t
  {
    NoneS = 0,
    TupleS = (std::uint8_t)MomVKind::TupleK,
    SetS = (std::uint8_t)MomVKind::SetK,
  };
  static constexpr bool _check_sequence_ = true;
  // the scanning stops as soon as f returns true; the result is true if the value has been fully scanned
  bool scan_objects(const std::function<bool(MomRefobj)>&f) const;

protected:
  const MomHash_t _hash;
  const unsigned _len;
  const MomRefobj *_seq;
  const SeqKind _skd;
  ~MomSequence()
  {
    *(const_cast<SeqKind *>(&_skd)) = SeqKind::NoneS;
    if (_seq)
      {
        for (unsigned ix = 0; ix < _len; ix++)
          (const_cast<MomRefobj *>(_seq))[ix].clear();
        delete[] _seq;
        _seq = nullptr;
      }
    *(const_cast<MomHash_t *>(&_hash)) = 0;
    *(const_cast<unsigned *>(&_len)) = 0;
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
  MomSequence(SeqKind k, MomHash_t h)
    : _hash(h), _len(0), _seq(nullptr), _skd(k) {};
  MomSequence(SeqKind k, MomHash_t h, unsigned ln, const MomRefobj *arr)
    : _hash(h), _len(check_length(ln)), _seq(new MomRefobj[ln]), _skd(k)
  {
    MOM_ASSERT(ln == 0 || arr != nullptr, "missing arr");
    for (unsigned ix = 0; ix < ln; ix++)
      (const_cast<MomRefobj *>(_seq))[ix] = arr[ix];
  };
  MomSequence(SeqKind k, MomHash_t h, unsigned ln, const MomRefobj *rawarr,
              RawTag)
    : _hash(h), _len(ln), _seq(rawarr), _skd(k)
  {
    MOM_ASSERT(ln <= MOM_SIZE_MAX, "too big length=" << ln);
    MOM_ASSERT(ln == 0 || rawarr != nullptr, "missing rawarr");
  }
  MomSequence(SeqKind k, MomHash_t h, const MomSequence &sq)
    : MomSequence(k, h, sq._len, sq._seq)
  {
    MOM_ASSERT(sq._skd != SeqKind::NoneS, "bad original sequence");
  };
  MomSequence(SeqKind k, MomHash_t h, const std::vector<MomRefobj> &vec)
    : MomSequence(k, h, check_length(vec.size()), vec.data()) {};
  MomSequence(SeqKind k, MomHash_t h, const std::vector<MomRefobj> &vec, RawTag)
    : MomSequence(k, h, vec.size(), vec.data()) {};
  MomSequence(SeqKind k, MomHash_t h, std::initializer_list<MomRefobj> il)
    : MomSequence(k, h, check_length(il.size()), il.begin()) {};
  MomSequence(SeqKind k, MomHash_t h, std::initializer_list<MomRefobj> il,
              RawTag)
    : MomSequence(k, h, check_length(il.size()), il.begin()) {};
  static bool good_array_refobj(const MomRefobj *arr, unsigned len)
  {
    if (len > 0 && !arr)
      return false;
    for (unsigned ix = 0; ix < len; ix++)
      if (!arr[ix])
        return false;
    return true;
  }
  static const MomRefobj *filter_array_refobj(const MomRefobj *arr,
      unsigned len)
  {
    if (!good_array_refobj(arr, len))
      {
        MOM_BACKTRACELOG("MomSequence::filter_array_refobj bad array "
                         << arr << " of length:" << len);
        throw std::runtime_error("MomSequence::filter_array_refobj bad array");
      }
    return arr;
  }
  static std::vector<MomRefobj> unique_sorted_vector_refobj(const std::vector<MomRefobj> &vec)
  {
    MOM_ASSERT(vec.size() <= MOM_SIZE_MAX, "unique_sorted_vector_refobj too big vec");
    MOM_ASSERT(good_vector_refobj(vec), "unique_sorted_vector_refobj not good vec");
    std::vector<MomRefobj> tmpvec = vec;
    std::sort(tmpvec.begin(),tmpvec.end());
    auto last = std::unique(tmpvec.begin(), tmpvec.end());
    tmpvec.erase(last,tmpvec.end());
    return tmpvec;
  }
  template<bool check=false>
  static std::vector<MomRefobj> vector_from_set_refobj(const MomSetRefobj &set)
  {
    std::vector<MomRefobj> vec;
    vec.reserve(set.size());
    for (auto rob : set)
      {
        if (check)
          {
            if (!rob)
              {
                MOM_BACKTRACELOG("MomSequence::vector_from_set_refobj null reference in set of size " << set.size());
                throw std::runtime_error("MomSequence::vector_from_set_refobj with null reference");
              }
          }
        else MOM_ASSERT(rob, "MomSequence::vector_from_set_refobj null reference in set of size " << set.size());
        vec.push_back(rob);
      }
    return vec;
  }
  static bool good_vector_refobj(const std::vector<MomRefobj> &vec)
  {
    for (MomRefobj ro : vec)
      if (!ro)
        return false;
    return true;
  };
  static bool good_sorted_array_refobj(unsigned len, const MomRefobj*arr)
  {
    if (len==0) return true;
    if (!arr) return false;
    if (!arr[0]) return false;
    for (unsigned ix=1; ix<len; ix++)
      {
        if (!arr[ix]) return false;
        if (arr[ix] <= arr[ix-1]) return false;
      }
    return true;
  }
  static const std::vector<MomRefobj> &
  filter_vector_refobj(const std::vector<MomRefobj> &vec)
  {
    if (!good_vector_refobj(vec))
      {
        MOM_BACKTRACELOG(
          "MomSequence::filter_array_refobj bad vector of size:" << vec.size());
        throw std::runtime_error("MomSequence::filter_vector_refobj bad vector");
      }
    return vec;
  }
  static bool
  good_initializer_list_refobj(std::initializer_list<MomRefobj> il)
  {
    for (MomRefobj ro : il)
      if (!ro)
        return false;
    return true;
  };
  static std::initializer_list<MomRefobj>
  filter_initializer_list_refobj(std::initializer_list<MomRefobj> il)
  {
    if (!good_initializer_list_refobj(il))
      {
        MOM_BACKTRACELOG("MomSequence::filter_initializer_list_refobj bad "
                         "initializer_list of size:"
                         << il.size());
        throw std::runtime_error(
          "MomSequence::filter_initializer_list_refobj bad initializer_list");
      }
    return il;
  }
  static std::vector<MomRefobj>
  vector_real_refs(const std::vector<MomRefobj> &vec);
  static std::vector<MomRefobj>
  vector_real_refs(const std::initializer_list<MomRefobj> il);
  template <unsigned hinit, unsigned k1, unsigned k2, unsigned k3, unsigned k4,
            bool check = false>
  static MomHash_t hash_vector_refobj(const std::vector<MomRefobj> &vec)
  {
    MomHash_t h = hinit;
    unsigned ln = vec.size();
    for (unsigned rk = 0; rk < ln; rk++)
      {
        auto ro = vec[rk];
        if (check)
          {
            if (!ro)
              {
                MOM_BACKTRACELOG("MomSequence::hash_vector_refobj null ref#"
                                 << rk << " ln=" << ln);
                throw std::runtime_error("MomSequence::hash_vector_refobj null ref");
              }
          }
        else
          MOM_ASSERT(ro, "hash_vector_refobj bad refobj#" << rk << ":" << ro);
        if (rk % 2 == 0)
          h = (h * k1) + (k3 * ro.hash());
        else
          h = (h * k2) ^ (k4 * ro.hash());
      }
    if (MOM_UNLIKELY(h == 0))
      {
        if (ln == 0)
          h = ((k1 + 7 * k2 + 19 * k3 + 317 * k4) & 0xfffff) + 11;
        else
          {
            auto firstro = vec[0];
            h = (((k1 + 19 * k2) * ln) & 0xfffff) +
                13 * (((k3 + 13 * k4) * firstro.hash()) & 0xffffff) + 1321;
          }
      }
    MOM_ASSERT(h != 0, "hash_vector_refobj zero h");
    return h;
  }; /* end hash_vector_refobj */
  template <unsigned hinit, unsigned k1, unsigned k2, unsigned k3, unsigned k4,
            bool check = false>
  static MomHash_t
  hash_initializer_list_refobj(const std::initializer_list<MomRefobj> il)
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
                MOM_BACKTRACELOG("MomSequence::hash_initializer_list_refobj null ref#"
                                 << rk << " ln=" << ln);
                throw std::runtime_error(
                  "MomSequence::hash_initializer_list_refobj null ref");
              }
          }
        else
          MOM_ASSERT(
            ro, "hash_initializer_list_refobj bad refobj#" << rk << ":" << ro);
        rk++;
        if (rk % 2 == 0)
          h = (h * k1) + (k3 * ro.hash());
        else
          h = (h * k2) ^ (k4 * ro.hash());
      }
    if (MOM_UNLIKELY(h == 0))
      {
        if (ln == 0)
          h = ((k1 + 7 * k2 + 19 * k3 + 317 * k4) & 0xfffff) + 11;
        else
          {
            auto firstro = *il.begin();
            h = (((k1 + 19 * k2) * ln) & 0xfffff) +
                13 * (((k3 + 13 * k4) * firstro.hash()) & 0xffffff) + 1321;
          }
      }
    MOM_ASSERT(h != 0, "hash_initializer_list_refobj zero h");
    return h;
  }; /* end hash_initializer_list_refobj */
  template <unsigned hinit, unsigned k1, unsigned k2, unsigned k3, unsigned k4,
            bool check = false>
  static MomHash_t hash_array_refobj(const MomRefobj *arr, unsigned len)
  {
    if (len == 0)
      return hinit;
    MOM_ASSERT(arr != nullptr, "hash_array_refobj null arr");
    MomHash_t h = hinit;
    for (unsigned rk = 0; rk < len; rk++)
      {
        auto ro = arr[rk];
        if (check)
          {
            if (!ro)
              {
                MOM_BACKTRACELOG("MomSequence::hash_array_refobj null ref#"
                                 << rk << " len=" << len);
                throw std::runtime_error("MomSequence::hash_array_refobj null ref");
              }
          }
        else
          MOM_ASSERT(ro, "hash_array_refobj bad refobj#" << rk << ":" << ro);
        if (rk % 2 == 0)
          h = (h * k1) + (k3 * ro.hash());
        else
          h = (h * k2) ^ (k4 * ro.hash());
      }
    if (MOM_UNLIKELY(h == 0))
      {
        if (len == 0)
          h = ((k1 + 7 * k2 + 19 * k3 + 317 * k4) & 0xfffff) + 11;
        else
          {
            auto firstro = arr[0];
            h = (((k1 + 19 * k2) * len) & 0xfffff) +
                13 * (((k3 + 13 * k4) * firstro.hash()) & 0xffffff) + 1321;
          }
      }
    MOM_ASSERT(h != 0, "hash_array_refobj zero h");
    return h;
  }; /* end hash_array_refobj */
public:
  typedef const MomRefobj *iterator;
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
  const MomRefobj *data() const
  {
    return _seq;
  };
  const MomRefobj *begin() const
  {
    return _seq;
  };
  const MomRefobj *end() const
  {
    return _seq + _len;
  };
  MomRefobj nth(int rk) const
  {
    if (rk < 0)
      rk += _len;
    if (rk >= 0 && rk < (int)_len)
      return _seq[rk];
    return nullptr;
  }
  MomRefobj nth(int rk, CheckTag) const
  {
    int origrk = 0;
    if (rk < 0)
      rk += _len;
    if (rk >= 0 && rk < (int)_len)
      return _seq[rk];
    MOM_BACKTRACELOG("MomSequence::nth bad rk=" << origrk
                     << " for length=" << _len);
    throw std::runtime_error("MomSequence::nth rank out of range");
  }
  MomRefobj at(unsigned ix, PlainTag) const
  {
    if (ix < _len)
      return _seq[ix];
    return nullptr;
  }
  MomRefobj at(unsigned ix) const
  {
    return at(ix, PlainTag{});
  };
  MomRefobj at(unsigned ix, RawTag) const
  {
    MOM_ASSERT(ix < _len,
               "MomSequence::at bad ix=" << ix << " for length=" << _len);
    return _seq[ix];
  }
  MomRefobj at(unsigned ix, CheckTag) const
  {
    if (ix < _len)
      return _seq[ix];
    MOM_BACKTRACELOG("MomSequence::at bad ix=" << ix << " for length=" << _len);
    throw std::runtime_error("MomSequence:at index out of range");
  }
  bool equal(const MomSequence &r) const
  {
    if (hash() != r.hash())
      return false;
    if (skind() != r.skind())
      return false;
    auto sz = size();
    if (sz != r.size())
      return false;
    for (unsigned ix = 0; ix < sz; ix++)
      if (at(ix, RawTag{}) != r.at(ix, RawTag{}))
        return false;
    return true;
  }
  bool less(const MomSequence &r) const
  {
    if (skind() < r.skind())
      return true;
    if (skind() > r.skind())
      return false;
    if (hash() == r.hash() && equal(r))
      return false;
    return std::lexicographical_compare(begin(), end(), r.begin(), r.end());
  }
  bool less_equal(const MomSequence &r) const
  {
    if (skind() < r.skind())
      return true;
    if (skind() > r.skind())
      return false;
    if (hash() == r.hash() && equal(r))
      return true;
    return std::lexicographical_compare(begin(), end(), r.begin(), r.end());
  }
  bool operator==(const MomSequence &r) const
  {
    return equal(r);
  };
  bool operator!=(const MomSequence &r) const
  {
    return equal(r);
  };
  bool operator<(const MomSequence &r) const
  {
    return less(r);
  };
  bool operator>(const MomSequence &r) const
  {
    return !less_equal(r);
  };
  bool operator<=(const MomSequence &r) const
  {
    return less_equal(r);
  };
  bool operator>=(const MomSequence &r) const
  {
    return !less(r);
  };
}; // end class MomSequence


////////////////
class MomTuple : public MomSequence
{
  static constexpr unsigned hinit = 100;
  static constexpr unsigned k1 = 233;
  static constexpr unsigned k2 = 1217;
  static constexpr unsigned k3 = 2243;
  static constexpr unsigned k4 = 139;

public:
  ~MomTuple() {};
private:
  MomTuple(void) : MomSequence(SeqKind::TupleS, hinit) {};
  MomTuple(std::nullptr_t) : MomTuple() {};
  MomTuple(std::initializer_list<MomRefobj> il, PlainTag)
    : MomSequence(SeqKind::TupleS,
                  hash_initializer_list_refobj<hinit, k1, k2, k3, k4,
                  _check_sequence_>(il),
                  il) {};
  MomTuple(std::initializer_list<MomRefobj> il, RawTag)
    : MomSequence(SeqKind::TupleS,
                  hash_initializer_list_refobj<hinit, k1, k2, k3, k4>(il),
                  il) {};
  MomTuple(std::initializer_list<MomRefobj> il, CheckTag)
    : MomSequence(SeqKind::TupleS,
                  hash_initializer_list_refobj<hinit, k1, k2, k3, k4,
                  _check_sequence_>(il),
                  il) {};
  MomTuple(const std::vector<MomRefobj> &vec, PlainTag)
    : MomSequence(
        SeqKind::TupleS,
        hash_vector_refobj<hinit, k1, k2, k3, k4, _check_sequence_>(vec),
        vec) {};
  MomTuple(const std::vector<MomRefobj> &vec, RawTag)
    : MomSequence(SeqKind::TupleS,
                  hash_vector_refobj<hinit, k1, k2, k3, k4>(vec), vec) {};
  MomTuple(const std::vector<MomRefobj> &vec, CheckTag)
    : MomSequence(
        SeqKind::TupleS,
        hash_vector_refobj<hinit, k1, k2, k3, k4, _check_sequence_>(vec),
        vec) {};
  MomTuple(const std::vector<MomRefobj> &vec) : MomTuple(vec, CheckTag{}) {};
  template <typename... Args>
  MomTuple(AnyTag, Args... args)
    : MomTuple(MomRefobj::reserve_vector(3 + (2 * sizeof...(args)), args...),
               CheckTag{}) {};
  MomTuple(const MomSequence &sq)
    : MomSequence(SeqKind::TupleS,
                  sq.is_tuple() ? sq.hash()
                  : hash_array_refobj<hinit, k1, k2, k3, k4>(
                    sq.data(), sq.size()),
                  sq.size(), sq.data()) {};
  MomTuple(const MomSequence &sq, CheckTag)
    : MomSequence(
        SeqKind::TupleS,
        sq.is_tuple()
        ? sq.hash()
        : hash_array_refobj<hinit, k1, k2, k3, k4, _check_sequence_>(
          sq.data(), sq.size()),
        sq.size(), sq.data()) {};
  MomTuple(CheckTag) : MomTuple() {};
  MomTuple(RawTag) : MomTuple() {};
  MomTuple(PlainTag) : MomTuple() {};
  template <typename... Args>
  MomTuple(CheckTag tag, Args... args)
    : MomTuple(std::initializer_list<MomRefobj>
  {
    args...
  }, tag) {};
  template <typename... Args>
  MomTuple(RawTag tag, Args... args)
    : MomTuple(std::initializer_list<MomRefobj>
  {
    args...
  }, tag) {};
  static void fill_vector(std::vector<MomRefobj>&) {};
  static void reserve_vector (std::vector<MomRefobj>&vec, size_t siz)
  {
    vec.reserve(vec.size() + siz);
  };
  static void add_to_vector(std::vector<MomRefobj>&vec, const MomVal val);
  template  <typename... Args>
  static void fill_vector(std::vector<MomRefobj>&vec, const MomRefobj rob, Args... args)
  {
    if (rob) vec.push_back(rob);
    fill_vector(vec,args...);
  };
  template  <typename... Args>
  static void fill_vector(std::vector<MomRefobj>&vec, const std::set<MomRefobj>& rset, Args... args)
  {
    reserve_vector(vec, rset.size()+sizeof...(args));
    for (auto rob: rset)
      if (rob) vec.push_back(rob);
    fill_vector(vec,args...);
  };
  template  <typename... Args>
  static void fill_vector(std::vector<MomRefobj>&vec, const std::vector<MomRefobj>& rvec, Args... args)
  {
    reserve_vector(vec, rvec.size()+sizeof...(args));
    for (auto rob: rvec)
      if (rob) vec.push_back(rob);
    fill_vector(vec,args...);
  };
  template  <typename... Args>
  static void fill_vector(std::vector<MomRefobj>&vec, const std::initializer_list<MomRefobj>& il, Args... args)
  {
    reserve_vector(vec, il.size()+sizeof...(args));
    for (auto rob: il)
      if (rob) vec.push_back(rob);
    fill_vector(vec,args...);
  }
  template  <typename... Args>
  static void inline fill_vector(std::vector<MomRefobj>&vec, const MomVal val, Args... args);
public:
  static const MomTuple*make(const std::vector<MomRefobj> &vec)
  {
    return new MomTuple(vec,CheckTag{});
  };
  static const MomTuple*make(std::initializer_list<MomRefobj> il)
  {
    return new MomTuple(il,CheckTag{});
  }
  static const MomTuple*make(const MomSequence&sq)
  {
    return new MomTuple(sq,CheckTag{});
  };
  static const MomTuple*make(const MomSequence*psq)
  {
    return psq?new MomTuple(*psq,CheckTag{}):new MomTuple(nullptr);
  };
  template <typename... Args>
  static const MomTuple*make(Args... args)
  {
    return new MomTuple(CheckTag{},args...);
  }
  template <typename... Args>
  static const MomTuple*make_any(Args... args)
  {
    std::vector<MomRefobj> vec;
    vec.reserve(4*sizeof...(args)/3+3);
    fill_vector(vec,args...);
    return make(vec);
  }
}; // end class MomTuple


////////////////
class MomSet  : public MomSequence
{
  friend class MomVSet;
  static constexpr unsigned hinit = 301;
  static constexpr unsigned k1 = 467;
  static constexpr unsigned k2 = 3671;
  static constexpr unsigned k3 = 1367;
  static constexpr unsigned k4 = 569;
  MomSet(MomSet&&set, RawTag tag)
    : MomSequence(SeqKind::SetS,
                  set._hash,
                  set._len,
                  set._seq,
                  tag)
  {
    *const_cast<SeqKind*>(&set._skd) = SeqKind::NoneS;
    *const_cast<MomRefobj**>(&set._seq) = nullptr;
    *const_cast<unsigned*>(&set._len) = 0;
    *const_cast<MomHash_t*>(&set._hash) = 0;
  };
  MomSet(MomSet&&set) : MomSet(std::move(set),RawTag{}) {};
  struct SortedTag {};
  MomSet(const std::vector<MomRefobj>& vec, SortedTag)
    : MomSequence(SeqKind::SetS,
                  hash_vector_refobj<hinit,k1,k2,k3,k4>(vec),
                  vec)
  {
    MOM_ASSERT(good_sorted_array_refobj(_len,_seq), "unsorted or bad MomSet of length:" << _len);
  };
  static MomSet make_from_vector(const std::vector<MomRefobj>& vec)
  {
    auto svec = unique_sorted_vector_refobj(filter_vector_refobj(vec));
    return MomSet(svec,SortedTag{});
  };
  static MomSet make_from_set(const MomSetRefobj& set)
  {
    auto svec = vector_from_set_refobj<_check_sequence_>(set);
    return MomSet(svec,SortedTag{});
  }
  MomSet(void) : MomSequence(SeqKind::SetS, hinit) {};
  MomSet(std::nullptr_t) : MomSet() {};
  MomSet(const std::vector<MomRefobj>& vec, PlainTag)
    : MomSet(std::move(make_from_vector(vec))) {};
  MomSet(std::initializer_list<MomRefobj> il, PlainTag)
    : MomSet(std::move(make_from_vector(std::vector<MomRefobj>
  {
    il
  }))) {};
  MomSet(const MomSetRefobj& set, PlainTag)
    : MomSet(std::move(make_from_set(set))) {};
  static void fill_set(std::set<MomRefobj>&);
  static void add_to_set(std::set<MomRefobj>&set, const MomVal val);
  template  <typename... Args>
  static void fill_set(std::set<MomRefobj>&set, const MomRefobj rob, Args... args)
  {
    if (rob) set.insert(rob);
    fill_set(set,args...);
  };
  template  <typename... Args>
  static void fill_set(std::set<MomRefobj>&set, const std::set<MomRefobj>& rset, Args... args)
  {
    for (auto rob: rset)
      if (rob) set.insert(rob);
    fill_set(set,args...);
  };
  template  <typename... Args>
  static void fill_set(std::set<MomRefobj>&set, const std::vector<MomRefobj>& rvec, Args... args)
  {
    for (auto rob: rvec)
      if (rob) set.insert(rob);
    fill_set(set,args...);
  };
  template  <typename... Args>
  static void fill_set(std::set<MomRefobj>&set, const std::initializer_list<MomRefobj>& il, Args... args)
  {
    for (auto rob: il)
      if (rob) set.insert(rob);
    fill_set(set,args...);
  }
  template  <typename... Args>
  static void inline fill_set(std::set<MomRefobj>&set, const MomVal val, Args... args);
public:
  ~MomSet() {};
  static const MomSet*make(void)
  {
    return new MomSet(nullptr);
  };
  static const MomSet*make(const std::vector<MomRefobj>& vec)
  {
    return new MomSet(vec,PlainTag{});
  };
  static const MomSet*make(const MomSetRefobj& set)
  {
    return new MomSet(set,PlainTag{});
  };
  static const MomSet*make(const std::initializer_list<MomRefobj>& il)
  {
    return new MomSet(il,PlainTag{});
  };
  template <typename... Args>
  static const MomSet*make(Args... args)
  {
    return new MomSet(std::initializer_list<MomRefobj> {args...},PlainTag());
  };
  template <typename... Args>
  static const MomSet*make_any(Args... args)
  {
    std::set<MomRefobj> set;
    fill_set(set,args...);
    return make(set);
  }
};        // end class MomSet


class MomString : public std::enable_shared_from_this<MomString>
{
  const std::string _str;
  const MomHash_t _hash;
  static constexpr unsigned hinit = 600;
  static constexpr unsigned k1 = 277;
  static constexpr unsigned k2 = 631;
  static constexpr unsigned k3 = 733;
  static constexpr unsigned k4 = 839;
  static constexpr unsigned kmod = 12157;
  friend class MomVal;
  friend class MomVString;
  void clear(void)
  {
    const_cast<std::string*>(&_str)->clear();
    *const_cast<MomHash_t*>(&_hash) = 0;
  }
  static unsigned normalize_length(const char*str, int len)
  {
    if (len<0) return str?strlen(str):0;
    else return len;
  };
public:
  size_t size() const
  {
    return _str.size();
  };
  char unsafe_at(unsigned ix) const
  {
    return _str[ix];
  };
  char nth(int ix, bool check=false) const
  {
    auto ln = size();
    if (ix<0) ix+=ln;
    if (ix>=0 && ix<(int)ln) return _str[ix];
    if (check)
      {
        MOM_BACKTRACELOG("MomString::nth invalid index " << ix);
        throw std::runtime_error("MomString::nth invalid index");
      }
    return (char)0;
  }
  const std::string to_string() const
  {
    return _str;
  };
  operator const std::string () const
  {
    return to_string();
  };
  const char* to_cstr() const
  {
    return _str.c_str();
  };
  static MomHash_t hash_of_cstr(const char*str, int len= -1)
  {
    if (!str) return 0;
    if (len<0) len = strlen(str);
    MomHash_t h1 = hinit, h2 = 0;
    for (unsigned ix=0; ix<(unsigned)len; ix++)
      {
        if (ix %2 == 0) h1 = (k1 * h1 + h2 % kmod) ^ (k2 * str[ix]);
        else h2 = ((k3 * h2 - h1 % kmod) ^ (k4 * str[ix])) + k1;
      }
    MomHash_t h = (11*h1) ^ (7*h2);
    if (MOM_UNLIKELY(h==0))
      h = (h1 & 0xffff) + (h2 & 0xffff) + (3*len&0xff) + hinit + 10;
    return h;
  };
  static MomHash_t hash_of_string(const std::string&s)
  {
    return hash_of_cstr(s.c_str(), s.size());
  };
  MomString(const std::string& s) : std::enable_shared_from_this<MomString>(),
    _str{s}, _hash(hash_of_string(s)) {};
  MomString(const MomString&ms): std::enable_shared_from_this<MomString>(),
    _str(ms._str), _hash(ms._hash) {};
  MomString(MomString&&ms): std::enable_shared_from_this<MomString>(),
    _str(std::move(ms._str)), _hash(ms._hash)
  {
    ms.clear();
  };
  MomString(const char*str, int len= -1)
    : std::enable_shared_from_this<MomString>(),
      _str{str,normalize_length(str,len)}, _hash(hash_of_string(_str)) {};
  MomHash_t hash() const
  {
    return _hash;
  };
  bool equal(const MomString&r) const
  {
    if (&r == this) return true;
    return _hash==r._hash && _str==r._str;
  };
  bool equal(const std::string& s) const
  {
    return _str==s;
  };
  bool equal(const char*cs) const
  {
    return cs && !strcmp(_str.c_str(),cs);
  };
  bool less(const MomString&r) const
  {
    return _str<r._str;
  };
  bool less(const std::string&rs) const
  {
    return _str<rs;
  };
  bool less_equal(const std::string&rs) const
  {
    return _str<=rs;
  };
  bool less_equal(const MomString&r) const
  {
    return _str<=r._str;
  };
  bool operator == (const MomString&r) const
  {
    return equal(r);
  };
  bool operator != (const MomString&r) const
  {
    return !equal(r);
  };
  bool operator < (const MomString&r) const
  {
    return less(r);
  };
  bool operator <= (const MomString&r) const
  {
    return less_equal(r);
  };
  bool operator > (const MomString&r) const
  {
    return !less_equal(r);
  };
  bool operator >= (const MomString&r) const
  {
    return !less(r);
  };
};        // end class MomString



class MomVNone : public MomVal
{
public:
  MomVNone() : MomVal(TagNone{}, nullptr) {};
  ~MomVNone() = default;
}; // end MomVNone

class MomVInt : public MomVal
{
public:
  MomVInt(int64_t i = 0) : MomVal(TagInt{}, i) {};
  ~MomVInt() = default;
}; // end MomVInt



class MomVString : public MomVal
{
  static std::ostringstream& out(std::ostringstream&o)
  {
    return o;
  };
  static void output(std::ostringstream&) {};
  template <typename T>
  static void output(std::ostringstream&o, const T x)
  {
    o << x;
  };
  template <typename T, typename... Args>
  static void output(std::ostringstream&o, const T x, Args... args)
  {
    output(o,x);
    output(o,args...);
  };
public:
  MomVString() : MomVal(nullptr) {};
  MomVString(const char *s, int l = -1)
    : MomVal(TagString{},new MomString(std::string{s,MomString::normalize_length(s,l)})) {};
  MomVString(const MomString &ms)
    : MomVal(TagString{},&ms) {};
  MomVString(const std::string &str)
    : MomVal(TagString{},new MomString(str)) {};
  MomVString(const std::ostringstream&os)
    : MomVal(TagString{},new MomString(os.str())) {};
  ~MomVString() = default;
  template <typename... Args>
  static MomVString make(Args... args)
  {
    std::ostringstream os;
    output(os,args...);
    return MomVString(os);
  };
}; // end MomVString


class MomVRef : public MomVal
{
public:
  MomVRef(const MomRefobj ro) : MomVal(TagRefobj{},ro)
  {
  };
  ~MomVRef() = default;
};        // end MomVRef

class MomVColoRef : public MomVal
{
public:
  MomVColoRef(const MomRefobj ob, const MomRefobj colorob)
    : MomVal(TagColoRef{},ob,colorob) {};
  ~MomVColoRef() = default;
}; // end MomVColoRef


class MomVSet : public MomVal
{
  friend class MomVal;
  friend class MomSet;
  friend class MomSequence;
public:
  ~MomVSet() = default;
  inline MomVSet(void);
  inline MomVSet(const MomSet &bs)
    : MomVal(TagSet{},&bs) {};
  inline MomVSet(const MomSetRefobj&setr)
    : MomVal(TagSet{}, MomSet::make_from_set(setr)) {};
  inline MomVSet(const MomUnorderedSetRefobj&);
  MomVSet(const std::vector<MomRefobj> &vec)
    : MomVal(TagSet{}, MomSet::make(vec)) {};
  MomVSet(const std::initializer_list<MomRefobj> &il)
    :  MomVal(TagSet{}, MomSet::make(il)) {};
  template <typename... Args>
  static MomVSet make_obj(Args... args)
  {
    return MomVSet(std::initializer_list<MomRefobj>(args...));
  };
  template <typename... Args>
  static MomVSet make_any(Args... args)
  {
    return MomVSet(MomSet::make_any(args...));
  };
}; // end MomVSet


class MomVTuple : public MomVal
{
  friend class MomVal;
  friend class MomTuple;
  friend class MomSequence;
public:
  ~MomVTuple() = default;
  inline MomVTuple(const MomTuple &);
  inline MomVTuple(void);
  MomVTuple(const std::vector<MomRefobj> &vec)
    : MomVal(TagTuple{}, MomTuple::make(vec)) {};
  MomVTuple(const std::initializer_list<MomRefobj>&il)
    : MomVal(TagTuple{}, MomTuple::make(il)) {};
  template <typename... Args>
  static MomVTuple make_obj(Args... args)
  {
    return MomVTuple(std::initializer_list<MomRefobj>(args...));
  };
  template <typename... Args>
  static MomVTuple make_any(Args... args)
  {
    return MomVTuple(MomTuple::make_any(args...));
  };
}; // end MomVTuple


class MomJsonParser
{
protected:
  MomJsonParser() {};
public:
  virtual ~MomJsonParser();
  virtual MomRefobj idstr_to_refobj(const std::string&) =0;
};        // end class MomJsonParser

class MomJsonEmitter
{
protected:
  MomJsonEmitter() {};
public:
  virtual ~MomJsonEmitter();
  virtual bool emittable_refobj(MomRefobj) =0;
};        // end class MomJsonEmitter




////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////
/***************** INLINE FUNCTIONS ****************/
MomSerial63::MomSerial63(uint64_t n, bool nocheck) : _serial(n)
{
  if (nocheck || n == 0)
    return;
  if (n < _minserial_)
    {
      MOM_BACKTRACELOG("MomSerial63 too small n:" << n);
      throw std::runtime_error("MomSerial63 too small n");
    };
  if (n > _maxserial_)
    {
      MOM_BACKTRACELOG("MomSerial63 too big n:" << n);
      throw std::runtime_error("MomSerial63 too big n");
    }
} /* end MomSerial63::MomSerial63 */


/// see also http://stackoverflow.com/a/28613483/841108
void MomVal::clear()
{
  auto k = _kind;
  *const_cast<MomVKind *>(&_kind) = MomVKind::NoneK;
  switch (k)
    {
    case MomVKind::NoneK:
      break;
    case MomVKind::IntK:
      _int = 0;
      break;
    case MomVKind::StringK:
      _str.~shared_ptr<const MomString>();
      ;
      break;
    case MomVKind::RefobjK:
      _ref.clear();
      break;
    case MomVKind::SetK:
      _set.~shared_ptr<const MomSet>();
    case MomVKind::TupleK:
      _tup.~shared_ptr<const MomTuple>();
      break;
    case MomVKind::ColoRefK:
      *const_cast<MomRefobj*>(&_coloref._cobref) = nullptr;
      *const_cast<MomRefobj*>(&_coloref._colorob) = nullptr;
      break;
    }
  _bothptr[0] = nullptr;
  _bothptr[1] = nullptr;
} // end MomVal::clear()

MomVal&
MomVal::operator =(const MomVal&sv)
{
  auto k = sv._kind;
  switch (k)
    {
    case MomVKind::NoneK:
      _ptr = nullptr;
      break;
    case MomVKind::IntK:
      _int = sv._int;
      break;
    case MomVKind::StringK:
      _str = (sv._str);
      break;
    case MomVKind::RefobjK:
      _ref = (sv._ref);
      break;
    case MomVKind::ColoRefK:
      _coloref = (sv._coloref);
      break;
    case MomVKind::SetK:
      _set = (sv._set);
      break;
    case MomVKind::TupleK:
      _tup = (sv._tup);
      break;
    }
  return *this;
} // end MomVal::operator =(const MomVal&sv)

MomVal::MomVal(MomVal&&sv) : _kind(sv._kind)
{
  switch (_kind)
    {
    case MomVKind::NoneK:
      _ptr = nullptr;
      break;
    case MomVKind::IntK:
      _int = sv._int;
      break;
    case MomVKind::StringK:
      _str = std::move(sv._str);
      break;
    case MomVKind::RefobjK:
      _ref = std::move(sv._ref);
      break;
    case MomVKind::ColoRefK:
      _coloref = std::move(sv._coloref);
      break;
    case MomVKind::SetK:
      _set = std::move(sv._set);
      break;
    case MomVKind::TupleK:
      _tup = std::move(sv._tup);
      break;
    }
  sv._bothptr[0] = nullptr;
  sv._bothptr[1] = nullptr;
  *const_cast<MomVKind *>(&sv._kind) =MomVKind::NoneK;
} // end MomVal::Momval(MomVal&&sv)

MomVal&
MomVal::operator=(MomVal&&sv)
{
  switch (_kind)
    {
    case MomVKind::NoneK:
      _ptr = nullptr;
      break;
    case MomVKind::IntK:
      _int = sv._int;
      break;
    case MomVKind::StringK:
      _str = std::move(sv._str);
      break;
    case MomVKind::RefobjK:
      _ref = std::move(sv._ref);
      break;
    case MomVKind::ColoRefK:
      _coloref = std::move(sv._coloref);
      break;
    case MomVKind::SetK:
      _set = std::move(sv._set);
      break;
    case MomVKind::TupleK:
      _tup = std::move(sv._tup);
      break;
    }
  sv._bothptr[0] = nullptr;
  sv._bothptr[1] = nullptr;
  *const_cast<MomVKind *>(&sv._kind) =MomVKind::NoneK;
  return *this;
}// end MomVal::operator=(MomVal&&sv)


bool
MomRefobj::less(const MomRefobj r) const
{
  if (!r)
    return false;
  if (unsafe_get_const() == r.unsafe_get_const())
    return false;
  if (!unsafe_get_const())
    return true;
  return unsafe_get_const()->less(r);
} // end MomRefobj::less

MomHash_t MomRefobj::hash(void) const
{
  auto pob = unsafe_get_const();
  if (!pob)
    return 0;
  return pob->hash();
} // end MomRefobj::hash

bool MomRefobj::less_equal(const MomRefobj r) const
{
  if (unsafe_get_const() == r.unsafe_get_const())
    return true;
  if (!unsafe_get_const())
    return true;
  return unsafe_get_const()->less_equal(r);
} // end MomRefobj::less_equal

bool MomRefobj::equal(const MomRefobj r) const
{
  return (unsafe_get_const() == r.unsafe_get_const());
} // end MomRefobj::equal

const MomPairid MomObject::random_id(void)
{
  return MomPairid{MomSerial63::make_random(), MomSerial63::make_random()};
} // end MomObject::random_id

const MomPairid MomObject::random_id_of_bucket(unsigned bucknum)
{
  return MomPairid{MomSerial63::make_random_of_bucket(bucknum),
                   MomSerial63::make_random()};
} // end MomObject::random_id_of_bucket

void MomRefobj::collect_vector_sequence(std::vector<MomRefobj> &vec,
                                        const MomSequence &seq)
{
  vec.reserve(vec.size() + seq.size());
  for (auto rob : seq)
    {
      MOM_ASSERT(rob, "collect_vector_sequence null rob");
      collect_vector_refobj(vec, rob);
    }
} // end MomRefobj::collect_vector_sequence

std::size_t
MomRefobj::longhash() const
{
  if (_ptrobj) return _ptrobj->longhash();
  return 0;
}

std::size_t
MomObject::longhash() const
{
  return std::hash<MomPairid>()(_obserpair);
}

void MomRefobj::add_set_sequence(std::set<MomRefobj> &set,
                                 const MomSequence &seq)
{
  for (auto rob : seq)
    if (rob)
      set.insert(rob);
} // end MomRefobj::add_set_sequence

template  <typename... Args>
void MomSet::fill_set(std::set<MomRefobj>&set, const MomVal val, Args... args)
{
  if (val) add_to_set(set, val);
  fill_set(set,args...);
} /* end MomSet::fill_set */

template  <typename... Args>
void MomTuple::fill_vector(std::vector<MomRefobj>&vec, const MomVal val, Args... args)
{
  reserve_vector(vec,4*sizeof...(args)/3 + 2);
  if (val) add_to_vector(vec, val);
  fill_vector(vec,args...);
} /* end MomTuple::fill_vector */

intptr_t MomVal::as_int(void) const
{
  if (!is_int())
    {
      MOM_BACKTRACELOG("MomVal::as_int not an int");
      throw std::runtime_error("MomVal::as_int not an int");
    };
  return unsafe_int();
}      // end MomVal::as_int

std::shared_ptr<const MomString>
MomVal::as_bstring(void) const
{
  if (!is_string())
    {
      MOM_BACKTRACELOG("MomVal::as_bstring not a string");
      throw std::runtime_error("MomVal::as_bstring is not a string");
    };
  return _str;
} // end MomVal::as_bstring

std::shared_ptr<const MomString>
MomVal::to_bstring(const std::shared_ptr<const MomString> &pbs) const
{
  if (!is_string()) return pbs;
  return _str;
} // end MomVal::to_bstring

const std::string
MomVal::as_string(void) const
{
  if (!is_string()) return nullptr;
  MOM_ASSERT(_str, "corrupted string value");
  return _str->to_string();
} // end MomVal::as_string

const std::string
MomVal::to_string(const std::string &str) const
{
  if (!is_string()) return str;
  MOM_ASSERT(_str, "corrupted string value");
  return _str->to_string();
} // end MomVal::to_string

const char*
MomVal::to_cstr(const char*defcstr) const
{
  if (!is_string()) return defcstr;
  MOM_ASSERT(_str, "corrupted string value");
  return _str->to_cstr();
} // end MomVal::to_cstr

bool
MomVal::equal(const MomVal&r) const
{
  if (this==&r) return true;
  auto k = kind();
  if (k != r.kind()) return false;
  switch (k)
    {
    case MomVKind::NoneK:
      return true;
    case MomVKind::IntK:
      return _int == r._int;
    case MomVKind::RefobjK:
      return _ref == r._ref;
    case MomVKind::ColoRefK:
      return _coloref._cobref == r._coloref._cobref &&  _coloref._colorob == r._coloref._colorob;
    case MomVKind::StringK:
    {
      MOM_ASSERT (_str, "MomVal::equal bad _str");
      MOM_ASSERT (r._str, "MomVal::equal bad r._str");
      return _str->equal(*r._str);
    }
    case MomVKind::TupleK:
    {
      MOM_ASSERT(_tup, "MomVal::equal bad tup");
      MOM_ASSERT(r._tup, "MomVal::equal bad r._tup");
      return _tup->equal(*r._tup);
    }
    case MomVKind::SetK:
    {
      MOM_ASSERT(_set, "MomVal::equal bad set");
      MOM_ASSERT(r._set, "MomVal::equal bad r._set");
      return _set->equal(*r._set);
    }
    }
}      // end MomVal::equal

MomVal::MomVal(const MomVal&sv) : MomVal()
{
  auto k = sv.kind();
  switch (k)
    {
    case MomVKind::NoneK:
      return;
    case MomVKind::IntK:
      _int = sv._int;
      break;
    case MomVKind::RefobjK:
      MOM_ASSERT(sv._ref, "bad source for MomVal");
      _ref = sv._ref;
      break;
    case MomVKind::ColoRefK:
      MOM_ASSERT(sv._coloref._cobref, "bad source coloref for MomVal");
      MOM_ASSERT(sv._coloref._colorob, "bad source colorob for MomVal");
      _coloref = sv._coloref;
      break;
    case MomVKind::StringK:
      MOM_ASSERT (sv._str, "bad source _str for MomVal");
      _str = sv._str;
      break;
    case MomVKind::SetK:
      MOM_ASSERT(sv._set, "bad source _set for MomVal");
      _set = sv._set;
      break;
    case MomVKind::TupleK:
      MOM_ASSERT(sv._tup, "bad source _tup for MomVal");
      _tup = sv._tup;
      break;
    }
  _kind = k;
} // end MomVal::MomVal(const MomVal&sv)


std::ostream&operator << (std::ostream&os, const MomObject& ob)
{
  os << ob.ident();
  return os;
}

std::ostream&operator << (std::ostream&os, const MomPairid pi)
{
  if (!pi) os << "__";
  else
    os << pi.first << pi.second;
  return os;
}
#endif /*MONIMELT_HEADER*/
