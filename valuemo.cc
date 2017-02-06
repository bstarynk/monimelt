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
    case MomVKind::ColoRefK:
      if (_coloref._colorob == r._coloref._colorob)
        return _coloref._cobref < r._coloref._colorob;
      else return _coloref._colorob < r._coloref._colorob;
    };
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
    case MomVKind::ColoRefK:
      if (_coloref._colorob == r._coloref._colorob)
        return _coloref._cobref <= r._coloref._colorob;
      else return _coloref._colorob < r._coloref._colorob;
    }
  MOM_BACKTRACELOG("MomVal::less_equal impossible case");
  throw std::runtime_error("MomVal::less_equal impossible case");
} // end MomVal::less_equal


MomHash_t
MomVal::hash() const
{
  auto k = kind();
  switch (k)
    {
    case MomVKind::NoneK:
      return 0;
    case MomVKind::IntK:
    {
      auto h = (MomHash_t)((1663L*_int) ^ (17L*(_int >> 28)));
      if (MOM_UNLIKELY(h==0))
        h=(MomHash_t(_int % 521363) & 0xfffff) + 310;
      MOM_ASSERT(h!=0, "MomVal::hash zero-hashed integer");
      return h;
    }
    case MomVKind::RefobjK:
    {
      MOM_ASSERT(_ref, "MomVal::hash null refobj");
      return _ref->hash();
    }
    case MomVKind::StringK:
    {
      MOM_ASSERT(_str, "MomVal::hash null str");
      return _str->hash();
    }
    case MomVKind::TupleK:
    {
      MOM_ASSERT(_tup, "MomVal::hash null tuple");
      return _tup->hash();
    }
    case MomVKind::SetK:
    {
      MOM_ASSERT(_set, "MomVal::hash null set");
      return _set->hash();
    }
    case MomVKind::ColoRefK:
    {
      auto cref = _coloref._cobref;
      auto colob = _coloref._colorob;
      MOM_ASSERT(cref, "MomVal::hash null cref");
      MOM_ASSERT(colob, "MomVal::hash null colob");
      MomHash_t href = cref->hash();
      MomHash_t h = href ^ MomHash_t(colob->lo_serial());
      if (MOM_UNLIKELY(h==0))
        {
          h = href;
          MOM_ASSERT(h!=0, "MomVal::hash zero coloref");
        }
      return h;
    }
    }
  MOM_BACKTRACELOG("MomVal::hash impossible case");
  throw std::runtime_error("MomVal::hash impossible case");
} // end of MomVal::hash



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
    case MomVKind::ColoRefK:
    {
      auto rob = val.unsafe_colorefobj();
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
    case MomVKind::ColoRefK:
    {
      auto colorob = val.unsafe_colorefobj();
      MOM_ASSERT (colorob, "add_to_vector no colorob");
      vec.push_back(colorob);
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
