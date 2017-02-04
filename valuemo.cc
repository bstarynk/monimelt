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


void
MomSet::add_to_set(std::set<MomRefobj>&set, const MomVal val)
{
  switch (val.kind())
    {
    case MomVKind::NoneK:
    case MomVKind::IntK:
    case MomVKind::StringK:
      return;
    case MomVKind::RefobjK:
    {
      auto rob = val.unsafe_refobj();
      MOM_ASSERT (rob, "add_to_set no rob");
      set.insert(rob);
    }
    return;
    case MomVKind::SetK:
    case MomVKind::TupleK:
    {
      auto seq = val.unsafe_sequence();
      MOM_ASSERT(seq != nullptr, "add_to_set no seq");
      for (auto rob: *seq)
        {
          MOM_ASSERT(rob, "add_to_set no rob in seq");
          set.insert(rob);
        }
    }
    return;
    }
} // end MomSet::add_to_set

void
MomTuple::add_to_vector(std::vector<MomRefobj>&vec, const MomVal val)
{
  switch (val.kind())
    {
    case MomVKind::NoneK:
    case MomVKind::IntK:
    case MomVKind::StringK:
      return;
    case MomVKind::RefobjK:
    {
      auto rob = val.unsafe_refobj();
      MOM_ASSERT (rob, "add_to_vector no rob");
      vec.push_back(rob);
    }
    return;
    case MomVKind::SetK:
    case MomVKind::TupleK:
    {
      auto seq = val.unsafe_sequence();
      MOM_ASSERT(seq != nullptr, "add_to_set no seq");
      vec.reserve(vec.size() + seq->size());
      for (auto rob: *seq)
        {
          MOM_ASSERT(rob, "add_to_vector no rob in seq");
          vec.push_back(rob);
        }
    }
    return;
    }
} // end MomTuple::add_to_vector
