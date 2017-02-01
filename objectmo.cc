// file objectmo.cc
#include "monimelt.h"

std::string
MomObject::id_to_string(const pairid_t pi)
{
  return pi.first.to_string() + pi.second.to_string();
} // end MomObject::id_to_string



const MomObject::pairid_t
MomObject::id_from_cstr(const char*s, const char*&end, bool fail)
{
  const char*endleft = nullptr;
  const char*endright = nullptr;
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
      return pairid_t{nullptr,nullptr};
    }
  end = endright;
  return pairid_t{sleft,sright};
} // end of MomObject::id_from_cstr


/// this function is rarely called, only when the simple hash computed
/// by hash_id - that is ls^(rs>>2) - is 0
MomHash_t
MomObject::hash0pairid(const pairid_t pi)
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
