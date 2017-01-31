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

class MomSerial63
{
  uint64_t _serial;
public:
  static constexpr const uint64_t _minserial_ = 1024;
  static constexpr const uint64_t _deltaserial_ =
    (uint64_t)10 * 62*62*62 * 62*62*62 * 62*62*62 * 62;
  static constexpr const uint64_t _maxserial_ =
    _minserial_ + _deltaserial_;
  static_assert(_maxserial_ < ((uint64_t)1<<63),
                "corrupted _maxserial_ in MomSerial63");
  static_assert(_deltaserial_ > ((uint64_t)1<<62),
                "corrupted _deltaserial_ in MomSerial63");
  static constexpr const unsigned _maxbucket_ = 10*62;
  inline MomSerial63(uint64_t n=0, bool nocheck=false);
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
  static const MomSerial63 make_random(void);
  static const MomSerial63 make_random_of_bucket(unsigned bun);
  MomSerial63(const MomSerial63&s) : _serial(s._serial) {};
  MomSerial63(MomSerial63&& s) : _serial(std::move(s._serial)) { };
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
class MomObject;
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
  MomObject* unsafe_get_const(void)
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
};    // end class MomObject

class MomSequence;

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
#endif /*MONIMELT_HEADER*/
