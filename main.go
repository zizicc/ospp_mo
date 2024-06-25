/*
 * Converts PDF to chunks using unipdf
 * Chunks to embeddings using LLAMA3
 * Store chunks and embedding in PgVector
 */

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

func init() {
	// Make sure to load your metered License API key prior to using the library.
	// If you need a key, you can sign up and create a free one at https://cloud.unidoc.io
	err := license.SetMeteredKey(`8452a63b8734238bc4290d82fd5fed36c318458e6add293dd7be03b516d55448`)
	if err != nil {
		panic(err)
	}
}

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("Usage: go run main.go input.pdf\n")
		os.Exit(1)
	}

	inputPath := os.Args[1]

	chunks, err := outputPdfText(inputPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	filteredChunks := filterChunks(chunks)
	//get the first two chunks for testing
	testChunks := filteredChunks[:2]
	//generate embeddings
	embeddings, err := getEmbeddings(testChunks)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println("Embeddings are created successfully!")

	fmt.Printf("Got %d embeddings:\n", len(embeddings))
	for i, emb := range embeddings {
		fmt.Printf("%d: len=%d; first few=%v\n", i, len(emb), emb[:4])
	}

	storeInPgV(testChunks, embeddings)

}

// outputPdfText returns chunks of PDF file
func outputPdfText(inputPath string) ([]string, error) {
	f, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return nil, err
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, err
	}

	var chunks []string
	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			continue
		}

		ex, err := extractor.New(page)
		if err != nil {
			continue
		}

		text, err := ex.ExtractText()
		if err != nil {
			continue
		}

		//split text to chunks by "\n\n"
		textChunks := strings.Split(text, "\n\n")
		chunks = append(chunks, textChunks...)
		/*
			fmt.Println("------------------------------")
			fmt.Printf("Page %d:\n", pageNum)
			fmt.Printf("\"%s\"\n", text)
			fmt.Println("------------------------------")
		*/
	}

	return chunks, nil
}

// filterChunks removes all the empty chunks and prints out the remaining
func filterChunks(chunks []string) []string {
	fmt.Printf("--------------------\n")
	fmt.Printf("PDF to chunks:\n")
	fmt.Printf("--------------------\n")

	filteredChunks := []string{}
	for i, str := range chunks {
		if str != "" {
			fmt.Printf("---------Chunk %d-----------\n", i)
			fmt.Println(str)
			fmt.Printf("---------Chunk %d-----------\n", i)
			filteredChunks = append(filteredChunks, str)
		}

	}

	return filteredChunks
}

// getEmbeddings return embeddings generated based on chunks
func getEmbeddings(chunks []string) ([][]float32, error) {
	llm, err := ollama.New(ollama.WithModel("llama3"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	fmt.Println("Attempting to create embeddings.")

	return llm.CreateEmbedding(ctx, chunks)

}

func storeInPgV(chunks []string, embeddings [][]float32) error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://yazi:wyz123@localhost/test")
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		fmt.Printf("Error extension: %v\n", err)
		panic(err)
	}

	_, err = conn.Exec(ctx, "DROP TABLE IF EXISTS documents")
	if err != nil {
		fmt.Println("Error drop table")
		panic(err)
	}

	_, err = conn.Exec(ctx, "CREATE TABLE documents (id bigserial PRIMARY KEY, content text, embedding vector(4096))")
	if err != nil {
		fmt.Println("Error create table")
		panic(err)
	}

	for i, content := range chunks {
		_, err := conn.Exec(ctx, "INSERT INTO documents (content, embedding) VALUES ($1, $2)", content, pgvector.NewVector(embeddings[i]))
		if err != nil {
			fmt.Printf("Error insert %d\n", i)
			panic(err)
		}
	}

	//use query to test
	documentId := 1
	rows, err := conn.Query(ctx, "SELECT id, content FROM documents WHERE id != $1 ORDER BY embedding <=> (SELECT embedding FROM documents WHERE id = $1) LIMIT 5", documentId)
	if err != nil {
		fmt.Println("Error select")
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var content string
		err = rows.Scan(&id, &content)
		if err != nil {
			fmt.Println("Error scan")
			panic(err)
		}
		fmt.Println(id, content)
	}

	if rows.Err() != nil {
		panic(rows.Err())
	}

	return err
}
