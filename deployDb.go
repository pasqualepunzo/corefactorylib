package lib

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime/debug"
	"strconv"

	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func DropMetadato(ctx context.Context, dbMetaName DbMetaConnMs, db *sql.DB) error {

	Logga(ctx, os.Getenv("JsonLog"), "Drop metadata :"+dbMetaName.MetaName)

	_, err := db.Exec("drop database if exists " + dbMetaName.MetaName)
	if err != nil {
		return err
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "Database "+dbMetaName.MetaName+" dropped")
	}

	return nil
}
func CreateDbMeta(ctx context.Context, dbMetaName DbMetaConnMs, db *sql.DB) error {

	query := "CREATE DATABASE " + dbMetaName.MetaName
	Logga(ctx, os.Getenv("JsonLog"), query)
	_, err := db.Exec(query)
	if err != nil {
		return err
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "Database "+dbMetaName.MetaName+" instance done")
	}

	// creo gli user
	CreateUser(ctx, dbMetaName, db)

	return nil
}
func CreateDbData(ctx context.Context, dbDataName DbDataConnMs, db *sql.DB) error {

	// creazione dei DATABASE

	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + dbDataName.DataName)
	if err != nil {
		return err
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "Create Database "+dbDataName.DataName+" instance done")
	}

	// creo gli user
	var cons DbMetaConnMs
	cons.MetaHost = dbDataName.DataHost
	cons.MetaName = dbDataName.DataName
	cons.MetaPass = dbDataName.DataPass
	cons.MetaUser = dbDataName.DataUser
	erro := CreateUser(ctx, cons, db)
	if erro != nil {
		return erro
	}

	return nil
}
func CreateUser(ctx context.Context, dbMetaName DbMetaConnMs, db *sql.DB) error {

	// FAC-753
	var _user string
	query := "SELECT user as _user FROM mysql.user where user = '" + dbMetaName.MetaUser + "'"

	row := db.QueryRow(query)
	errUser := row.Scan(&_user)

	if errUser != nil {

		if errUser.Error() == "sql: no rows in result set" {

			// create user
			query := "CREATE USER   '" + dbMetaName.MetaUser + "'@'%' IDENTIFIED BY '" + dbMetaName.MetaPass + "'"
			Logga(ctx, os.Getenv("JsonLog"), query)
			_, err := db.Exec(query)
			if err != nil {
				return err
			} else {
				Logga(ctx, os.Getenv("JsonLog"), "CREATE USER    "+dbMetaName.MetaUser+" done")
			}

			// grant su metadati
			query = "GRANT ALL PRIVILEGES ON " + dbMetaName.MetaName + ".* TO '" + dbMetaName.MetaUser + "'@'%' WITH GRANT OPTION "
			Logga(ctx, os.Getenv("JsonLog"), query)
			_, err = db.Exec(query)
			//Logga(ctx, os.Getenv("JsonLog"), "GRANT ALL PRIVILEGES ON " + ires.MetaName + ".* TO '" + ires.MetaUser + "'@'%'")
			if err != nil {
				return err
			} else {
				Logga(ctx, os.Getenv("JsonLog"), "GRANT ON "+dbMetaName.MetaName+" created")
			}

			// grant su data
			query = "FLUSH PRIVILEGES"
			Logga(ctx, os.Getenv("JsonLog"), query)
			_, err = db.Exec(query)
			if err != nil {
				return err
			} else {
				Logga(ctx, os.Getenv("JsonLog"), "FLUSH PRIVILEGES "+dbMetaName.MetaName+" done")
			}
		} else {
			return errUser
		}

	} else {

		Logga(ctx, os.Getenv("JsonLog"), "User: "+dbMetaName.MetaUser+" already exists")
	}

	return nil
}
func DropDbData(ctx context.Context, dbDataName DbDataConnMs, db *sql.DB) error {

	_, err := db.Exec("drop database if exists " + dbDataName.DataName)
	if err != nil {
		return err
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "Database "+dbDataName.DataName+" dropped")
	}
	return nil
}
func Comparedb(ctx context.Context, ires IstanzaMicro, dbDataName DbDataConnMs, db *sql.DB, db2 *sql.DB, doQueryExec bool) ([]string, []string, error) {

	var allCompareSql []string
	var allCompareSqlError []string

	var dbDataSrc string

	dbDataSrc = dbDataName.DataName + "_ccd_nuovo"
	dbDataDst := dbDataName.DataName

	// se facciamo il compare sui monoliti
	if ires.Monolith == 1 {
		dbDataSrc = dbDataName.DataName + "_ccd_nuovo"
		dbDataDst = dbDataName.DataName + "_ccd_prod"
	}

	Logga(ctx, os.Getenv("JsonLog"), "")
	Logga(ctx, os.Getenv("JsonLog"), "*********")
	Logga(ctx, os.Getenv("JsonLog"), "Source Database: "+dbDataSrc)
	Logga(ctx, os.Getenv("JsonLog"), "Destination Database: "+dbDataDst)
	Logga(ctx, os.Getenv("JsonLog"), "*********")
	Logga(ctx, os.Getenv("JsonLog"), "")

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
		return allCompareSql, allCompareSqlError, errors.New(err.Error() + " - " + sqlSel)
	}

	var srcTbls []CompareDbRes
	tablesList := "(\""
	for selDB.Next() {

		var tbl CompareDbRes

		err = selDB.Scan(&table_name, &column_name, &columns)
		if err != nil {
			return allCompareSql, allCompareSqlError, err
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
	//fmt.Println(sqlSel)
	// os.Exit(0)
	selDB, err = db.Query(sqlSel)
	if err != nil {
		return allCompareSql, allCompareSqlError, err
	}

	var dstTbls []CompareDbRes
	for selDB.Next() {
		err = selDB.Scan(&table_name, &column_name, &columns)
		if err != nil {
			return allCompareSql, allCompareSqlError, err
		}
		var tbl CompareDbRes
		tbl.Tbl = table_name
		tbl.Column_name = column_name
		tbl.Columns = columns
		dstTbls = append(dstTbls, tbl)
	}

	// fmt.Println("SRC")
	// LogJson(srcTbls)
	// fmt.Println("DST")
	// LogJson(dstTbls)
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

			// mi segno se la colonna è uguale
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
	// fmt.Println("DIFF")
	// LogJson(diffTbls)
	// fmt.Println(diffTbls)
	// os.Exit(0)

	Logga(ctx, os.Getenv("JsonLog"), "Get all diff")
	Logga(ctx, os.Getenv("JsonLog"), "")
	Logga(ctx, os.Getenv("JsonLog"), "STO PER APPLICARE LE DIFF")
	Logga(ctx, os.Getenv("JsonLog"), "Change Database Structure on "+dbDataName.DataName)
	Logga(ctx, os.Getenv("JsonLog"), dbDataName.DataHost+"|"+dbDataName.DataName)
	//fmt.Println(missingTbls)

	// **************************************************************************
	// **************************************************************************
	// **************************************************************************
	// da qui in poi si applica cio che e stato calcolato

	for k := range missingTbls {
		// sqlCompare := "CREATE TABLE IF NOT EXISTS " + dbDataDst + "." + k + " like " + dbDataSrc + "." + k
		sqlCompare := "CREATE TABLE IF NOT EXISTS " + k + " like " + dbDataSrc + "." + k

		// popolo un array con tutte le query da fare
		allCompareSql = append(allCompareSql, sqlCompare)

		// TODO - PORTA FUORI
		if doQueryExec {
			_, err = db2.Exec(sqlCompare)
			if err != nil {
				allCompareSqlError = append(allCompareSqlError, sqlCompare)
			} else {
				// Logga(ctx, os.Getenv("JsonLog"), sqlCompare+" ok")
			}
		}
	}

	var sqlCompare string
	var columnExists bool
	columnExists = false

	fmt.Println(" =============================== diffTbls")
	fmt.Println(diffTbls)

	for _, vv := range diffTbls {

		// poiche la madonna di mysql non contempla add column if not exist sono costretto a tirare le madonne ...
		var COLUMN_NAME string
		sqlCheck := "SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS "
		sqlCheck += "WHERE TABLE_SCHEMA='" + dbDataDst + "' "
		sqlCheck += "AND TABLE_NAME='" + vv.Tbl + "' "
		sqlCheck += "AND COLUMN_NAME='" + vv.Column_name + "' "
		sqlCheckRes, errcheck := db.Query(sqlCheck)
		if errcheck != nil {
			return allCompareSql, allCompareSqlError, errors.New(err.Error() + " - " + sqlCheck)
		}

		for sqlCheckRes.Next() {
			err = sqlCheckRes.Scan(&COLUMN_NAME)
			if err != nil {
				return allCompareSql, allCompareSqlError, err
			}
			if COLUMN_NAME == vv.Column_name {
				columnExists = true
			}
		}

		// !!! questo blocco è clone del seguente !!!
		xxx := strings.Split(vv.Columns, ":")
		if vv.Tipo == "CHANGE" {
			if columnExists {
				if xxx[1] != "" && xxx[1] != "NULL" {
					// sqlCompare = "ALTER TABLE " + vv.Tbl + " MODIFY " + vv.Column_name + " " + vv.Column_name + " " + xxx[0] + " DEFAULT " + xxx[1]
					sqlCompare = "ALTER TABLE " + vv.Tbl + " MODIFY " + vv.Column_name + " " + xxx[0] + " DEFAULT " + xxx[1]
				} else {
					// sqlCompare = "ALTER TABLE " + vv.Tbl + " MODIFY " + vv.Column_name + " " + vv.Column_name + " " + xxx[0]
					sqlCompare = "ALTER TABLE " + vv.Tbl + " MODIFY " + vv.Column_name + " " + xxx[0]
				}
			} else {
				allCompareSqlError = append(allCompareSqlError, sqlCompare+" - Cannot MODIFY "+vv.Column_name+" column missing")
				fmt.Println("Cannot MODIFY " + vv.Column_name + " column missing")
			}

		} else {
			if !columnExists {
				if xxx[1] != "" {
					sqlCompare = "ALTER TABLE " + vv.Tbl + " ADD " + vv.Column_name + " " + xxx[0] + " DEFAULT " + xxx[1]
				} else {
					sqlCompare = "ALTER TABLE " + vv.Tbl + " ADD " + vv.Column_name + " " + xxx[0]
				}
			} else {
				allCompareSqlError = append(allCompareSqlError, sqlCompare+" - Cannot add "+vv.Column_name+" column exists")
				fmt.Println("Cannot add " + vv.Column_name + " column exists")
			}
		}
		//	fmt.Println(sqlCompare)

		// popolo un array con tutte le query da fare
		allCompareSql = append(allCompareSql, sqlCompare)

		// SE DEVO APPLICARE LE DIFFERENCE (ddcm)
		if doQueryExec {
			_, err = db2.Exec(sqlCompare)
			if err != nil {
				allCompareSqlError = append(allCompareSqlError, err.Error()+" - "+sqlCompare)

			} else {
				Logga(ctx, os.Getenv("JsonLog"), sqlCompare+"  ok")
			}
		}
		// !!! fine blocco !!!

	}

	Logga(ctx, os.Getenv("JsonLog"), "Compare Database terminated")
	Logga(ctx, os.Getenv("JsonLog"), "")

	return allCompareSql, allCompareSqlError, nil
}
func Compareidx(dbDataName DbDataConnMs, dbMetaName DbMetaConnMs, db *sql.DB, db2 *sql.DB, db3 *sql.DB, doQueryExec bool) ([]string, []string, error) {
	fmt.Println()
	fmt.Println("Compare Index")

	var allCompareIdx []string
	var allCompareIdxError []string

	dbData := dbDataName.DataName

	fmt.Println("Seek indexes on all tables on " + dbData)
	fmt.Println("For each table compare indexes between TB_INDEX and on information_schema on " + dbData)
	fmt.Println()

	sqlSel := " SELECT table_name as tableName "
	sqlSel += " FROM information_schema.tables where 1>0 "
	sqlSel += " and table_schema = '" + dbData + "' "
	sqlSel += " AND TABLE_TYPE = 'BASE TABLE'  AND table_name like 'TB_ANAG_%00' "
	sqlSel += " ORDER BY table_name"
	//fmt.Println(sqlSel)
	selDB, err := db.Query(sqlSel)
	if err != nil {
		return allCompareIdx, allCompareIdxError, errors.New(err.Error() + " - " + sqlSel)
	}

	var tableName string
	var idxsSrc []CompareIndex
	var idxsDst []CompareIndex
	var idxsMiss []CompareIndex
	for selDB.Next() {
		err = selDB.Scan(&tableName)
		if err != nil {
			return allCompareIdx, allCompareIdxError, err
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
					return allCompareIdx, allCompareIdxError, errors.New(err.Error() + " - " + sqlIdx)
				}
				var tbIndex, NAME_IDX, UNIQUE_IDX, COLUMN_NAME string

				for selDB2.Next() {
					err = selDB2.Scan(&NAME_IDX, &UNIQUE_IDX, &COLUMN_NAME, &tbIndex)
					if err != nil {
						return allCompareIdx, allCompareIdxError, err
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
				selDB2, err = db.Query(sqlIdx)
				if err != nil {
					return allCompareIdx, allCompareIdxError, errors.New(err.Error() + " - " + sqlIdx)
				}
				var tbSchema, INDEX_NAME string
				for selDB2.Next() {
					err = selDB2.Scan(&INDEX_NAME, &UNIQUE_IDX, &COLUMN_NAME, &tbSchema)
					if err != nil {
						return allCompareIdx, allCompareIdxError, err
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

	if len(idxsMiss) > 0 {

		fmt.Println("Find unique indexes to drop when adding new one")
		fmt.Println()

		var tbls = make(map[string]string)
		for _, v := range idxsMiss {
			if v.Unique == "1" {
				tbls[v.Tbl] = v.Tbl
			}
		}

		var idxsToDrop []CompareIndex
		for _, v := range idxsDst {

			if v.Unique == "1" {
				_, ok := tbls[v.Tbl]
				if ok {
					var find bool
					find = false

					for _, vv := range idxsSrc {
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
						fmt.Println("Append " + idxMiss.NomeIdx + " to list idxsToDrop")
						idxsToDrop = append(idxsToDrop, idxMiss)
					}
				}

			}
		}

		if len(idxsToDrop) > 0 {
			fmt.Println("Dropping Unique Indexes that will be replaced")
			fmt.Println()
			var iddi = make(map[string]string)
			for _, v := range idxsToDrop {
				iddi[v.Tbl+"."+v.NomeIdx] = v.Tbl + "." + v.NomeIdx
			}

			for _, v := range iddi {

				nomeIndiceArr := strings.Split(iddi[v], ".")

				// cerco se esiste l'indice

				sqlCheck := " SELECT INDEX_NAME as Key_name "
				sqlCheck += " FROM INFORMATION_SCHEMA.STATISTICS WHERE 1>0 "
				sqlCheck += " and TABLE_SCHEMA = '" + dbDataName.DataName + "' "
				sqlCheck += " and INDEX_NAME='" + nomeIndiceArr[1] + "' "
				sqlCheck += " and  TABLE_NAME = '" + nomeIndiceArr[0] + "' "

				fmt.Println(sqlCheck)
				sqlCheckRes, errcheck := db.Query(sqlCheck)
				// LogJson(sqlCheckRes)
				// LogJson(errcheck)
				if errcheck != nil {
					return allCompareIdx, allCompareIdxError, errors.New(err.Error() + " - " + sqlCheck)
				}

				indexExists := false
				var Key_name string
				for sqlCheckRes.Next() {
					err = sqlCheckRes.Scan(&Key_name)
					if err != nil {
						return allCompareIdx, allCompareIdxError, err
					}
					fmt.Printf("++++++++++++++++++")
					fmt.Printf("Key_name: " + Key_name)
					if Key_name != "" {
						indexExists = true
					}
				}

				//fmt.Println(iddi[v])
				dropIdx := "DROP INDEX  " + nomeIndiceArr[1] + " on " + dbDataName.DataName + "." + nomeIndiceArr[0]

				if indexExists {
					allCompareIdx = append(allCompareIdx, dropIdx)
					if doQueryExec {
						_, err = db2.Exec(dropIdx)
						if err != nil {
							allCompareIdxError = append(allCompareIdxError, dropIdx)
						} else {
							//	fmt.Println(dropIdx + "  ok")
						}
					}
				} else {
					allCompareIdx = append(allCompareIdx, dropIdx+" - KO")
					allCompareIdxError = append(allCompareIdxError, dropIdx+" - KO NON DROPPATO PERCHE MANCA")
				}
			}
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

			// cerco se esiste l'indice
			indexExistsONCreate := false

			sqlCheckIdx := " SELECT INDEX_NAME as Key_name "
			sqlCheckIdx += " FROM INFORMATION_SCHEMA.STATISTICS WHERE 1>0 "
			sqlCheckIdx += " and TABLE_SCHEMA = '" + dbDataName.DataName + "' "
			sqlCheckIdx += " and INDEX_NAME='" + nomeIndiceArr[1] + "' "
			sqlCheckIdx += " and  TABLE_NAME = '" + nomeIndiceArr[0] + "' "

			fmt.Println(sqlCheckIdx)
			sqlCheckResIdx, errcheckIdx := db.Query(sqlCheckIdx)
			if errcheckIdx != nil {
				return allCompareIdx, allCompareIdxError, errors.New(err.Error() + " - " + sqlCheckIdx)
			}

			var Key_name string
			for sqlCheckResIdx.Next() {
				err = sqlCheckResIdx.Scan(&Key_name)
				if err != nil {
					return allCompareIdx, allCompareIdxError, err
				}
				fmt.Printf("++++++++++++++++++")
				fmt.Printf("Key_name: " + Key_name)
				if Key_name != "" {
					indexExistsONCreate = true
				}
			}

			sqlIdx := "select COD_DIM, NAME_IDX, UNIQUE_IDX, CODICE_IDX as COLUMN_NAME "
			sqlIdx += "from TB_INDEX where 1>0 "
			sqlIdx += "and COD_DIM = '" + codDim + "' "
			sqlIdx += "and NAME_IDX = '" + nomeIndiceArr[1] + "' "
			sqlIdx += " order by SEQUENCE_IDX"
			fmt.Println(sqlIdx)
			selDB2, err := db3.Query(sqlIdx)
			if err != nil {
				return allCompareIdx, allCompareIdxError, errors.New(err.Error() + " - " + sqlIdx)
			}
			var COD_DIM, NAME_IDX, UNIQUE_IDX, COLUMN_NAME string

			dropIdx := "DROP INDEX  "
			createIdx := "CREATE "
			idx := 0

			var culumnExists bool
			culumnExists = true
			var columnMissing []string
			indexFieldList := ""

			for selDB2.Next() {
				err = selDB2.Scan(&COD_DIM, &NAME_IDX, &UNIQUE_IDX, &COLUMN_NAME)
				if err != nil {
					return allCompareIdx, allCompareIdxError, err
				}

				// poiche la madonna di mysql non contempla add column if not exist sono costretto a tirare le madonne ...
				var COLUMN_NAME_IDX string
				sqlCheck := "SELECT COLUMN_NAME as COLUMN_NAME_IDX FROM INFORMATION_SCHEMA.COLUMNS "
				sqlCheck += "WHERE TABLE_SCHEMA='" + dbData + "' "
				sqlCheck += "and TABLE_NAME = 'TB_ANAG_" + COD_DIM + "00' "
				sqlCheck += "AND COLUMN_NAME='" + COLUMN_NAME + "' "
				fmt.Println(sqlCheck)
				sqlCheckRes, errcheck := db.Query(sqlCheck)
				//LogJson(sqlCheckRes)
				if errcheck != nil {
					return allCompareIdx, allCompareIdxError, errors.New(err.Error() + " - " + sqlCheck)
				}

				for sqlCheckRes.Next() {
					err = sqlCheckRes.Scan(&COLUMN_NAME_IDX)

					if err != nil {
						return allCompareIdx, allCompareIdxError, err
					}
					if COLUMN_NAME_IDX == "" {
						culumnExists = false
						columnMissing = append(columnMissing, COLUMN_NAME_IDX)
						fmt.Println("The column for key is missing")
					} else {
						indexFieldList += COLUMN_NAME_IDX + ", "
						fmt.Println("The column " + COLUMN_NAME_IDX + " for key exists")
					}
				}
				// -------------------------
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

			// facciamo un check sui record
			if UNIQUE_IDX == "1" {
				var UNIQUE_CHECK_NUM int
				sqlIndexUniqueCheck := " select count(*) as num "
				sqlIndexUniqueCheck += " from " + dbDataName.DataName + "." + nomeIndiceArr[0]
				sqlIndexUniqueCheck += " group by " + indexFieldList[:len(indexFieldList)-2] + " "
				sqlIndexUniqueCheck += " having num > 1"

				fmt.Println(sqlIndexUniqueCheck)
				sqlCheckUniqueRes, errcheckUnique := db.Query(sqlIndexUniqueCheck)
				if errcheckUnique != nil {

				}
				idxUni := 0
				for sqlCheckUniqueRes.Next() {
					err = sqlCheckUniqueRes.Scan(&UNIQUE_CHECK_NUM)
					idxUni++
					break
				}
				if idxUni > 0 {
					allCompareIdxError = append(allCompareIdxError, "THE UNIQUE INDEX "+NAME_IDX+" CANNOT BE CREATED DUE unique index constraint violation ERROR ")
				}
			}
			// FINE facciamo un check sui record

			createIdx = createIdx[:len(createIdx)-2] + " ) "
			fmt.Println(dropIdx)
			fmt.Println(createIdx)
			fmt.Println("culumnExists", culumnExists)
			fmt.Println("indexExistsONCreate", indexExistsONCreate)

			// se esistono le colonne
			if culumnExists {
				// se l'indice esiste
				if indexExistsONCreate {
					allCompareIdx = append(allCompareIdx, dropIdx)
					if doQueryExec {
						_, err = db.Exec(dropIdx)
						if err != nil {
							allCompareIdxError = append(allCompareIdxError, err.Error()+" - "+dropIdx)
						} else {
							//	fmt.Println(dropIdx + "  ok")
						}
					}
				} else {
					fmt.Println("CANT DROP INDEX BEACAUSE IT DOES NOT EXISTS:" + dropIdx)
				}

				allCompareIdx = append(allCompareIdx, createIdx)
				//fmt.Println("CUSTOM PERFORM CREATE INDEX:" + createIdx)
				if doQueryExec {
					_, err = db.Exec(createIdx)
					if err != nil {
						return allCompareIdx, allCompareIdxError, errors.New(err.Error() + " - " + createIdx)
					} else {
						//	fmt.Println(createIdx + "  ok")
					}
				}
			} else {
				allCompareIdx = append(allCompareIdx, dropIdx+" - KO")
				allCompareIdx = append(allCompareIdx, createIdx+" - KO")
				var clmnmiss string
				for _, mm := range columnMissing {
					clmnmiss += mm + ", "
				}

				return allCompareIdx, allCompareIdxError, errors.New("cannot add or drop indexs because some columns: " + clmnmiss + " are missing on " + dbDataName.DataName)
			}

		}
	} else {
		fmt.Println("Indexes are OK")
		fmt.Println()
	}

	return allCompareIdx, allCompareIdxError, nil
}
func RenameDatabases(ctx context.Context, dbMetaName DbMetaConnMs, masterDb MasterConn, db *sql.DB) {

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

	mysqldumpParameters := "--complete-insert=true --extended-insert=false --skip-comments"
	// dump metadato
	queryCommand := "mysqldump -u" + masterDb.User + " -p" + masterDb.Pass + " -h" + dbMetaName.MetaHost + " " + mysqldumpParameters + " " + dbMetaName.MetaName + " > /tmp/" + dbMetaName.MetaName + ".sql"
	fmt.Println(queryCommand)
	cmd := exec.Command("bash", "-c", queryCommand)
	_, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	//dump canarino
	queryCommand = "mysqldump -u" + masterDb.User + " -p" + masterDb.Pass + " -h" + dbMetaName.MetaHost + " " + mysqldumpParameters + " " + dbMetaName.MetaName + "_canary > /tmp/" + dbMetaName.MetaName + "_canary.sql"
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
	CreateUser(ctx, dbMetaName, db)

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

	query = "DROP DATABASE IF EXISTS " + dbMetaName.MetaName + "_METAOLD"
	_, err = db.Exec(query)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(query + "  ok")
	}
}
func GetMasterConn(ctx context.Context, gruppoDeveloper, cluster, devopsToken, devopsTokenDst, enviro, dominio, coreApiVersion string, monolith int32) (MasterConn, error) {

	Logga(ctx, os.Getenv("JsonLog"), "getMasterConn")
	Logga(ctx, os.Getenv("JsonLog"), "Cluster: "+cluster)
	Logga(ctx, os.Getenv("JsonLog"), "Gruppo: "+gruppoDeveloper)

	devops := "devops"
	if monolith == 1 {
		devops = "devopsmono"
	}

	if gruppoDeveloper == "" && cluster == "" {
		Logga(ctx, os.Getenv("JsonLog"), "BOTH GROUP AND CLUSTER MISSING")
		debug.PrintStack()
		//os.Exit(0)
	}

	var erro error

	var master MasterConn

	/*
		se al metodo NON passo il cluster
		lo cerco partendo dal gruppo developer
	*/
	if cluster == "" {
		// ottengo lo stage
		gruppo, erro := GetUserGroup(ctx, devopsToken, gruppoDeveloper, dominio, coreApiVersion)
		if erro != nil {
			Logga(ctx, os.Getenv("JsonLog"), "getUserGroup")
			Logga(ctx, os.Getenv("JsonLog"), erro.Error())
		}
		cluster = gruppo["stage"]
	}

	/* ************************************************************************************************ */
	// KUBECLUSTER
	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBECLUSTER MASTER CONN")

	argsClu := make(map[string]string)
	argsClu["source"] = "devops-8"
	argsClu["$select"] = "XKUBECLUSTER12,XKUBECLUSTER06,XKUBECLUSTER08,XKUBECLUSTER09,XKUBECLUSTER10,XKUBECLUSTER11,XKUBECLUSTER15,XKUBECLUSTER20"
	argsClu["center_dett"] = "dettaglio"
	argsClu["$filter"] = "equals(XKUBECLUSTER03,'" + cluster + "') "

	restyKubeCluRes, errKubeCluRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsClu, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBECLUSTER", devopsTokenDst, dominio, coreApiVersion)
	if errKubeCluRes != nil {
		return master, errKubeCluRes

	}

	if len(restyKubeCluRes.BodyJson) > 0 {
		master.Host = restyKubeCluRes.BodyJson["XKUBECLUSTER08"].(string)
		master.MetaName = restyKubeCluRes.BodyJson["XKUBECLUSTER09"].(string)
		master.User = restyKubeCluRes.BodyJson["XKUBECLUSTER10"].(string)
		master.Pass = restyKubeCluRes.BodyJson["XKUBECLUSTER11"].(string)
		master.Domain = restyKubeCluRes.BodyJson["XKUBECLUSTER15"].(string)
		master.AccessToken = restyKubeCluRes.BodyJson["XKUBECLUSTER20"].(string)
		master.Cluster = cluster
		Logga(ctx, os.Getenv("JsonLog"), "KUBECLUSTER MASTER CONN OK")

		/**
		Andiamo a vedere se esiste un record in KUBECLUSTERENV che fa l'overwrite di alcune proprietà di
		KUBECLUSTER in base all'env
		**/

		ambienteFloat := restyKubeCluRes.BodyJson["XKUBECLUSTER12"].(float64)

		argsCluEnv := make(map[string]string)
		argsCluEnv["source"] = "devops-8"
		argsCluEnv["center_dett"] = "dettaglio"
		argsCluEnv["$select"] = "XKUBECLUSTERENV07"
		argsCluEnv["$filter"] = "equals(XKUBECLUSTERENV03,'" + cluster + "') "
		argsCluEnv["$filter"] += " and equals(XKUBECLUSTERENV04,'" + restyKubeCluRes.BodyJson["XKUBECLUSTER06"].(string) + "') "
		argsCluEnv["$filter"] += " and XKUBECLUSTERENV05 eq " + strconv.Itoa(int(ambienteFloat)) + " "
		argsCluEnv["$filter"] += " and equals(XKUBECLUSTERENV06,'" + enviro + "') "

		restyKubeCluEnvRes, errKubeCluEnvRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsCluEnv, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBECLUSTERENV", devopsTokenDst, dominio, coreApiVersion)
		if errKubeCluEnvRes != nil {
			return master, errKubeCluEnvRes
		}

		if len(restyKubeCluEnvRes.BodyJson) > 0 {
			metanameCluEnv, _ := restyKubeCluEnvRes.BodyJson["XKUBECLUSTERENV07"].(string)
			if metanameCluEnv != "" {
				master.MetaName = metanameCluEnv
			}
			Logga(ctx, os.Getenv("JsonLog"), "KUBECLUSTERENV MASTER CONN OK")
		}
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "KUBECLUSTER MASTER CONN MISSING")
	}
	Logga(ctx, os.Getenv("JsonLog"), "")
	/* ************************************************************************************************ */

	if cluster == "" {
		fmt.Println("CLUSTER MISSING")
		//debug.PrintStack()

	}
	return master, erro
}
