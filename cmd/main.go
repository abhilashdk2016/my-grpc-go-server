package main

import (
	"database/sql"
	"log"

	"github.com/abhilashdk2016/my-grpc-go-server/db"
	mygrpc "github.com/abhilashdk2016/my-grpc-go-server/internal/adapter/grpc"
	app "github.com/abhilashdk2016/my-grpc-go-server/internal/application"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(logwriter{})

	sqlDB, err := sql.Open("pgx", "postgres://postgres:abhi@2024@localhost:5432/grpc_bank?sslmode=disable")

	if err != nil {
		log.Fatal("Unable to connect to database : ", err)
	}
	db.Migrate(sqlDB)
	// databaseAdapter, err := database.NewDatabaseAdapter(sqlDB)
	// if err != nil {
	// 	log.Fatal("Unable to create database adapter : ", err)
	// }

	// runDummyOrm(databaseAdapter)

	bs := &app.BankService{}
	grpcAdapter := mygrpc.NewGrpcAdapter(bs, 8080)
	grpcAdapter.Run()
}

// func runDummyOrm(da *database.DatabaseAdapter) {
// 	now := time.Now()

// 	uuid, _ := da.Save(
// 		&database.DummyOrm{
// 			UserId:   uuid.New(),
// 			UserName: "user" + now.Format("15:04:05"),
// 		},
// 	)

// 	res, _ := da.GetByUuid(&uuid)

// 	log.Println("res : ", res)
// }
