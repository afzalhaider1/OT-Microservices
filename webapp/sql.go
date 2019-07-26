package webapp

import (
    "database/sql"
    log "github.com/sirupsen/logrus"
    "github.com/magiconair/properties"
    "os"
    "time"
    "net/http"
    "text/template"
    _ "github.com/go-sql-driver/mysql"
)

type Employee struct {
    Id    int
    Name  string
    City  string
}

var tmpl = template.Must(template.New("Employee Management Template").Parse(htmltemplate))

func dbConn() (db *sql.DB) {
    dbDriver := "mysql"
    dbName := "employeedb"
    var dbUser string
    var dbPass string
    var dbUrl string
    var dbPort string
    propertyfile := "/etc/conf.d/ot-go-webapp/database.properties"

    if fileExists(propertyfile) {
        vaules := properties.MustLoadFiles([]string{propertyfile}, properties.UTF8, true)
        dbUser = vaules.GetString("DB_USER", "DB_USER")
        dbPass = vaules.GetString("DB_PASSWORD", "DB_PASSWORD")
        dbUrl  = vaules.GetString("DB_URL", "DB_URL")
        dbPort = vaules.GetString("DB_PORT", "DB_PORT")
        log.Info("READING properties from /etc/conf.d/ot-go-webapp/database.properties")
    } else {
        dbUser = os.Getenv("DB_USER")
        dbPass = os.Getenv("DB_PASSWORD")
        dbUrl  = os.Getenv("DB_URL")
        dbPort = os.Getenv("DB_PORT")
        log.Info("NO PROPERTY found in /etc/conf.d/ot-go-webapp/database.properties, USING environment variables")
    }

    db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp("+dbUrl+":"+dbPort+")/"+dbName)
    if err != nil {
        log.Error(err.Error())
    }
    return db
}

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
    for i := 0; ; i++ {
        err = f()
        if err == nil {
            return
        }

        if i >= (attempts - 1) {
            break
        }

        time.Sleep(sleep)

        log.Println("retrying after error:", err)
    }
    return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func createDatabaseTable() {
	db := dbConn()
	_,err := db.Exec("CREATE DATABASE IF NOT EXISTS employeedb")
	if err != nil {
		log.Error(err.Error())
	} else {
		log.Info("DATABASE is created with name employeedb")
	}

    return attempt < 5, err

	_,err = db.Exec("USE employeedb")
	if err != nil {
		log.Error(err.Error())
	} else {
		log.Info("USING employeedb database")
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Employee ( id int(6) NOT NULL AUTO_INCREMENT, name varchar(50) NOT NULL, city varchar(50) NOT NULL, PRIMARY KEY (id) )")
	if err != nil {
		log.Error(err.Error())
	} else {
		log.Info("TABLE is created with name Employee")
	}
	defer db.Close()
}

if err != nil {
    log.Error("Error:", err)
}

func Index(w http.ResponseWriter, r *http.Request) {
    db := dbConn()
    selDB, err := db.Query("SELECT * FROM Employee ORDER BY id DESC")
    if err != nil {
        panic(err.Error())
    }
    emp := Employee{}
    res := []Employee{}
    for selDB.Next() {
        var id int
        var name, city string
        err = selDB.Scan(&id, &name, &city)
        if err != nil {
            panic(err.Error())
        }
        emp.Id = id
        emp.Name = name
        emp.City = city
        res = append(res, emp)
        log.Info("GET request on the /index page")
    }
    tmpl.ExecuteTemplate(w, "Index", res)
    defer db.Close()
}

func Show(w http.ResponseWriter, r *http.Request) {
    db := dbConn()
    nId := r.URL.Query().Get("id")
    selDB, err := db.Query("SELECT * FROM Employee WHERE id=?", nId)
    if err != nil {
        panic(err.Error())
    }
    emp := Employee{}
    for selDB.Next() {
        var id int
        var name, city string
        err = selDB.Scan(&id, &name, &city)
        if err != nil {
            panic(err.Error())
        }
        emp.Id = id
        emp.Name = name
        emp.City = city
        log.Info("GET request on the /show for "+ emp.Name)
    }
    tmpl.ExecuteTemplate(w, "Show", emp)
    defer db.Close()
}

func New(w http.ResponseWriter, r *http.Request) {
    tmpl.ExecuteTemplate(w, "New", nil)
}

func Edit(w http.ResponseWriter, r *http.Request) {
    db := dbConn()
    nId := r.URL.Query().Get("id")
    selDB, err := db.Query("SELECT * FROM Employee WHERE id=?", nId)
    if err != nil {
        panic(err.Error())
    }
    emp := Employee{}
    for selDB.Next() {
        var id int
        var name, city string
        err = selDB.Scan(&id, &name, &city)
        if err != nil {
            panic(err.Error())
        }
        emp.Id = id
        emp.Name = name
        emp.City = city
        log.Info("POST request on the /edit for "+ emp.Name)
    }
    tmpl.ExecuteTemplate(w, "Edit", emp)
    defer db.Close()
}

func Insert(w http.ResponseWriter, r *http.Request) {
    db := dbConn()
    if r.Method == "POST" {
        name := r.FormValue("name")
        city := r.FormValue("city")
        insForm, err := db.Prepare("INSERT INTO Employee(name, city) VALUES(?,?)")
        if err != nil {
            panic(err.Error())
        }
        insForm.Exec(name, city)
        log.Info("POST request on the /insert for "+ name)
    }
    defer db.Close()
    http.Redirect(w, r, "/", 301)
}

func Update(w http.ResponseWriter, r *http.Request) {
    db := dbConn()
    if r.Method == "POST" {
        name := r.FormValue("name")
        city := r.FormValue("city")
        id := r.FormValue("uid")
        insForm, err := db.Prepare("UPDATE Employee SET name=?, city=? WHERE id=?")
        if err != nil {
            panic(err.Error())
        }
        insForm.Exec(name, city, id)
        log.Info("POST request on the /update for "+ name)
    }
    defer db.Close()
    http.Redirect(w, r, "/", 301)
}

func Delete(w http.ResponseWriter, r *http.Request) {
    db := dbConn()
    emp := r.URL.Query().Get("id")
    delForm, err := db.Prepare("DELETE FROM Employee WHERE id=?")
    if err != nil {
        panic(err.Error())
    }
    delForm.Exec(emp)
    log.Println("DELETE")
    log.Info("POST request on the /delete")
    defer db.Close()
    http.Redirect(w, r, "/", 301)
}
