package lib

import (
	"os/exec"
	"runtime/debug"

	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func DropMetadato(dbMetaName DbMetaConnMs, db *sql.DB) LoggaErrore {

	var loggaErrore LoggaErrore
	loggaErrore.Errore = 0

	Logga("Drop metadata :" + dbMetaName.MetaName)

	_, err := db.Exec("drop database if exists " + dbMetaName.MetaName)
	if err != nil {
		loggaErrore.Log = err.Error()
		loggaErrore.Errore = -1
		return loggaErrore
	} else {
		Logga("Database " + dbMetaName.MetaName + " dropped")
	}

	loggaErrore.Log = ""
	loggaErrore.Errore = 1
	return loggaErrore
}
func CreateDbMeta(dbMetaName DbMetaConnMs, db *sql.DB) LoggaErrore {

	var loggaErrore LoggaErrore
	loggaErrore.Errore = 0

	query := "CREATE DATABASE " + dbMetaName.MetaName
	Logga(query)
	_, err := db.Exec(query)
	if err != nil {
		loggaErrore.Log = err.Error()
		loggaErrore.Errore = -1
		return loggaErrore
	} else {
		Logga("Database " + dbMetaName.MetaName + " instance done")
	}

	// creo gli user
	CreateUser(dbMetaName, db)

	return loggaErrore
}
func CreateDbData(dbDataName DbDataConnMs, db *sql.DB) LoggaErrore {

	// creazione dei DATABASE
	var loggaErrore LoggaErrore

	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + dbDataName.DataName)
	if err != nil {
		loggaErrore.Log = err.Error()
		loggaErrore.Errore = -1
		return loggaErrore
	} else {
		Logga("Create Database " + dbDataName.DataName + " instance done")
	}

	// creo gli user
	var cons DbMetaConnMs
	cons.MetaHost = dbDataName.DataHost
	cons.MetaName = dbDataName.DataName
	cons.MetaPass = dbDataName.DataPass
	cons.MetaUser = dbDataName.DataUser
	CreateUser(cons, db)

	//os.Exit(0)

	loggaErrore.Log = ""
	loggaErrore.Errore = 1
	return loggaErrore
}
func CreateUser(dbMetaName DbMetaConnMs, db *sql.DB) LoggaErrore {

	var loggaErrore LoggaErrore
	loggaErrore.Errore = 0

	// FAC-753
	var _user string
	query := "SELECT user as _user FROM mysql.user where user = '" + dbMetaName.MetaUser + "'"
	fmt.Println(query)
	row := db.QueryRow(query)
	errUser := row.Scan(&_user)

	if errUser != nil {

		if errUser.Error() == "sql: no rows in result set" {

			// create user
			query := "CREATE USER   '" + dbMetaName.MetaUser + "'@'%' IDENTIFIED BY '" + dbMetaName.MetaPass + "'"
			Logga(query)
			_, err := db.Exec(query)
			if err != nil {
				loggaErrore.Log = err.Error()
				loggaErrore.Errore = -1
				return loggaErrore
			} else {
				Logga("CREATE USER    " + dbMetaName.MetaUser + " done")
			}

			// grant su metadati
			query = "GRANT ALL PRIVILEGES ON " + dbMetaName.MetaName + ".* TO '" + dbMetaName.MetaUser + "'@'%' WITH GRANT OPTION "
			Logga(query)
			_, err = db.Exec(query)
			//Logga("GRANT ALL PRIVILEGES ON " + ires.MetaName + ".* TO '" + ires.MetaUser + "'@'%'")
			if err != nil {
				loggaErrore.Log = err.Error()
				loggaErrore.Errore = -1
				return loggaErrore
			} else {
				Logga("GRANT ON " + dbMetaName.MetaName + " created")
			}

			// grant su data
			query = "FLUSH PRIVILEGES"
			Logga(query)
			_, err = db.Exec(query)
			if err != nil {
				loggaErrore.Log = err.Error()
				loggaErrore.Errore = -1
				return loggaErrore
			} else {
				Logga("FLUSH PRIVILEGES " + dbMetaName.MetaName + " done")
			}
		} else {
			loggaErrore.Log = errUser.Error()
			loggaErrore.Errore = -1
			return loggaErrore
		}

	} else {

		Logga("User: " + dbMetaName.MetaUser + " already exists")
	}

	loggaErrore.Log = ""
	loggaErrore.Errore = 0
	return loggaErrore
}
func DropDbData(dbDataName DbDataConnMs, db *sql.DB) LoggaErrore {

	var loggaErrore LoggaErrore

	_, err := db.Exec("drop database if exists " + dbDataName.DataName)
	if err != nil {
		loggaErrore.Log = err.Error()
		loggaErrore.Errore = -1
		return loggaErrore
	} else {
		Logga("Database " + dbDataName.DataName + " dropped")
	}
	loggaErrore.Log = ""
	loggaErrore.Errore = 1
	return loggaErrore
}
func Comparedb(ires IstanzaMicro, dbDataName DbDataConnMs, db *sql.DB, db2 *sql.DB) (LoggaErrore, []string) {

	var loggaErrore LoggaErrore

	var allCompareSql []string

	var dbDataSrc string
	if strings.Contains(dbDataName.DataName, "_compare") {
		dbDataSrc = dbDataName.DataName
	} else {
		dbDataSrc = dbDataName.DataName + "_compare"
	}
	dbDataDst := dbDataName.DataName

	// se facciamo il compare sui monoliti
	if ires.Monolith == 1 {
		dbDataSrc = dbDataName.DataName + "_compare_canary_monolith"
		dbDataDst = dbDataName.DataName + "_compare_production_monolith"
	}

	Logga("")
	Logga("*********")
	Logga("Source Database: " + dbDataSrc)
	Logga("Destination Database: " + dbDataDst)
	Logga("*********")
	Logga("")

	var table_name, column_name, columns string

	sqlSel := " SELECT "
	sqlSel += " table_name, column_name,  "
	sqlSel += " concat(column_type, ':', case when column_default is null then 'NULL' when column_default = '0.00' then 0 when column_default = '' then \"''\" when column_default='N' then \"'N'\" else  column_default end) as columns "
	sqlSel += " FROM information_schema.columns where 1>0 "
	sqlSel += " and table_schema = '" + dbDataSrc + "' "
	sqlSel += " ORDER BY table_name, column_name"
	//fmt.Println(sqlSel)
	// os.Exit(0)
	selDB, err := db.Query(sqlSel)
	if err != nil {
		loggaErrore.Log = err.Error()
		loggaErrore.Errore = -1
		return loggaErrore, allCompareSql
	}

	var srcTbls []CompareDbRes
	tablesList := "(\""
	for selDB.Next() {

		var tbl CompareDbRes

		err = selDB.Scan(&table_name, &column_name, &columns)
		if err != nil {
			loggaErrore.Log = err.Error()
			loggaErrore.Errore = -1
			return loggaErrore, allCompareSql
		}

		tbl.Tbl = table_name
		tbl.Column_name = column_name
		tbl.Columns = columns
		srcTbls = append(srcTbls, tbl)

		tablesList += table_name + "\", "

	}

	sqlSel = " SELECT "
	sqlSel += " table_name, column_name, "
	sqlSel += " concat(column_type, ':', case when column_default is null then 'NULL'  when column_default = '0.00' then 0 when column_default = '' then \"''\" when column_default='N' then \"'N'\" else  column_default end) as columns "
	sqlSel += " FROM information_schema.columns where 1>0 "
	sqlSel += " and table_schema = '" + dbDataDst + "' "
	sqlSel += " ORDER BY table_name, column_name"
	// fmt.Println(sqlSel)
	// os.Exit(0)
	selDB, err = db.Query(sqlSel)
	if err != nil {
		loggaErrore.Log = err.Error()
		loggaErrore.Errore = -1
		return loggaErrore, allCompareSql
	}

	var dstTbls []CompareDbRes
	for selDB.Next() {
		err = selDB.Scan(&table_name, &column_name, &columns)
		if err != nil {
			loggaErrore.Log = err.Error()
			loggaErrore.Errore = -1
			return loggaErrore, allCompareSql
		}
		var tbl CompareDbRes
		tbl.Tbl = table_name
		tbl.Column_name = column_name
		tbl.Columns = columns
		dstTbls = append(dstTbls, tbl)
	}

	// fmt.Println(srcTbls)
	// fmt.Println(dstTbls)
	// os.Exit(0)

	//fmt.Println("Get all info")
	// --------------------------------------------------
	// RACCOLTO LE INFO PROCEDO ALLA COMPARE

	// mi scorro l'oggetto sorgente e confrontandolo con il secondo
	// ne creo un terzo con le differnze
	var diffTbls []CompareDbRes
	var strSrc, strDst string

	// conterra' tutte le tabelle mancanti nel db dest
	missingTbls := make(map[string]interface{})

	for _, v := range srcTbls {

		strSrc = v.Tbl + ":" + v.Column_name + ":" + v.Columns
		//fmt.Println("check: " + v.column_name)

		var find bool
		var tblFind bool
		var colFind bool
		tblFind = false
		colFind = false

		// qui cerco le tabelle mancanti
		tblFind = false
		for _, vv := range dstTbls {
			if v.Tbl == vv.Tbl {
				tblFind = true
				break
			}
		}
		if tblFind == false {
			missingTbls[v.Tbl] = 1
			continue
		}

		for _, vv := range dstTbls {

			strDst = vv.Tbl + ":" + vv.Column_name + ":" + vv.Columns

			// mi segno se la colonna esiste
			if v.Column_name == vv.Column_name {
				colFind = true
			}

			// mi segno se la colonna ?? uguale
			if strSrc == strDst {
				find = true
				break
			} else {
				find = false
			}
		}

		if !colFind {
			var diffTbl CompareDbRes
			diffTbl.Tbl = v.Tbl
			diffTbl.Columns = v.Columns
			diffTbl.Column_name = v.Column_name
			diffTbl.Tipo = "ADD"

			diffTbls = append(diffTbls, diffTbl)
		} else {

			if !find {
				// // fmt.Println("elimino" + v.column_name)
				var diffTbl CompareDbRes
				diffTbl.Tbl = v.Tbl
				diffTbl.Columns = v.Columns
				diffTbl.Column_name = v.Column_name
				diffTbl.Tipo = "CHANGE"

				diffTbls = append(diffTbls, diffTbl)
			}
		}

		// fine ricerca tabelle mancanti
	}
	// fmt.Println(diffTbls)
	// os.Exit(0)

	Logga("Get all diff")
	Logga("")
	Logga("STO PER APPLICARE LE DIFF")
	Logga("Change Database Structure on " + dbDataName.DataName)
	Logga(dbDataName.DataHost + "|" + dbDataName.DataName)
	//fmt.Println(missingTbls)

	// **************************************************************************
	// **************************************************************************
	// **************************************************************************
	// da qui in poi si applica cio che e stato calcolato

	for k, _ := range missingTbls {
		sqlCompare := "CREATE TABLE " + k + " like " + dbDataSrc + "." + k

		// popolo un array con tutte le query da fare
		allCompareSql = append(allCompareSql, sqlCompare)

		_, err = db2.Exec(sqlCompare)
		if err != nil {

			loggaErrore.Log += err.Error() + "\n"
			loggaErrore.Errore = -1

		} else {
			Logga(sqlCompare + " ok")
		}
	}

	// fmt.Println(diffTbls)
	// os.Exit(0)

	var sqlCompare string
	for _, vv := range diffTbls {

		// !!! questo blocco ?? clone del seguente !!!
		xxx := strings.Split(vv.Columns, ":")
		if vv.Tipo == "CHANGE" {
			if xxx[1] != "" {
				sqlCompare = "ALTER TABLE " + vv.Tbl + " CHANGE " + vv.Column_name + " " + vv.Column_name + " " + xxx[0] + " DEFAULT " + xxx[1]
			} else {
				sqlCompare = "ALTER TABLE " + vv.Tbl + " CHANGE " + vv.Column_name + " " + vv.Column_name + " " + xxx[0]
			}
		} else {
			if xxx[1] != "" {
				sqlCompare = "ALTER TABLE " + vv.Tbl + " ADD " + vv.Column_name + " " + xxx[0] + " DEFAULT " + xxx[1]
			} else {
				sqlCompare = "ALTER TABLE " + vv.Tbl + " ADD " + vv.Column_name + " " + xxx[0]
			}
		}
		//	fmt.Println(sqlCompare)

		// popolo un array con tutte le query da fare
		allCompareSql = append(allCompareSql, sqlCompare)

		_, err = db2.Exec(sqlCompare)
		if err != nil {
			loggaErrore.Log += err.Error() + "\n"
			loggaErrore.Errore = -1

		} else {
			Logga(sqlCompare + "  ok")
		}
		// !!! fine blocco !!!

	}
	//os.Exit(0)

	// se facciamo il compare sui monoliti
	if ires.Monolith == 1 {
		//_, err = db.Exec("DROP DATABASE " + dbDataSrc)
		if err != nil {
			loggaErrore.Log = err.Error()
			loggaErrore.Errore = -1

		} else {
			Logga("DROP DATABASE " + dbDataSrc + "  ok")
		}
		//_, err = db.Exec("DROP DATABASE " + dbDataDst)
		if err != nil {
			loggaErrore.Log += err.Error() + "\n"
			loggaErrore.Errore = -1

		} else {
			Logga("DROP DATABASE " + dbDataSrc + "  ok")
		}
	} else {
		_, err = db.Exec("DROP DATABASE " + dbDataSrc)
		if err != nil {
			loggaErrore.Log += err.Error() + "\n"
			loggaErrore.Errore = -1

		} else {
			Logga("DROP DATABASE " + dbDataSrc + "  ok")
		}
	}

	Logga("Compare Database terminated")
	Logga("")

	return loggaErrore, allCompareSql
}
func Compareidx(dbDataName DbDataConnMs, dbMetaName DbMetaConnMs, db *sql.DB, db2 *sql.DB, db3 *sql.DB) (LoggaErrore, []string) {
	fmt.Println()
	fmt.Println("Compare Index")

	var loggaErrore LoggaErrore
	loggaErrore.Errore = 0

	var allCompareIdx []string

	dbData := dbDataName.DataName

	fmt.Println("Seek indexes on all tables on " + dbData)
	fmt.Println("For each table compare indexes between TB_INDEX and on information_schema on " + dbData)
	fmt.Println()

	sqlSel := " SELECT DISTINCT(table_name) as tableName "
	sqlSel += " FROM information_schema.columns where 1>0 "
	sqlSel += " and table_schema = '" + dbData + "' "
	sqlSel += " ORDER BY table_name"
	//fmt.Println(sqlSel)
	selDB, err := db.Query(sqlSel)
	if err != nil {
		loggaErrore.Log = err.Error()
		loggaErrore.Errore = -1
		return loggaErrore, allCompareIdx
	}

	var tableName string
	var idxsSrc []CompareIndex
	var idxsDst []CompareIndex
	var idxsMiss []CompareIndex
	for selDB.Next() {
		err = selDB.Scan(&tableName)
		if err != nil {
			loggaErrore.Log = err.Error()
			loggaErrore.Errore = -1
			return loggaErrore, allCompareIdx
		}

		codDimArr := strings.Split(tableName, "_")

		if len(codDimArr) > 2 {

			if codDimArr[1] == "ANAG" {
				codDim := strings.Replace(codDimArr[2], "00", "", -1)

				// cerco indici
				sqlIdx := " select NAME_IDX, UNIQUE_IDX, CODICE_IDX as COLUMN_NAME ,  concat (NAME_IDX,\":\", UNIQUE_IDX,\":\", SEQUENCE_IDX) as tbIndex "
				sqlIdx += " from TB_INDEX where COD_DIM = '" + codDim + "' "
				sqlIdx += " order by SEQUENCE_IDX"
				//fmt.Println(sqlIdx)
				selDB2, err := db3.Query(sqlIdx)
				if err != nil {
					loggaErrore.Log = err.Error()
					loggaErrore.Errore = -1
					return loggaErrore, allCompareIdx
				}
				var tbIndex, NAME_IDX, UNIQUE_IDX, COLUMN_NAME string

				for selDB2.Next() {
					err = selDB2.Scan(&NAME_IDX, &UNIQUE_IDX, &COLUMN_NAME, &tbIndex)
					if err != nil {
						loggaErrore.Log = err.Error()
						loggaErrore.Errore = -1
						return loggaErrore, allCompareIdx
					}

					var idxSrc CompareIndex
					idxSrc.Tbl = tableName
					idxSrc.Index = tbIndex
					idxSrc.NomeIdx = NAME_IDX
					idxSrc.NomeColonna = COLUMN_NAME
					idxSrc.Unique = UNIQUE_IDX
					idxsSrc = append(idxsSrc, idxSrc)
				}

				sqlIdx = " SELECT INDEX_NAME, case when NON_UNIQUE = 0 then 1 else 0 end  as UNIQUE_IDX, COLUMN_NAME, concat(INDEX_NAME,\":\",case when NON_UNIQUE = 0 then 1 else 0 end ,\":\",SEQ_IN_INDEX) as tbSchema "
				sqlIdx += " FROM INFORMATION_SCHEMA.STATISTICS WHERE 1>0 "
				sqlIdx += " and TABLE_SCHEMA = '" + dbData + "' "
				sqlIdx += " and INDEX_NAME!='PRIMARY' "
				sqlIdx += " and  TABLE_NAME = '" + tableName + "' "
				sqlIdx += " order by SEQ_IN_INDEX"
				// fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++")
				// fmt.Println(sqlIdx)
				selDB2, err = db3.Query(sqlIdx)
				if err != nil {
					loggaErrore.Log = err.Error()
					loggaErrore.Errore = -1
					return loggaErrore, allCompareIdx
				}
				var tbSchema, INDEX_NAME string
				for selDB2.Next() {
					err = selDB2.Scan(&INDEX_NAME, &UNIQUE_IDX, &COLUMN_NAME, &tbSchema)
					if err != nil {
						loggaErrore.Log = err.Error()
						loggaErrore.Errore = -1
						return loggaErrore, allCompareIdx
					}
					var idxDst CompareIndex
					idxDst.Tbl = tableName
					idxDst.NomeIdx = INDEX_NAME
					idxDst.Index = tbSchema
					idxDst.NomeColonna = COLUMN_NAME
					idxDst.Unique = UNIQUE_IDX
					idxsDst = append(idxsDst, idxDst)

				}
			}
		}
	}

	// fmt.Println(idxsSrc)
	// fmt.Println(idxsDst)

	// FA FEDE CIO CHE E DICHIARATO SU TB_INDEX
	fmt.Println("Elaborating differences ....")
	fmt.Println()
	for _, v := range idxsSrc {

		var find bool
		find = false
		for _, vv := range idxsDst {
			//fmt.Println(v.index + "|" + vv.index)
			if v.Index == vv.Index {
				find = true
			}

		}

		if find == false {
			var idxMiss CompareIndex
			idxMiss.Tbl = v.Tbl
			idxMiss.NomeIdx = v.NomeIdx
			idxMiss.Index = v.Index
			idxMiss.Unique = v.Unique
			idxMiss.NomeColonna = v.NomeColonna
			idxsMiss = append(idxsMiss, idxMiss)
		}
	}

	fmt.Println("Creating new indexes and editing the different ones")
	fmt.Println()

	if len(idxsMiss) > 0 {
		// fmt.Println(idxsMiss)
		// fmt.Println()
		// fmt.Println()

		var iddi = make(map[string]string)
		for _, v := range idxsMiss {
			iddi[v.Tbl+"."+v.NomeIdx] = v.Tbl + "." + v.NomeIdx
		}

		for _, v := range iddi {
			//fmt.Println(iddi[v])

			nomeIndiceArr := strings.Split(iddi[v], ".")
			codDimArr := strings.Split(nomeIndiceArr[0], "_")
			codDim := strings.Replace(codDimArr[2], "00", "", -1)

			sqlIdx := "select NAME_IDX, UNIQUE_IDX, CODICE_IDX as COLUMN_NAME "
			sqlIdx += "from TB_INDEX where 1>0 "
			sqlIdx += "and COD_DIM = '" + codDim + "' "
			sqlIdx += "and NAME_IDX = '" + nomeIndiceArr[1] + "' "
			sqlIdx += " order by SEQUENCE_IDX"
			//fmt.Println(sqlIdx)
			selDB2, err := db2.Query(sqlIdx)
			if err != nil {
				loggaErrore.Log = err.Error()
				loggaErrore.Errore = -1
				return loggaErrore, allCompareIdx
			}
			var NAME_IDX, UNIQUE_IDX, COLUMN_NAME string

			dropIdx := "DROP INDEX "
			createIdx := "CREATE "
			idx := 0
			for selDB2.Next() {
				err = selDB2.Scan(&NAME_IDX, &UNIQUE_IDX, &COLUMN_NAME)
				if err != nil {
					loggaErrore.Log = err.Error()
					loggaErrore.Errore = -1
					return loggaErrore, allCompareIdx
				}
				if idx == 0 {
					if UNIQUE_IDX == "1" {
						createIdx += " UNIQUE "
					}
					dropIdx += NAME_IDX + " on " + nomeIndiceArr[0]
					createIdx += " INDEX " + NAME_IDX + " on " + nomeIndiceArr[0] + " ( "
				}
				createIdx += COLUMN_NAME + ", "

				idx++
			}
			createIdx = createIdx[:len(createIdx)-2] + " ) "
			//fmt.Println(dropIdx)
			//fmt.Println(createIdx)

			allCompareIdx = append(allCompareIdx, dropIdx)
			_, err = db2.Exec(dropIdx)
			if err != nil {
				loggaErrore.Log = err.Error()
				loggaErrore.Errore = -1
				//return loggaErrore, allCompareIdx
			} else {
				//	fmt.Println(dropIdx + "  ok")
			}

			allCompareIdx = append(allCompareIdx, createIdx)
			_, err = db2.Exec(createIdx)
			if err != nil {

				loggaErrore.Log = err.Error()
				loggaErrore.Errore = -1
				return loggaErrore, allCompareIdx
			} else {
				//	fmt.Println(createIdx + "  ok")
			}

		}
	} else {
		fmt.Println("Indexes are OK")
		fmt.Println()
	}

	return loggaErrore, allCompareIdx
}
func RenameDatabases(dbMetaName DbMetaConnMs, masterDb MasterConn, db *sql.DB) {

	query := "DROP DATABASE IF EXISTS " + dbMetaName.MetaName + "_METAOLD"
	_, err := db.Exec(query)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(query + "  ok")
	}
	query = "CREATE DATABASE  " + dbMetaName.MetaName + "_METAOLD"
	_, err = db.Exec(query)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(query + "  ok")
	}

	// dump metadato
	queryCommand := "mysqldump -u" + masterDb.User + " -p" + masterDb.Pass + " -h" + dbMetaName.MetaHost + " " + dbMetaName.MetaName + " > /tmp/" + dbMetaName.MetaName + ".sql"
	fmt.Println(queryCommand)
	cmd := exec.Command("bash", "-c", queryCommand)
	_, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	//dump canarino
	queryCommand = "mysqldump -u" + masterDb.User + " -p" + masterDb.Pass + " -h" + dbMetaName.MetaHost + " " + dbMetaName.MetaName + "_canary > /tmp/" + dbMetaName.MetaName + "_canary.sql"
	fmt.Println(queryCommand)
	cmd = exec.Command("bash", "-c", queryCommand)
	_, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	query = "DROP DATABASE IF EXISTS " + dbMetaName.MetaName
	_, err = db.Exec(query)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(query + "  ok")
	}
	query = "CREATE DATABASE  " + dbMetaName.MetaName
	_, err = db.Exec(query)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(query + "  ok")
	}

	// creo gli user
	CreateUser(dbMetaName, db)

	//dump IN metadato
	queryCommand = "mysql -u" + masterDb.User + " -p" + masterDb.Pass + " -h" + dbMetaName.MetaHost + " " + dbMetaName.MetaName + " < /tmp/" + dbMetaName.MetaName + "_canary.sql"
	fmt.Println(queryCommand)
	cmd = exec.Command("bash", "-c", queryCommand)
	_, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	//dump IN METAOLD
	queryCommand = "mysql -u" + masterDb.User + " -p" + masterDb.Pass + " -h" + dbMetaName.MetaHost + " " + dbMetaName.MetaName + "_METAOLD < /tmp/" + dbMetaName.MetaName + ".sql"
	fmt.Println(queryCommand)
	cmd = exec.Command("bash", "-c", queryCommand)
	_, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
}
func GetMasterConn(gruppoDeveloper, cluster, devopsToken string) (MasterConn, LoggaErrore) {

	Logga("getMasterConn")
	Logga("Cluster: " + cluster)
	Logga("Gruppo: " + gruppoDeveloper)

	if gruppoDeveloper == "" && cluster == "" {
		Logga("BOTH GROUP AND CLUSTER MISSING")
		debug.PrintStack()
		//os.Exit(0)
	}

	var erro LoggaErrore
	erro.Errore = 0

	var master MasterConn

	/*
		se al metodo NON passo il cluster
		lo cerco partendo dal gruppo developer
	*/
	if cluster == "" {
		// ottengo lo stage
		gruppo, erro := GetUserGroup(devopsToken, gruppoDeveloper)
		if erro.Errore < 0 {
			Logga("getUserGroup")
			Logga(erro.Log)
		}
		cluster = gruppo["stage"]
	}

	/* ************************************************************************************************ */
	// KUBECLUSTER
	Logga("Getting KUBECLUSTER MASTER CONN")

	argsClu := make(map[string]string)
	argsClu["source"] = "devops-8"
	argsClu["$select"] = "XKUBECLUSTER08,XKUBECLUSTER09,XKUBECLUSTER10,XKUBECLUSTER11,XKUBECLUSTER15,XKUBECLUSTER20"
	argsClu["center_dett"] = "dettaglio"
	argsClu["$filter"] = "equals(XKUBECLUSTER03,'" + cluster + "') "

	restyKubeCluRes := ApiCallGET(false, argsClu, "msdevops", "/devops/KUBECLUSTER", devopsToken, "")
	if restyKubeCluRes.Errore < 0 {
		erro.Errore = -1
		erro.Log = restyKubeCluRes.Log
		return master, erro

	}

	if len(restyKubeCluRes.BodyJson) > 0 {
		master.Host = restyKubeCluRes.BodyJson["XKUBECLUSTER08"].(string)
		master.MetaName = restyKubeCluRes.BodyJson["XKUBECLUSTER09"].(string)
		master.User = restyKubeCluRes.BodyJson["XKUBECLUSTER10"].(string)
		master.Pass = restyKubeCluRes.BodyJson["XKUBECLUSTER11"].(string)
		master.Domain = restyKubeCluRes.BodyJson["XKUBECLUSTER15"].(string)
		master.AccessToken = restyKubeCluRes.BodyJson["XKUBECLUSTER20"].(string)
		master.Cluster = cluster
		Logga("KUBECLUSTER MASTER CONN OK")
	} else {
		Logga("KUBECLUSTER MASTER CONN MISSING")
	}
	Logga("")
	/* ************************************************************************************************ */

	if cluster == "" {
		fmt.Println("CLUSTER MISSING")
		//debug.PrintStack()

	}
	return master, erro
}
