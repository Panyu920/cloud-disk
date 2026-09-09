package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Panyu920/cloud-disk/utils"
	"github.com/stretchr/testify/require"
)

func createRandomFile(t *testing.T) File {
	// Create a new file
	file := CreateFileParams{
		FileSha1: utils.RandomSha1(),
		FileName: utils.RandomString(10),
		FileSize: utils.RandomInt64(100000, 100),
		FileAddr: utils.RandomString(20),
	}

	// Call the CreateFile method
	result, err := testQueries.CreateFile(context.Background(), file)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify that the file was created successfully
	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), rowsAffected)

	// Retrieve the created file
	fileID, err := result.LastInsertId()
	require.NoError(t, err)
	require.NotZero(t, fileID)
	retrievedFile, err := testQueries.GetFileById(context.Background(), fileID)
	require.NoError(t, err)

	return retrievedFile
}
func TestCreateFile(t *testing.T) {
	// Create a new file
	_ = createRandomFile(t)
}

func TestGetFileBySha1(t *testing.T) {
	// Create a new file
	file := createRandomFile(t)

	// Call the GetFileBySha1 method
	retrievedFile, err := testQueries.GetFileBySha1(context.Background(), file.FileSha1)
	require.NoError(t, err)

	// Verify that the retrieved file matches the created file
	require.Equal(t, file.FileSha1, retrievedFile.FileSha1)
	require.Equal(t, file.FileName, retrievedFile.FileName)
	require.Equal(t, file.FileSize, retrievedFile.FileSize)
	require.Equal(t, file.FileAddr, retrievedFile.FileAddr)
	require.Equal(t, int8(0), retrievedFile.Status)
	require.WithinDuration(t, retrievedFile.CreateAt, time.Now(), time.Second*5)
	require.WithinDuration(t, retrievedFile.UpdateAt, time.Now(), time.Second*5)
}

func TestUpdateFile(t *testing.T) {
	// Create a new file
	file := createRandomFile(t)

	// Update the file's status
	updateParams := UpdateFileParams{
		ID:     file.ID,
		Status: sql.NullInt16{Int16: 1, Valid: true}, // Set status to 1 (disabled)
	}
	result, err := testQueries.UpdateFile(context.Background(), updateParams)
	require.NoError(t, err)

	// Verify that the file was updated successfully
	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), rowsAffected)

	// Retrieve the updated file
	updatedFile, err := testQueries.GetFileById(context.Background(), file.ID)
	require.NoError(t, err)

	// Verify that the updated file's status matches the expected value
	require.Equal(t, int8(1), updatedFile.Status)
}
