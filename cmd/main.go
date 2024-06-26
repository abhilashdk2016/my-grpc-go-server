package main

import (
	"database/sql"
	"log"
	"math/rand"
	"time"

	"github.com/abhilashdk2016/my-grpc-go-server/db"
	"github.com/abhilashdk2016/my-grpc-go-server/internal/adapter/database"
	mygrpc "github.com/abhilashdk2016/my-grpc-go-server/internal/adapter/grpc"
	app "github.com/abhilashdk2016/my-grpc-go-server/internal/application"
	"github.com/abhilashdk2016/my-grpc-go-server/internal/application/domain/bank"
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
	databaseAdapter, err := database.NewDatabaseAdapter(sqlDB)
	if err != nil {
		log.Fatal("Unable to create database adapter : ", err)
	}

	// runDummyOrm(databaseAdapter)

	bs := app.NewBankService(databaseAdapter)

	go generateExcahngeRates(bs, "USD", "INR", time.Second*5)
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

func generateExcahngeRates(bs *app.BankService, fromCurrency, toCurrency string, duration time.Duration) {
	ticker := time.NewTicker(duration)

	for range ticker.C {
		now := time.Now()
		validFrom := now.Truncate(time.Second).Add(3 * time.Second)
		validTo := validFrom.Add(duration).Add(-1 * time.Millisecond)
		dummyRate := bank.ExchangeRate{
			FromCurrency:       fromCurrency,
			ToCurrency:         toCurrency,
			ValidFromTimestamp: validFrom,
			ValidToTimestamp:   validTo,
			Rate:               2000 + float64(rand.Intn(300)),
		}

		bs.CreateExchangeRate(dummyRate)
	}
}
