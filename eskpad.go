package main

import (
    "fmt"
    "os"
    "time"
    "errors"
    "strings"

    "net/http"
    "github.com/gin-gonic/gin"
    "html/template"
    
    "database/sql"
    _ "github.com/mattn/go-sqlite3"

    "github.com/gosimple/slug"
)

type Detalhes struct {
    id int
    url, content, title, tipo, updated_at string
    visits int
}

const version = 0.4

func main() {
    fmt.Println("EskPad", version)

    if _, err := os.Stat("database_v2.db"); errors.Is(err, os.ErrNotExist) {
        fmt.Println("DB does not exists, creating...")
        initDb()
    }
    db, err := sql.Open("sqlite3", "database_v2.db")
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    defer db.Close()

    // Create a Gin router with default middleware (logger and recovery)
    gin.SetMode(gin.ReleaseMode)
    r := gin.Default()

    r.LoadHTMLGlob("templates/*")

    r.GET("/", func(c *gin.Context) {

        links := getLast10Links(db)

        c.HTML(http.StatusOK, "main.tmpl", gin.H{
            "url": "",
            "title": "title",
            "tipo": "",
            "content": "",
            "updated_at": "",
            "last_links": template.HTML(links),
            "version": version,
        })
    })

    r.POST("/search", func(c *gin.Context){
        search := c.PostForm("search")
        results := seachLinks(search, db)

        c.String(http.StatusOK, results)
    })

    r.POST("/update", func(c *gin.Context) {
        url_post := c.PostForm("url")
        tipo := c.DefaultPostForm("tipo", "")
        content := c.DefaultPostForm("content", "")
        title := c.DefaultPostForm("title", "Text")
        url := Slugify(url_post)
        if url_post == "" {
            url = Slugify(title)
        }       

        s, _ := getUrl(url, db)

        if s == false {
            url = createUrl(url, title, tipo, content, db)
            status, _ := getUrl(url, db)

            if status == false {
                c.JSON(200, gin.H{
                    "status":  "failed",
                    "message": "Failed to create db entry",
                    "url":    url,
                })
            } else {
                c.Redirect(http.StatusFound, "/" + url)
            }
        } else {
            _ = updateUrl(url, content, title, tipo, db)
            c.Redirect(http.StatusFound, "/" + url)
        }        
    })

    r.GET("/:name", func(c *gin.Context) {
        link := c.Params.ByName("name")
        ok, detalhes := getUrl(link, db)
        if ok {
            c.HTML(http.StatusOK, "view.tmpl", gin.H{
                "url": detalhes.url,
                "title": detalhes.title,
                "tipo": detalhes.tipo,
                "content": detalhes.content,
                "updated_at": detalhes.updated_at,
                "visits": detalhes.visits,
            })
        } else {
            c.HTML(http.StatusOK, "main.tmpl", gin.H{
                "url": detalhes.url,
                "title": "My Title",
                "tipo": "",
                "content": "",
                "updated_at": "",
                "last_links": "",
                "visits": 0,
                "version": version,
            })
        }
    })

    r.GET("/:name/update", func(c *gin.Context) {
        link := c.Params.ByName("name")
        ok, detalhes := getUrl(link, db)
        if ok {
            c.HTML(http.StatusOK, "main.tmpl", gin.H{
                "url": detalhes.url,
                "title": detalhes.title,
                "tipo": detalhes.tipo,
                "content": detalhes.content,
                "updated_at": detalhes.updated_at,
                "visits": detalhes.visits,
                "last_links": "",
                "version": version,
            })
        } else {
            c.HTML(http.StatusOK, "main.tmpl", gin.H{
                "url": detalhes.url,
                "title": "",
                "tipo": "",
                "content": "",
                "updated_at": "",
                "last_links": "",
                "visits": 0,
                "version": version,
            })
        }
    })

    r.GET("/:name/raw", func(c *gin.Context) {
        link := c.Params.ByName("name")
        ok, detalhes := getUrl(link, db)
        if ok {
            c.HTML(http.StatusOK, "raw.tmpl", gin.H{
                "url": detalhes.url,
                "title": detalhes.title,
                "tipo": detalhes.tipo,
                "content": detalhes.content,
                "updated_at": detalhes.updated_at,
                "visits": detalhes.visits,
            })
        } else {
            c.HTML(http.StatusOK, "main.tmpl", gin.H{
                "url": detalhes.url,
                "title": "",
                "tipo": "",
                "content": "",
                "updated_at": "",
                "last_links": "",
                "visits": 0,
                "version": version,
            })
        }
    })

    r.GET("/:name/delete", func(c *gin.Context) {
        link := c.Params.ByName("name")
        _ = deleteUrl(link, db)
        c.Redirect(http.StatusFound, "/")
    })

    r.Run()
}

func initDb() (bool){
    db, err := sql.Open("sqlite3", "./database_v2.db")
        if err != nil {
            fmt.Println(err)
        os.Exit(1)
    }
    defer db.Close()
    sqlStat := "create table urls (id integer not null primary key, url text UNIQUE, title text, tipo text, content text, updated_at string, visits integer);"
    _, err = db.Exec(sqlStat)
    if err != nil {
        fmt.Printf("%q: %s\n", err, sqlStat)
        return false
    }
    fmt.Println("DB initiated")
    return true
}

func createUrl(url string, title string, tipo string, content string, db *sql.DB) (string) {
    fmt.Printf("Create Url: %s\n", url)
    db.Exec("INSERT INTO urls(url, content, title, tipo, updated_at, visits) values (?, ?, ?, ?, ?, ?)", url, content, title, tipo, time.Now().Format("2006-01-02 15:04:05"), 0)
    return url
}

func getLast10Links(db *sql.DB) (string) {
    rows, err := db.Query("SELECT url, title FROM urls ORDER BY id DESC LIMIT 0, 10")
    if err !=  nil {
        fmt.Println("ERROR Getting Url!")
        return ""
        //os.Exit(1)
    }
    defer rows.Close()

    var returnText = ""
    for rows.Next() {
        var url string
        var title string
        err = rows.Scan(&url, &title)
        if err != nil {
            fmt.Println(err)
        }
        returnText = returnText + fmt.Sprintf("<li><a href=\"\\%s\">%s</a></li>", url, title)
    }
    if err = rows.Err(); err != nil {
        fmt.Println(err)
    }
    return returnText
}

func seachLinks(search string, db *sql.DB) (string) {
    search = "%" + search + "%"
    sql := fmt.Sprintf("SELECT url, title FROM urls WHERE title LIKE '%s' ORDER BY title ASC", search)
    fmt.Println(sql)
    rows, err := db.Query(sql)
    if err !=  nil {
        fmt.Println("ERROR Getting Url!")
        return ""
        //os.Exit(1)
    }
    defer rows.Close()

    var returnText = ""
    for rows.Next() {
        var url string
        var title string
        err = rows.Scan(&url, &title)
        if err != nil {
            fmt.Println(err)
        }
        returnText = returnText + fmt.Sprintf("<li><a href=\"\\%s\">%s</a></li>", url, title)
    }
    if err = rows.Err(); err != nil {
        fmt.Println(err)
    }
    return returnText
}

func getUrl(q_url string, db *sql.DB) (bool, Detalhes) {
    stmt, err := db.Prepare("SELECT id, url, content, title, tipo, updated_at, visits FROM urls WHERE url = ?")
    if err !=  nil {
        fmt.Println("ERROR Getting Url!")
        return false, Detalhes{}
        //os.Exit(1)
    }
    detalhes := Detalhes{}
    err = stmt.QueryRow(q_url).Scan(&detalhes.id, &detalhes.url, &detalhes.content, &detalhes.title, &detalhes.tipo, &detalhes.updated_at, &detalhes.visits)
    if err != nil {
        return false, Detalhes{}
    }
    fmt.Printf("%v\n", detalhes.updated_at)
    visits := detalhes.visits + 1
    db.Exec("UPDATE urls SET visits = ? WHERE url = ?", visits, q_url)
    return true, detalhes
}

func updateUrl(url string, content string, title string, tipo string, db *sql.DB) (bool) {
    db.Exec("UPDATE urls SET content = ?, title = ?, tipo = ?, updated_at = ? WHERE url = ?", content, title, tipo, time.Now().Format("2006-01-02 15:04:05"), url)
    return true
}

func deleteUrl(url string, db *sql.DB) (bool) {
    db.Exec("DELETE FROM urls WHERE url = ?", url)
    return true
}

func stats(url string, db *sql.DB) (int) {
    stmt, err := db.Prepare("SELECT visits FROM urls WHERE url = ?")
    if err !=  nil {
        fmt.Println("Url not found!")
        return 0
    }
    var visits int
    err = stmt.QueryRow(url).Scan(&visits)
    if err != nil {
        fmt.Println("Url not found!")
        return 0
    }
    return visits
}

func formatAsDate(t time.Time) string {
    year, month, day := t.Date()
    return fmt.Sprintf("%d%02d/%02d", year, month, day)
}

func Slugify(s string) string {
    // Convert to lowercase
    s = strings.ToLower(s)
    s = slug.Make(s)
    return s
}