// file statemo.cc
#include "monimelt.h"

#include <QtSql>
#include <QSqlDatabase>
#include <QSqlQuery>
#include <QProcess>
#include <QFileInfo>

class MomLoader final : public MomJsonParser ////
{
  std::string _ld_dirname;
  std::unordered_map<MomPairid,MomRefobj> _ld_objmap;
  QSqlDatabase* _ld_sqldb;
  double _ld_startelapsedtime;
  double _ld_startprocesstime;
public:
  MomLoader(std::string dir);
  ~MomLoader();
  MomRefobj idstr_to_refobj(const std::string&);
  MomRefobj make_object_of_idstr(const std::string&);
  void create_objects();
};    // end class MomLoader

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
  : _ld_dirname(dir), _ld_objmap(), _ld_sqldb(nullptr), _ld_startelapsedtime(0.0), _ld_startprocesstime(0.0)
{
  if (dir.empty())
    _ld_dirname=".";
  _ld_startelapsedtime = mom_elapsed_real_time();
  _ld_startprocesstime = mom_process_cpu_time();
  if (!QSqlDatabase::drivers().contains("QSQLITE"))
    {
      MOM_BACKTRACELOG("load: missing QSQLITE driver");
      throw std::runtime_error("MomLoader::load missing QSQLITE");
    }
  _ld_sqldb = new QSqlDatabase(QSqlDatabase::addDatabase("QSQLITE","momloader"));
  QString sqlitepath((_ld_dirname+"/"+monimelt_statebase+".sqlite").c_str());
  QString sqlpath((_ld_dirname+"/"+monimelt_statebase+".sql").c_str());
  if (!QFileInfo::exists(sqlitepath) || !QFileInfo::exists(sqlpath))
    {
      MOM_BACKTRACELOG("load: missing " << sqlitepath.toStdString()
                       << " or " << sqlpath.toStdString());
      throw std::runtime_error("MomLoader::load missing file");
    }
  if (QFileInfo(sqlitepath).lastModified() > QFileInfo(sqlpath).lastModified())
    {
      MOM_BACKTRACELOG("load: " << sqlitepath.toStdString()
                       << " younger than " << sqlpath.toStdString());
      throw std::runtime_error("MomLoader::load .sqlite youger");
    }
  _ld_sqldb->setDatabaseName(sqlitepath);
  if (!_ld_sqldb->open())
    {
      MOM_BACKTRACELOG("load " << sqlitepath.toStdString()
                       << " failed to open: " << _ld_sqldb->lastError().text().toStdString());
      throw std::runtime_error("MomLoader::load open failure");
    }
} // end MomLoader::MomLoader


MomLoader::~MomLoader()
{
  delete _ld_sqldb;
  _ld_sqldb = nullptr;
} // end of MomLoader::~MomLoader

MomRefobj
MomLoader::idstr_to_refobj(const std::string&idstr)
{
  auto pi = MomObject::id_from_cstr(idstr.c_str());
  if (!pi) return nullptr;
  auto it = _ld_objmap.find(pi);
  if (it == _ld_objmap.end()) return nullptr;
  return it->second;
} // end MomLoader::idstr_to_refobj


void
MomLoader::create_objects(void)
{
  MOM_ASSERT(_ld_sqldb, "create_objects no sqldb");
  QSqlQuery query(*_ld_sqldb);
  enum { ResixId, Resix_LAST };
  if (!query.exec("SELECT ob_id FROM t_objects"))
    {
      MOM_BACKTRACELOG("create_objects Sql query failure: " <<  _ld_sqldb->lastError().text().toStdString());
      throw std::runtime_error("MomLoader::create_objects query failure");
    }
  while (query.next())
    {
      std::string idstr = query.value(ResixId).toString().toStdString();
      auto rob = make_object_of_idstr(idstr);
      _ld_objmap.insert({rob->ident(),rob});
    }
} // end MomLoader::create_objects


MomRefobj
MomLoader::make_object_of_idstr(const std::string&ids)
{
  auto pi = MomObject::id_from_cstr(ids.c_str());
  if (!pi)
    {
      MOM_BACKTRACELOG("make_object_of_idstr bad ids:" << ids);
      return nullptr;
    }
  auto pob = MomObject::find_object_of_id(pi);
  if (pob)
    return pob;
  pob = new MomObject(MomObject::TagNewObject{},pi,MomSpace::GlobalSp);
  return pob;
} // end MomLoader::make_object_of_idstr

////////////////
void mom_initial_load(const std::string&dir)
{
  MomLoader ld(dir);
} // end mom_initial_load

////////////////
