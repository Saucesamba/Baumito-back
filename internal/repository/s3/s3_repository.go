package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FileRepository struct {
	client     *minio.Client
	bucketName string
}

func NewFileRepository(endpoint, accessKey, secretKey, bucket string) (*FileRepository, error) {
	// Инициализируем клиент MinIO
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false, // В докере используем HTTP
	})
	if err != nil {
		return nil, err
	}

	// Проверяем/создаем бакет (папку для фото)
	ctx := context.Background()
	exists, _ := client.BucketExists(ctx, bucket)
	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &FileRepository{client: client, bucketName: bucket}, nil
}

// Upload сохраняет файл и возвращает путь
func (r *FileRepository) Upload(ctx context.Context, fileName string, reader io.Reader, size int64) (string, error) {
	_, err := r.client.PutObject(ctx, r.bucketName, fileName, reader, size, minio.PutObjectOptions{
		ContentType: "image/jpeg", // Для простоты пока так
	})
	if err != nil {
		return "", err
	}

	// Возвращаем путь к файлу (потом будем его отдавать через API)
	return fmt.Sprintf("/%s/%s", r.bucketName, fileName), nil
}
