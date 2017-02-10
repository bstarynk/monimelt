// file statemo.cc
#include "monimelt.h"

MomDumper::MomDumper(const std::string&dir)
  : _dustate(IdleDu), _dudir(dir), _duobjset(), _duqueue()
{
  if (dir.empty()) _dudir = ".";
  struct stat ds = {};
  errno = 0;
  if (access(_dudir.c_str(), F_OK) && errno == ENOENT)
    {
      errno = 0;
      if (mkdir(_dudir.c_str(), 0750))
        {
          MOM_BACKTRACELOG("MomDumper mkdir " << _dudir << " failed: "
                           << strerror(errno));
          throw std::runtime_error("MomDumper mkdir failure");
        }
    };
  errno = 0;
  if (stat(_dudir.c_str(), &ds) || !S_ISDIR(ds.st_mode))
    {
      MOM_BACKTRACELOG("MomDumper bad directory " << _dudir << " : " << strerror(errno));
      throw std::runtime_error("MomDumper bad directory");
    }
} // end MomDumper::MomDumper

MomDumper::~MomDumper()
{
  MOM_ASSERT(_dustate == IdleDu,"MomDumper in " << _dudir
             << " still active when destroyed");
  _duobjset.clear();
  _duqueue.clear();
}

void
MomDumper::scan_refobj(const MomRefobj rob)
{
  MOM_ASSERT(_dustate == ScanDu, "MomDumper is not scanning");
  if (!rob) return;
  if (rob->space() == MomSpace::NoneSp) return;
  if (_duobjset.find(rob) != _duobjset.end()) return;
  _duobjset.insert(rob);
  _duqueue.push_back(rob);
} // end MomDumper::scan_refobj

void
MomDumper::scan_value(const MomVal val)
{
  MOM_ASSERT(_dustate == ScanDu, "MomDumper is not scanning");
  if (!val) return;
  auto thisdumper = this;
  val.scan_objects([=](MomRefobj ro)
  {
    thisdumper->scan_refobj(ro);
    return false;
  });
} // end MomDumper::scan_value

void
MomDumper::scan_inside_dumped_object(const MomObject*pob)
{
  MOM_ASSERT(_dustate == ScanDu, "MomDumper is not scanning");
  if (!pob) return;
  if (!dumpable_refobj(pob)) return;
  auto thisdumper = this;
  MomSharedReadObjLock _gu(pob);
  pob->scan_inside_object([=](MomRefobj ro)
  {
    thisdumper->scan_refobj(ro);
    return false;
  },
  [=](MomRefobj insiderob, MomRefobj rob)
  {
    MOM_ASSERT(insiderob, "no insiderob");
    return dumpable_refobj(rob);
  });
} // end MomDumper::scan_inside_dumped_object

void
MomDumper::scan_loop(void)
{
  MOM_ASSERT(_dustate == ScanDu, "MomDumper is not scanning");
  unsigned long nbscan = 0;
  while (!_duqueue.empty())
    {
      MomRefobj curob = _duqueue.front();
      _duqueue.pop_front();
      MOM_ASSERT(curob, "MomDumper nil curob");
      MOM_ASSERT(dumpable_refobj(curob),"MomDumper non dumpable curob=" << curob);
      scan_inside_dumped_object(curob);
      nbscan++;
    }
  MOM_VERBOSELOG("MomDumper::scan_loop nbscan=" << nbscan);
} // end MomDumper::scan_loop


void
MomDumper::begin_scan(void)
{
  MOM_ASSERT(_dustate == IdleDu, "MomDumper is not idle when begin_scan");
  _dustate = ScanDu;
  scan_value(MomObject::set_of_predefined());
} // end MomDumper::scan_loop


////////////////////////////////////////////////////////////////

MomLoader::MomLoader(std::string dir)
  : _lddir(dir), _ldobjset()
{
  if (dir.empty())
    _lddir=".";
} // end MomLoader::MomLoader


MomLoader::~MomLoader()
{
} // end of MomLoader::~MomLoader

MomRefobj
MomLoader::idstr_to_refobj(const std::string&idstr)
{
  auto pi = MomObject::id_from_cstr(idstr.c_str());
  if (!pi) return nullptr;
  return MomObject::find_object_of_id(pi);
} // end MomLoader::idstr_to_refobj
