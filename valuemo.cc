// file valuemo.cc
#include "monimelt.h"


std::vector<MomRefobj>
MomSequence::vector_real_refs(const std::vector<MomRefobj>& vec)
{
  std::vector<MomRefobj> res;
  res.reserve(vec.size());
  for (auto ro : vec)
    {
      if (ro)
        res.push_back(ro);
    }
  res.shrink_to_fit();
  return res;
} // end of  MomSequence::vector_real_refs

std::vector<MomRefobj>
MomSequence::vector_real_refs(const std::initializer_list<MomRefobj> il)
{
  std::vector<MomRefobj> res;
  res.reserve(il.size());
  for (auto ro : il)
    {
      if (ro)
        res.push_back(ro);
    }
  res.shrink_to_fit();
  return res;
} // end of MomSequence::vector_real_refs
