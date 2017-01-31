-- monimelt_state.sql dump 2017 Jan 31 from monimelt_state.sqlite dumped by ./monimelt-dump-state.sh .....

 --   Copyright (C) 2017 Basile Starynkevitch.
 --  This sqlite3 dump file monimelt_state.sql is part of MONIMELT.
 --
 --  MONIMELT is free software; you can redistribute it and/or modify
 --  it under the terms of the GNU General Public License as published by
 --  the Free Software Foundation; either version 3, or (at your option)
 --  any later version.
 --
 --  MONIMELT is distributed in the hope that it will be useful,
 --  but WITHOUT ANY WARRANTY; without even the implied warranty of
 --  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 --  GNU General Public License for more details.
 --  You should have received a copy of the GNU General Public License
 --  along with MONIMELT; see the file COPYING3.   If not see
 --  <http://www.gnu.org/licenses/>.

PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE t_params (par_name VARCHAR(35) PRIMARY KEY ASC NOT NULL UNIQUE,  par_value TEXT NOT NULL);
COMMIT;
-- monimelt-dump-state end dump monimelt_state.sqlite
