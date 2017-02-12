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
  : _dustate(IdleDu), _dudir(dir), _duobjset(), _duqueue(), _dusqldb(nullptr), _duqueryinsobj(nullptr)
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
  if (_duqueryinsobj)
    {
      delete _duqueryinsobj;
      _duqueryinsobj = nullptr;
    }
  if (_dusqldb)
    {
      delete _dusqldb;
      _dusqldb = nullptr;
    }
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


MomVal
MomDumper::begin_scan(void)
{
  MOM_ASSERT(_dustate == IdleDu, "MomDumper is not idle when begin_scan");
  _dustate = ScanDu;
  auto vsetpredef = MomObject::set_of_predefined();
  MOM_VERBOSELOG("MomDumper::begin_scan vsetpredef=" << vsetpredef);
  auto vglobal = MomObject::set_of_globals();
  MOM_VERBOSELOG("MomDumper::begin_scan vglobal=" << vglobal);
  MomSetRefobj set;
  for (auto pob : *vsetpredef.as_set())
    {
      set.insert(pob);
      scan_refobj(pob);
    }
  for (auto pob : *vglobal.as_set())
    {
      set.insert(pob);
      scan_refobj(pob);
    }
  return MomVSet(set);
} // end MomDumper::begin_scan



void
MomDumper::emit_loop(void)
{
  MOM_ASSERT(_dustate == ScanDu, "MomDumper is not scanning when emit_loop");
  MOM_ASSERT(_duqueryinsobj == nullptr, "MomDumper with non-nil _duqueryinsobj");
  _dustate = EmitDu;
  _duqueryinsobj = new QSqlQuery(*_dusqldb);
  /// clear the t_objects table
  {
    QSqlQuery dquery(*_dusqldb);
    dquery.prepare("DELETE FROM t_objects");
    if (!dquery.exec())
      {
        MOM_BACKTRACELOG("emit_loop: SQL failure for t_objects delete :"
                         <<  _dusqldb->lastError().text().toStdString());
        throw std::runtime_error("MomDumper::emit_loop SQL failure t_objects delete");
      }
  }
  _duqueryinsobj->prepare(_insert_object_sql_);
  unsigned long nbemit = 0;
  for (auto rob : _duobjset)
    {
      emit_dumped_object(rob);
      nbemit++;
    }
  delete _duqueryinsobj;
  _duqueryinsobj = nullptr;
  MOM_VERBOSELOG("emit_loop emitted " << nbemit << " objects");
} // end MomDumper::emit_loop



void
MomDumper::emit_dumped_object(const MomObject*pob)
{
  MOM_ASSERT(_dustate == EmitDu, "MomDumper is not emitting when emit_dumped_object");
  MOM_ASSERT(pob, "MomDumper nil emitted dumped object");
  MOM_ASSERT(_duqueryinsobj, "MomDumper nil queryinsobj");
  {
    MomSharedReadObjLock _gu(pob);
    _duqueryinsobj->bindValue((int)InsobIdIx, pob->idstr().c_str());
    _duqueryinsobj->bindValue((int)InsobMtimIx, (qlonglong) pob->mtime());
    {
      const MomJson& jcont= pob->json_for_content(*this);
      Json::StyledWriter jwr;
      _duqueryinsobj->bindValue((int)InsobJsoncontIx, jwr.write(jcont).c_str());
    }
    auto payl = pob->get_payload_ptr();
    bool dumpedpayload = false;
    if (payl)
      {
        MOM_ASSERT(payl->owner() == MomRefobj(pob), "MomDumper corrupted payload");
        if (payl->emittable_payload(*this))
          {
            {
              const MomJson& jpayl = payl->payload_json(*this);
              Json::StyledWriter jwr;
              _duqueryinsobj->bindValue((int)InsobPaylcontIx, jwr.write(jpayl).c_str());
            }
            dumpedpayload = true;
            _duqueryinsobj->bindValue((int)InsobPaylkindIx, payl->payload_name());
          }
      };
    if (!dumpedpayload)
      {
        _duqueryinsobj->bindValue((int)InsobPaylkindIx,"");
        _duqueryinsobj->bindValue((int)InsobPaylcontIx,"");
      }
  }
  if (!_duqueryinsobj->exec())
    {
      MOM_BACKTRACELOG("emit_dumped_object: SQL failure for " <<  pob->idstr()
                       << " :" <<  _dusqldb->lastError().text().toStdString());
      throw std::runtime_error("MomDumper::emit_dumped_object SQL failure");
    }
} // end MomDumper::emit_dumped_object

void
MomDumper::write_file_content(const std::string&basepath, const std::string&content)
{
  MOM_ASSERT(_dustate == EmitDu, "MomDumper is not emitting when write_file_content " << basepath);
  if (basepath.empty() || basepath[0] == '/' || basepath.find(".."))
    {
      MOM_BACKTRACELOG("MomDumper::write_file_content invalid basepath:" << basepath);
      throw std::runtime_error("MomDumper::write_file_content invalid basepath");
    }
  std::string fullpath = _dudir + "/" + basepath;
  if (!::access(fullpath.c_str(), R_OK))
    {
      QFile fil(fullpath.c_str());
      if ((unsigned)fil.size() == (unsigned)content.size())
        {
          QByteArray by= fil.readAll();
          if (!strcmp(by.data(),content.c_str()))
            return;
        }
      std::string backupath = fullpath + "~";
      fil.rename(backupath.c_str());
    }
  std::ofstream outf(fullpath);
  outf << content << std::flush;
  outf.close();
}// end MomDumper::write_file_content

void
MomDumper::emit_predefined_header(const MomVal vset)
{
  MOM_ASSERT(vset.kind() == MomVKind::SetK, "emit_predefined_header bad vset");
  MOM_ASSERT(_dustate == EmitDu, "MomDumper is not emitting when emit_predefined_header");
  std::ostringstream outs;
  outs << "// generated file of predefined " << _predefined_header_ << " - DO NOT EDIT" << std::endl << std::endl;
  outs << "#" "ifndef MOM_HAS_PREDEF" << std::endl;
  outs << "#" "error missing MOM_HAS_PREDEF" << std::endl;
  outs << "#" "endif /*no MOM_HAS_PREDEF*/" << std::endl << std::endl;
  outs << "///MOM_HAS_PREDEF(Id,S1,S2,H)" << std::endl;
  unsigned count = 0;
  for (auto rob : *vset.as_set())
    {
      MOM_ASSERT(rob, "MomDumper emit_predefined_header no rob");
      if (rob->space() != MomSpace::PredefinedSp)
        continue;
      outs << "MOM_HAS_PREDEF(" << rob->idstr() << ","
           << rob->hi_serial() << "," << rob->lo_serial() << ","
           << rob->hash() << ")" << std::endl;
      count++;
    }
  outs << std::endl;
  outs << "#" << "undef MOM_HAS_PREDEF" << std::endl;
  outs << "#" << "undef MOM_NB_PREDEF" << std::endl;
  outs << "#" "define MOM_NB_PREDEF" << " " << count << std::endl << std::endl;
  outs << "// eof " << _predefined_header_ << std::endl;
  write_file_content(_predefined_header_, outs.str());
}// end MomDumper::emit_predefined_header


void
MomDumper::emit_globals(void)
{
  std::set<std::string> globalset;
  /// collect the globals
  for (const char*const*psrcfile= monimelt_cxxsources; *psrcfile; psrcfile++)
    {
      std::string srcfilpath = std::string(monimelt_directory) + "/" + std::string(*psrcfile);
      std::ifstream inp {srcfilpath};
      int lineno = 0;
      std::string clin;
      do
        {
          clin.clear();
          std::getline(inp,clin);
          lineno++;
          auto p = clin.find(_global_prefix_);
          while (p != std::string::npos)
            {
              auto e = p;
              while (isalnum(clin[e]) || (clin[e] == '_' && clin[e-1] != '_')) e++;
              auto nam = clin.substr(p+sizeof(_global_prefix_), e);
              if (mom_valid_name(nam))
                globalset.insert(nam);
              p = clin.find(_global_prefix_, e);
            }
        }
      while(inp);
    }
  /// emit the global header
  {
    std::ostringstream outs;
    outs << "// generated file of globals " << _global_header_ << " - DO NOT EDIT" << std::endl << std::endl;
    outs << "#" "ifndef MOM_HAS_GLOBAL" << std::endl;
    outs << "#" "error missing MOM_HAS_GLOBAL" << std::endl;
    outs << "#" "endif /*no MOM_HAS_GLOBAL*/" << std::endl << std::endl;
    outs << "///MOM_HAS_GLOBAL(GlobNam,Num)" << std::endl;
    unsigned count = 0;
    for (auto globnam : globalset)
      {
        outs << "MOM_HAS_GLOBAL(" << globnam << "," << count << ")" << std::endl;
        count++;
      }
    outs << std::endl;
    outs << "#" << "undef MOM_HAS_GLOBAL" << std::endl;
    outs << "#" << "undef MOM_NB_GLOBAL" << std::endl;
    outs << "#" "define MOM_NB_GLOBAL" << " " << count << std::endl << std::endl;
    outs << "// eof " << _global_header_ << std::endl;
    write_file_content(_global_header_, outs.str());
  }
  /// clear the t_globals table
  {
    QSqlQuery dquery(*_dusqldb);
    dquery.prepare("DELETE FROM t_globals");
    if (!dquery.exec())
      {
        MOM_BACKTRACELOG("emit_globals: SQL failure for t_globals delete :"
                         <<  _dusqldb->lastError().text().toStdString());
        throw std::runtime_error("MomDumper::emit_globals SQL failure t_globals delete");
      }
  }
  /// fill the t_globals table
  {
    QSqlQuery iquery(*_dusqldb);
    iquery.prepare("INSERT INTO t_globals (glo_name, glo_id) VALUES (?, ?)");
    enum { InsGloNamIx, InsGloIdIx, InsGlo_LasrtIx };
    for (auto glonam: globalset)
      {
        std::string cgloname = std::string{_globalatom_prefix_} + glonam;
        auto gloptr = reinterpret_cast<volatile std::atomic<MomObject*>*>(dlsym(mom_dlh,cgloname.c_str()));
        MomRefobj glrob;
        if (gloptr && (glrob=atomic_load(gloptr)))
          {
            iquery.bindValue(InsGloNamIx,glonam.c_str());
            iquery.bindValue(InsGloIdIx,glrob->idstr().c_str());
            if (!iquery.exec())
              {
                MOM_BACKTRACELOG("emit_globals: SQL failure for t_globals insert " <<  glonam.c_str()
                                 << " :" <<  _dusqldb->lastError().text().toStdString());
                throw std::runtime_error("MomDumper::emit_globals SQL failure t_globals insert");
              }
          }
      }
  }
} // end of MomDumper::emit_globals


void
MomDumper::create_tables(void)
{
  static constexpr const unsigned _obidwidth_ = 2*(MomSerial63::_nbdigits_+2);
  static constexpr const unsigned _kindwidth_ = 40;
  QSqlQuery query(*_dusqldb);
  {
    std::ostringstream obouts;
    obouts << " CREATE TABLE IF NOT EXISTS t_objects "
           << "(ob_id VARCHAR(" << _obidwidth_ << ") PRIMARY KEY ASC NOT NULL UNIQUE,"
           << " ob_mtime DATETIME NOT NULL,"
           << " ob_jsoncont TEXT NOT NULL,"
           << " ob_paylkid VARCHAR(" << _kindwidth_ << ") NOT NULL,"
           << " ob_paylcont TEXT NOT NULL)"
           << std::endl;
    if (!query.exec(obouts.str().c_str()))
      {
        MOM_BACKTRACELOG("create_tables Sql query t_objects failure: " <<  _dusqldb->lastError().text().toStdString());
        throw std::runtime_error("MomLoader::create_tables  t_objects query failure");
      }
  }
  {
    std::ostringstream glouts;
    glouts << " CREATE TABLE IF NOT EXISTS t_globals "
           << " (glo_name VARCHAR(256)  PRIMARY KEY ASC NOT NULL UNIQUE,"
           << "  glo_id VARCHAR(" << _obidwidth_ << ") NOT NULL)"
           << std::endl;
    if (!query.exec(glouts.str().c_str()))
      {
        MOM_BACKTRACELOG("create_tables Sql query t_globals failure: " <<  _dusqldb->lastError().text().toStdString());
        throw std::runtime_error("MomLoader::create_tables  t_global query failure");
      }
  }
} // end MomDumper::create_tables


void
mom_full_dump(const std::string&dir)
{
  MomDumper du(dir);
  MOM_VERBOSELOG("mom_full_dump start in " << du.dir());
  auto vroots = du.begin_scan();
  du.scan_loop();
  du.create_tables();
  du.emit_loop();
  du.emit_predefined_header(vroots);
  du.emit_globals();
  MOM_VERBOSELOG("mom_full_dump done in " << du.dir());
} // end mom_full_dump

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
  ld.create_objects();
} // end mom_initial_load

////////////////
