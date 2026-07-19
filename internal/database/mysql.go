package database

import(
	"fmt"
	"database/sql"
	"IM-system/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

func NewMySQL(cfg config.MySQLConfig)(*sql.DB, error){
	dsn := fmt.Sprintf(
		//用户名:密码@tcp(主机:端口)/数据库名
		"%s:%s@tcp(%s:%d)/%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,    
		cfg.Port ,   
		cfg.DataName,
	)
	//连接池对象是否创建成功
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	//能不能真正连接到 MySQL
	if err := db.Ping(); err != nil{
		db.Close()
		return nil, err
	}
	return db, nil
}