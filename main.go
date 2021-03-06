package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/writer"
)

const (
	filePath     = "./output.parquet"
	recordNumber = 10000
)

var data []*user

type user struct {
	ID        string    `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	FirstName string    `parquet:"name=firstname, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	LastName  string    `parquet:"name=lastname, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Email     string    `parquet:"name=email, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Phone     string    `parquet:"name=phone, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Blog      string    `parquet:"name=blog, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Username  string    `parquet:"name=username, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Score     float64   `parquet:"name=score, type=DOUBLE"`
	CreatedAt time.Time // wont be saved in the parquet file
}

func main() {
	log.Println("")

	log.Println("create fake data")
	createFakeData()

	log.Println("generate parquet file")
	err := generateParquet(data)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(2 * time.Second)
	log.Println("")

	log.Println("Print page 1 only")
	page1, err := readPartialParquet(10, 1)
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range page1 {
		fmt.Println(a)
	}

	time.Sleep(2 * time.Second)
	log.Println("")

	log.Println("Print page 2 only")
	page2, err := readPartialParquet(10, 2)
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range page2 {
		fmt.Println(a)
	}

	time.Sleep(2 * time.Second)
	log.Println("")

	log.Println("Print all data")
	all, err := readParquet()
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range all {
		fmt.Println(a)
	}

	time.Sleep(2 * time.Second)
	log.Println("")

	log.Println("Print column 'Firstname' only")
	names, err := readParquetColumn("firstname")
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range names {
		fmt.Println(a)
	}

	time.Sleep(2 * time.Second)
	log.Println("")

	avg, err := calcScoreAVG()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(fmt.Sprintf("Calculate score average: %.3f", avg))

	log.Println("")
	log.Println("")
}

func createFakeData() {
	for i := 0; i < recordNumber; i++ {
		u := &user{
			ID:        faker.UUIDDigit(),
			FirstName: faker.FirstName(),
			LastName:  faker.LastName(),
			Email:     faker.Email(),
			Phone:     faker.Phonenumber(),
			Blog:      faker.URL(),
			Username:  faker.Username(),
			Score:     float64(i),
			CreatedAt: time.Now(),
		}
		data = append(data, u)
	}
}

func generateParquet(data []*user) error {
	fw, err := local.NewLocalFileWriter(filePath)
	if err != nil {
		return err
	}

	// parameters: writer, type of struct, size
	pw, err := writer.NewParquetWriter(fw, new(user), int64(len(data)))
	if err != nil {
		return err
	}

	// compression type
	pw.CompressionType = parquet.CompressionCodec_GZIP
	defer fw.Close()
	for _, d := range data {
		if err = pw.Write(d); err != nil {
			return err
		}
	}
	if err = pw.WriteStop(); err != nil {
		return err
	}

	return nil
}

func readParquet() ([]*user, error) {
	fr, err := local.NewLocalFileReader(filePath)
	if err != nil {
		return nil, err
	}

	pr, err := reader.NewParquetReader(fr, new(user), recordNumber)
	if err != nil {
		return nil, err
	}

	u := make([]*user, recordNumber)
	if err = pr.Read(&u); err != nil {
		return nil, err
	}

	pr.ReadStop()
	fr.Close()

	return u, nil
}

func readPartialParquet(pageSize, page int) ([]*user, error) {
	fr, err := local.NewLocalFileReader(filePath)
	if err != nil {
		return nil, err
	}

	pr, err := reader.NewParquetReader(fr, new(user), int64(pageSize))
	if err != nil {
		return nil, err
	}

	err = pr.SkipRows(int64(pageSize * page))
	if err != nil {
		return nil, err
	}

	u := make([]*user, pageSize)
	if err = pr.Read(&u); err != nil {
		return nil, err
	}

	pr.ReadStop()
	fr.Close()

	return u, nil
}

func readParquetColumn(name string) ([]string, error) {
	fr, err := local.NewLocalFileReader(filePath)
	if err != nil {
		return nil, err
	}

	pr, err := reader.NewParquetColumnReader(fr, recordNumber)
	if err != nil {
		return nil, err
	}
	num := pr.GetNumRows()

	colData, _, _, err := pr.ReadColumnByPath("parquet_go_root."+name, num)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, i := range colData {
		result = append(result, i.(string))
	}

	return result, nil
}

func calcScoreAVG() (float64, error) {
	fr, err := local.NewLocalFileReader(filePath)
	if err != nil {
		return 0.0, err
	}

	pr, err := reader.NewParquetColumnReader(fr, recordNumber)
	if err != nil {
		return 0.0, err
	}
	num := pr.GetNumRows()

	colData, _, _, err := pr.ReadColumnByPath("parquet_go_root.score", num)
	if err != nil {
		return 0.0, err
	}
	var result float64
	for _, i := range colData {
		result += i.(float64)
	}

	return result / float64(num), nil
}
