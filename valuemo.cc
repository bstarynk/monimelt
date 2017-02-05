// file valuemo.cc
#include "monimelt.h"


bool
MomVal::less(const MomVal&r) const
{
  if (this==&r) return false;
  auto k = kind();
  auto rk = r.kind();
  if (k<rk) return true;
  if (k>rk) return false;
  switch(k)
    {
    case MomVKind::NoneK:
      return false;
    case MomVKind::IntK:
      return _int < r._int;
    case MomVKind::RefobjK:
      return _ref < r._ref;
    case MomVKind::StringK:
    {
      MOM_ASSERT (_str, "MomVal::less bad _str");
      MOM_ASSERT (r._str, "MomVal::less bad r._str");
      return _str->less(*r._str);
    }
    case MomVKind::TupleK:
    {
      MOM_ASSERT(_tup, "MomVal::less bad tup");
      MOM_ASSERT(r._tup, "MomVal::less bad r._tup");
      return _tup->less(*r._tup);
    }
    case MomVKind::SetK:
    {
      MOM_ASSERT(_set, "MomVal::less bad set");
      MOM_ASSERT(r._set, "MomVal::less bad r._set");
      return _set->less(*r._set);
    }
    }
  MOM_BACKTRACELOG("MomVal::less impossible case");
  throw std::runtime_error("MomVal::less impossible case");
} // end MomVal::less

bool
MomVal::less_equal(const MomVal&r) const
{
  if (this==&r) return true;
  auto k = kind();
  auto rk = r.kind();
  if (k<rk) return true;
  if (k>rk) return false;
  switch(k)
    {
    case MomVKind::NoneK:
      return true;
    case MomVKind::IntK:
      return _int <= r._int;
    case MomVKind::RefobjK:
      return _ref <= r._ref;
    case MomVKind::StringK:
    {
      MOM_ASSERT (_str, "MomVal::less_equal bad _str");
      MOM_ASSERT (r._str, "MomVal::less_equal bad r._str");
      return _str->less_equal(*r._str);
    }
    case MomVKind::TupleK:
    {
      MOM_ASSERT(_tup, "MomVal::less_equal bad tup");
      MOM_ASSERT(r._tup, "MomVal::less_equal bad r._tup");
      return _tup->less_equal(*r._tup);
    }
    case MomVKind::SetK:
    {
      MOM_ASSERT(_set, "MomVal::less_equal bad set");
      MOM_ASSERT(r._set, "MomVal::less_equal bad r._set");
      return _set->less_equal(*r._set);
    }
    }
  MOM_BACKTRACELOG("MomVal::less_equal impossible case");
  throw std::runtime_error("MomVal::less_equal impossible case");
} // end MomVal::less_equal


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
