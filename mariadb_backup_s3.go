package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	svc  *s3.S3
	sess *session.Session
)

const (
	BUCKET_NAME = "mariadb-backup"
	REGION      = "hcm"
	ENDPOINT    = "https://s3-hcm.sds.vnpaycloud.vn"
)

func init() {
	svc = s3.New(session.Must(session.NewSession(&aws.Config{
		Region:   aws.String(REGION),
		Endpoint: aws.String(ENDPOINT),
	})))
}

func listAllBuckets(sess *session.Session) (*s3.ListBucketsOutput, error) {

	resp, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func listObject() (resp *s3.ListObjectsV2Output) {
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(BUCKET_NAME),
	})
	if err != nil {
		panic(err)
	}
	return resp
}

func getObject(filename string) {
	fmt.Println("Downloading: ", filename)
	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(BUCKET_NAME),
		Key:    aws.String(filename),
	})
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	err = ioutil.WriteFile(filename, body, 0644)
	if err != nil {
		panic(err)
	}
}

func uploadObject(filename string) (resp *s3.PutObjectOutput) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	fmt.Println("Uploading:", filename)
	resp, err = svc.PutObject(&s3.PutObjectInput{
		Body:   file,
		Bucket: aws.String(BUCKET_NAME),
		Key:    aws.String(strings.Split(filename, "/")[1]),
	})

	if err != nil {
		panic(err)
	}

	return resp
}

func main() {
	/*
		resp, err := listAllBuckets(sess)
				if err != nil {
					fmt.Println("Got an error retrieving buckets:")
					fmt.Println(err)
					return
				}

				fmt.Println("Buckets:")

				for _, bucket := range resp.Buckets {
					fmt.Println(*bucket.Name + ": " + bucket.CreationDate.Format("15:04:05 Monday 2006-01-02"))
				}

			//	fmt.Println(listObject())
			for _, object := range listObject().Contents {
				getObject(*object.Key)
				fmt.Println("Name:   ", *object.Key)
				fmt.Println("Last modified:   ", *object.LastModified)
				fmt.Println("")
			}
			//	fmt.Println(listObject())
	*/
	fmt.Print("Enter file_path (example mariadb_backup/): ")
	var folder_path string
	fmt.Scanf("%s", &folder_path)
	fileinfo, err := os.Stat(folder_path)
	if os.IsNotExist(err) {
		log.Fatal("File does not exist.")
	}
	log.Println(fileinfo)

	logWriter, err := syslog.New(syslog.LOG_SYSLOG, "S3")
	if err != nil {
		log.Fatalln("Unable to set logfile:", err.Error())
	}
	log.SetOutput(logWriter)
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	files, _ := ioutil.ReadDir(folder_path)

	fmt.Println(files)
	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			uploadObject(folder_path + file.Name())
			log.Println("UPLOAD FILE MARIADB_BACKUP DONE")
		}
	}

}
