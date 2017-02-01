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
