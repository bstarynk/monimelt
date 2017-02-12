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

void
MomVal::out(std::ostream& os) const
{
  auto k = kind();
  char endc = 0;
  switch(k)
    {
    case MomVKind::NoneK:
      os << "~";
      break;
    case MomVKind::IntK:
      os << _int;
      break;
    case MomVKind::StringK:
      os << '"' << MomUtf8Out(as_string()) << '"';
      break;
    case MomVKind::RefobjK:
      os << as_refobj();
      break;
    case MomVKind::SetK:
      endc = '}';
      os << '{';
      goto outseq;
    case MomVKind::TupleK:
      endc = ']';
      os << '[';
      goto outseq;
outseq:
      {
        const MomSequence& seq = *as_sequence();
        unsigned ln = seq.size();
        if (ln>0)
          for (unsigned ix=0; ix<ln; ix++)
            {
              if (ix>0) os << ' ';
              os << seq.at(ix);
            }
        os << endc;
      }
      break;
    case MomVKind::ColoRefK:
      os << '%' << as_colorob() << '!' << as_colorefobj();
      break;
    }
} // end MomVal::out

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



MomVal
MomVal::parse_json(const MomJson&js, MomJsonParser&jp)
{
  if (js.isNull()) return nullptr;
  else if (js.isInt()) return MomVInt(js.asInt64());
  else if (js.isString()) return MomVString(js.asString());
  else if (js.isObject())
    {
      MomJson jcomp;
      MomJson jcolor;
      if ((jcomp= js["ref"]).isString())
        {
          auto id = MomObject::id_from_cstr(jcomp.asCString(),true);
          auto pob = MomObject::find_object_of_id(id);
          if (!pob)
            pob = jp.idstr_to_refobj(jcomp.asCString());
          if (!pob)
            {
              MOM_BACKTRACELOG("parse_json bad id:" << id);
              throw std::runtime_error("parse_json bad id");
            }
          return MomVRef{pob};
        }
      else if ((jcomp = js["cref"]).isString() && (jcolor= js["color"]).isString())
        {
          auto cid = MomObject::id_from_cstr(jcomp.asCString(),true);
          auto prob = MomObject::find_object_of_id(cid);
          if (!prob)
            prob = jp.idstr_to_refobj(jcomp.asCString());
          if (!prob)
            {
              MOM_BACKTRACELOG("parse_json bad cref:" << cid << " in:" << js);
              throw std::runtime_error("parse_json bad cref");
            };
          auto colorid = MomObject::id_from_cstr(jcolor.asCString(),true);
          auto pobcolor = MomObject::find_object_of_id(colorid);
          if (!pobcolor)
            pobcolor = jp.idstr_to_refobj(jcolor.asCString());
          if (!pobcolor)
            {
              MOM_BACKTRACELOG("parse_json bad color:" << colorid << " in:" << js);
              throw std::runtime_error("parse_json bad color");
            };
          return MomVColoRef{prob,pobcolor};
        } // end when "cref" && "color"
      else if ((jcomp = js["set"]).isArray())
        {
          auto ln = jcomp.size();
          MomSetRefobj setr;
          for (unsigned ix=0; ix<ln; ix++)
            {
              auto jid = jcomp[ix];
              if (jid.isString())
                {
                  auto cid = MomObject::id_from_cstr(jid.asCString(),true);
                  auto elrob = MomObject::find_object_of_id(cid);
                  if (!elrob)
                    elrob = jp.idstr_to_refobj(jid.asCString());
                  if (!elrob)
                    {
                      MOM_BACKTRACELOG("parse_json bad element id:" << jid << " in:" << js);
                      throw std::runtime_error("parse_json bad element id");
                    }
                  setr.insert(elrob);
                }
            }
          return MomVSet(setr);
        } // end when "set"
      else if ((jcomp = js["tup"]).isArray())
        {
          auto ln = jcomp.size();
          std::vector<MomRefobj> vec;
          for (unsigned ix=0; ix<ln; ix++)
            {
              auto jid = jcomp[ix];
              if (jid.isString())
                {
                  auto cid = MomObject::id_from_cstr(jid.asCString(),true);
                  auto comprob = MomObject::find_object_of_id(cid);
                  if (!comprob)
                    comprob = jp.idstr_to_refobj(jid.asCString());
                  if (!comprob)
                    {
                      MOM_BACKTRACELOG("parse_json bad component id:" << jid << " in:" << js);
                      throw std::runtime_error("parse_json bad component id");
                    }
                  vec.push_back(comprob);
                }
            }
          return MomVTuple(vec);
        } // end when "tup"
    }
  MOM_BACKTRACELOG("parse_json bad js:" << js);
  throw std::runtime_error("MomVal::parse_json bad js");
} // end MomVal::parse_json


MomJson
MomVal::emit_json(MomJsonEmitter&jem) const
{
  switch (_kind)
    {
    case MomVKind::NoneK:
      return nullptr;
    case MomVKind::IntK:
      return MomJson::Value::Int64(_int);
    case MomVKind::StringK:
      MOM_ASSERT(_str, "bad string to emit_json");
      return _str->to_string();
    case MomVKind::RefobjK:
    {
      MOM_ASSERT(_ref, "bad reference to emit_json");
      if (jem.emittable_refobj(_ref))
        {
          auto job = MomJson{Json::objectValue};
          job["ref"]= _ref->idstr();
          return job;
        }
      else
        return nullptr;
    }
    case MomVKind::ColoRefK:
    {
      auto cref = _coloref._cobref;
      auto colorob = _coloref._colorob;
      MOM_ASSERT(cref, "bad colored reference to emit_json");
      MOM_ASSERT(colorob, "bad color to emit_json");
      if (jem.emittable_refobj(colorob) && jem.emittable_refobj(cref))
        {
          auto job = MomJson{Json::objectValue};
          job["cref"] = cref->idstr();
          job["color"] = colorob->idstr();
          return job;
        }
    }
    case MomVKind::SetK:
    {
      MOM_ASSERT(_set, "bad set to emit_json");
      auto job = MomJson{Json::objectValue};
      auto jseq = MomJson{Json::arrayValue};
      for (auto elrob : *_set)
        {
          MOM_ASSERT(elrob, "bad element in set to emit_json");
          if (jem.emittable_refobj(elrob))
            jseq.append(elrob->idstr());
        }
      job["set"] = jseq;
      return job;
    }
    case MomVKind::TupleK:
    {
      MOM_ASSERT(_tup, "bad tuple to emit_json");
      auto job = MomJson{Json::objectValue};
      auto jseq = MomJson{Json::arrayValue};
      for (auto comprob : *_tup)
        {
          MOM_ASSERT(comprob, "bad component in tuple to emit_json");
          if (jem.emittable_refobj(comprob))
            jseq.append(comprob->idstr());
        }
      job["tup"] = jseq;
      return job;
    }
    }
  MOM_ASSERT(false,"impossible value to emit_json:" << *this);
} // end MomVal::emit_json



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


bool
MomSequence::scan_objects(const std::function<bool(MomRefobj)>&f) const
{
  for (MomRefobj ro : *this)
    if (f(ro)) return false;
  return true;
} // end MomSequence::scan_objects


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

bool
MomVal::scan_objects(const std::function<bool(MomRefobj)>&f) const
{
  switch(kind())
    {
    case MomVKind::NoneK:
    case MomVKind::IntK:
    case MomVKind::StringK:
      return true;
    case MomVKind::RefobjK:
    {
      auto po = unsafe_refobj();
      MOM_ASSERT(po, "nil refobj in MomVal::scan_objects");
      return !f(po);
    }
    case MomVKind::TupleK:
    case MomVKind::SetK:
      MOM_ASSERT(_seq, "nil sequence in MomVal::scan_objects");
      return _seq->scan_objects(f);
    case MomVKind::ColoRefK:
    {
      auto cref = unsafe_colorefobj();
      MOM_ASSERT(cref, "nil coloref in  MomVal::scan_objects");
      auto colorob = unsafe_colorob();
      MOM_ASSERT(colorob, "nil colorob in MomVal::scan_objects");
      return !f(cref) && !f(colorob);
    }
    }
  MOM_BACKTRACELOG("MomVal::scan_objects unexpected value");
  throw std::runtime_error("MomVal::scan_objects unexpected value");
} // end MomVal::scan_objects
