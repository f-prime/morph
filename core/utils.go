package core

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

func Perr(e error) {
  perr(e)
}

func perr(e error) {
  if e != nil {
    panic(e)
  }
}

func HashString(s string) string {
  h := sha1.New()
  h.Write([]byte(s))
  r := h.Sum(nil)
  return hex.EncodeToString(r)
}

func HashFile(p string) string {
  data, e := os.ReadFile(p)
  perr(e)
  sdata := string(data)
  return HashString(sdata) 
}


func DoesPathExist(path string) bool {
  _, err := os.Stat(path)

  if os.IsNotExist(err) {
    return false
  }

  return true
}

func ListAllFilesInPath(path string) []string {
  var files []string
  err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }

    if !info.IsDir() && strings.HasSuffix(path, ".sql") {
      files = append(files, path)
    }

    return nil 
  })

  perr(err)
  return files
}

type statements struct {
  stmt *string
  stmtName *string
  stmtHash *string
  stmtType int
}

func getListOfStatementsInDb(ctx *Context) []statements {
  var stmts []statements

  allStmts, e := ctx.Db.Query("select stmt, stmt_name, stmt_hash, stmt_type from schemaflow.statements")
  perr(e)

  for allStmts.Next() {
    var stmt, stmt_name, stmt_hash *string
    var stmt_type int

    perr(allStmts.Scan(&stmt, &stmt_name, &stmt_hash, &stmt_type))
    stmts = append(stmts, statements { stmt, stmt_name, stmt_hash, stmt_type})
  }

  return stmts
}

type executedMigration struct {
  fileName string
  fileHash string
}

func getListOfExecutedMigrationFiles(ctx *Context) []executedMigration{
  var executedMigrations []executedMigration

  migrations, e := ctx.Db.Query("select file_name, file_hash from schemaflow.migrations")
  perr(e)

  for migrations.Next() {
    var file_name, file_hash string;

    migrations.Scan(&file_name, &file_hash)

    executedMigrations = append(executedMigrations, executedMigration { file_name, file_hash })
  }

  return executedMigrations
}

func removeStmtByHash(ctx *Context, hash string) {
  _, e := ctx.DbTx.Exec("delete from schemaflow.statements where stmt_hash=$1", hash)
  perr(e)
}

func updateStmtInDb(ctx *Context, stmt *ParsedStmt) {
  if stmt.HasName {
    _, err := ctx.DbTx.Exec("delete from schemaflow.statements where stmt_name=$1 and stmt_type=$2", stmt.Name, stmt.StmtType)
    perr(err)
  }

  addStmtToDb(ctx, stmt)
}

func addStmtToDb(ctx *Context, stmt *ParsedStmt) {
  if stmt.HasName {
    _, err := ctx.DbTx.Exec("insert into schemaflow.statements (stmt, stmt_hash, stmt_type, stmt_name) values ($1, $2, $3, $4) on conflict (stmt_hash) do nothing", stmt.Deparsed, stmt.Hash, stmt.StmtType, stmt.Name) 
    perr(err)
  } else {
    _, err := ctx.DbTx.Exec("insert into schemaflow.statements (stmt, stmt_hash, stmt_type) values ($1, $2, $3) on conflict (stmt_hash) do nothing", stmt.Deparsed, stmt.Hash, stmt.StmtType) 

    if err != nil {
      fmt.Printf("ERROR WITH: %s %s %d\n", stmt.Deparsed, stmt.Hash, stmt.StmtType)
    }

    perr(err)
  }

}

func isStmtHashFoundInDb(ctx *Context, stmt *ParsedStmt) bool {
  r, e := ctx.Db.Query("select * from schemaflow.statements where stmt_hash=$1", stmt.Hash)
  defer r.Close()
  perr(e)
  return r.Next()
}

func isStmtNameFoundInDb(ctx *Context, stmt *ParsedStmt) bool {
  if !stmt.HasName {
    return false
  }

  r, e := ctx.Db.Query("select * from schemaflow.statements where stmt_name=$1 and stmt_type=$2", stmt.Name, stmt.StmtType);
  defer r.Close()
  perr(e)
  return r.Next()
}

func getPrevStmtVersion(ctx *Context, stmt *ParsedStmt) *pg_query.RawStmt {
  var prev_stmt_text string
  e := ctx.Db.QueryRow("select stmt from schemaflow.statements where stmt_name=$1 and stmt_type=$2", stmt.Name, stmt.StmtType).Scan(&prev_stmt_text);
  perr(e)
  parsed, e := pg_query.Parse(prev_stmt_text)
  perr(e)

  stmts := parsed.GetStmts()

  if len(stmts) == 0 {
    return nil
  }

  return stmts[0]
}

func readFileToString(ctx *Context, file string) string {
  data, err := os.ReadFile(file)
  perr(err)
  return string(data)
}

func extractFileFromPath(path string) string {
  return filepath.Base(path)
}
