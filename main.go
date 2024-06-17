package main

import (
    "database/sql"
    "flag"
    "fmt"
    "time"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    _ "github.com/go-sql-driver/mysql"
)

type Queue struct {
    Code  string `json:"code"`
    Name  string `json:"name"`
}

func main() {
    port := flag.String("port", "8080", "port to run the server on")
    flag.Parse()
    db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/ale_project")
    if err != nil {
        log.Fatal("Error connecting to the database:", err)
    }
    defer db.Close()
    err = db.Ping()
    if err != nil {
        log.Fatal("Error pinging database:", err)
    }
    r := gin.Default()

    r.GET("/queue", func(c *gin.Context) {
        var queue []Queue

        rows, err := db.Query("SELECT code, name FROM tbl_queue")
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()

        for rows.Next() {
            var item Queue
            if err := rows.Scan(&item.Code, &item.Name); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
            queue = append(queue, item)
        }

        if err := rows.Err(); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, queue)
    })

    r.POST("/generate_code", func(c *gin.Context) {
        var newQueue Queue
    
        if err := c.BindJSON(&newQueue); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
            return
        }

        var count_name int
        err = db.QueryRow("SELECT COUNT(name) FROM tbl_queue WHERE DATE(date) = CURDATE() AND name LIKE ?", "%"+newQueue.Name+"%").Scan(&count_name)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        if count_name > 0 {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "DUPLICATE DATA!"})
            return
        }
    
        stmt, err := db.Prepare("INSERT INTO tbl_queue (code, name, date) VALUES (?, ?, ?)")
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer stmt.Close()
    
        now := time.Now()
        dateString := now.Format("20060102")
    
        var count int
        err = db.QueryRow("SELECT COUNT(*) FROM tbl_queue WHERE DATE(date) = CURDATE()").Scan(&count)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
    
        code := fmt.Sprintf("QUE%s%04d", dateString, count+1)
    
        _, err = stmt.Exec(code, newQueue.Name, now)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
    
        c.JSON(http.StatusOK, gin.H{"message": "Data inserted successfully"})
    })

    r.POST("/delete_queue", func(c *gin.Context) {
        var queue Queue
    
        if err := c.BindJSON(&queue); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
            return
        }

        var count_code int
        err = db.QueryRow("SELECT COUNT(name) FROM tbl_queue WHERE code LIKE ?", "%"+queue.Code+"%").Scan(&count_code)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        if count_code == 0 {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "NO DATA TO DELETE!"})
            return
        }
    
        stmt, err := db.Prepare("DELETE FROM tbl_queue WHERE code = ?")
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer stmt.Close()
    
        _, err = stmt.Exec(queue.Code)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
    
        c.JSON(http.StatusOK, gin.H{"message": "Data deleted successfully"})
    })

    r.POST("/update_queue", func(c *gin.Context) {
        var queue Queue
    
        if err := c.BindJSON(&queue); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
            return
        }

        var count_code int
        err = db.QueryRow("SELECT COUNT(name) FROM tbl_queue WHERE code LIKE ?", "%"+queue.Code+"%").Scan(&count_code)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        if count_code == 0 {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "NO DATA TO UPDATE!"})
            return
        }

        var count_name int
        err = db.QueryRow("SELECT COUNT(name) FROM tbl_queue WHERE DATE(date) = CURDATE() AND name LIKE ? AND code <> ?", "%"+queue.Name+"%", queue.Code).Scan(&count_name)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        if count_name > 0 {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "DUPLICATE DATA!"})
            return
        }
    
        stmt, err := db.Prepare("UPDATE tbl_queue SET name = ? WHERE code = ?")
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer stmt.Close()
    
        _, err = stmt.Exec(queue.Name, queue.Code)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
    
        c.JSON(http.StatusOK, gin.H{"message": "Data updated successfully"})
    })

    addr := fmt.Sprintf(":%s", *port)
    log.Printf("Server running on %s", addr)
    r.Run(addr)
}
