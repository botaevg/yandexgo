package repositories

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"time"
)

func NewClient(ctx context.Context, maxAttempts int, dsn string) (pool *pgxpool.Pool, err error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err = pgxpool.Connect(ctx, dsn)

	if err != nil {
		log.Print("error do with tries postgresql")
	}
	//pool.Ping()
	q := `CREATE TABLE cookies(
    idEncrypt VARCHAR(100),
	key VARCHAR(100),
	nonce VARCHAR(100)
);`
	_, err = pool.Exec(context.Background(), q)
	if err != nil {
		log.Print("ТАБЛИЦА НЕ СОЗДАНА")
		log.Print(err)
	}
	q = `CREATE TABLE urls(
    idEncrypt VARCHAR(100),
	shortURL VARCHAR(100),
	fullURL VARCHAR(100)
);`
	_, err = pool.Exec(context.Background(), q)
	if err != nil {
		log.Print("ТАБЛИЦА НЕ СОЗДАНА")
		log.Print(err)
	}

	/*q2 := `
		INSERT INTO cookies
		    (name)
		VALUES
		       ($1) ;
	`

	_, err = pool.Query(context.Background(), q2, "botaev garik")
	if err != nil {
		log.Print("Запись не создана")
		log.Print(err)
	} else {
		log.Print("Запись создана")
	}*/

	return pool, nil
}
