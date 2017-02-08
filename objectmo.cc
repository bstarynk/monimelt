// file objectmo.cc
#include "monimelt.h"

std::array<MomObject::ObjBucket,MomSerial63::_maxbucket_> MomObject::_buckarr_;

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
