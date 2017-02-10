// file objectmo.cc
#include "monimelt.h"

std::array<MomObject::ObjBucket,MomSerial63::_maxbucket_> MomObject::_buckarr_;

MomSetRefobj MomObject::_predefined_set_;
std::mutex MomObject::_predefined_mtx_;

MomObject*
MomObject::ObjBucket::find_object_in_bucket(const MomPairid id) const
{
  std::lock_guard<std::mutex> _gu(_bumtx);
  auto it = _bumap.find(id);
  if (it == _bumap.end()) return nullptr;
  return it->second;
} // end MomObject::ObjBucket::find_object_in_bucket

void
MomObject::ObjBucket::register_object_in_bucket(MomObject*ob)
{
  MOM_ASSERT(ob != nullptr, "register_object_in_bucket no ob");
  auto id = ob->ident();
  std::lock_guard<std::mutex> _gu(_bumtx);
  MOM_ASSERT(_bumap.find(id) == _bumap.end(), "already registered id " << id);
  _bumap.insert({id, ob});
}

void
MomObject::ObjBucket::unregister_object_in_bucket(MomObject*ob)
{
  MOM_ASSERT(ob != nullptr, "unregister_object_in_bucket no ob");
  auto id = ob->ident();
  std::lock_guard<std::mutex> _gu(_bumtx);
  auto it = _bumap.find(id);
  MOM_ASSERT(it != _bumap.end(), "already registered id " << id);
  MOM_ASSERT(it->second == ob, "corrupted bucket");
  _bumap.erase(it);
}

MomObject::MomObject(TagNewObject, MomPairid id, MomSpace sp)
  : _obserpair(id), _obmtx{}, _obmtime{0}, _obspace(sp), _obattrmap{}, _obcompvec{}, _obpayload(nullptr)
{
  _buckarr_[id_bucketnum(id)].register_object_in_bucket(this);
};

void
MomObject::initialize_predefined(void)
{
  static bool inited;
  MOM_ASSERT(inited==false, "already inited");
  inited=true;
#define MOM_HAS_PREDEF(Id,S1,S2,H) do {         \
  mompredef##Id =                               \
    new MomObject(TagNewObject{},               \
      MomPairid{S1,S2},                         \
      MomSpace::PredefinedSp);                  \
  MOM_ASSERT(mompredef##Id->hash() == (H),      \
       "corrupted hash for predefined");        \
  } while(0);
#include "_mompredef.h"
} // end MomObject::initialize_predefined

std::string
MomObject::id_to_string(const MomPairid pi)
{
  if (!pi) return std::string{"__"};
  return pi.first.to_string() + pi.second.to_string();
} // end MomObject::id_to_string



const MomPairid
MomObject::id_from_cstr(const char*s, const char*&end, bool fail)
{
  const char*endleft = nullptr;
  const char*endright = nullptr;
  if (s&&s[0] == '_' && s[1]=='_')
    {
      if (end) end= s+2;
      return MomPairid{nullptr,nullptr};
    }
  auto sleft = MomSerial63::make_from_cstr(s,endleft,fail);
  auto sright = MomSerial63::make_from_cstr(endleft,endright,fail);
  if (!sleft && sright)
    {
      if (fail)
        {
          std::string str{s};
          if (str.size() >= 2*MomSerial63::_nbdigits_+2)
            str.resize(2*MomSerial63::_nbdigits_+2);
          MOM_BACKTRACELOG("MomObject::id_from_cstr bad str="<<str);
          throw std::runtime_error("MomObject::id_from_cstr bad str");
        }
      end = s;
      return MomPairid{nullptr,nullptr};
    }
  end = endright;
  return MomPairid{sleft,sright};
} // end of MomObject::id_from_cstr


/// this function is rarely called, only when the simple hash computed
/// by hash_id - that is ls^(rs>>2) - is 0
MomHash_t
MomObject::hash0pairid(const MomPairid pi)
{
  auto ls = pi.first.serial();
  auto rs = pi.second.serial();
  MOM_ASSERT(ls != 0 || rs != 0, "hash0pairid zero pi");
  MomHash_t h {(MomHash_t)((ls<<3) ^ (rs*5147))};
  if (MOM_LIKELY(h != 0))
    return h;
  h = 17*(MomHash_t)(ls % 504677LL) + (MomHash_t)(rs % 11716949LL) + 31;
  MOM_ASSERT(h!=0, "hash0pairid zero h");
  return h;
} // end MomObject::hash0pairid


bool
MomObject::scan_inside_object(const std::function<bool(MomRefobj)>&f,
                              const std::function<bool(MomRefobj,MomRefobj)>&filterf) const
{
  for (auto p : _obattrmap)
    {
      const MomRefobj atob = p.first;
      MomVal aval = p.second;
      if (!filterf(this,atob)) continue;
      if (!f(atob))
        if (!aval.scan_objects(f))
          return false;
    }
  for (auto cval : _obcompvec)
    if (!cval.scan_objects(f))
      return false;
  auto py = get_payload_ptr();
  if (py && !py->scan_objects(f))
    return false;
  return true;
} // end MomObject::scan_inside_objects

void
MomObject::add_predefined(MomRefobj ob)
{
  std::lock_guard<std::mutex> _gu(_predefined_mtx_);
  _predefined_set_.insert(ob);
} // end of MomObject::add_predefined

void
MomObject::remove_predefined(MomRefobj ob)
{
  std::lock_guard<std::mutex> _gu(_predefined_mtx_);
  _predefined_set_.erase(ob);
} // end of MomObject::remove_predefined


MomVal
MomObject::set_of_predefined(void)
{
  std::lock_guard<std::mutex> _gu(_predefined_mtx_);
  return MomVSet( _predefined_set_);
} // end  MomObject::set_of_predefined

void
MomObject::set_space(MomSpace sp)
{
  if (sp == _obspace) return;
  if (_obspace == MomSpace::PredefinedSp) remove_predefined(this);
  if (sp == MomSpace::PredefinedSp) add_predefined(this);
  _obspace = sp;
} // end of MomObject::set_space

MomJson
MomObject::json_for_content(MomJsonEmitter&jem) const
{
#warning MomObject::json_for_content to be coded
} // end of MomObject::json_for_content

void
MomObject::fill_content_from_json(const MomJson&job, MomJsonParser&jp)
{
#warning MomObject::fill_content_from_json to be coded
} // end MomObject::fill_content_from_json

#define MOM_HAS_PREDEF(Id,S1,S2,H) MomRefobj mompredef##Id;
#include "_mompredef.h"
