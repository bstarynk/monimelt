## file Makefile in monimelt

## by Basile Starynkevitch <basile@starynkevitch.net>
CXX=g++
CC=gcc
GCC=gcc
INDENT= indent
ASTYLE= astyle
MD5SUM= md5sum
SQLITE= sqlite3
PACKAGES= sqlite3 jsoncpp Qt5Core Qt5Widgets Qt5Gui Qt5Sql
PKGCONFIG= pkg-config
OPTIMFLAGS= -g2 -O1
CXXOPTIMFLAGS= $(OPTIMFLAGS)
COPTIMFLAGS= $(OPTIMFLAGS)
CXXWARNFLAGS= -Wall -Wextra
CXXPREPROFLAGS= -I /usr/local/include  $(shell $(PKGCONFIG) --cflags $(PACKAGES))  -I $(shell $(GCC) -print-file-name=include/)
## -fPIC is required by Qt5, -fPIE is not enough
CXXCOMMONFLAGS= -fPIC
QTMOC= moc
CXXFLAGS= -std=gnu++14 $(CXXCOMMONFLAGS) $(CXXOPTIMFLAGS) $(CXXWARNFLAGS) $(CXXPREPROFLAGS)
CFLAGS= -Wall $(COPTIMFLAGS)
ASTYLEFLAGS= --style=gnu -s2  --convert-tabs
INDENTFLAGS= --gnu-style --no-tabs --honour-newlines
CXXSOURCES= $(wildcard [a-z]*.cc)
CSOURCES= $(wildcard [a-z]*.c)
## this monimelt_state basename is "sacred", don't change it
MONIMELT_STATE=monimelt_state
SHELLSOURCES= $(sort $(wildcard [a-z]*.sh))
OBJECTS= $(patsubst %.cc,%.o,$(CXXSOURCES)) $(patsubst %.c,%.o,$(CSOURCES))
GENERATED_HEADERS= $(wildcard _*.h)
LIBES= -L/usr/local/lib  $(shell $(PKGCONFIG) --libs $(PACKAGES)) -pthread  $(shell $(GCC) -print-file-name=libbacktrace.a) -ldl
.PHONY: all checkgithooks installgithooks clean dumpstate restorestate indent
all: checkgithooks monimelt

_timestamp.c: Makefile | $(OBJECTS)
	@echo "/* generated file _timestamp.c - DONT EDIT */" > _timestamp.tmp
	@date +'const char monimelt_timestamp[]="%c";' >> _timestamp.tmp
	@(echo -n 'const char monimelt_lastgitcommit[]="' ; \
	   git log --format=oneline --abbrev=12 --abbrev-commit -q  \
	     | head -1 | tr -d '\n\r\f\"' ; \
	   echo '";') >> _timestamp.tmp
	@(echo -n 'const char monimelt_lastgittag[]="'; (git describe --abbrev=0 --all || echo '*notag*') | tr -d '\n\r\f\"'; echo '";') >> _timestamp.tmp
	@(echo -n 'const char*const monimelt_cxxsources[]={'; for sf in $(CXXSOURCES) ; do \
	  echo -n "\"$$sf\", " ; done ; echo '(const char*)0};' ; \
	echo -n 'const char*const monimelt_csources[]={'; for sf in $(CSOURCES) ; do \
	  echo -n "\"$$sf\", " ; done ; echo '(const char*)0};' ; \
	echo -n 'const char*const monimelt_shellsources[]={'; for sf in $(SHELLSOURCES) ; do \
	  echo -n "\"$$sf\", " ; done ; \
	echo '(const char*)0};') >> _timestamp.tmp
	@(echo -n 'const char monimelt_directory[]="'; echo -n $(realpath .); echo '";') >> _timestamp.tmp
	@(echo -n 'const char monimelt_statebase[]="'; echo -n $(MONIMELT_STATE); echo '";') >> _timestamp.tmp
	@echo >> _timestamp.tmp
	@echo >> _timestamp.tmp
	mv _timestamp.tmp _timestamp.c

monimelt: $(OBJECTS) _timestamp.o
	@if [ -f $@ ]; then echo -n makebackup old executable: ' ' ; mv -v $@ $@~ ; fi
	$(LINK.cc)  $(LINKFLAGS) $(OPTIMFLAGS) -rdynamic $(OBJECTS)  _timestamp.o $(LIBES) -o $@ 

%.o: %.cc monimelt.h $(GENERATED_HEADERS) 
# implicitly COMPILE.cc = $(CXX) $(CXXFLAGS) $(CPPFLAGS) $(TARGET_ARCH) -c & OUTPUT_OPTION = -o $@
	$(CXX)  $(CXXFLAGS) $(CPPFLAGS)  -MF $(patsubst %.cc,_%.mkd,$<) -MT $@ -MMD  $(TARGET_ARCH) -c -o $@ $<


%.ii: %.cc monimelt.h $(GENERATED_HEADERS)
	$(COMPILE.cc) -C -E $< -o -  | sed s:^#://#:g > $@

%.moc.h: %.cc
	$(QTMOC) -o $@ $<

clean:
	$(RM) *~ *% *.o *.so */*.so *.log */*~ */*.orig *.i *.ii *.orig README.html *#
	$(RM) *.moc.h
	$(RM) _*.mkd _mocdepend.mk
	$(RM) core*
	$(RM) _timestamp.* bxmo

checkgithooks:
	@for hf in *-githook.sh ; do \
	  [ ! -d .git -o -L .git/hooks/$$(basename $$hf "-githook.sh") ] \
	    || (echo uninstalled git hook $$hf "(run: make installgithooks)" >&2 ; exit 1) ; \
	done
installgithooks:
	for hf in *-githook.sh ; do \
	  ln -sv  "../../$$hf" .git/hooks/$$(basename $$hf "-githook.sh") ; \
	done

dumpstate: $(MONIMELT_STATE).sqlite | monimelt-dump-state.sh
	./monimelt-dump-state.sh $(MONIMELT_STATE).sqlite $(MONIMELT_STATE).sql

restorestate: | $(MONIMELT_STATE).sql
	@if [ -f $(MONIMELT_STATE).sqlite ]; then \
	  echo makebackup old: ' ' ; mv -b -v  $(MONIMELT_STATE).sqlite  $(MONIMELT_STATE).sqlite~ ; fi
	$(SQLITE) $(MONIMELT_STATE).sqlite < $(MONIMELT_STATE).sql
	touch -r $(MONIMELT_STATE).sql -c $(MONIMELT_STATE).sqlite

indent:
	$(ASTYLE) $(ASTYLEFLAGS) monimelt.h
	for g in $(wildcard [a-z]*.cc) ; do \
	  $(ASTYLE) $(ASTYLEFLAGS) $$g ; \
	done
	for c in $(wildcard [a-z]*.c) ; do \
	  $(INDENT) $(INDENTFLAGS) $$c ; \
	done


_mocdepend.mk: | $(CXXSOURCES)
	date +"#$< generated _mocdepend.mk %c%n" > _mocdepend.tmp
	for f in $(CXXSOURCES) ; do				\
	  b=$$(basename $$f .cc) ;				\
	  m=$$b.moc.h ;						\
	  grep -q "$$m" $$f					\
	     && printf "%s: %s\n" $$b.o $$m >> _mocdepend.tmp ;	\
	  true ; \
	done
	echo '#eof ' $@ >> _mocdepend.tmp
	mv _mocdepend.tmp _mocdepend.mk

include _mocdepend.mk
-include $(wildcard _*.mkd)
