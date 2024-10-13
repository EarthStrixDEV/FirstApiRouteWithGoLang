package main

import (
	"database/sql"
	"fmt"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

// define struct User as schema object
type User struct {
    Id int `json:"user_id"`
    Name string `json:"user_name"`
    Age int `json:"user_age"`
}

func connectDB() (*sql.DB, error) {
    // connect to database
    // config Postgres database
    db ,err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=123456 dbname=Fiber sslmode=disable")
    if err != nil {
        return nil ,err
    }

    // check connection status
    err = db.Ping()
    if err != nil {
        db.Close()
        return nil, err
    }

    return db , nil
}

func main() {
    app := fiber.New()
    
    // test route
    app.Get("/" ,func(c *fiber.Ctx) error {
        return c.SendString("Hello ,Fiber!")
    })

    // get all user rows
    app.Get("/user" ,func(c *fiber.Ctx) error {
        db, err := connectDB()

        if err != nil {
            panic(err)
        }
        rows, err := db.Query("SELECT user_id, user_name, user_age FROM public.\"user\"")
        if err != nil {
            panic(err)
        }

        defer rows.Close()

        var users []User

        for rows.Next() {
            var user User
            if err := rows.Scan(
            	&user.Id,
            	&user.Name,
            	&user.Age,
            ); err != nil {
                return err
            }
            users = append(users, user)
        }

        if err := rows.Err(); err != nil {
            return err
        }

        return c.JSON(users)
    })

    // get user data with user id
    app.Get("/user/:id" ,func(c *fiber.Ctx) error {
        userId := c.Params("id")
        db, err := connectDB()

        if err != nil {
            panic(err)
        }
        rows, err := db.Query("SELECT user_id, user_name, user_age FROM public.\"user\" WHERE user_id = $1", userId)
        if err != nil {
            panic(err)
        }

        defer rows.Close()

        var users []User

        for rows.Next() {
            var user User
            if err := rows.Scan(
            	&user.Id,
            	&user.Name,
            	&user.Age,
            ); err != nil {
                return err
            }
            users = append(users, user)
        }

        if err := rows.Err(); err != nil {
            return err
        }

        return c.JSON(users)
    })

    // insert a new user
    app.Post("/user/create", func(c *fiber.Ctx) error {
        user := new(User)
        db, err := connectDB()

        if err != nil {
            panic(err)
        }

        if err := c.BodyParser(user); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "Invalid JSON format",
            })
        }

        stmt ,err := db.Prepare("INSERT INTO public.\"user\" (user_name, user_age) VALUES ($1 ,$2)")
        if err != nil {
            panic(err)
        }
        defer stmt.Close()

        result ,err := stmt.Exec(user.Name ,user.Age)
        if err != nil {
            panic(err)
        }

        fmt.Println(result)

        return c.SendStatus(fiber.StatusCreated)
    })

    // delete a user with user id
    app.Delete("/user/:id", func(c *fiber.Ctx) error {
        userId := c.Params("id")
        db, err := connectDB()

        if err != nil {
            panic(err)
        }
        defer db.Close()

        stmt, err := db.Prepare("DELETE FROM public.\"user\" WHERE user_id = $1")
        if err != nil {
            panic(err)
        }
        defer stmt.Close()

        result ,err := stmt.Exec(userId)
        if err != nil {
            return c.SendStatus(fiber.StatusInternalServerError)
        }


        rowAffected ,err := result.RowsAffected()
        if err != nil {
            return c.SendStatus(fiber.StatusInternalServerError)
        }

        if rowAffected == 0 {
            return c.SendStatus(fiber.StatusNotFound)
        }

        return c.SendStatus(fiber.StatusOK)
    })
    
    // server listening on http://localhost:5000
    err := app.Listen(fmt.Sprintf(":5000"))
    if err != nil {
        fmt.Printf("Error starting server: %s\n", err)
    }
}